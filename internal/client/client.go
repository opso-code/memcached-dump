package client

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"memcached-dump/conn"
	"memcached-dump/utils"
)

var (
	resultOK        = []byte("OK\r\n")
	resultError     = []byte("ERROR\r\n")
	resultStored    = []byte("STORED\r\n")
	resultNotStored = []byte("NOT_STORED\r\n")
	resultEnd       = []byte("END\r\n")

	resultClientErrorPrefix = []byte("CLIENT_ERROR ")
)

const SupportMetaDumpVersion = "1.4.31"

type Client struct {
	conn *conn.Conn
}

func NewClient(addr net.Addr) (*Client, error) {
	cn, err := conn.NewConn(addr)
	if err != nil {
		return nil, err
	}
	c := new(Client)
	c.conn = cn
	return c, err
}

func (c *Client) Close() {
	if c.conn.Nc != nil {
		err := c.conn.Nc.Close()
		if err != nil {
			return
		}
	}
}

func (c *Client) Version() (string, error) {
	cmd := []byte("version\r\n")
	_, err := c.conn.Rw.Write(cmd)
	if err != nil {
		return "", err
	}
	err = c.conn.Rw.Flush()
	if err != nil {
		return "", err
	}
	line, err := c.conn.Rw.ReadSlice('\n')
	if err != nil {
		return "", err
	}
	if bytes.Equal(line, resultError) {
		return "", errors.New("ERROR: invalid cmd")
	}
	pattern := "VERSION %s\r\n"
	var ver string
	_, err = fmt.Sscanf(string(line), pattern, &ver)
	if err != nil {
		return "", err
	}
	return ver, nil
}

// stats结果信息
type StatsInfo struct {
	Pid             int      // pid
	Uptime          int64    // 服务已运行秒数
	Time            int64    // 服务器当前Unix时间戳
	Version         string   // memcached版本号
	CurrConnections int      // 当前连接数
	CurrItems       int      // 当前存储的数据条数
	Raw             []string // 原始数据
}

func (c *Client) Stats() (*StatsInfo, error) {
	var info StatsInfo
	cmd := []byte("stats\r\n")
	_, err := c.conn.Rw.Write(cmd)
	if err != nil {
		return nil, err
	}
	err = c.conn.Rw.Flush()
	if err != nil {
		return nil, err
	}
	for {
		line, err := c.conn.Rw.ReadSlice('\n')
		if err != nil {
			return nil, err
		}
		if bytes.Equal(line, resultError) {
			return nil, errors.New("ERROR: invalid cmd")
		}
		if bytes.Equal(line, resultEnd) {
			break
		}
		lineStr := strings.TrimRight(string(line), "\r\n")
		if !strings.Contains(lineStr, "STAT ") {
			continue
		}
		p := strings.Split(lineStr, " ")
		var k string
		var v string
		if len(p) >= 3 {
			k = p[1]
			v = p[2]
		} else {
			continue
		}
		info.Raw = append(info.Raw, lineStr)

		switch k {
		case "pid":
			info.Pid, _ = strconv.Atoi(v)
		case "curr_connections":
			info.CurrConnections, _ = strconv.Atoi(v)
		case "curr_items":
			info.CurrItems, _ = strconv.Atoi(v)
		case "uptime":
			info.Uptime, _ = strconv.ParseInt(v, 10, 64)
		case "time":
			info.Time, _ = strconv.ParseInt(v, 10, 64)
		case "version":
			info.Version = v
		}
	}
	return &info, nil
}

func (c *Client) GetSlabs() ([]int, int, error) {
	cmd := []byte("stats items\r\n")
	_, err := c.conn.Rw.Write(cmd)
	if err != nil {
		return nil, 0, err
	}
	err = c.conn.Rw.Flush()
	if err != nil {
		return nil, 0, err
	}

	var data []int
	var total int
	for {
		line, err := c.conn.Rw.ReadSlice('\n')
		if err != nil {
			return nil, 0, err
		}
		if bytes.Equal(line, resultEnd) {
			break
		}
		pattern := "STAT items:%d:number %d\r\n"
		var slab int
		var num int
		_, err = fmt.Sscanf(string(line), pattern, &slab, &num)
		if err != nil {
			continue
		}
		data = append(data, slab)
		total += num
	}
	return data, total, err
}

type Key struct {
	Name     string // 键名
	Size     int    // 大小
	ExpireAt int64  // 过期时间戳
}

func (c *Client) GetKeysByCacheDump() (map[string]Key, error) {
	slabs, total, err := c.GetSlabs()
	if err != nil {
		return nil, err
	}
	log.Printf("Get total %d keys with slabs %v \n", total, slabs)

	info, err := c.Stats()
	if err != nil {
		return nil, err
	}
	log.Printf("Get total keys from stats command %d\n", info.CurrItems)

	keys := make(map[string]Key)
	for _, slab := range slabs {
		cmd := fmt.Sprintf("stats cachedump %d 0\r\n", slab)
		_, err = c.conn.Rw.Write([]byte(cmd))
		if err != nil {
			return nil, err
		}
		err = c.conn.Rw.Flush()
		if err != nil {
			return nil, err
		}
		for {
			line, err := c.conn.Rw.ReadSlice('\n')
			if err != nil {
				return nil, err
			}
			if bytes.Equal(line, resultEnd) {
				break
			}
			if bytes.Equal(line, resultError) {
				return nil, errors.New("command error")
			}
			// log.Println(slab, string(line[:len(line)-2]))
			pattern := "ITEM %s [%d b; %d s]\r\n"
			var key Key
			_, err = fmt.Sscanf(string(line), pattern, &key.Name, &key.Size, &key.ExpireAt)
			if err != nil {
				log.Printf("parse line failed, line %s, error %s \n", line, err)
				continue
			}
			if key.ExpireAt == info.Time {
				key.ExpireAt = 0 // 不过期
			}
			// log.Printf("time %d exp %d\n", info.Time, key.ExpireAt)
			keys[key.Name] = key
		}
	}
	return keys, nil
}

func (c *Client) GetKeysByCrawler() (map[string]Key, error) {
	err := c.conn.Nc.SetReadDeadline(time.Now().Add(6 * time.Second))
	if err != nil {
		return nil, err
	}
	defer c.conn.Nc.SetReadDeadline(time.Time{})
	cmd := []byte("lru_crawler metadump all\r\n")
	_, err = c.conn.Rw.Write(cmd)
	if err != nil {
		return nil, err
	}
	err = c.conn.Rw.Flush()
	if err != nil {
		return nil, err
	}
	m := make(map[string]Key)
	for {
		line, err := c.conn.Rw.ReadSlice('\n')
		if err != nil {
			return nil, err
		}
		if bytes.Equal(line, resultEnd) {
			break
		}
		if bytes.Equal(line, resultOK) {
			break
		}
		if bytes.Equal(line, resultError) {
			return nil, errors.New("CLIENT_ERROR ERROR")
		}
		if bytes.HasPrefix(line, resultClientErrorPrefix) {
			return nil, errors.New(string(line[:len(line)-2]))
		}
		// key=testeb4623222de13dcb2a3055fb390e exp=1637987620 la=1637823086 cas=493360 fetch=no cls=3 size=123
		pattern := "key=%s exp=%d la=%d cas=%d"
		var (
			k   Key
			la  int
			cas int
		)
		lineStr, err := url.QueryUnescape(strings.TrimRight(string(line), "\r\n")) // 转义处理
		if err != nil {
			return nil, err
		}
		_, err = fmt.Sscanf(lineStr, pattern, &k.Name, &k.ExpireAt, &la, &cas)
		if err != nil {
			return nil, fmt.Errorf("parse string %s, error %s", lineStr, err)
		}
		m[k.Name] = k
	}
	return m, nil
}

func (c *Client) GetKeys() (map[string]Key, error) {
	ver, err := c.Version()
	if err != nil {
		return nil, err
	}
	log.Println("memcached version", ver)
	metaDumpSupport := utils.CompareVer(ver, SupportMetaDumpVersion) >= 0 // 是否支持`lru_crawler metadump`命令
	var keys map[string]Key
	if !metaDumpSupport {
		log.Println("\033[7;37;40m[memcached版本过久，正在使用stats cachedump命令查询所有key（受限1M大小数据）]\033[0m")
		keys, err = c.GetKeysByCacheDump()
		if err != nil {
			log.Println(err)
			return nil, err
		}
	} else {
		log.Println("\033[7;37;40m[使用lru_crawler metadump all命令查询所有key]\033[0m")
		keys, err = c.GetKeysByCrawler()
		if err != nil {
			log.Println("\033[7;37;40m[命令超时或memcached实例启动时未开启-o lru_crawler]\033[0m")
			return nil, err
		}
	}
	return keys, nil
}

type Item struct {
	Key    string
	Flags  int
	Size   int
	Casid  int
	Value  string
	Expire int64
}

func (c *Client) Store() (int, error) {
	keys, err := c.GetKeys()
	if err != nil {
		return 0, err
	}
	length := len(keys)
	if length == 0 {
		return 0, nil
	}
	filename := utils.GetDumpFilename(c.conn.Addr)
	var f *os.File
	f, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0o666)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	var keyArr []Key
	for _, key := range keys {
		keyArr = append(keyArr, key)
	}
	chunks := keysChunk(keyArr, 100) // 切割为小数组
	var bar utils.Bar
	bar.NewOption(0, int64(length))
	wl := 0
	var failedKeys []string
	for _, chunk := range chunks {
		items, keys, err := c.Gets(chunk)
		if err != nil {
			return 0, err
		}
		for _, item := range items {
			val := fmt.Sprintf("%s##%d##%d##%s", item.Key, item.Flags, item.Expire, item.Value)
			if _, err = f.WriteString(val + "\n"); err != nil {
				return 0, err
			}
			wl++
		}
		bar.Play(int64(wl))
		for _, key := range keys {
			failedKeys = append(failedKeys, key)
		}
	}
	bar.Finish()
	if wl > 0 {
		log.Printf("Save %d keys to file %s\n", wl, filename)
	}
	if len(failedKeys) > 0 {
		log.Printf("Save failed keys %v\n", failedKeys)
	}
	return wl, nil
}

func (c *Client) DumpTo(addr net.Addr) (int, error) {
	keys, err := c.GetKeys()
	if err != nil {
		return 0, err
	}
	length := len(keys)
	if length == 0 {
		return 0, nil
	}
	// 导入的memcache实例
	cli, err := NewClient(addr)
	if err != nil {
		return 0, err
	}
	defer cli.Close()

	var keyArr []Key
	for _, key := range keys {
		keyArr = append(keyArr, key)
	}
	chunks := keysChunk(keyArr, 100) // 切割为小数组
	var bar utils.Bar
	bar.NewOption(0, int64(length))
	wl := 0
	var failedKeys []string
	for _, chunk := range chunks {
		items, keys, err := c.Gets(chunk)
		if err != nil {
			return 0, err
		}
		for _, item := range items {
			err := cli.Add(item)
			if err != nil {
				log.Printf("Add error:%s key:%s\n", err.Error(), item.Key)
				return 0, err
			}
			wl++
		}
		for _, key := range keys {
			failedKeys = append(failedKeys, key)
		}
		bar.Play(int64(wl))
	}
	bar.Finish()
	if wl > 0 {
		log.Printf("Dump %d keys to memcached %s success!\n", wl, cli.conn.Addr.String())
	}
	if len(failedKeys) > 0 {
		log.Printf("Dump failed keys %v\n", failedKeys)
	}
	return wl, nil
}

func (c *Client) Gets(keys []Key) (items []Item, failedKeys []string, err error) {
	var (
		names   []string       // 查询的所有key
		keysMap map[string]Key // 查询结果 string => Key
		temp    map[string]int // 执行成功标识
	)

	keysMap = make(map[string]Key, len(keys))
	temp = make(map[string]int, len(keys))
	for _, k := range keys {
		keysMap[k.Name] = k
		names = append(names, k.Name)
	}
	if len(names) == 0 {
		return nil, nil, errors.New("memcached is empty")
	}
	cmd := fmt.Sprintf("gets %s\r\n", strings.Join(names, " "))
	_, err = c.conn.Rw.Write([]byte(cmd))
	if err != nil {
		return nil, nil, err
	}
	err = c.conn.Rw.Flush()
	if err != nil {
		return nil, nil, err
	}
	for {
		line, err := c.conn.Rw.ReadSlice('\n')
		if err != nil {
			return nil, nil, err
		}
		if bytes.Equal(line, resultEnd) {
			break
		}
		if bytes.Equal(line, resultError) {
			return nil, nil, errors.New("command error")
		}
		pattern := "VALUE %s %d %d %d\r\n"
		var item Item
		n, err := fmt.Sscanf(string(line), pattern, &item.Key, &item.Flags, &item.Size, &item.Casid)
		if err != nil {
			log.Println(err, n, item)
			continue
		}
		line2, err := c.conn.Rw.ReadSlice('\n')
		if err != nil {
			return nil, nil, err
		}
		if len(line2) < item.Size+2 {
			continue
		}
		item.Value = strings.TrimRight(string(line2), "\r\n")
		item.Expire = keysMap[item.Key].ExpireAt
		items = append(items, item)
		temp[item.Key] = 1
	}
	for _, it := range keysMap {
		if _, ok := temp[it.Name]; !ok {
			failedKeys = append(failedKeys, it.Name)
		}
	}
	return
}

func (c *Client) Add(item Item) error {
	cmd := fmt.Sprintf("add %s %d %d %d\r\n%s\r\n", item.Key, item.Flags, item.Expire, len(item.Value), item.Value)
	_, err := c.conn.Rw.WriteString(cmd)
	if err != nil {
		return err
	}
	err = c.conn.Rw.Flush()
	if err != nil {
		return err
	}
	line, err := c.conn.Rw.ReadSlice('\n')
	if err != nil {
		return err
	}
	if bytes.Equal(line, resultStored) {
		return nil
	}
	if bytes.Equal(line, resultNotStored) {
		// log.Println("add failed, key exists " + item.Key)
		return nil
	}
	if bytes.Equal(line, resultEnd) {
		return nil
	}
	return errors.New("add failed: " + strings.TrimRight(string(line), "\r\n"))
}

// 切片分隔
func keysChunk(keys []Key, size int) [][]Key {
	length := len(keys)
	chunks := int(math.Ceil(float64(length) / float64(size)))
	var chunkArr [][]Key
	for i, end := 0, 0; chunks > 0; chunks-- {
		end = (i + 1) * size
		if end > length {
			end = length
		}
		chunkArr = append(chunkArr, keys[i*size:end])
		i++
	}
	return chunkArr
}

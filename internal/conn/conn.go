package conn

import (
	"bufio"
	"net"
	"time"
)

type Conn struct {
	Nc   net.Conn
	Rw   *bufio.ReadWriter
	Addr net.Addr
}

func NewConn(addr net.Addr) (*Conn, error) {
	conn := new(Conn)
	nc, err := net.DialTimeout(addr.Network(), addr.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}
	conn.Nc = nc
	conn.Rw = bufio.NewReadWriter(bufio.NewReader(nc), bufio.NewWriter(nc))
	conn.Addr = addr
	return conn, nil
}

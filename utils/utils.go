package utils

import (
	"net"
	"strconv"
	"strings"
)

func CompareVer(a, b string) (ret int) {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	loopMax := len(bs)
	if len(as) > len(bs) {
		loopMax = len(as)
	}
	for i := 0; i < loopMax; i++ {
		var x, y string
		if len(as) > i {
			x = as[i]
		}
		if len(bs) > i {
			y = bs[i]
		}
		xi, _ := strconv.Atoi(x)
		yi, _ := strconv.Atoi(y)
		if xi > yi {
			ret = 1
		} else if xi < yi {
			ret = -1
		}
		if ret != 0 {
			break
		}
	}
	return
}

func GetDumpFilename(addr net.Addr) string {
	filename := strings.Replace(addr.String(), ".", "_", -1)
	filename = strings.Replace(filename, ":", "_", -1) + ".txt"
	return filename
}

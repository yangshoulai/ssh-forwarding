package util

import (
	"net"
	"strconv"
)

func IsPortUsed(port int) bool {
	l, e := net.Listen("tcp", ":"+strconv.Itoa(port))
	if e != nil {
		return true
	}
	defer func() { _ = l.Close() }()
	return false
}

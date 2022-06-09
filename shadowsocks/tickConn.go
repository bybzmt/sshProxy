package shadowsocks

import (
	"net"
	"time"
)

type tickConn struct {
	net.Conn
	timeout time.Duration
}

func (c *tickConn) Read(b []byte) (n int, err error) {
	c.Conn.SetDeadline(time.Now().Add(c.timeout))
	return c.Conn.Read(b)
}

func (c *tickConn) Write(b []byte) (n int, err error) {
	c.Conn.SetDeadline(time.Now().Add(c.timeout))
	return c.Conn.Write(b)
}

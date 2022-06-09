package shadowsocks

import (
	"net"
)

type trafficConn struct {
	net.Conn
	traffic Traffic
	now     *Traffic
	end     *Traffic
}

func (c *trafficConn) Read(b []byte) (n int, err error) {
	defer func() {
		if n > 0 {
			c.traffic.Incoming += int64(n)
			if c.now != nil {
				c.now.AddIncoming(int64(n))
			}
		}
	}()

	return c.Conn.Read(b)
}

func (c *trafficConn) Write(b []byte) (n int, err error) {
	defer func() {
		if n > 0 {
			c.traffic.Outgoing += int64(n)
			if c.now != nil {
				c.now.AddOutgoing(int64(n))
			}
		}
	}()

	return c.Conn.Write(b)
}

func (c *trafficConn) Close() (err error) {
	if c.end != nil {
		c.end.Add(&c.traffic)
	}
	return c.Conn.Close()
}

package ui

import (
	"io"
	"net"
)

type listener struct {
	buf chan net.Conn
}

func (this *listener) Accept() (net.Conn, error) {
	c, ok := <-this.buf
	if ok {
		return c, nil
	}
	return nil, io.EOF
}

func (this *listener) Close() error {
	close(this.buf)
	return nil
}

func (this *listener) Addr() net.Addr {
	return &addr{}
}

type addr struct{}

func (t *addr) Network() string {
	return "tcp"
}
func (t *addr) String() string {
	return "shadowsocks"
}

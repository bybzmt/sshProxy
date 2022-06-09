package shadowsocks

import (
	"io"
	"net"
	"strings"
)

func Relay(a, b net.Conn) (err error) {
	if t, ok := a.(*net.TCPConn); ok {
		t.SetNoDelay(true)
		t.SetKeepAlive(true)
	}
	if t, ok := b.(*net.TCPConn); ok {
		t.SetNoDelay(true)
		t.SetKeepAlive(true)
	}

	ch := make(chan error, 1)

	go func() {
		_, e := io.Copy(a, b)
		ch <- e
	}()
	go func() {
		_, e := io.Copy(b, a)
		ch <- e
	}()

	//first err
	return <-ch
}

func StrSplit(str string) (out []string) {
	for _, tm := range strings.Split(str, "\n") {
		for _, t := range strings.Split(tm, "\r") {
			t = strings.TrimSpace(t)
			if len(t) > 0 {
				if t[0] != '#' {
					out = append(out, t)
				}
			}
		}
	}
	return
}

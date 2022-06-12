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
	for _, t1 := range strings.Split(str, "\n") {
		for _, t2 := range strings.Split(t1, "\r") {
			t2 = strings.TrimSpace(t2)
			if len(t2) > 0 {
				if t2[0] != '#' {
					for _, t3 := range strings.Split(t2, ",") {
						t3 = strings.TrimSpace(t3)
						if len(t3) > 0 {
							out = append(out, t3)
						}
					}
				}
			}
		}
	}
	return
}

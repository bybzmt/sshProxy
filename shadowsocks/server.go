package shadowsocks

import (
	"net"
	"time"
)

type Server struct {
	shadow *Shadow

	Watcher Watcher

	Traffic     Traffic
	idleTimeout time.Duration
	timeout     time.Duration
}

func NewServer(addr, cipher, passwd string, timeout int) (*Server, error) {
	c := &Server{}

	t, err := NewShadow("tcp", addr, cipher, passwd)
	if err != nil {
		return nil, err
	}

	c.shadow = t
	c.timeout = 3 * time.Second
	c.idleTimeout = time.Duration(timeout) * time.Second
	c.Watcher = DefaultWatcher

	return c, nil
}

func (s *Server) ListenAndServe() error {
	l, err := s.shadow.Listen()
	if err != nil {
		return err
	}

	for {
		c, e := l.Accept()
		if e != nil {
			continue
		}
		go s.Serve(c)
	}
	return nil
}

func (s *Server) Serve(c net.Conn) {
	defer c.Close()

	c.SetReadDeadline(time.Now().Add(s.timeout))

	from := s.shadow.Shadow(s.trafficConn(c))

	addr, err := ReadRawAddr(from)
	if err != nil {
		s.Watcher.OnShadowInvalid(c.RemoteAddr(), err)
		return
	}

	s.Watcher.OnProxyStart(false, c.RemoteAddr(), addr)
	defer func() {
		s.Watcher.OnProxyStop(false, c.RemoteAddr(), addr, err)
	}()

	to, err := net.DialTimeout(addr.Network(), addr.String(), s.timeout)
	if err != nil {
		Debug.Println("Dial", err)
		return
	}
	defer to.Close()

	err = Relay(s.tickConn(from, time.Second), s.tickConn(to, 0))
	return
}

func (s *Server) trafficConn(c net.Conn) *trafficConn {
	return &trafficConn{
		Conn: c,
		now:  &s.Traffic,
		end:  &s.shadow.Traffic,
	}
}

func (s *Server) tickConn(c net.Conn, sc time.Duration) *tickConn {
	return &tickConn{
		Conn:    c,
		timeout: s.idleTimeout + sc,
	}
}

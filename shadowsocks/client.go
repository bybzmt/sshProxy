package shadowsocks

import (
	"errors"
	"net"
	"strconv"
	"sync/atomic"
	"time"
)

var ErrAllServerUnavailable = errors.New("Failed connect to all available shadowsocks server")
var ErrDial = errors.New("Dial Error")

type Client struct {
	addr     string
	idx      uint32
	shadows  []*Shadow
	listener net.Listener

	localDNS  *dns
	remoteDNS *dns

	Watcher Watcher

	Traffic Traffic

	idleTimeout time.Duration
	timeout     time.Duration
}

func NewClient(addr string, timeout, idleTimeout int) *Client {
	c := &Client{}
	c.addr = addr

	c.timeout = time.Duration(timeout) * time.Second
	c.idleTimeout = time.Duration(idleTimeout) * time.Second
	c.Watcher = DefaultWatcher

	return c
}

func (c *Client) AddServer(id uint64, addr, cipher, user, passwd string) error {
	t, err := NewShadow("tcp", addr, cipher, user, passwd)
	if err != nil {
		return err
	}
	t.ID = id

	c.shadows = append(c.shadows, t)
	return nil
}

func (s *Client) SetLocalDNS(dns string) {
	var ips []string
	for _, t := range StrSplit(dns) {
		if ip := net.ParseIP(t); ip != nil {
			ips = append(ips, ip.String())
		}
	}

	if len(ips) == 0 {
		return
	}

	s.localDNS = NewDNS(ips)
}

func (s *Client) SetRemoteDNS(dns string) {
	var ips []string
	for _, t := range StrSplit(dns) {
		if ip := net.ParseIP(t); ip != nil {
			ips = append(ips, ip.String())
		}
	}

	if len(ips) == 0 {
		return
	}

	s.remoteDNS = NewDNS(ips)
	s.remoteDNS.Dial = func(n, addr string) (net.Conn, error) {
		host, _, _ := net.SplitHostPort(addr)

		server := s.match(host)
		if server != nil {
			return server.Dial(addr, s.timeout)
		} else {
			return net.DialTimeout(n, addr, s.timeout)
		}
	}
}

func (c *Client) AddRules(itmes, serverIds string) {
	var ids []uint64

	for _, r := range StrSplit(serverIds) {
		id, err := strconv.ParseUint(r, 10, 64)
		if err == nil {
			ids = append(ids, id)
		}
	}

	if len(ids) < 1 {
		for _, s := range c.shadows {
			ids = append(ids, s.ID)
		}
	}

	for _, id := range ids {
		for _, s := range c.shadows {
			if s.ID == id {
				for _, r := range StrSplit(itmes) {
					s.Rule.Add(r)
				}
			}
		}
	}
}

func (s *Client) ListenAndServe() (e error) {
	s.listener, e = net.Listen("tcp", s.addr)
	if e != nil {
		return e
	}

	for {
		c, e := s.listener.Accept()
		if e != nil {
			Debug.Println("Accept", e)
			return e
		}
		go s.Serve(c)
	}
}

func (s *Client) Close() {
	s.listener.Close()
	if s.localDNS != nil {
		s.localDNS.Close()
	}
	if s.remoteDNS != nil {
		s.remoteDNS.Close()
	}
}

func (s *Client) Serve(from net.Conn) {
	defer from.Close()

	from.SetReadDeadline(time.Now().Add(s.timeout))

	addr, err := HandShake(from)
	if err != nil {
		s.Watcher.OnSocksInvalid(from.RemoteAddr(), err)
		return
	}

	host := addr.Host()

	if s.Watcher.Hijacker(host, from) {
		return
	}

	from = s.trafficConn(from, &s.Traffic, nil)
	to, ac, err := s.dial(addr)

	s.Watcher.OnProxyStart(ac, from.RemoteAddr(), addr)
	defer func() {
		s.Watcher.OnProxyStop(ac, from.RemoteAddr(), addr, err)
	}()

	if err != nil {
		Debug.Println("Dial", err)
		return
	}
	defer to.Close()

	err = Relay(s.tickConn(from, time.Second), s.tickConn(to, 0))
	if err != nil {
		Debug.Println("Relay", err)
	}
}

func (c *Client) match(addr string) *Shadow {
	for i := 0; i < len(c.shadows); i++ {
		idx := int(atomic.AddUint32(&c.idx, 1) % uint32(len(c.shadows)))

		s := c.shadows[idx]

		if s.Rule.Match(addr) {
			return s
		}
	}

	return nil
}

func (s *Client) dial(addr RawAddr) (conn net.Conn, ac bool, err error) {
	server := s.match(addr.String())

	Debug.Println("Match", server != nil, addr.String())

	if server != nil {
		conn, err = s.dialShadow(server, addr)
		ac = true
		return
	}

	return s.dialLocal(addr)
}

func (c *Client) dialShadow(s *Shadow, addr RawAddr) (net.Conn, error) {
	var addrs []RawAddr

	if addr.ToIP() == nil && c.remoteDNS != nil {
		ipaddr, err := c.remoteDNS.LookupIPAddr(addr.Host())
		if err != nil {
			return nil, err
		}

		t := make([]RawAddr, len(ipaddr))
		for i, p := range ipaddr {
			t[i] = IP2RawAddr(p.IP, addr.Port())
		}
		addrs = t
	} else {
		addrs = append(addrs, addr)
	}

	for _, add := range addrs {
		to, err := s.Dial(add.String(), c.timeout)
		if err == nil {
			return to, nil
		}
	}

	return nil, ErrDial
}

func (c *Client) dialLocal(addr RawAddr) (net.Conn, bool, error) {
	var addrs []RawAddr

	if addr.ToIP() == nil && c.localDNS != nil {
		ipaddr, err := c.localDNS.LookupIPAddr(addr.Host())
		if err != nil {
			return nil, false, err
		}

		t := make([]RawAddr, len(ipaddr))
		for i, p := range ipaddr {
			t[i] = IP2RawAddr(p.IP, addr.Port())
		}
		addrs = t

		for _, add := range addrs {
			to, ac, err := c.dial(add)
			if err == nil {
				return to, ac, nil
			}
		}
	}

	to, err := net.DialTimeout("tcp", addr.String(), c.timeout)
	return to, false, err
}

func (s *Client) trafficConn(c net.Conn, now, end *Traffic) *trafficConn {
	return &trafficConn{
		Conn: c,
		now:  now,
		end:  end,
	}
}

func (s *Client) tickConn(c net.Conn, sc time.Duration) *tickConn {
	return &tickConn{
		Conn:    c,
		timeout: s.idleTimeout + sc,
	}
}

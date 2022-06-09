package shadowsocks

import (
	"errors"
	"net"
	"strconv"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/ssh"
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
	rule    Rules
	Proxy   bool

	idleTimeout time.Duration
	timeout     time.Duration
}

func NewClient(addr string, timeout, idleTimeout int) *Client {
	c := &Client{}
	c.addr = addr

	c.timeout = time.Duration(timeout) * time.Second
	c.idleTimeout = time.Duration(idleTimeout) * time.Second
	c.Watcher = DefaultWatcher

	c.rule.Init()

	return c
}

func (c *Client) AddServer(addr, cipher, passwd string) error {
	t, err := NewShadow("tcp", addr, cipher, passwd)
	if err != nil {
		return err
	}

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
		host, port, _ := net.SplitHostPort(addr)

		if s.match(host) {
			i, _ := strconv.ParseUint(port, 10, 16)
			raw := IP2RawAddr(net.ParseIP(host), uint16(i))
			return s.dialShadow(raw)
		} else {
			return net.Dial(n, addr)
		}
	}
}

func (s *Client) AddRules(itmes string) {
	for _, r := range StrSplit(itmes) {
		s.rule.Add(r)
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

	ac := s.match(host)

	s.Watcher.OnProxyStart(ac, from.RemoteAddr(), addr)
	defer func() {
		s.Watcher.OnProxyStop(ac, from.RemoteAddr(), addr, err)
	}()

	to, err := s.dial(ac, addr)
	if err != nil {
		Debug.Println("Dial", err)
		return
	}
	defer to.Close()

	err = Relay(s.tickConn(from, time.Second), s.tickConn(to, 0))
	return
}

func (s *Client) dial(ac bool, addr RawAddr) (conn net.Conn, err error) {
	if ac {
		if addr.ToIP() == nil && s.remoteDNS != nil {
			ipaddr, err := s.remoteDNS.LookupIPAddr(addr.Host())
			if err != nil {
				return nil, err
			}

			t := make([]RawAddr, len(ipaddr))
			for i, p := range ipaddr {
				t[i] = IP2RawAddr(p.IP, addr.Port())
			}

			return s.dialShadow(t...)
		} else {
			return s.dialShadow(addr)
		}
	} else {
		if addr.ToIP() == nil && s.localDNS != nil {
			ipaddr, err := s.localDNS.LookupIPAddr(addr.Host())
			if err != nil {
				return nil, err
			}

			for _, ip := range ipaddr {
				rel := net.JoinHostPort(ip.String(), addr.PortString())

				conn, err = net.DialTimeout("tcp", rel, s.timeout)
				if err != nil {
					Debug.Println("Dial", rel, err)
					continue
				}
				return conn, err
			}

			return nil, errors.New("Dial " + addr.Host() + " Error")
		} else {
			return net.DialTimeout(addr.Network(), addr.String(), s.timeout)
		}
	}
}

func (c *Client) dialShadow(addr ...RawAddr) (net.Conn, error) {
	for i := 0; i < len(c.shadows); i++ {
		for _, add := range addr {
			idx := int(atomic.AddUint32(&c.idx, 1) % uint32(len(c.shadows)))

			s := c.shadows[idx]

			to, err := c.dialShadowSingle(s, add)
			if err == nil {
				return to, nil
			}
			if err == ErrDial {
				break
			}
		}
	}

	return nil, ErrAllServerUnavailable
}

var sshclient *ssh.Client

func (c *Client) dialShadowSingle(s *Shadow, addr RawAddr) (net.Conn, error) {

	/*
		if sshclient == nil {
			config := &ssh.ClientConfig{
				User: "123",
				Auth: []ssh.AuthMethod{
					ssh.Password("345"),
				},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			}
			var err error
			sshclient, err = ssh.Dial("tcp", "0.0.0.0:22", config)
			if err != nil {
				log.Println(1, err)
				return nil, err
			}
		}

		n, err := sshclient.Dial("tcp", addr.String())
		if err != nil {
			log.Println(2, err)
		}
		return n, err

		dialSocksProxy := socks.Dial("socks4://" + s.Address + "?timeout=5s")
		return dialSocksProxy("", addr.String())
	*/

	to, err := s.Dial(c.timeout)
	if err != nil {
		Debug.Println("Dial Shadow", s.Address, "To", addr.String(), err)
		return nil, ErrDial
	}

	to.SetReadDeadline(time.Now().Add(c.timeout))

	t2 := s.Shadow(c.trafficConn(to, nil, &s.Traffic))

	if _, err = t2.Write(addr); err != nil {
		t2.Close()
		Debug.Println("Dial Shadow", s.Address, "To", addr.String(), err)
		return nil, err
	}

	return t2, nil
}

func (s *Client) match(host string) bool {
	if s.rule.Match(host) {
		return !s.Proxy
	}
	return s.Proxy
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

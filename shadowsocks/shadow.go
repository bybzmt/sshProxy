package shadowsocks

import (
	"errors"
	"net"
	"time"

	shadow "sshProxy/shadowsocks/shadow"
)

var errEmptyPassword = errors.New("empty key")

type Creater func(net.Conn) net.Conn

type Shadow struct {
	Network string
	Address string
	Shadow  Creater
	Traffic Traffic
}

func AllCiphers() []string {
	return shadow.ListCipher()
}

func NewShadow(network, addr, cipher, password string) (*Shadow, error) {
	/*
		if password == "" {
			return nil, errEmptyPassword
		}
	*/

	var key []byte
	c, err := shadow.PickCipher(cipher, key, password)
	if err != nil {
		return nil, err
	}

	s := &Shadow{
		Network: network,
		Address: addr,
		Shadow:  c.StreamConn,
	}

	return s, nil
}

// func (s *Shadow) Shadow(c net.Conn) *ShadowConn {
// tc := &TrafficConn{
// Conn: c,
// s:    s.Traffic,
// }

// return &ShadowConn{
// Conn:    s.Creater(tc),
// traffic: &tc.traffic,
// }
// }

func (s *Shadow) Dial(timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout(s.Network, s.Address, timeout)
}

func (s *Shadow) Listen() (l net.Listener, err error) {
	return net.Listen(s.Network, s.Address)
}

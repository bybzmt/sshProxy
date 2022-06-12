package shadowsocks

import (
	"errors"
	"net"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
	"h12.io/socks"

	shadow "sshProxy/shadowsocks/shadow"
)

var errEmptyPassword = errors.New("empty key")

type Creater func(net.Conn) net.Conn

type Shadow struct {
	ID        uint64
	Network   string
	Address   string
	Rule      Rules
	Shadow    Creater
	Traffic   Traffic
	sshConfig *ssh.ClientConfig
	ssh       *ssh.Client
	ss        shadow.Cipher
	Dial      func(string, time.Duration) (net.Conn, error)
}

func AllCiphers() []string {
	return shadow.ListCipher()
}

func NewShadow(network, addr, cipher, user, password string) (*Shadow, error) {
	s := &Shadow{
		Network: network,
		Address: addr,
	}

	if cipher == "SOCKS4" {
		s.Network = "socks4"
		s.Dial = s.DialSocks
	} else if cipher == "SOCKS5" {
		s.Network = "socks5"
		s.Dial = s.DialSocks
	} else if cipher == "SSH(Password)" {
		s.Dial = s.DialSSH

		s.sshConfig = &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				ssh.Password(password),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	} else if cipher == "SSH(PublicKeys)" {
		s.Dial = s.DialSSH

		signer, err := ssh.ParsePrivateKey([]byte(password))
		if err != nil {
			Debug.Println("unable to parse private key:", err)
			return nil, err
		}

		s.sshConfig = &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	} else {
		var key []byte
		c, err := shadow.PickCipher(cipher, key, password)
		if err != nil {
			return nil, err
		}

		s.ss = c
		s.Dial = s.DialSS
	}

	s.Rule.Init()

	return s, nil
}

func (s *Shadow) Listen() (l net.Listener, err error) {
	return net.Listen(s.Network, s.Address)
}

func (s *Shadow) DialSocks(addr string, timeout time.Duration) (net.Conn, error) {
	num := strconv.Itoa(int(timeout / time.Second))

	dialSocksProxy := socks.Dial(s.Network + "//" + s.Address + "?timeout=" + num + "s")
	return dialSocksProxy("", addr)
}

func (s *Shadow) DialSS(addr string, timeout time.Duration) (net.Conn, error) {
	raw, err := Parse2RawAddr(addr)
	if err != nil {
		return nil, err
	}

	c, err := net.DialTimeout(s.Network, s.Address, timeout)
	if err != nil {
		return nil, err
	}

	c.SetReadDeadline(time.Now().Add(timeout))

	var t2 net.Conn
	t2 = s.ss.StreamConn(c)

	if _, err = t2.Write(raw); err != nil {
		t2.Close()
		Debug.Println("Dial Shadow", s.Address, "To", addr, err)
		return nil, err
	}

	return t2, nil
}

func (s *Shadow) DialSSH(addr string, timeout time.Duration) (net.Conn, error) {
	if s.ssh == nil {
		var err error
		s.ssh, err = ssh.Dial("tcp", s.Address, s.sshConfig)
		if err != nil {
			Debug.Println("SSH Dial", err)
			return nil, err
		}
	}

	n, err := s.ssh.Dial(s.Network, addr)
	if err != nil {
		Debug.Println("Dial From SSH", err)
	}
	return n, err
}

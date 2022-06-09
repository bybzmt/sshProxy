package shadowsocks

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strconv"
)

var (
	errVer           = errors.New("socks version not supported")
	errMethod        = errors.New("socks only support 1 method now")
	errAuthExtraData = errors.New("socks authentication get extra data")
	errReqExtraData  = errors.New("socks request get extra data")
	errCmd           = errors.New("socks command not supported")
	errAddrType      = errors.New("socks addr type not supported")
)

const (
	SOCKS_VER         = 5
	SOCKS_CMD_CONNECT = 1

	SOCKS_ATYP_IPV4       = 1
	SOCKS_ATYP_DOMAINNAME = 3
	SOCKS_ATYP_IPV6       = 4
)

func HandShake(rw io.ReadWriter) (RawAddr, error) {
	idVer := 0
	idNMethods := 1
	h := [2]byte{}

	// The client connects to the server, and sends a version
	// identifier/method selection message:
	// +----+----------+----------+
	// |VER | NMETHODS | METHODS  |
	// +----+----------+----------+
	// | 1  |    1     | 1 to 255 |
	// +----+----------+----------+
	if _, err := io.ReadFull(rw, h[:]); err != nil {
		return nil, err
	}

	if h[idVer] != SOCKS_VER {
		return nil, errVer
	}

	methods := make([]byte, int(h[idNMethods]))
	if _, err := io.ReadFull(rw, methods); err != nil {
		return nil, err
	}

	// send confirmation: version 5, no authentication required
	// write VER METHOD
	if _, err := rw.Write([]byte{SOCKS_VER, 0}); err != nil {
		return nil, err
	}

	// The SOCKS request is formed as follows:
	// +----+-----+-------+------+----------+----------+
	// |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	// +----+-----+-------+------+----------+----------+
	// | 1  |  1  | X'00' |  1   | Variable |    2     |
	// +----+-----+-------+------+----------+----------+
	//VER + CMD + RSV
	idCmd := 1
	h2 := [3]byte{}

	if _, err := io.ReadFull(rw, h2[:]); err != nil {
		return nil, err
	}

	// check version and cmd
	if h2[idVer] != SOCKS_VER {
		return nil, errVer
	}
	if h2[idCmd] != SOCKS_CMD_CONNECT {
		return nil, errCmd
	}

	addr, err := ReadRawAddr(rw)
	if err != nil {
		return nil, err
	}

	// SOCKS v5, reply succeeded
	if _, err := rw.Write([]byte{SOCKS_VER, 0, 0, 1, 0, 0, 0, 0, 0, 0}); err != nil {
		return nil, err
	}

	return addr, nil
}

// The RawAddr is formed as follows:
// +------+----------+----------+
// | ATYP | DST.ADDR | DST.PORT |
// +------+----------+----------+
// |  1   | Variable |    2     |
// +------+----------+----------+
type RawAddr []byte

func ReadRawAddr(r io.Reader) (RawAddr, error) {
	idAtyp := 0
	idLen := 1
	h := [2]byte{}

	if _, err := io.ReadFull(r, h[:]); err != nil {
		return nil, err
	}

	var l int
	switch h[idAtyp] {
	case SOCKS_ATYP_IPV4:
		l = net.IPv4len
	case SOCKS_ATYP_IPV6:
		l = net.IPv6len
	case SOCKS_ATYP_DOMAINNAME:
		l = int(h[idLen]) + 1
	default:
		return nil, errAddrType
	}

	//atype(1) + len + prot(2)
	buf := make([]byte, 1+l+2)
	copy(buf, h[:])

	if _, err := io.ReadFull(r, buf[2:]); err != nil {
		return nil, err
	}

	return RawAddr(buf), nil
}

func (r RawAddr) ToIP() net.IP {
	if len(r) < 5 {
		return nil
	}

	switch r[0] {
	case SOCKS_ATYP_IPV4, SOCKS_ATYP_IPV6:
		return net.IP(r[1 : len(r)-2])
	}

	return nil
}

func (r RawAddr) Host() string {
	//atype(1) + len(min 1) + port(2)
	if len(r) < 4 {
		return ""
	}

	var host string
	switch r[0] {
	case SOCKS_ATYP_IPV4, SOCKS_ATYP_IPV6:
		host = net.IP(r[1 : len(r)-2]).String()
	case SOCKS_ATYP_DOMAINNAME:
		host = string(r[2 : len(r)-2])
	}
	return host
}

func (r RawAddr) Port() uint16 {
	if len(r) < 2 {
		return 0
	}
	return binary.BigEndian.Uint16(r[len(r)-2:])
}

func (r RawAddr) PortString() string {
	if len(r) < 2 {
		return ""
	}
	return strconv.Itoa(int(r.Port()))
}

func (r RawAddr) String() string {
	return net.JoinHostPort(r.Host(), r.PortString())
}

//interface net.Addr
func (r RawAddr) Network() string {
	return "tcp"
}

func IP2RawAddr(ip net.IP, port uint16) RawAddr {
	var raw []byte

	if t := ip.To4(); t != nil {
		raw = make([]byte, 1+net.IPv4len+2)
		raw[0] = SOCKS_ATYP_IPV4
		copy(raw[1:], t)
	} else {
		raw = make([]byte, 1+net.IPv6len+2)
		raw[0] = SOCKS_ATYP_IPV6
		copy(raw[1:], ip)
	}

	raw[len(raw)-2] = uint8((port & 0xff00) >> 8)
	raw[len(raw)-1] = uint8((port & 0xff))

	return RawAddr(raw)
}

func Addr2RawAddr(a net.Addr) (RawAddr, error) {
	if rel, ok := a.(RawAddr); ok {
		return rel, nil
	}
	return Parse2RawAddr(a.String())
}

func Parse2RawAddr(addr string) (RawAddr, error) {
	h, p, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	port, _ := strconv.ParseUint(p, 10, 16)

	if ip := net.ParseIP(h); ip != nil {
		return IP2RawAddr(ip, uint16(port)), nil
	}

	raw := make([]byte, 1+1+len(h)+2)
	raw[0] = SOCKS_ATYP_DOMAINNAME
	raw[1] = uint8(len(h))
	raw[len(raw)-2] = uint8((port & 0xff00) >> 8)
	raw[len(raw)-1] = uint8(port & 0xff)

	return RawAddr(raw), nil
}

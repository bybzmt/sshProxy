package pac

import (
	"net"
	"sort"
)

type IPNets interface {
	Match(net.IP) bool
	Add(net.IPNet)
	Each(func(net.IPNet))
}

func NewIPNets() IPNets {
	return new(ipNets)
}

type ipv4 [net.IPv4len]byte
type ipv6 [net.IPv6len]byte
type ipv6s map[ipv6]struct{}
type ipv4s map[ipv4]struct{}

type ipv4Nets struct {
	ips  ipv4s
	mask net.IPMask
}

type ipv6Nets struct {
	ips  ipv6s
	mask net.IPMask
}

type ipNets struct {
	v4 []ipv4Nets
	v6 []ipv6Nets
}

func (s ipv6s) add(ip net.IP) {
	var t ipv6
	copy(t[:], ip)
	s[t] = struct{}{}
}

func (s ipv6s) match(ip net.IP) bool {
	var t ipv6
	copy(t[:], ip)
	_, ok := s[t]
	return ok
}

func (s ipv4s) add(ip net.IP) {
	var t ipv4
	copy(t[:], ip)
	s[t] = struct{}{}
}

func (s ipv4s) match(ip net.IP) bool {
	var t ipv4
	copy(t[:], ip)
	_, ok := s[t]
	return ok
}

func (s *ipNets) Add(n net.IPNet) {
	o1, _ := n.Mask.Size()

	if ip := n.IP.To4(); ip != nil && len(n.Mask) == net.IPv4len {
		for _, m := range s.v4 {
			o2, _ := m.mask.Size()
			if o1 == o2 {
				m.ips.add(ip)
				return
			}
		}

		t := ipv4Nets{
			ips:  make(ipv4s),
			mask: n.Mask,
		}
		t.ips.add(ip)

		s.v4 = append(s.v4, t)
		sort.Sort(ipv4NetsSort(s.v4))
	}

	if ip := n.IP.To16(); ip != nil && len(n.Mask) == net.IPv6len {
		for _, m := range s.v6 {
			o2, _ := m.mask.Size()
			if o1 == o2 {
				m.ips.add(ip)
				return
			}
		}

		t := ipv6Nets{
			ips:  make(ipv6s),
			mask: n.Mask,
		}
		t.ips.add(ip)

		s.v6 = append(s.v6, t)
		sort.Sort(ipv6NetsSort(s.v6))
	}
}

func (s *ipNets) Match(ip net.IP) bool {
	if t := ip.To4(); t != nil {
		for _, n := range s.v4 {
			t = t.Mask(n.mask)
			if ok := n.ips.match(t); ok {
				return true
			}
		}
	} else if t := ip.To16(); t != nil {
		for _, n := range s.v6 {
			t = t.Mask(n.mask)
			if ok := n.ips.match(t); ok {
				return true
			}
		}
	}
	return false
}

func (s *ipNets) Each(fn func(net.IPNet)) {
	for _, n := range s.v4 {
		for ip, _ := range n.ips {
			t := net.IPNet{
				IP:   net.IP(ip[:]),
				Mask: n.mask,
			}
			fn(t)
		}
	}
	for _, n := range s.v6 {
		for ip, _ := range n.ips {
			t := net.IPNet{
				IP:   net.IP(ip[:]),
				Mask: n.mask,
			}
			fn(t)
		}
	}
}

package shadowsocks

import (
	"net"
	"regexp"
	"regexp/syntax"
	pac "sshProxy/shadowsocks/pac"
	"strings"
)

type Rules struct {
	Host     pac.Domain
	IP       pac.IPNets
	Regs     []regexp.Regexp
	ServerID []uint64
}

func (r *Rules) Init() {
	r.Host = make(pac.Domain)
	r.IP = pac.NewIPNets()
}

func (r *Rules) Add(rule string) {
	_, ipnet, err := net.ParseCIDR(rule)
	if err == nil {
		r.IP.Add(*ipnet)
	} else if len(rule) > 2 && rule[0] == '%' {
		reg, err := syntax.Parse(rule[1:len(rule)-1], syntax.Perl)
		if err != nil {
			Debug.Println("Rule Regexp", err)
		} else {
			r.Regs = append(r.Regs, *regexp.MustCompile(reg.String()))
		}
	} else {
		r.Host.Add(rule)
	}
}

func (r *Rules) Match(host string) bool {
	tmp := strings.Split(host, ":")
	host = tmp[0]

	ip := net.ParseIP(host)

	if ip != nil {
		return r.IP.Match(ip)
	}
	if r.Host.Match(host) {
		return true
	}
	for _, reg := range r.Regs {
		if reg.MatchString(host) {
			return true
		}
	}

	return false
}

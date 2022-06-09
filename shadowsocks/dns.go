package shadowsocks

import (
	"context"
	lru "github.com/hashicorp/golang-lru"
	"net"
	"sync"
	"time"
)

type dnsVal struct {
	ipaddr []net.IPAddr
	expire time.Time
}

type dnsLocker struct {
	l   sync.Mutex
	num int
}

type dns struct {
	dns      []string
	lru      *lru.Cache
	job      chan string
	idx      uint16
	resolver net.Resolver
	l        sync.Mutex
	lh       map[string]*dnsLocker

	Dial func(n, addr string) (net.Conn, error)
}

func NewDNS(ips []string) *dns {
	lru, _ := lru.New(2000)

	d := &dns{
		dns: ips,
		lru: lru,
		job: make(chan string, 100),
		lh:  make(map[string]*dnsLocker),
	}

	d.resolver.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
		ip := d.dns[int(d.idx)%len(d.dns)]
		d.idx++

		_, port, _ := net.SplitHostPort(address)
		addr := net.JoinHostPort(ip, port)

		if d.Dial != nil {
			return d.Dial(network, addr)
		} else {
			return net.Dial(network, addr)
		}
	}

	go d.run()

	return d
}

func (d *dns) Close() {
	close(d.job)
}

func (d *dns) run() {
	for {
		host, ok := <-d.job
		if !ok {
			break
		}

		if t, ok := d.lru.Get(host); ok {
			if time.Now().Before(t.(dnsVal).expire) {
				continue
			}
		}

		go d._lookupIPAddr(host)
	}
}

func (d *dns) _lookupIPAddr(host string) ([]net.IPAddr, error) {
	ipaddr, err := d.resolver.LookupIPAddr(context.Background(), host)
	if err != nil {
		Debug.Println("Lookup", host, err)
		return nil, err
	}

	Debug.Println("Lookup", host, ipaddr)

	v := dnsVal{
		ipaddr: ipaddr,
		expire: time.Now().Add(2 * time.Minute),
	}

	d.lru.Add(host, v)

	return ipaddr, nil
}

func (d *dns) LookupIPAddr(host string) ([]net.IPAddr, error) {
	d.l.Lock()
	l, ok := d.lh[host]
	if !ok {
		l = &dnsLocker{}
		d.lh[host] = l
	}
	l.num++
	d.l.Unlock()

	defer func() {
		d.l.Lock()
		defer d.l.Unlock()

		l, ok := d.lh[host]
		if ok {
			l.num--
			if l.num <= 0 {
				delete(d.lh, host)
			}
		}
	}()

	l.l.Lock()
	defer l.l.Unlock()

	if t, ok := d.lru.Get(host); ok {
		v := t.(dnsVal)

		if time.Now().After(v.expire) {
			d.job <- host
			Debug.Println("Lookup Refresh", host, v.ipaddr)
		} else {
			Debug.Println("Lookup Cached", host, v.ipaddr)
		}

		return v.ipaddr, nil
	}

	return d._lookupIPAddr(host)
}

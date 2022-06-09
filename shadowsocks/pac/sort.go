package pac

import (
// "sort"
)

type ipv4NetsSort []ipv4Nets

func (p ipv4NetsSort) Len() int {
	return len(p)
}

func (p ipv4NetsSort) Less(i, j int) bool {
	o1, _ := p[i].mask.Size()
	o2, _ := p[j].mask.Size()
	if o1 < o2 {
		return true
	}
	return false
}

func (p ipv4NetsSort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

type ipv6NetsSort []ipv6Nets

func (p ipv6NetsSort) Len() int {
	return len(p)
}

func (p ipv6NetsSort) Less(i, j int) bool {
	o1, _ := p[i].mask.Size()
	o2, _ := p[j].mask.Size()
	if o1 < o2 {
		return true
	}
	return false
}

func (p ipv6NetsSort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

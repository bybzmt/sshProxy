package pac

import (
	"strings"
)

type Domain map[string]Domain

func (d Domain) Match(host string) bool {
	part := strings.Split(strings.Trim(strings.TrimSpace(host), "."), ".")

	top := d
	for i := len(part) - 1; i >= 0; i-- {
		if h, ok := top[part[i]]; ok {
			if h == nil {
				return true
			} else {
				top = h
			}
		} else {
			return false
		}
	}

	return false
}

func (d Domain) Add(host string) {
	part := strings.Split(strings.Trim(strings.TrimSpace(host), "."), ".")

	top := d
	for i := len(part) - 1; i >= 0; i-- {
		if h, ok := top[part[i]]; ok {
			if i == 0 {
				top[part[i]] = nil
			} else if h != nil {
				top = h
			} else {
				break
			}
		} else {
			if i == 0 {
				top[part[i]] = nil
			} else {
				top[part[i]] = Domain{}
				top = top[part[i]]
			}
		}
	}
}

func (d Domain) Each(fn func(string)) {
	for k, v := range d {
		if v == nil {
			fn(k)
		} else {
			_domainEach(v, k, fn)
		}
	}
}

func _domainEach(d Domain, p string, fn func(string)) {
	for k, v := range d {
		if v == nil {
			fn(k + "." + p)
		} else {
			_domainEach(v, k+"."+p, fn)
		}
	}
}

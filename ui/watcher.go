package ui

import (
	"io"
	"net"
	ss "sshProxy/shadowsocks"
	"strings"
	"sync/atomic"
	"time"
)

type uiWatcher struct {
	l       *listener
	counter int32
	buf     chan *LogMsg
	host    string
}

func (this *uiWatcher) OnSocksInvalid(from net.Addr, err error) {
	this.buf <- &LogMsg{
		Now:  time.Now(),
		From: from.String(),
		Msg:  "SocksInvalid: " + err.Error(),
	}
}

func (this *uiWatcher) OnShadowInvalid(from net.Addr, err error) {
	this.buf <- &LogMsg{
		Now:  time.Now(),
		From: from.String(),
		Msg:  "ShadowInvalid: " + err.Error(),
	}
}

func (this *uiWatcher) OnProxyStart(proxy bool, from, to net.Addr) {
	this.buf <- &LogMsg{
		Now:   time.Now(),
		Proxy: proxy,
		From:  from.String(),
		To:    to.String(),
	}

	atomic.AddInt32(&this.counter, 1)
}

func (this *uiWatcher) OnProxyStop(proxy bool, from, to net.Addr, err error) {
	msg := "success"
	if err != nil && err != io.EOF {
		msg = err.Error()
	}

	this.buf <- &LogMsg{
		Proxy: proxy,
		From:  from.String(),
		To:    to.String(),
		Msg:   msg,
	}

	atomic.AddInt32(&this.counter, -1)
}

func (this *uiWatcher) Hijacker(host string, c net.Conn) bool {
	if strings.ToLower(host) != this.host {
		return false
	}

	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()

	this.l.buf <- a

	ss.Relay(b, c)
	return true
}

package shadowsocks

import (
	"sync/atomic"
)

type Traffic struct {
	Incoming int64
	Outgoing int64
}

func (t *Traffic) Clone() Traffic {
	incoming := atomic.LoadInt64(&t.Incoming)
	outgoing := atomic.LoadInt64(&t.Outgoing)

	return Traffic{
		Incoming: incoming,
		Outgoing: outgoing,
	}
}

func (t *Traffic) Sub(u *Traffic) {
	atomic.AddInt64(&t.Incoming, -u.Incoming)
	atomic.AddInt64(&t.Outgoing, -u.Outgoing)
}

func (t *Traffic) Add(u *Traffic) {
	atomic.AddInt64(&t.Incoming, u.Incoming)
	atomic.AddInt64(&t.Outgoing, u.Outgoing)
}

func (t *Traffic) AddIncoming(i int64) {
	atomic.AddInt64(&t.Incoming, i)
}

func (t *Traffic) AddOutgoing(i int64) {
	atomic.AddInt64(&t.Outgoing, i)
}

package ui

import (
	"time"
)

type LogMsg struct {
	ID    uint64 `bolthold:"key"`
	Now   time.Time
	Proxy bool
	From  string
	To    string
	Msg   string
}

type ClientConfig struct {
	Addr        string
	Timeout     int
	IdleTimeout int
	Proxy       bool
	LDNS        string
	RDNS        string
}

type ServerConfig struct {
	ID     uint64 `bolthold:"key"`
	Addr   string
	Passwd string
	Cipher string
	Note   string
	Enable bool
}

type Rules struct {
	ID     uint64 `bolthold:"key"`
	Proxy  bool
	Note   string
	Enable bool
	Items  string
}

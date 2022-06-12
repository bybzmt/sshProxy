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
	LDNS        string
	LDNSEnable  bool
	RDNS        string
	RDNSEnable  bool
}

type ServerConfig struct {
	ID     uint64 `bolthold:"key"`
	Addr   string
	Cipher string
	User   string
	Passwd string
	Note   string
	Enable bool
}

type Rules struct {
	ID      uint64 `bolthold:"key"`
	Note    string
	Enable  bool
	Items   string
	Servers string
}

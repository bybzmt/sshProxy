package shadowsocks

import (
	"log"
	"os"
)

type DebugLog bool

var Debug DebugLog

var dbgLog = log.New(os.Stdout, "[DEBUG] ", log.Ltime)

func SetDebug(d DebugLog) {
	Debug = d
}

func (d DebugLog) Printf(format string, args ...interface{}) {
	if d {
		dbgLog.Printf(format, args...)
	}
}

func (d DebugLog) Println(args ...interface{}) {
	if d {
		dbgLog.Println(args...)
	}
}

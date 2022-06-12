package main

import (
	"log"
	"os"
	"strings"
	"sync/atomic"
	"time"

	ss "sshProxy/shadowsocks"
	"sshProxy/ui"

	flaggy "github.com/integrii/flaggy"
)

var debug ss.DebugLog

type flg struct {
	addr, cipher string
	user, passwd string
	c_addr, host string
	db           string
	rules        []string
	timeout      int
	LDNS         []string
	RDNS         []string
}

func main() {
	log.SetOutput(os.Stdout)

	flaggy.SetVersion("v1.0")
	flaggy.Bool((*bool)(&debug), "d", "debug", "print debug log")
	flaggy.SetDescription("shadowsocks client and server")

	var f flg

	server := flaggy.NewSubcommand("server")
	client := flaggy.NewSubcommand("client")
	ui := flaggy.NewSubcommand("clientUI")

	server.Description = "shadowsocks server"
	client.Description = "shadowsocks client"
	ui.Description = "shadowsocks client with http ui. default"

	server.String(&f.addr, "a", "addr", "shadowsocks listen on addr:port. default :1080")
	server.String(&f.cipher, "c", "cipher", "cipher: "+strings.Join(ss.AllCiphers(), " "))
	server.String(&f.user, "u", "user", "user")
	server.String(&f.passwd, "p", "passwd", "password")
	server.Int(&f.timeout, "t", "timeout", "timeout in seconds. default 65s")

	client.String(&f.c_addr, "a", "addr", "socks5 listen on addr:port. default :1080")
	client.String(&f.addr, "s", "server", "server addr:port")
	client.String(&f.cipher, "c", "cipher", "server cipher: "+strings.Join(ss.AllCiphers(), " "))
	client.String(&f.user, "u", "user", "server user")
	client.String(&f.passwd, "p", "passwd", "server password")
	client.Int(&f.timeout, "t", "timeout", "timeout in seconds. default 65s")
	client.StringSlice(&f.rules, "", "rule", "pac rule file")
	client.StringSlice(&f.LDNS, "", "LDNS", "local direct dns")
	client.StringSlice(&f.RDNS, "", "RDNS", "remote proxy dns")

	ui.String(&f.addr, "a", "addr", "shadowsocks listen on addr:port")
	ui.String(&f.db, "", "db", "database file. default: ./shadowsocks.db")
	ui.String(&f.host, "", "host", "web ui host default: shadowsocks")

	flaggy.AttachSubcommand(client, 1)
	flaggy.AttachSubcommand(server, 1)
	flaggy.AttachSubcommand(ui, 1)
	flaggy.DefaultParser.HelpTemplate = newHelpTemplate()

	flaggy.Parse()

	ss.SetDebug(debug)

	if f.timeout < 3 {
		f.timeout = 65
	}
	if f.addr == "" && server.Used {
		f.addr = ":1080"
	}
	if f.c_addr == "" {
		f.c_addr = ":1080"
	}
	if f.db == "" {
		f.db = "shadowsocks.db"
	}
	if f.host == "" {
		f.host = "shadowsocks"
	}

	if client.Used {
		runClient(f)
	} else if server.Used {
		runServer(f)
	} else {
		runClientUI(f)
	}
}

func runServer(f flg) {
	server, err := ss.NewServer(f.addr, f.cipher, f.user, f.passwd, f.timeout)
	if err != nil {
		log.Println("NewServer Error:", err)
		os.Exit(1)
	}

	log.Println("Starting Shadowsocks Server At", f.addr)

	err = server.ListenAndServe()
	if err != nil {
		log.Println("ListenAndServe Error:", err)
		os.Exit(1)
	}
}

func runClient(f flg) {

	client := ss.NewClient(f.c_addr, 3, f.timeout)

	if f.addr != "" {
		err := client.AddServer(0, f.addr, f.cipher, f.user, f.passwd)
		if err != nil {
			log.Println("Server Error", err)
			os.Exit(1)
		}
	}

	for _, t := range f.rules {
		ts, err := loadFile(t)
		if err != nil {
			log.Println("Load Rule", err)
			os.Exit(1)
		}
		client.AddRules(ts, "")
	}
	for _, t := range f.LDNS {
		client.SetLocalDNS(t)
	}
	for _, t := range f.RDNS {
		client.SetRemoteDNS(t)
	}

	log.Println("Starting Shadowsocks Client At", f.c_addr)

	go cliState(client)

	err := client.ListenAndServe()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func cliState(c *ss.Client) {
	for _ = range time.Tick(10 * time.Second) {
		num := atomic.LoadInt32(&ss.DefaultWatcher.Counter)

		t := c.Traffic.Clone()
		c.Traffic.Sub(&t)

		Outgoing := ui.FmtSize(10*time.Second, t.Outgoing)
		Incoming := ui.FmtSize(10*time.Second, t.Incoming)

		log.Println("state", "Conn", num, "Outgoing", Incoming, "Incoming", Outgoing)
	}
}

func runClientUI(f flg) {
	log.Println("Starting Shadowsocks Database", f.db)

	u := ui.NewUI(f.db, f.addr, f.host)

	err := u.Run()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

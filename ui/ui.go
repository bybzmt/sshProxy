package ui

import (
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bybzmt/bolthold"

	"embed"
	"io/fs"
	ss "sshProxy/shadowsocks"
)

//go:embed dist/*
var uifiles embed.FS

type ui struct {
	l   sync.Mutex
	now time.Time

	listener listener
	reStart  bool
	watcher  uiWatcher

	handler    *http.ServeMux
	httpServer http.Server

	ssServer *ss.Client

	storeFile string
	cliAddr   string
	store     *bolthold.Store
}

func NewUI(cfg, addr, host string) *ui {
	u := &ui{}

	u.listener.buf = make(chan net.Conn, 5)

	u.watcher.buf = make(chan *LogMsg, 100)
	u.watcher.l = &u.listener
	u.watcher.host = strings.ToLower(host)

	u.handler = http.NewServeMux()
	u.httpServer.Handler = u.handler

	u.init()
	u.storeFile = cfg
	u.cliAddr = addr

	return u
}

func (this *ui) init() {

	tfs, _ := fs.Sub(uifiles, "dist")

	this.handler.Handle("/", http.FileServer(http.FS(tfs)))

	this.handler.HandleFunc("/api/state", this.cross(this.apiState))
	this.handler.HandleFunc("/api/log", this.cross(this.apiLog))
	this.handler.HandleFunc("/api/rules", this.cross(this.apiRules))
	this.handler.HandleFunc("/api/ruleAdd", this.cross(this.apiRuleAdd))
	this.handler.HandleFunc("/api/ruleEdit", this.cross(this.apiRuleEdit))
	this.handler.HandleFunc("/api/ruleDel", this.cross(this.apiRuleDel))
	this.handler.HandleFunc("/api/clientConfig", this.cross(this.apiClientConfig))
	this.handler.HandleFunc("/api/clientConfigSave", this.cross(this.apiClientConfigSave))
	this.handler.HandleFunc("/api/serverConfigs", this.cross(this.apiServerConfigs))
	this.handler.HandleFunc("/api/serverConfigAdd", this.cross(this.apiServerConfigAdd))
	this.handler.HandleFunc("/api/serverConfigEdit", this.cross(this.apiServerConfigEdit))
	this.handler.HandleFunc("/api/serverConfigDel", this.cross(this.apiServerConfigDel))
	this.handler.HandleFunc("/api/restart", this.cross(this.apiRestart))
}

func (this *ui) Run() error {
	var err error
	this.store, err = bolthold.Open(this.storeFile, 0644, nil)
	if err != nil {
		return err
	}

	go this.runStore()
	go func() {
		e := this.httpServer.Serve(&this.listener)
		if e != nil {
			ss.Debug.Println("httpServer", e)
		}
	}()

	return this.runClient()
}

func (this *ui) Close() {
	this.ssServer.Close()
}

func (this *ui) Restart() {
	this.reStart = true
	this.ssServer.Close()
}

func (this *ui) initClient() {
	var rs ClientConfig

	err := this.store.Get("ClientConfig", &rs)
	if err != nil {
		ss.Debug.Println("initClientConfig", err)
	}
	if rs.Addr == "" {
		rs.Addr = "127.0.0.1:1080"
	}
	if rs.Timeout < 1 {
		rs.Timeout = 5
	}
	if rs.IdleTimeout < 5 {
		rs.IdleTimeout = 60
	}

	if this.cliAddr != "" && rs.Addr != this.cliAddr {
		ss.Debug.Println("cli addr cover config")
		rs.Addr = this.cliAddr
	}

	log.Println("Starting Client At", rs.Addr)
	log.Println("open http://" + this.watcher.host)

	this.ssServer = ss.NewClient(rs.Addr, rs.Timeout, rs.IdleTimeout)
	if rs.LDNSEnable && rs.LDNS != "" {
		this.ssServer.SetLocalDNS(rs.LDNS)
	}
	if rs.RDNSEnable && rs.RDNS != "" {
		this.ssServer.SetRemoteDNS(rs.RDNS)
	}
	this.ssServer.Watcher = &this.watcher
}

func (this *ui) initServer() {
	var rs []ServerConfig

	err := this.store.Find(&rs, bolthold.Where("Enable").Eq(true))
	if err != nil {
		ss.Debug.Println("initServerConfig", err)
	}

	for _, r := range rs {
		err := this.ssServer.AddServer(r.ID, r.Addr, r.Cipher, r.User, r.Passwd)
		if err != nil {
			ss.Debug.Println("AddServer", err)
		}
	}
}

func (this *ui) initRules() {
	var rs []Rules

	err := this.store.Find(&rs,
		bolthold.Where("Enable").Eq(true))
	if err != nil {
		ss.Debug.Println("initRules", err)
	}

	for _, r := range rs {
		this.ssServer.AddRules(r.Items, r.Servers)
	}
}

func (this *ui) runClient() error {
	for {
		this.initClient()
		this.initServer()
		this.initRules()

		err := this.ssServer.ListenAndServe()

		if this.reStart {
			this.reStart = false

			if err != nil {
				ss.Debug.Println("runClient", err)
			}

			log.Println("Restart")

			time.Sleep(100 * time.Millisecond)
			continue
		}

		return err
	}
}

func (this *ui) gcStore() {
	err := this.store.DeleteMatching(LogMsg{}, new(bolthold.Query).SortBy("ID").Reverse().Skip(2000))
	if err != nil {
		ss.Debug.Println("gcStore LogMsg", err)
	}
}

func (this *ui) runStore() {
	err := this.store.UpdateMatching(LogMsg{}, bolthold.Where("Msg").Eq(""), func(t interface{}) error {
		if rs, ok := t.(*LogMsg); ok {
			rs.Msg = "interrupt"
		}
		return nil
	})
	if err != nil {
		ss.Debug.Println("runStore reset", err)
	}

	this.gcStore()
	go func() {
		c := time.Tick(10 * time.Minute)
		for _ = range c {
			this.gcStore()
		}
	}()

	for t := range this.watcher.buf {
		if t.Now.IsZero() {
			var t1 []LogMsg
			err := this.store.Find(&t1, bolthold.Where("From").Eq(t.From).And("To").Eq(t.To).Limit(1))
			if err != nil {
				ss.Debug.Println("log find", err)
			} else if len(t1) > 0 {
				t1[0].Msg = t.Msg
				t1[0].Proxy = t.Proxy
				this.store.Update(t1[0].ID, t1[0])
			} else {
				ss.Debug.Println("Not Found Log", t.From, t.To)
			}
		} else {
			this.store.Insert(bolthold.NextSequence(), t)
		}
	}
}

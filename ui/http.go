package ui

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"regexp/syntax"
	ss "sshProxy/shadowsocks"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/bybzmt/bolthold"
)

type ui_state struct {
	ConnNum  int32
	Incoming string
	Outgoing string
}

func (this *ui) cross(fn http.HandlerFunc) http.HandlerFunc {

	reg, err := syntax.Parse("^(http|https)://(127.0.0.1|localhost|shadowsocks)(:\\d+)?", syntax.Perl)
	if err != nil {
		log.Panicln(err)
	}
	perl, err := regexp.Compile(reg.String())
	if err != nil {
		log.Panicln(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if !perl.MatchString(origin) {
				w.WriteHeader(403)
				return
			}

			//w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Origin", origin)

			if r.Method == "OPTIONS" {
				w.WriteHeader(204)
				return
			}
		}

		fn(w, r)
	}

}

//读取状态
func (this *ui) apiState(w http.ResponseWriter, r *http.Request) {
	this.l.Lock()
	defer this.l.Unlock()

	num := atomic.LoadInt32(&this.watcher.counter)

	t := this.ssServer.Traffic.Clone()
	this.ssServer.Traffic.Sub(&t)

	now := time.Now()
	diff := now.Sub(this.now)
	this.now = now

	s := ui_state{
		ConnNum:  num,
		Outgoing: FmtSize(diff, t.Incoming),
		Incoming: FmtSize(diff, t.Outgoing),
	}

	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(&s)
}

//读取日志
func (this *ui) apiLog(w http.ResponseWriter, r *http.Request) {
	rs := make([]LogMsg, 0)

	typ := r.FormValue("type")
	id, _ := strconv.Atoi(r.FormValue("id"))
	length, _ := strconv.Atoi(r.FormValue("length"))
	if length < 1 {
		length = 100
	}

	q := bolthold.Where("ID").Gt(uint64(id))
	if typ == "1" {
		q = q.And("Msg").Eq("")
	} else if typ == "2" {
		q = q.And("Msg").Ne("")
	}
	q.SortBy("ID").Reverse().Limit(length)

	err := this.store.Find(&rs, q)
	if err != nil {
		ss.Debug.Println("store find", err)
	}

	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(&rs)
}

//读取规则列表
func (this *ui) apiRules(w http.ResponseWriter, r *http.Request) {
	rs := make([]Rules, 0)

	err := this.store.Find(&rs, nil)
	if err != nil {
		ss.Debug.Println("apiRules", err)
	}

	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(&rs)
}

//添加规则
func (this *ui) apiRuleAdd(w http.ResponseWriter, r *http.Request) {
	var rs Rules

	rs.Note = r.FormValue("Note")
	rs.Enable = r.FormValue("Enable") == "1"
	rs.Items = r.FormValue("Items")
	rs.Servers = r.FormValue("Servers")

	err := this.store.Insert(bolthold.NextSequence(), &rs)
	if err != nil {
		ss.Debug.Println("apiRuleAdd", err)
	}

	w.Write([]byte("ok"))
}

//保存规则
func (this *ui) apiRuleEdit(w http.ResponseWriter, r *http.Request) {
	var rs Rules

	id, _ := strconv.Atoi(r.FormValue("ID"))

	rs.ID = uint64(id)
	rs.Note = r.FormValue("Note")
	rs.Enable = r.FormValue("Enable") == "1"
	rs.Items = r.FormValue("Items")
	rs.Servers = r.FormValue("Servers")

	err := this.store.Update(rs.ID, rs)
	if err != nil {
		ss.Debug.Println("apiRuleEdit", err)
	}

	w.Write([]byte("ok"))
}

//删除规则
func (this *ui) apiRuleDel(w http.ResponseWriter, r *http.Request) {
	var rs Rules

	id, _ := strconv.Atoi(r.FormValue("ID"))
	rs.ID = uint64(id)

	err := this.store.Delete(rs.ID, rs)
	if err != nil {
		ss.Debug.Println("apiRuleDel", err)
	}

	w.Write([]byte("ok"))
}

//读取配置
func (this *ui) apiClientConfig(w http.ResponseWriter, r *http.Request) {
	var rs ClientConfig

	err := this.store.Get("ClientConfig", &rs)
	if err != nil {
		ss.Debug.Println("apiClientConfig", err)
	}

	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(&rs)
}

//保存配置
func (this *ui) apiClientConfigSave(w http.ResponseWriter, r *http.Request) {
	var rs ClientConfig
	rs.Addr = r.FormValue("Addr")
	rs.Timeout, _ = strconv.Atoi(r.FormValue("Timeout"))
	rs.IdleTimeout, _ = strconv.Atoi(r.FormValue("IdleTimeout"))
	rs.LDNS = r.FormValue("LDNS")
	rs.LDNSEnable = r.FormValue("LDNSEnable") == "1"
	rs.RDNS = r.FormValue("RDNS")
	rs.RDNSEnable = r.FormValue("RDNSEnable") == "1"

	err := this.store.Upsert("ClientConfig", &rs)
	if err != nil {
		ss.Debug.Println("apiClientConfigSave", err)
	}

	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(&rs)
}

func (this *ui) apiServerConfigs(w http.ResponseWriter, r *http.Request) {
	rs := make([]ServerConfig, 0)

	err := this.store.Find(&rs, nil)
	if err != nil {
		ss.Debug.Println("apiServerConfigs", err)
	}

	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(&rs)
}

func (this *ui) apiServerConfigAdd(w http.ResponseWriter, r *http.Request) {
	var rs ServerConfig

	rs.Addr = r.FormValue("Addr")
	rs.User = r.FormValue("User")
	rs.Passwd = r.FormValue("Passwd")
	rs.Cipher = r.FormValue("Cipher")
	rs.Note = r.FormValue("Note")
	rs.Enable = r.FormValue("Enable") == "1"

	err := this.store.Insert(bolthold.NextSequence(), &rs)
	if err != nil {
		ss.Debug.Println("apiServerConfigAdd", err)
	}

	w.Write([]byte("ok"))
}

func (this *ui) apiServerConfigEdit(w http.ResponseWriter, r *http.Request) {
	var rs ServerConfig

	id, _ := strconv.Atoi(r.FormValue("ID"))

	rs.ID = uint64(id)
	rs.Addr = r.FormValue("Addr")
	rs.User = r.FormValue("User")
	rs.Passwd = r.FormValue("Passwd")
	rs.Cipher = r.FormValue("Cipher")
	rs.Note = r.FormValue("Note")
	rs.Enable = r.FormValue("Enable") == "1"

	err := this.store.Update(rs.ID, rs)
	if err != nil {
		ss.Debug.Println("apiServerConfigEdit", err)
	}

	w.Write([]byte("ok"))
}

func (this *ui) apiServerConfigDel(w http.ResponseWriter, r *http.Request) {
	var rs ServerConfig

	id, _ := strconv.Atoi(r.FormValue("ID"))
	rs.ID = uint64(id)

	err := this.store.Delete(rs.ID, rs)
	if err != nil {
		ss.Debug.Println("apiServerConfigDel", err)
	}

	w.Write([]byte("ok"))
}

func (this *ui) apiRestart(w http.ResponseWriter, r *http.Request) {
	this.Restart()

	w.Write([]byte("ok"))
}

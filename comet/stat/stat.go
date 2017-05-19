package stat

import (
	"encoding/json"
	"im/pkg/log"
	"net/http"
	"os"
	"os/user"
	"runtime"
	"sync/atomic"
	"time"
)

var (
	// server
	startTime int64 // process start unixnano
	// message
	MsgStat = &MessageStat{}
	RStat   = &RoutineStat{}
	SvrZones *ZonesStat
)

// Message stat info
type MessageStat struct {
	Succeed uint64 // total push message succeed count
	Failed  uint64 // total push message failed count
}

func (s *MessageStat) IncrSucceed(delta uint64) {
	atomic.AddUint64(&s.Succeed, delta)
}

func (s *MessageStat) IncrFailed(delta uint64) {
	atomic.AddUint64(&s.Failed, delta)
}

// Stat get the message stat info
func (s *MessageStat) Stat() []byte {
	res := map[string]interface{}{}
	res["succeed"] = s.Succeed
	res["failed"] = s.Failed
	res["total"] = s.Succeed + s.Failed
	return jsonRes(res)
}

type RoutineStat struct {
	Read  int32
	Write int32
}

func (rs *RoutineStat) IncRead() {
	atomic.AddInt32(&rs.Read, 1)
}

func (rs *RoutineStat) DescRead() {
	atomic.AddInt32(&rs.Read, -1)
}

func (rs *RoutineStat) IncWrite() {
	atomic.AddInt32(&rs.Write, 1)
}

func (rs *RoutineStat) DescWrite() {
	atomic.AddInt32(&rs.Write, -1)
}

func (rs *RoutineStat) Stat() []byte {
	res := make(map[string]interface{})
	res["read"] = rs.Read
	res["write"] = rs.Write
	
	return jsonRes(res)
}

// zone stat info
type ZoneInfo struct {
	Add    uint64
	Remove uint64
}

func (s *ZoneInfo) IncrAdd() {
	atomic.AddUint64(&s.Add, 1)
}

func (s *ZoneInfo) IncrRemove() {
	atomic.AddUint64(&s.Remove, 1)
}

type ZonesStat struct {
	Zones []*ZoneInfo
}

func NewZonesStat(zsize int) *ZonesStat {
	return &ZonesStat{Zones: make([]*ZoneInfo, zsize, zsize)}
}

func (sz *ZonesStat) IncrAdd(id int) {
	sz.Zones[id].IncrAdd()
}

func (sz *ZonesStat) IncrRemove(id int) {
	sz.Zones[id].IncrRemove()
}

func (sz *ZonesStat) Stat() []byte {
	res := make([]interface{}, len(sz.Zones))
	for idx, zone := range sz.Zones {
		st := make(map[string]interface{})
		st["id"] = idx
		st["add"] = zone.Add
		st["remove"] = zone.Remove
		st["current"] = zone.Add - zone.Remove
		
		res = append(res, st)
	}
	
	return jsonRes(res)
}

func (sz *ZonesStat) Connection() []byte {
	var total uint64
	for _, zone := range sz.Zones {
		total += zone.Add - zone.Remove
	}
	
	return jsonRes(map[string]interface{}{"total":total})
}

// start stats, called at process start
func StartStats(bind []string, zsize int) {
	startTime = time.Now().UnixNano()
	SvrZones = NewZonesStat(zsize)
	for _, bind := range bind {
		log.Info("start stat listen addr:\"%s\"", bind)
		go statListen(bind)
	}
}

func statListen(bind string) {
	httpServeMux := http.NewServeMux()
	httpServeMux.HandleFunc("/stat", handle)
	if err := http.ListenAndServe(bind, httpServeMux); err != nil {
		log.Error("http.ListenAdServe(\"%s\") error(%v)", bind, err)
		panic(err)
	}
}

// memory stats
func memStats() []byte {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	// general
	res := map[string]interface{}{}
	res["alloc"] = m.Alloc
	res["total_alloc"] = m.TotalAlloc
	res["sys"] = m.Sys
	res["lookups"] = m.Lookups
	res["mallocs"] = m.Mallocs
	res["frees"] = m.Frees
	// heap
	res["heap_alloc"] = m.HeapAlloc
	res["heap_sys"] = m.HeapSys
	res["heap_idle"] = m.HeapIdle
	res["heap_inuse"] = m.HeapInuse
	res["heap_released"] = m.HeapReleased
	res["heap_objects"] = m.HeapObjects
	// low-level fixed-size struct alloctor
	res["stack_inuse"] = m.StackInuse
	res["stack_sys"] = m.StackSys
	res["mspan_inuse"] = m.MSpanInuse
	res["mspan_sys"] = m.MSpanSys
	res["mcache_inuse"] = m.MCacheInuse
	res["mcache_sys"] = m.MCacheSys
	res["buckhash_sys"] = m.BuckHashSys
	// GC
	res["next_gc"] = m.NextGC
	res["last_gc"] = m.LastGC
	res["pause_total_ns"] = m.PauseTotalNs
	res["pause_ns"] = m.PauseNs
	res["num_gc"] = m.NumGC
	res["enable_gc"] = m.EnableGC
	res["debug_gc"] = m.DebugGC
	res["by_size"] = m.BySize
	return jsonRes(res)
}

// golang stats
func goStats() []byte {
	res := map[string]interface{}{}
	res["compiler"] = runtime.Compiler
	res["arch"] = runtime.GOARCH
	res["os"] = runtime.GOOS
	res["max_procs"] = runtime.GOMAXPROCS(-1)
	res["root"] = runtime.GOROOT()
	res["cgo_call"] = runtime.NumCgoCall()
	res["goroutine_num"] = runtime.NumGoroutine()
	res["version"] = runtime.Version()
	return jsonRes(res)
}

// server stats
func serverStats() []byte {
	res := map[string]interface{}{}
	res["uptime"] = time.Now().UnixNano() - startTime
	hostname, _ := os.Hostname()
	res["hostname"] = hostname
	wd, _ := os.Getwd()
	res["wd"] = wd
	res["ppid"] = os.Getppid()
	res["pid"] = os.Getpid()
	res["pagesize"] = os.Getpagesize()
	if usr, err := user.Current(); err != nil {
		log.Error("user.Current() error(%v)", err)
		res["group"] = ""
		res["user"] = ""
	} else {
		res["group"] = usr.Gid
		res["user"] = usr.Uid
	}
	return jsonRes(res)
}

// jsonRes format the output
func jsonRes(res interface{}) []byte {
	byteJson, err := json.MarshalIndent(res, "", "    ")
	if err != nil {
		log.Error("json.MarshalIndent(\"%v\", \"\", \"    \") error(%v)", res, err)
		return nil
	}
	return byteJson
}

// StatHandle get stat info by http
func handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	params := r.URL.Query()
	types := params.Get("type")
	res := []byte{}
	switch types {
	case "memory":
		res = memStats()
	case "server":
		res = serverStats()
	case "golang":
		res = goStats()
	//case "config":
	//	res = configInfo()
	case "zone":
		res = SvrZones.Stat()
	case "message":
		res = MsgStat.Stat()
	case "connection":
		res = SvrZones.Connection()
	case "routine":
		res = SvrZones.Connection()
	default:
		http.Error(w, "Not Found", 404)
	}
	if res != nil {
		if _, err := w.Write(res); err != nil {
			log.Error("w.Write(\"%s\") error(%v)", string(res), err)
		}
	}
}

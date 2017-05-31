package stat

import (
	"im/pkg/log"
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



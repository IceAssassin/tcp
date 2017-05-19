package zone

import (
	"fmt"
	"im/comet/proto"
	"sync"
	"im/comet/stat"
)

type ZoneOptions struct {
	CacheSize int
}

type Zone struct {
	rLock    sync.RWMutex
	Id  int
	sessions map[uint64]*Session
}

// NewZone new a zone struct, store session zone info.
func NewZone(i int, zoption ZoneOptions) (r *Zone) {
	r = new(Zone)
	r.Id = i
	r.sessions = make(map[uint64]*Session, zoption.CacheSize) //
	return
}

// Get get session from the zone.
func (r *Zone) Session(id uint64) (session *Session, e error) {
	r.rLock.Lock()
	if s, ok := r.sessions[id]; ok {
		session = s
	} else {
		e = fmt.Errorf("not found %v session", id)
	}
	r.rLock.Unlock()
	return
}

// Put put session into the zone.
func (r *Zone) Put(session *Session) {
	r.rLock.Lock()
	r.sessions[session.Id] = session
	stat.SvrZones.IncrAdd(r.Id)
	r.rLock.Unlock()
	return
}

// Del delete session from the zone.
func (r *Zone) Del(id uint64) {
	r.rLock.Lock()
	if _, ok := r.sessions[id]; ok {
		delete(r.sessions, id)
		stat.SvrZones.IncrRemove(r.Id)
	}
	r.rLock.Unlock()
}

// Push push msg
func (r *Zone) Push(id uint64, p *proto.Proto) {
	r.rLock.RLock()
	if session, ok := r.sessions[id]; ok {
		session.Push(p)
	}
	r.rLock.RUnlock()
}

// Close close the room.
func (r *Zone) Close() {
	r.rLock.RLock()
	for _, session := range r.sessions {
		session.Close()
	}
	r.rLock.RUnlock()
}

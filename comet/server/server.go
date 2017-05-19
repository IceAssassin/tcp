package server

import (
	"im/comet/handle"
	"im/comet/utils"
	"im/comet/zone"
	"im/pkg/log"
	"time"
)

var (
	maxInt           = 1<<31 - 1
	emptyJSONBody    = []byte("{}")
	DefaultServer    *Server
	DefaultWhitelist *Whitelist
)

type ServerOptions struct {
	CliProto         int
	SvrProto         int
	HandshakeTimeout time.Duration
	TCPKeepalive     bool
	TCPRcvbufSize    int
	TCPSndbufSize    int
}

type Server struct {
	Zones   []*zone.Zone // subkey bucket
	round   *utils.Round // accept round store
	handle  []handle.Handle
	Options ServerOptions
}

// NewServer returns a new Server.
func NewServer(z []*zone.Zone, r *utils.Round, h []handle.Handle, options ServerOptions) *Server {
	s := new(Server)
	s.Zones = z
	s.round = r
	s.handle = h
	s.Options = options
	return s
}

func (server *Server) Zone(id uint64) *zone.Zone {
	zid := uint8(id >> 56)
	log.Debug("%v hit zone index: %d", id, zid)
	return server.Zones[zid]
}

func (server *Server) Disconect(id uint64) error {
	zid := uint8(id >> 56)
	server.Zones[zid].Del(id)
	return nil
}

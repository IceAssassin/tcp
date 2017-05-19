package zone

import (
	"errors"
	"im/comet/proto"
	"im/comet/utils"
	"im/pkg/bufio"
	"im/pkg/log"
)

var (
	ErrSessionFull = errors.New("session cache full")
)

// Session used by message pusher send msg to write goroutine.
type Session struct {
	Id       uint64
	ZoneId   int
	CliProto utils.Ring
	signal   chan *proto.Proto
	Writer   bufio.Writer
	Reader   bufio.Reader
}

// cli: recv cache size, svr: send cache size
func NewSession(id uint64, zid int, cli, svr int) *Session {
	c := new(Session)
	c.Id = id
	c.ZoneId = zid
	c.CliProto.Init(cli)
	c.signal = make(chan *proto.Proto, svr)
	return c
}

// Push server push message.
func (c *Session) Push(p *proto.Proto) (e error) {
	select {
	case c.signal <- p:
	default:
		e = ErrSessionFull
		log.Error("Session Cache Full %v:%v", c.ZoneId, c.Id)
	}
	return
}

// Ready check the session ready or close?
func (c *Session) Ready() *proto.Proto {
	return <-c.signal
}

// Signal send signal to the session, protocol ready.
func (c *Session) Signal() {
	c.signal <- proto.ProtoReady
}

// Close close the session.
func (c *Session) Close() {
	c.signal <- proto.ProtoFinish
}

package server

import (
	"fmt"
	"im/comet/proto"
	"im/comet/zone"
	"im/pkg/bufio"
	"im/pkg/bytes"
	"im/pkg/log"
	itime "im/pkg/time"
	"net"
	"time"
	"encoding/json"
	"math/rand"
	"math"
	"im/comet/stat"
)

// InitTCP listen all tcp.bind and start accept connections.
func InitTCP(addrs []string, accept int) (err error) {
	var (
		bind     string
		listener *net.TCPListener
		addr     *net.TCPAddr
	)
	for _, bind = range addrs {
		if addr, err = net.ResolveTCPAddr("tcp4", bind); err != nil {
			log.Error("net.ResolveTCPAddr(tcp4, %s) error(%v)\n", bind, err)
			return
		}

		if listener, err = net.ListenTCP("tcp4", addr); err != nil {
			log.Error("net.ListenTCP(tcp4, %s) error(%v)\n", bind, err)
			return
		}

		log.Debug("start tcp listen: %s:%d\n", bind, accept)
		// split N core accept
		for i := 0; i < accept; i++ {
			go acceptTCP(DefaultServer, listener)
		}
	}
	return
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.  Accept blocks; the caller typically
// invokes it in a go statement.
func acceptTCP(server *Server, lis *net.TCPListener) {
	var (
		conn *net.TCPConn
		err  error
		r    int
	)
	for {
		if conn, err = lis.AcceptTCP(); err != nil {
			// if listener close then return
			log.Error("listener.Accept(\"%s\") error(%v)", lis.Addr().String(), err)
			return
		}
		if err = conn.SetKeepAlive(server.Options.TCPKeepalive); err != nil {
			log.Error("conn.SetKeepAlive() error(%v)", err)
			return
		}
		if err = conn.SetReadBuffer(server.Options.TCPRcvbufSize); err != nil {
			log.Error("conn.SetReadBuffer() error(%v)", err)
			return
		}
		if err = conn.SetWriteBuffer(server.Options.TCPSndbufSize); err != nil {
			log.Error("conn.SetWriteBuffer() error(%v)", err)
			return
		}

		go serveTCP(server, conn, r)
		if r++; r == maxInt {
			r = 0
		}
	}
}

func serveTCP(server *Server, conn *net.TCPConn, r int) {
	var (
		// timer
		tr = server.round.Timer(r)
		rp = server.round.Reader(r)
		wp = server.round.Writer(r)
		// ip addr
		lAddr = conn.LocalAddr().String()
		rAddr = conn.RemoteAddr().String()
	)

	log.Debug("start tcp serve %s with %s", lAddr, rAddr)
	server.serveTCP(conn, rp, wp, tr)
}

// TODO linger close?
func (server *Server) serveTCP(conn *net.TCPConn, rp, wp *bytes.Pool, tr *itime.Timer) {
	var (
		err  error
		id   uint64
		hb   time.Duration // heartbeat
		p    *proto.Proto
		z    *zone.Zone
		trd  *itime.TimerData
		rb   = rp.Get()
		wb   = wp.Get()
		sion = zone.NewSession(0, -1, server.Options.CliProto, server.Options.SvrProto)
		rr   = &sion.Reader
		wr   = &sion.Writer
	)

	sion.Reader.ResetBuffer(conn, rb.Bytes())
	sion.Writer.ResetBuffer(conn, wb.Bytes())

	// handshake
	trd = tr.Add(server.Options.HandshakeTimeout, func() {
		conn.Close()
	})

	// must not setadv, only used in auth
	if p, err = sion.CliProto.Set(); err == nil {
		if id, sion.ZoneId, hb, err = server.authTCP(rr, wr, p); err == nil {
			z = server.Zone(id)
			z.Put(sion)
		}
	}

	if err != nil {
		conn.Close()
		rp.Put(rb)
		wp.Put(wb)
		tr.Del(trd)
		log.Error("key: %s handshake failed error(%v)", id, err)
		return
	}

	trd.Key = id
	tr.Set(trd, hb)

	// hanshake ok start dispatch goroutine
	go server.dispatchTCP(id, conn, wr, wp, wb, sion)
	stat.RStat.IncRead()
	defer stat.RStat.DescRead()
	for {
		if p, err = sion.CliProto.Set(); err != nil {
			log.Error("id: %v serve tcp cliptoto.Set error(%v)", id, err)
			break
		}

		if err = p.ReadTCP(rr); err != nil {
			log.Error("id: %v serve tcp read tcp error(%v)", id, err)
			break
		}

		if int(p.Type) >= len(server.handle) {
			log.Error("id: %v, server tcp invalid proto %v", id, p.Type)
			break
		}

		// TODO handle proto msg
		if err = server.handle[p.Type](id, p); err != nil {
			log.Error("id: %v, server handle proto %v", id, p)
			break
		}

		if p.Type == proto.C2S_HEART_BEAT { // heart beat set expired
			tr.Set(trd, hb)
		}

		sion.CliProto.SetAdv()
		sion.Signal()
	}

	z.Del(id)
	tr.Del(trd)
	rp.Put(rb)
	conn.Close()
	sion.Close()
	// TODO Disconnect
	if err = server.Disconect(id); err != nil {
		log.Error("id: %s do disconnect error(%v)", id, err)
	}

	return
}

// dispatch accepts connections on the listener and serves requests
// for each incoming connection.  dispatch blocks; the caller typically
// invokes it in a go statement.
func (server *Server) dispatchTCP(id uint64, conn *net.TCPConn, wr *bufio.Writer, wp *bytes.Pool, wb *bytes.Buffer, session *zone.Session) {
	var (
		err    error
		finish bool
	)
	
	stat.RStat.IncWrite()
	defer stat.RStat.DescWrite()

	for {
		var p = session.Ready()
		switch p {
		case proto.ProtoFinish:
			finish = true
			goto failed
		case proto.ProtoReady:
			// fetch message from svrbox(client send)
			for {
				if p, err = session.CliProto.Get(); err != nil {
					err = nil // must be empty error
					break
				}
				if err = p.WriteTCP(wr); err != nil {
					goto failed
				}
				p.Body = nil // avoid memory leak
				session.CliProto.GetAdv()
			}
		default:
			// server send
			if err = p.WriteTCP(wr); err != nil {
				goto failed
			}
		}

		// only hungry flush response
		if err = wr.Flush(); err != nil {
			break
		}
	}

failed:
	if err != nil {
		log.Error("id: %v dispatch tcp error(%v)", id, err)
	}
	conn.Close()
	wp.Put(wb)
	// must ensure all channel message discard, for reader won't blocking Signal
	for !finish {
		finish = (session.Ready() == proto.ProtoFinish)
	}
	return
}

// auth for handshake with client, use rsa & aes.
func (server *Server) authTCP(rr *bufio.Reader, wr *bufio.Writer, p *proto.Proto) (id uint64, zid int, heartbeat time.Duration, e error) {
	if e = p.ReadTCP(rr); e != nil {
		return
	}

	if p.Type != proto.C2S_AUTH {
		log.Warn("auth operation not valid: %d", p.Type)
		e = fmt.Errorf("invalid type %v", p.Type)
		return
	}
	
	auth := proto.Auth{}
	if e = json.Unmarshal([]byte(p.Body), &auth); e != nil {
		return
	}
	
	log.Debug("uid = %v, code %v", auth.Uid, auth.Code)
	// TODO auth from router, Return from router
	NodeId := uint8(0)
	ZondId := rand.Int() % math.MaxUint8
	HeartBeat := 5
	
	//server.handle[p.Type](p)
	if e = p.WriteTCP(wr); e != nil {
		return
	}
	
	id = uint64(NodeId<<56) | uint64(ZondId<<48) | uint64(auth.Uid)
	zid = int(ZondId)
	heartbeat = time.Duration(HeartBeat)
	
	e = wr.Flush()
	return
}

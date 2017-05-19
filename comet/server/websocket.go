package server

import (
	"crypto/tls"
	"github.com/gorilla/websocket"
	"im/comet/proto"
	"im/comet/zone"
	"im/pkg/log"
	itime "im/pkg/time"
	"math/rand"
	"net"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func InitWebsocket(addrs []string) (err error) {
	var (
		bind         string
		listener     *net.TCPListener
		addr         *net.TCPAddr
		httpServeMux = http.NewServeMux()
		server       *http.Server
	)
	httpServeMux.HandleFunc("/sub", ServeWebSocket)

	for _, bind = range addrs {
		if addr, err = net.ResolveTCPAddr("tcp4", bind); err != nil {
			log.Error("net.ResolveTCPAddr(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		if listener, err = net.ListenTCP("tcp4", addr); err != nil {
			log.Error("net.ListenTCP(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		server = &http.Server{Handler: httpServeMux}
		go func(host string) {
			if err = server.Serve(listener); err != nil {
				log.Error("server.Serve(\"%s\") error(%v)", host, err)
				panic(err)
			}
		}(bind)
	}
	return
}

func InitWebsocketWithTLS(addrs []string, cert, priv string) (err error) {
	var (
		httpServeMux = http.NewServeMux()
	)
	httpServeMux.HandleFunc("/sub", ServeWebSocket)
	config := &tls.Config{}
	config.Certificates = make([]tls.Certificate, 1)
	if config.Certificates[0], err = tls.LoadX509KeyPair(cert, priv); err != nil {
		return
	}
	for _, bind := range addrs {
		server := &http.Server{Addr: bind, Handler: httpServeMux}
		server.SetKeepAlivesEnabled(true)
		log.Debug("start websocket wss listen: \"%s\"", bind)
		go func(host string) {
			ln, err := net.Listen("tcp", host)
			if err != nil {
				return
			}

			tlsListener := tls.NewListener(ln, config)
			if err = server.Serve(tlsListener); err != nil {
				log.Error("server.Serve(\"%s\") error(%v)", host, err)
				return
			}
		}(bind)
	}
	return
}

func ServeWebSocket(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	ws, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Error("Websocket Upgrade error(%v), userAgent(%s)", err, req.UserAgent())
		return
	}
	defer ws.Close()
	var (
		lAddr = ws.LocalAddr()
		rAddr = ws.RemoteAddr()
		tr    = DefaultServer.round.Timer(rand.Int())
	)
	log.Debug("start websocket serve \"%s\" with \"%s\"", lAddr, rAddr)
	DefaultServer.serveWebsocket(ws, tr)
}

func (server *Server) serveWebsocket(conn *websocket.Conn, tr *itime.Timer) {
	var (
		err  error
		id   uint64
		hb   time.Duration // heartbeat
		p    *proto.Proto
		z    *zone.Zone
		trd  *itime.TimerData
		sion = zone.NewSession(0, -1, server.Options.CliProto, server.Options.SvrProto)
	)
	// handshake
	trd = tr.Add(server.Options.HandshakeTimeout, func() {
		conn.Close()
	})
	// must not setadv, only used in auth
	if p, err = sion.CliProto.Set(); err == nil {
		if id, sion.ZoneId, hb, err = server.authWebsocket(conn, p); err == nil {
			z = server.Zone(id)
			z.Put(sion)
		}
	}
	if err != nil {
		conn.Close()
		tr.Del(trd)
		log.Error("handshake failed error(%v)", err)
		return
	}
	trd.Key = id
	tr.Set(trd, hb)
	// hanshake ok start dispatch goroutine
	go server.dispatchWebsocket(id, conn, sion)
	for {
		if p, err = sion.CliProto.Set(); err != nil {
			break
		}
		if err = p.ReadWebsocket(conn); err != nil {
			break
		}
		//p.Time = *globalNowTime
		//if p.Type == proto.OP_HEARTBEAT {
		//	// heartbeat
		//	tr.Set(trd, hb)
		//	p.Body = nil
		//	p.Type = proto.OP_HEARTBEAT_REPLY
		//} else {
		//	// process message
		//	if err = server.operator.Operate(p); err != nil {
		//		break
		//	}
		//}
		sion.CliProto.SetAdv()
		sion.Signal()
	}

	tr.Del(trd)
	conn.Close()
	sion.Close()
	z.Del(id)
	// TODO disconnect
	//if err = server.operator.Disconnect(key, sion.ZoneId); err != nil {
	//	log.Error("key: %s operator do disconnect error(%v)", key, err)
	//}

	return
}

// dispatch accepts connections on the listener and serves requests
// for each incoming connection.  dispatch blocks; the caller typically
// invokes it in a go statement.
func (server *Server) dispatchWebsocket(id uint64, conn *websocket.Conn, sion *zone.Session) {
	var (
		p   *proto.Proto
		err error
	)

	log.Debug("key: %s start dispatch websocket goroutine", id)
	for {
		p = sion.Ready()
		switch p {
		case proto.ProtoFinish:
			log.Debug("key: %s wakeup exit dispatch goroutine", id)
			goto failed
		case proto.ProtoReady:
			for {
				if p, err = sion.CliProto.Get(); err != nil {
					err = nil // must be empty error
					break
				}
				if err = p.WriteWebsocket(conn); err != nil {
					goto failed
				}
				p.Body = nil // avoid memory leak
				sion.CliProto.GetAdv()
			}
		default:
			// TODO room-push support
			// just forward the message
			if err = p.WriteWebsocket(conn); err != nil {
				goto failed
			}
		}
	}
failed:
	if err != nil {
		log.Error("key: %s dispatch websocket error(%v)", id, err)
	}
	conn.Close()
	// must ensure all channel message discard, for reader won't blocking Signal
	for {
		if p == proto.ProtoFinish {
			break
		}
		p = sion.Ready()
	}

	return
}

func (server *Server) authWebsocket(conn *websocket.Conn, p *proto.Proto) (id uint64, zid int, heartbeat time.Duration, err error) {
	if err = p.ReadWebsocket(conn); err != nil {
		return
	}
	//if p.Type != proto.OP_AUTH {
	//	err = ErrOperation
	//	return
	//}
	//if key, rid, heartbeat, err = server.operator.Connect(p); err != nil {
	//	return
	//}
	//p.Body = emptyJSONBody
	//p.Type = proto.OP_AUTH_REPLY
	err = p.WriteWebsocket(conn)
	return
}

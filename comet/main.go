package main

import (
	"fmt"
	"im/comet/handle"
	"im/comet/server"
	"im/comet/utils"
	"im/comet/zone"
	"im/pkg/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
	"im/comet/config"
	"im/comet/stat"
)

var Conf *config.Config = nil

func main() {
	Conf = &config.Config{}
	if e := Conf.Load("./comet-config.yaml"); e != nil {
		fmt.Printf("config init error %v\n", e)
		return
	}

	Conf.Print()

	// set max routine
	runtime.GOMAXPROCS(Conf.MaxProc)

	// new server
	zones := make([]*zone.Zone, Conf.Zone.ZoneNum)
	for i := 0; i < Conf.Zone.ZoneNum; i++ {
		zones[i] = zone.NewZone(i, zone.ZoneOptions{
			CacheSize: Conf.Zone.CacheSize,
		})
	}
	round := utils.NewRound(utils.RoundOptions{
		ReaderNum:    Conf.TCP.ReaderNum,
		ReadbufNum:   Conf.TCP.ReadbufNum,
		ReadbufSize:  Conf.TCP.ReadbufSize,
		WriterNum:    Conf.TCP.WriterNum,
		WritebufNum:  Conf.TCP.WritebufNum,
		WritebufSize: Conf.TCP.WritebufSize,
		TimerNum:     Conf.Timer.TimerNum,
		TimerSize:    Conf.Timer.TimerSize,
	})
	server.DefaultServer = server.NewServer(zones, round, []handle.Handle{}, server.ServerOptions{
		CliProto:         Conf.Proto.CliProto,
		SvrProto:         Conf.Proto.SvrProto,
		HandshakeTimeout: time.Duration(Conf.Proto.HandshakeTimeout),
		TCPKeepalive:     Conf.TCP.Keepalive,
		TCPRcvbufSize:    Conf.TCP.RcvbufSize,
		TCPSndbufSize:    Conf.TCP.SndbufSize,
	})

	// white list TODO

	pprof.Init(Conf.PprofBind.StringSlice())
	stat.StartStats(Conf.StatBind.StringSlice(), Conf.Zone.ZoneNum)

	// tcp comet
	if e := server.InitTCP(Conf.TCP.Bind.StringSlice(), Conf.MaxProc); e != nil {
		panic(e)
	}

	// websocket comet
	if e := server.InitWebsocket(Conf.Websocket.Bind.StringSlice()); e != nil {
		panic(e)
	}

	// wss comet
	if Conf.Websocket.TLSOpen {
		if e := server.InitWebsocketWithTLS(Conf.Websocket.TLSBind.StringSlice(), Conf.Websocket.CertFile, Conf.Websocket.PrivateFile); e != nil {
			panic(e)
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSTOP)
	for {
		s := <-c
		fmt.Printf("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			return
		case syscall.SIGHUP:
			// TODO reload
			//return
		default:
			return
		}
	}

}

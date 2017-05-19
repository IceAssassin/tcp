package config

import (
	"fmt"
	"im/pkg/yaml"
)

type Config struct {
	// base section
	PidFile string "pidfile"
	MaxProc int    "max_proc"
	//Whitelist []string "white_list"
	//WhiteLog  string   "white_log"
	StatBind  yaml.Addresses "stat_bind"
	PprofBind yaml.Addresses "pprof_bind"

	// tcp
	TCP struct {
		Bind         yaml.Addresses "bind"
		SndbufSize   int            "sndbuf_size"
		RcvbufSize   int            "rcvbuf_size"
		Keepalive    bool           "keepalive"
		ReaderNum    int            "reader_num"
		ReadbufNum   int            "readbuf_num"
		ReadbufSize  int            "readbuf_size"
		WriterNum    int            "writer_num"
		WritebufNum  int            "writebuf_num"
		WritebufSize int            "writebuf_size"
	} "tcp"

	// websocket
	Websocket struct {
		Bind        yaml.Addresses "bind"
		TLSOpen     bool           "tls_open"
		TLSBind     yaml.Addresses "tls_bind"
		CertFile    string         "cert_file"
		PrivateFile string         "private_file"
	} "websocket"

	//// flash safe policy
	//FlashPolicyOpen bool     `:"flash:policy.open"`
	//FlashPolicyBind []string `:"flash:policy.bind:,"`
	// proto section
	Proto struct {
		HandshakeTimeout int "handshake_timeout"
		WriteTimeout     int "write_timeout"
		SvrProto         int "svr_proto"
		CliProto         int "cli_proto"
	} "proto"

	// timer
	Timer struct {
		TimerNum  int "timer_num"
		TimerSize int "timer_size"
	} "timer"

	Zone struct {
		ZoneNum   int "zone_num"
		CacheSize int "cache_size"
	} "zone"

	//EtcdAddr   yaml.Address "etcd_addr"
	//// push
	//RPCPushAddrs []string `:"push:rpc.addrs:,"`
	//// logic
	//LogicAddrs []string `:"logic:rpc.addrs:,"`
	//// monitor
	//MonitorOpen  bool     `:"monitor:open"`
	//MonitorAddrs []string `:"monitor:addrs:,"`

	// Log
	Log struct {
		Dir     string "dir"
		Level   string "level"
		BufSize int32  "buf_size"
	} "log"
}

func (c *Config) Load(path string) error {
	return yaml.Load(c, path)
}

func (c Config) Print() {
	fmt.Printf("%v", c)
}

//func NewConfig() *Config {
//	return &Config{
//		// base section
//		PidFile:   "/tmp/goim-comet.pid",
//		Dir:       "./",
//		Log:       "./comet-log.xml",
//		MaxProc:   runtime.NumCPU(),
//		PprofBind: []string{"localhost:6971"},
//		StatBind:  []string{"localhost:6972"},
//		Debug:     true,
//		// tcp
//		TCPBind:      []string{"0.0.0.0:8080"},
//		TCPSndbuf:    1024,
//		TCPRcvbuf:    1024,
//		TCPKeepalive: false,
//		// websocket
//		WebsocketBind: []string{"0.0.0.0:8090"},
//		// websocket tls
//		WebsocketTLSOpen:     false,
//		WebsocketTLSBind:     []string{"0.0.0.0:8095"},
//		WebsocketCertFile:    "../source/cert.pem",
//		WebsocketPrivateFile: "../source/private.pem",
//		// flash safe policy
//		FlashPolicyOpen: false,
//		FlashPolicyBind: []string{"0.0.0.0:843"},
//		// proto section
//		HandshakeTimeout: 5 * time.Second,
//		WriteTimeout:     5 * time.Second,
//		TCPReadBuf:       1024,
//		TCPWriteBuf:      1024,
//		TCPReadBufSize:   1024,
//		TCPWriteBufSize:  1024,
//		// timer
//		Timer:     runtime.NumCPU(),
//		TimerSize: 1000,
//		// bucket
//		CliProto:      5,
//		SvrProto:      80,
//		// push
//		RPCPushAddrs: []string{"localhost:8083"},
//	}
//}

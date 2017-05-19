package utils

// Noed info
type ServerInfo struct {
	ID         int32  `json:"id"`          // serverID
	Root       string `json:"root"`        // etcd Path
	LocalIP    string `json:"local_ip"`    // 本地服务IP
	LocalPort  int  `json:"local_port"`  // 本地服务端口
	PublicIP   string `json:"public_ip"`   // 公网服务IP
	PublicPort int32  `json:"public_port"` // 公网端口
	NodeSta           // 本机负载
}

// Nodestatistics
type NodeSta struct {
	ConnNum int32 `json:"conn_num"`
	CpuLoad int32 `json:"cpu_load"`
	NetLoad int32 `json:"net_load"`
}

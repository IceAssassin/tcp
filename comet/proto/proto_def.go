package proto

const (
	PROTO_READY  = 2048
	PROTO_FINISH = 2049
)

const (
	C2S_RC = iota
	C2S_HEART_BEAT
	C2S_AUTH
	C2S_MAX
)

const (
	S2C_BASE       = 1024
	S2C_RC         = S2C_BASE + C2S_RC
	S2C_HEART_BEAT = S2C_BASE + C2S_HEART_BEAT
	S2C_AUTH       = S2C_BASE + C2S_AUTH
	S2C_MAX
)

type Auth struct {
	Uid  uint32  `json:"uid"`
	Code string  `json:"code"`
}

type HeartBeat struct {
	Uid  uint32  `json:"uid"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}

type RC struct {
	Uid  uint32 `json:"uid"`
	Mid  uint64 `json:"mid"`
}

package handle

import (
	"im/comet/proto"
)

type Handle func(id uint64, p *proto.Proto) (e error)

var Handles []Handle

func init() {
	Handles = make([]Handle, proto.C2S_MAX, proto.C2S_MAX)
	Handles[proto.C2S_RC] = handle_rc
	Handles[proto.C2S_HEART_BEAT] = handle_heartbeat
}

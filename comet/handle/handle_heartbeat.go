package handle

import (
	"im/comet/proto"
	"encoding/json"
	"im/pkg/log"
)

func handle_heartbeat(id uint64, p *proto.Proto) (e error) {
	heartbeat := &proto.HeartBeat{}
	if e = json.Unmarshal([]byte(p.Body), &heartbeat); e != nil {
		return
	}
	
	log.Debug("id %v, beat %v\n", id, heartbeat)
	
	p.Body = nil
	p.Type = proto.S2C_HEART_BEAT

	return nil
}

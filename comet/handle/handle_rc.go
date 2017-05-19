package handle

import (
	"fmt"
	"im/comet/proto"
	"encoding/json"
)

func handle_rc(id uint64, p *proto.Proto) (e error) {
	rc := &proto.RC{}
	if e = json.Unmarshal([]byte(p.Body), &rc); e != nil {
		return
	}
	
	fmt.Printf("id %v, rc = %v\n", id, rc)
	p.Body = nil
	p.Type = proto.C2S_RC

	return nil
}

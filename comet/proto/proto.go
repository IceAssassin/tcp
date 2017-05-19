package proto

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"im/pkg/bufio"
	"im/pkg/bytes"
	"im/pkg/encoding/binary"
)

// for tcp
const (
	MaxBodySize = int32(1 << 10)
)

const (
	// size
	PackSize      = 4
	XSize         = 1
	VerSize       = 1
	TypeSize      = 2
	SeqIdSize     = 4
	RawHeaderSize = PackSize + XSize + VerSize + TypeSize + SeqIdSize
	MaxPackSize   = MaxBodySize + int32(RawHeaderSize)
	// offset
	PackOffset  = 0
	XOffset     = PackOffset + PackSize
	VerOffset   = XOffset + XSize
	TypeOffset  = VerOffset + VerSize
	SeqIdOffset = TypeOffset + TypeSize
)

var (
	emptyProto    = Proto{}
	emptyJSONBody = []byte("{}")

	ErrProtoPackLen = errors.New("default server codec pack length error")
)

var (
	ProtoReady  = &Proto{Type: PROTO_READY}
	ProtoFinish = &Proto{Type: PROTO_FINISH}
)

// Proto is a request&response written before every goim connect.  It is used internally
// but documented here as an aid to debugging, such as when analyzing
// network traffic.
// tcp:
// binary codec
// websocket & http:
// raw codec, with http header stored ver, operation, seqid

// |--len--|--x--|--ver--|--type--|--SeqId--|--Body--|
//     4      1      1        2        4        x
type Proto struct {
	Ver   int8            `json:"ver"`  // protocol version
	Type  int16           `json:"type"` // operation for request
	SeqId int32           `json:"seq"`  // sequence number chosen by client
	Body  json.RawMessage `json:"body"` // binary body bytes(json.RawMessage is []byte)
}

func (p *Proto) Reset() {
	*p = emptyProto
}

func (p *Proto) String() string {
	return fmt.Sprintf("\n-------- proto --------\nver: %d\nop: %d\nseq: %d\nbody: %v\n-----------------------", p.Ver, p.Type, p.SeqId, p.Body)
}

func (p *Proto) WriteTo(b *bytes.Writer) {
	var (
		packLen = RawHeaderSize + int32(len(p.Body))
		buf     = b.Peek(RawHeaderSize)
	)
	binary.BigEndian.PutInt32(buf[PackOffset:], packLen)
	binary.BigEndian.PutInt8(buf[XOffset:], int8(0))
	binary.BigEndian.PutInt8(buf[VerOffset:], p.Ver)
	binary.BigEndian.PutInt16(buf[TypeOffset:], p.Type)
	binary.BigEndian.PutInt32(buf[SeqIdOffset:], p.SeqId)
	if p.Body != nil {
		b.Write(p.Body)
	}
}

func (p *Proto) ReadTCP(rr *bufio.Reader) (e error) {
	var (
		bodyLen int
		packLen int32
		buf     []byte
	)

	if buf, e = rr.Pop(RawHeaderSize); e != nil {
		return
	}

	packLen = binary.BigEndian.Int32(buf[PackOffset:XOffset])
	if packLen > MaxPackSize {
		return ErrProtoPackLen
	}

	p.Ver = binary.BigEndian.Int8(buf[VerOffset:TypeOffset])
	p.Type = binary.BigEndian.Int16(buf[TypeOffset:SeqIdOffset])
	p.SeqId = binary.BigEndian.Int32(buf[SeqIdOffset:])

	if bodyLen = int(packLen - int32(RawHeaderSize)); bodyLen > 0 {
		p.Body, e = rr.Pop(bodyLen)
	} else {
		p.Body = nil
	}
	return
}

func (p *Proto) WriteTCP(wr *bufio.Writer) (e error) {
	var (
		buf     []byte
		packLen int32
	)

	//if p.Operation == define.OP_RAW {
	//	// write without buffer, job concact proto into raw buffer
	//	_, e = wr.WriteRaw(p.Body)
	//	return
	//}
	packLen = RawHeaderSize + int32(len(p.Body))
	if buf, e = wr.Peek(RawHeaderSize); e != nil {
		return
	}
	binary.BigEndian.PutInt32(buf[PackOffset:], packLen)
	binary.BigEndian.PutInt8(buf[XOffset:], 0)
	binary.BigEndian.PutInt8(buf[VerOffset:], p.Ver)
	binary.BigEndian.PutInt16(buf[TypeOffset:], p.Type)
	binary.BigEndian.PutInt32(buf[SeqIdOffset:], p.SeqId)
	if p.Body != nil {
		_, e = wr.Write(p.Body)
	}
	return
}

func (p *Proto) ReadWebsocket(wr *websocket.Conn) (e error) {
	return wr.ReadJSON(p)
}

func (p *Proto) WriteBodyTo(b *bytes.Writer) (e error) {
	var (
		ph  Proto
		js  []json.RawMessage
		j   json.RawMessage
		jb  []byte
		bts []byte
	)
	offset := int32(PackOffset)
	buf := p.Body[:]
	for {
		if (len(buf[offset:])) < RawHeaderSize {
			// should not be here
			break
		}
		packLen := binary.BigEndian.Int32(buf[offset:PackSize])
		packBuf := buf[offset : offset+packLen]
		// packet
		ph.Ver = binary.BigEndian.Int8(packBuf[VerOffset:TypeOffset])
		ph.Type = binary.BigEndian.Int16(packBuf[TypeOffset:SeqIdOffset])
		ph.SeqId = binary.BigEndian.Int32(packBuf[SeqIdOffset:RawHeaderSize])
		ph.Body = packBuf[RawHeaderSize:]
		if jb, e = json.Marshal(&ph); e != nil {
			return
		}
		j = json.RawMessage(jb)
		js = append(js, j)
		offset += packLen
	}
	if bts, e = json.Marshal(js); e != nil {
		return
	}
	b.Write(bts)
	return
}

func (p *Proto) WriteWebsocket(wr *websocket.Conn) (e error) {
	if p.Body == nil {
		p.Body = emptyJSONBody
	}
	// [{"ver":1,"op":8,"seq":1,"body":{}}, {"ver":1,"op":3,"seq":2,"body":{}}]
	//if p.Operation == define.OP_RAW {
	//	// batch mod
	//	var b = bytes.NewWriterSize(len(p.Body) + 40*RawHeaderSize)
	//	if e = p.WriteBodyTo(b); e != nil {
	//		return
	//	}
	//	return wr.WriteMessage(websocket.TextMessage, b.Buffer())
	//}

	return wr.WriteJSON([]*Proto{p})
}

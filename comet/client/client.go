package main

import (
	"net"
	"time"
	"math/rand"
	"im/comet/proto"
	"math"
	"encoding/json"
	"encoding/binary"
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	//tcp_url := "127.0.0.1"
	tcp_url := "123.59.187.13"
	port := 10050
	addr := &net.TCPAddr{net.ParseIP(tcp_url), port, ""}
	c, e := net.DialTCP("tcp", nil, addr)
	if e != nil {
		fmt.Print("connect error %v\n", e)
		return
	}
	
	uid := uint32(int64(rand.Int31n(math.MaxInt16)) << 16 | time.Now().UnixNano() & 0xffffffff)
	
	c.SetDeadline(time.Now().Add(time.Duration(120)))
	
	var sqid uint32 = 0
	var rLock    sync.RWMutex
	// read
	go func() {
		head := make([]byte, 12)
		for {
			// |--len--|--x--|--ver--|--type--|--SeqId--|--Body--|
			//     4      1      1        2        4        x
			data, tp, e := Read(c, head, uid)
			if e != nil {
				break
			}
			
			if tp != proto.S2C_HEART_BEAT {
				resp := make(map[string]interface{})
				if e := json.Unmarshal(data, &resp); e != nil {
					fmt.Print("uid %v unmarshal error %v\n", uid, e)
					break
				}
			} else {
				fmt.Print("heartbeat %d", uid)
			}
			
			if rand.Int() % 100 > 30 {
				str := time.Now().Format(time.RFC3339Nano)
				data := fmt.Sprint("%s%s", str, str)
				if e := Send(c, []byte(data), uint16(proto.C2S_CALCULATE), &sqid, &rLock); e != nil {
					fmt.Print("uid %d send calculate err %v\n", uid, e)
					break
				}
			}
		}
	}()
	
	// write
	for {
		time.Sleep(time.Duration(time.Second * 3))
		
		x := rand.Int31()
		y := rand.Int31()
		hbt := proto.HeartBeat{Uid:uid, X:float64(x), Y:float64(y)}
		d, e := json.Marshal(hbt)
		if e != nil {
			break
		}
		
		if e := Send(c, d, uint16(proto.C2S_HEART_BEAT), &sqid, &rLock); e != nil {
			fmt.Print("uid %d send error %v\n", uid, e)
			break
		}
	}
}

func Read(c *net.TCPConn, head[]byte, uid uint32) (data []byte, tp uint16, e error) {
	if e = ReadAll(c, head); e != nil {
		fmt.Printf("uid[%d] error = %v\n", uid, e)
		return
	}
	
	var l int32
	b_buf := bytes.NewBuffer(head[0:4])
	if e = binary.Read(b_buf, binary.BigEndian, &l); e != nil {
		fmt.Print("uid %d error len %v\n", uid, e)
		return
	}
	
	b_buf2 := bytes.NewBuffer(head[6:8])
	if e = binary.Read(b_buf2, binary.BigEndian, &tp); e != nil {
		fmt.Print("uid %v error type %v\n", uid, e)
		return
	}
	
	if tp == proto.S2C_HEART_BEAT {
		c.SetDeadline(time.Now().Add(time.Duration(120)))
		return
	}
	
	data = make([]byte, l)
	if e = ReadAll(c, data); e != nil {
		fmt.Print("uid %d read body error %v\n", uid, e)
		return
	}
	
	return
}

func ReadAll(c *net.TCPConn, data []byte)(e error) {
	start := 0
	length := len(data)
	for length > 0 {
		l, e := c.Read(data[start:])
		if e != nil {
			return e
		}
		length -= l
		start += l
	}
	
	return nil
}

func Send(c *net.TCPConn, d []byte, tp uint16, sqid *uint32, rlock *sync.RWMutex) (e error) {
	atomic.AddUint32(sqid, 1)
	
	l := len(d) + 12
	data := make([]byte, l)
	
	b_buf := new(bytes.Buffer)
	
	binary.Write(b_buf, binary.BigEndian, int32(l))
	copy(data[0:4], b_buf.Bytes()[0:4])
	
	binary.Write(b_buf, binary.BigEndian, int16(tp))
	copy(data[6:8], b_buf.Bytes()[0:2])
	
	binary.Write(b_buf, binary.BigEndian, uint32(*sqid))
	copy(data[8:12], b_buf.Bytes()[0:4])
	copy(data[12:], d)
	
	rlock.Lock()
	_, e = c.Write(data)
	rlock.Unlock()
	return
}

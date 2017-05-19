package utils

import (
	"im/pkg/bytes"
	"im/pkg/time"
)

type RoundOptions struct {
	TimerNum     int
	TimerSize    int
	ReaderNum    int
	ReadbufNum   int
	ReadbufSize  int
	WriterNum    int
	WritebufNum  int
	WritebufSize int
}

// Ronnd userd for connection round-robin get a reader/writer/timer for split big lock.
type Round struct {
	readers   []bytes.Pool
	writers   []bytes.Pool
	timers    []time.Timer
	options   RoundOptions
	readerIdx int
	writerIdx int
	timerIdx  int
}

// NewRound new a round struct.
func NewRound(options RoundOptions) (r *Round) {
	var i int
	r = new(Round)
	r.options = options
	// reader
	r.readers = make([]bytes.Pool, options.ReaderNum)
	for i = 0; i < options.ReaderNum; i++ {
		r.readers[i].Init(options.ReadbufNum, options.ReadbufSize)
	}
	// writer
	r.writers = make([]bytes.Pool, options.WriterNum)
	for i = 0; i < options.WriterNum; i++ {
		r.writers[i].Init(options.WritebufNum, options.WritebufSize)
	}
	// timer
	r.timers = make([]time.Timer, options.TimerNum)
	for i = 0; i < options.TimerNum; i++ {
		r.timers[i].Init(options.TimerSize)
	}
	return
}

// Timer get a timer.
func (r *Round) Timer(rn int) *time.Timer {
	return &(r.timers[rn%r.options.TimerNum])
}

// Reader get a reader memory buffer.
func (r *Round) Reader(rn int) *bytes.Pool {
	return &(r.readers[rn%r.options.ReaderNum])
}

// Writer get a writer memory buffer pool.
func (r *Round) Writer(rn int) *bytes.Pool {
	return &(r.writers[rn%r.options.WriterNum])
}

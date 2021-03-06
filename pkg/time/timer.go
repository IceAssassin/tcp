package time

import (
	"sync"
	itime "time"
	"fmt"
	"im/pkg/log"
)

const (
	timerFormat      = "2006-01-02 15:04:05"
	infiniteDuration = itime.Duration(1<<63 - 1)
)

var (
	Now itime.Time
	Local *itime.Location
	timerLazyDelay = 300 * itime.Millisecond
)

func init() {
	Now = itime.Now().Round(itime.Second)
	Local, _ = itime.LoadLocation("Local")
	go refresh()
}

func refresh() {
	for {
		Now = itime.Now().Round(itime.Second)
		itime.Sleep(100 * itime.Millisecond)
	}
}

// 获取时间界限，如：today  返回stm: 2015-05-01 00:00:00  etm: 2015-05-02: 00:00:00
func TmLime(tmflag string) (stm, etm string) {
	stm = "1970-01-01 00:00:00"
	etm = "2070-01-01 00:00:00"
	if "today" == tmflag {
		stm = Now.Format("2006-01-02") + " 00:00:00"
		etm_tm := Now.AddDate(0, 0, 1)
		etm = etm_tm.Format("2006-01-02") + " 00:00:00"
	} else if "yesterday" == tmflag {
		stm_tm := Now.AddDate(0, 0, -1)
		stm = stm_tm.Format("2006-01-02") + " 00:00:00"
		etm = Now.Format("2006-01-02") + " 00:00:00"
	}
	return
}

type TimerData struct {
	Key    uint64
	expire itime.Time
	fn     func()
	index  int
	next   *TimerData
}

func (td *TimerData) Delay() itime.Duration {
	return td.expire.Sub(itime.Now())
}

func (td *TimerData) ExpireString() string {
	return td.expire.Format(timerFormat)
}

type Timer struct {
	lock   sync.Mutex
	free   *TimerData
	timers []*TimerData
	signal *itime.Timer
	num    int
}

// A heap must be initialized before any of the heap operations
// can be used. Init is idempotent with respect to the heap invariants
// and may be called whenever the heap invariants may have been invalidated.
// Its complexity is O(n) where n = h.Len().
//
func NewTimer(num int) (t *Timer) {
	t = new(Timer)
	t.init(num)
	return t
}

// Init init the timer.
func (t *Timer) Init(num int) {
	t.init(num)
}

func (t *Timer) init(num int) {
	t.signal = itime.NewTimer(infiniteDuration)
	t.timers = make([]*TimerData, 0, num)
	t.num = num
	t.grow()
	go t.start()
}

func (t *Timer) grow() {
	var (
		i   int
		td  *TimerData
		tds = make([]TimerData, t.num)
	)
	t.free = &(tds[0])
	td = t.free
	for i = 1; i < t.num; i++ {
		td.next = &(tds[i])
		td = td.next
	}
	td.next = nil
	return
}

// get get a free timer data.
func (t *Timer) get() (td *TimerData) {
	if td = t.free; td == nil {
		t.grow()
		td = t.free
	}
	t.free = td.next
	return
}

// put put back a timer data.
func (t *Timer) put(td *TimerData) {
	td.fn = nil
	td.next = t.free
	t.free = td
}

// Push pushes the element x onto the heap. The complexity is
// O(log(n)) where n = h.Len().
func (t *Timer) Add(expire itime.Duration, fn func()) (td *TimerData) {
	t.lock.Lock()
	td = t.get()
	td.expire = itime.Now().Add(expire)
	td.fn = fn
	t.add(td)
	t.lock.Unlock()
	return
}

// Del removes the element at index i from the heap.
// The complexity is O(log(n)) where n = h.Len().
func (t *Timer) Del(td *TimerData) {
	t.lock.Lock()
	t.del(td)
	t.put(td)
	t.lock.Unlock()
	return
}

// Push pushes the element x onto the heap. The complexity is
// O(log(n)) where n = h.Len().
func (t *Timer) add(td *TimerData) {
	var d itime.Duration
	td.index = len(t.timers)
	// add to the minheap last node
	t.timers = append(t.timers, td)
	t.up(td.index)
	if td.index == 0 {
		// if first node, signal start goroutine
		d = td.Delay()
		t.signal.Reset(d)
		if Debug {
			fmt.Printf("timer: add reset delay %d ms", int64(d)/int64(itime.Millisecond))
		}
	}
	log.Debug("timer: push item key: %v, expire: %s, index: %d", td.Key, td.ExpireString(), td.index)
	return
}

func (t *Timer) del(td *TimerData) {
	var (
		i    = td.index
		last = len(t.timers) - 1
	)
	if i < 0 || i > last || t.timers[i] != td {
		// already remove, usually by expire
		if Debug {
			fmt.Printf("timer del i: %d, last: %d, %p", i, last, td)
		}
		return
	}
	if i != last {
		t.swap(i, last)
		t.down(i, last)
		t.up(i)
	}
	// remove item is the last node
	t.timers[last].index = -1 // for safety
	t.timers = t.timers[:last]
	log.Debug("timer: remove item key: %v, expire: %s, index: %d", td.Key, td.ExpireString(), td.index)
	return
}

// Set update timer data.
func (t *Timer) Set(td *TimerData, expire itime.Duration) {
	t.lock.Lock()
	t.del(td)
	td.expire = itime.Now().Add(expire)
	t.add(td)
	t.lock.Unlock()
	return
}

// start start the timer.
func (t *Timer) start() {
	for {
		t.expire()
		<-t.signal.C
	}
}

// expire removes the minimum element (according to Less) from the heap.
// The complexity is O(log(n)) where n = max.
// It is equivalent to Del(0).
func (t *Timer) expire() {
	var (
		fn func()
		td *TimerData
		d  itime.Duration
	)
	t.lock.Lock()
	for {
		if len(t.timers) == 0 {
			d = infiniteDuration
			if Debug {
				fmt.Printf("timer: no other instance")
			}
			break
		}
		td = t.timers[0]
		if d = td.Delay(); d > 0 {
			break
		}
		fn = td.fn
		// let caller put back, usually by Del()
		t.del(td)
		t.lock.Unlock()
		if fn == nil {
			fmt.Printf("expire timer no fn")
		} else {
			log.Debug("timer key: %v, expire: %s, index: %d expired, call fn", td.Key, td.ExpireString(), td.index)
			fn()
		}
		t.lock.Lock()
	}
	t.signal.Reset(d)
	if Debug {
		fmt.Printf("timer: expire reset delay %d ms", int64(d)/int64(itime.Millisecond))
	}
	t.lock.Unlock()
	return
}

func (t *Timer) up(j int) {
	for {
		i := (j - 1) / 2 // parent
		if j <= 0 || !t.less(j, i) {
			break
		}
		t.swap(i, j)
		j = i
	}
}

func (t *Timer) down(i, n int) {
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && !t.less(j1, j2) {
			j = j2 // = 2*i + 2  // right child
		}
		if !t.less(j, i) {
			break
		}
		t.swap(i, j)
		i = j
	}
}

func (t *Timer) less(i, j int) bool {
	return t.timers[i].expire.Before(t.timers[j].expire)
}

func (t *Timer) swap(i, j int) {
	t.timers[i], t.timers[j] = t.timers[j], t.timers[i]
	t.timers[i].index = i
	t.timers[j].index = j
}

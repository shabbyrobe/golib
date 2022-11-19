package synctools

import (
	"fmt"
	"sync"
)

type ChanGroup struct {
	c       chan struct{}
	signal  chan struct{}
	active  int
	closed  bool
	waiting bool
	lock    sync.Mutex
}

func NewChanGroup() *ChanGroup {
	cg := &ChanGroup{
		signal: make(chan struct{}),
	}
	return cg
}

func (wg *ChanGroup) Signal() <-chan struct{} {
	wg.lock.Lock()
	defer wg.lock.Unlock()
	wg.waiting = true
	if wg.active == 0 && !wg.closed {
		close(wg.signal)
		wg.closed = true
	}
	return wg.signal
}

func (wg *ChanGroup) Done() {
	wg.lock.Lock()
	defer wg.lock.Unlock()
	if wg.active == 0 {
		panic(fmt.Errorf("done without add"))
	}
	if wg.closed {
		panic(fmt.Errorf("closed"))
	}
	wg.active--
	if wg.active == 0 && wg.waiting && !wg.closed {
		close(wg.signal)
		wg.closed = true
	}
}

func (wg *ChanGroup) Add(n int) {
	wg.lock.Lock()
	defer wg.lock.Unlock()
	if wg.closed {
		panic(fmt.Errorf("reused completed changroup"))
	}
	wg.active += n
}

func (wg *ChanGroup) Wait() {
	<-wg.Signal()
}

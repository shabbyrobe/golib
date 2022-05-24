package panicmerchant

import (
	"runtime/debug"
	"sync/atomic"
)

type Panic struct {
	Stack []byte
	Value any
}

var (
	panics = make(chan Panic, 1024)
	done   int64
)

func Panics() <-chan Panic {
	return panics
}

func Drain() []Panic {
	atomic.StoreInt64(&done, 1)
	out := make([]Panic, 0, 16)
loop:
	for {
		select {
		case p := <-panics:
			out = append(out, p)
		default:
			break loop
		}
	}
	return out
}

func Capture() {
	if r := recover(); r != nil {
		if atomic.LoadInt64(&done) == 1 {
			return
		}
		select {
		case panics <- Panic{
			Stack: debug.Stack(),
			Value: r,
		}:
		default:
		}
	}
}

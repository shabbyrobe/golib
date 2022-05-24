package panicmerchant

import (
	"runtime/debug"
	"sync/atomic"
)

type Panic struct {
	Stack []byte
	Value any
}

func (p Panic) String() string {
	return fmt.Sprintf("Panic!\n%s\nValue: %#v\n", string(p.Stack), p.Value)
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

// Call this in a defer to capture any unhandled panic and send it to the Panics()
// channel.
//
// Consume the Panics() channel in your main() so you can centralise graceful shutdowns.
//
// Example:
//
//  go func() {
//      defer panicmerchant.DeferCapture()
//      panic("FLEEB")
//  }()
//
func DeferCapture() {
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

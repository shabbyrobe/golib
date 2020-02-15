package synctools

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestWaitGroup(t *testing.T) {
	for i := 0; i < 100; i++ {
		wg := NewWaitGroup()
		var done int32
		var n int32 = 100
		for j := 0; j < int(n); j++ {
			wg.Add(1)
			go func() {
				atomic.AddInt32(&done, 1)
				wg.Done()
			}()
		}
		wg.Wait()

		dv := atomic.LoadInt32(&done)
		if n != dv {
			t.Fatal(n, "!=", dv)
		}
	}
}

func TestWaitGroupRecycle(t *testing.T) {
	wg := NewWaitGroup()
	for i := 0; i < 100; i++ {
		var done int32
		var n int32 = 100
		for j := 0; j < int(n); j++ {
			wg.Add(1)
			go func() {
				atomic.AddInt32(&done, 1)
				wg.Done()
			}()
		}
		wg.Wait()

		dv := atomic.LoadInt32(&done)
		if n != dv {
			t.Fatal(n, "!=", dv)
		}
	}
}

func TestWaitGroupIncreaseWhileWaiting(t *testing.T) {
	wg := NewWaitGroup()
	wg.Add(1)
	go func() {
		time.Sleep(50 * time.Millisecond)
		wg.Add(1)
		wg.Done()
		wg.Done()
	}()
	wg.Wait()
}

func TestWaitGroupWaitOnNothing(t *testing.T) {
	wg := NewWaitGroup()
	tm := time.Now()
	wg.Wait()
	if time.Since(tm) >= time.Millisecond {
		t.Fail()
	}
}

func TestWaitGroupMultipleWaiters(t *testing.T) {
	wg := NewWaitGroup()
	for i := 0; i < 100; i++ {
		var done int32
		var n int32 = 1000
		for j := 0; j < int(n); j++ {
			wg.Add(1)
			go func() {
				atomic.AddInt32(&done, 1)
				wg.Done()
			}()
		}
		for k := 0; k < 50; k++ {
			go wg.Wait()
		}

		wg.Wait()
		dv := atomic.LoadInt32(&done)
		if n != dv {
			t.Fatal(n, "!=", dv)
		}
	}
}

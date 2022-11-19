package synctools

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestChanGroup(t *testing.T) {
	for i := 0; i < 100; i++ {
		wg := NewChanGroup()
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

func TestChanGroupWaitOnNothing(t *testing.T) {
	wg := NewChanGroup()
	tm := time.Now()
	wg.Wait()
	if time.Since(tm) >= time.Millisecond {
		t.Fail()
	}
}

func TestChanGroupMultipleWaiters(t *testing.T) {
	for i := 0; i < 100; i++ {
		wg := NewChanGroup()
		var done int32
		var n int32 = 1000
		for j := 0; j < int(n); j++ {
			wg.Add(1)
			go func() {
				atomic.AddInt32(&done, 1)
				wg.Done()
			}()
		}

		wt := make(chan time.Time, 50)
		for k := 0; k < 50; k++ {
			go func() {
				wg.Wait()
				wt <- time.Now()
			}()
		}

		wg.Wait()
		comp := time.Now()
		for k := 0; k < 50; k++ {
			tm := <-wt
			diff := tm.Sub(comp)
			if diff < 0 {
				diff = -diff
			}
			if diff > 1*time.Millisecond {
				t.Fatal()
			}
		}

		wg.Wait()
		dv := atomic.LoadInt32(&done)
		if n != dv {
			t.Fatal(n, "!=", dv)
		}
	}
}

func TestChanGroupMultipleSignals(t *testing.T) {
	for i := 0; i < 100; i++ {
		wg := NewChanGroup()
		var done int32
		var n int32 = 1000
		for j := 0; j < int(n); j++ {
			wg.Add(1)
			go func() {
				atomic.AddInt32(&done, 1)
				wg.Done()
			}()
		}

		wt := make(chan time.Time, 50)
		for k := 0; k < 50; k++ {
			go func() {
				<-wg.Signal()
				wt <- time.Now()
			}()
		}

		<-wg.Signal()
		comp := time.Now()
		for k := 0; k < 50; k++ {
			tm := <-wt
			diff := tm.Sub(comp)
			if diff < 0 {
				diff = -diff
			}
			if diff > 1*time.Millisecond {
				t.Fatal()
			}
		}

		dv := atomic.LoadInt32(&done)
		if n != dv {
			t.Fatal(n, "!=", dv)
		}
	}
}

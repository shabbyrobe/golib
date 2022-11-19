package synctools

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type TestTask struct {
	do      func(ctx context.Context)
	dropped func()
}

func (t *TestTask) Do(ctx context.Context) { t.do(ctx) }
func (t *TestTask) Dropped()               { t.dropped() }

func failAfter(t *testing.T, dur time.Duration, f func(t *testing.T)) {
	var done = make(chan struct{}, 0)
	go func() {
		f(t)
		close(done)
	}()
	select {
	case <-done:
		return
	case <-time.After(dur):
		t.Fatal("timeout")
	}
}

func TestTaskSlotRunsATask(t *testing.T) {
	failAfter(t, 2*time.Second, func(t *testing.T) {
		tr := NewTaskSlot()
		defer tr.Stop()

		var done = make(chan struct{}, 0)
		tr.Next(&TestTask{
			do:      func(ctx context.Context) { close(done) },
			dropped: func() {},
		})

		<-done
		tr.Stop()
	})
}

func TestTaskSlotCancelsATask(t *testing.T) {
	failAfter(t, 2*time.Second, func(t *testing.T) {
		tr := NewTaskSlot()
		defer tr.Stop()

		var cancelled, completed bool
		started := make(chan struct{}, 0)
		tr.Next(&TestTask{
			do: func(ctx context.Context) {
				close(started)
				<-ctx.Done()
				cancelled = true
			},
			dropped: func() {},
		})

		<-started
		done := make(chan struct{}, 0)
		tr.Next(&TestTask{
			do: func(ctx context.Context) {
				close(done)
				completed = true
			},
			dropped: func() {},
		})

		<-done
		tr.Stop()

		if !cancelled || !completed {
			t.Fatal()
		}
	})
}

func TestTaskSlotStopCancelsTask(t *testing.T) {
	failAfter(t, 2*time.Second, func(t *testing.T) {
		tr := NewTaskSlot()
		defer tr.Stop()

		var cancelled bool
		started := make(chan struct{}, 0)
		tr.Next(&TestTask{
			do: func(ctx context.Context) {
				close(started)
				<-ctx.Done()
				cancelled = true
			},
			dropped: func() {},
		})

		<-started
		tr.Stop()

		if !cancelled {
			t.Fatal()
		}
	})
}

func TestTaskSlotSpam(t *testing.T) {
	failAfter(t, 10*time.Second, func(t *testing.T) {
		tr := NewTaskSlot()
		defer tr.Stop()

		var workers = 100
		var iters = 10000

		var n int64
		var wg sync.WaitGroup
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < iters; j++ {
					tr.Next(&TestTask{
						do: func(ctx context.Context) {
							atomic.AddInt64(&n, 1)
						},
						dropped: func() {},
					})
				}
			}()
		}
		wg.Wait()
		tr.Stop()
	})
}

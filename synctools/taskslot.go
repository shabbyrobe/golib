package synctools

import (
	"context"
	"sync"
)

type Task interface {
	// Do MUST NOT panic.
	Do(ctx context.Context)
	Dropped()
}

// This is a bit of boilerplate for a very specific type of task running scenario.
//
//   - Only one task can run at a time.
//   - It may be cancelled and replaced with a task at any time.
//   - If you need to wait for the task, you need to implement that yourself. This
//     does not handle serialisation, only mutual exclusion.
//
// --
type TaskSlot struct {
	stopped chan struct{}
	wake    *sync.Cond
	mu      sync.Mutex

	locked struct {
		stop    bool
		next    Task
		current Task
		ctx     context.Context
		cancel  func()
	}
}

func NewTaskSlot() *TaskSlot {
	tr := &TaskSlot{
		stopped: make(chan struct{}, 0),
	}
	tr.wake = sync.NewCond(&tr.mu)
	go tr.worker()
	return tr
}

// Inform the TaskSlot of the next task to run. Cancel any task currently running. The
// task may be nil to indicate "cancel, but there is no next task".
func (slot *TaskSlot) Next(task Task) {
	slot.mu.Lock()
	defer slot.mu.Unlock()
	if slot.locked.ctx != nil {
		slot.locked.cancel()
	}
	if slot.locked.next != nil {
		slot.locked.next.Dropped()
	}
	slot.locked.next = task
	slot.wake.Broadcast()
}

func (slot *TaskSlot) Stop() {
	func() {
		slot.mu.Lock()
		defer slot.mu.Unlock()
		if slot.locked.cancel != nil {
			slot.locked.cancel()
		}
		slot.locked.stop = true
		slot.wake.Broadcast()
	}()
	<-slot.stopped
}

func (slot *TaskSlot) worker() {
	defer close(slot.stopped)

	prepareTask := func() (task Task, ctx context.Context, cancel func(), done bool) {
		slot.mu.Lock()
		defer slot.mu.Unlock()

		for task == nil {
			if slot.locked.stop {
				return nil, nil, nil, true
			} else if slot.locked.next != nil {
				task = slot.locked.next
			} else {
				slot.wake.Wait()
			}
		}
		slot.locked.next, slot.locked.current = nil, task
		ctx, cancel = context.WithCancel(context.Background())
		slot.locked.ctx, slot.locked.cancel = ctx, cancel
		return task, ctx, cancel, false
	}

	doTask := func(task Task, ctx context.Context, cancel func()) {
		defer cancel()
		task.Do(ctx)
	}

	for {
		task, ctx, cancel, done := prepareTask()
		if done {
			break
		}
		doTask(task, ctx, cancel)
	}
}

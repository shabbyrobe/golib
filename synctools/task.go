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
type TaskRunner struct {
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

func NewTaskRunner() *TaskRunner {
	tr := &TaskRunner{
		stopped: make(chan struct{}, 0),
	}
	tr.wake = sync.NewCond(&tr.mu)
	go tr.worker()
	return tr
}

// Inform the TaskRunner of the next task to run. Cancel any task currently running. The
// task may be nil to indicate "cancel, but there is no next task".
func (runner *TaskRunner) Next(task Task) {
	runner.mu.Lock()
	defer runner.mu.Unlock()
	if runner.locked.ctx != nil {
		runner.locked.cancel()
	}
	if runner.locked.next != nil {
		runner.locked.next.Dropped()
	}
	runner.locked.next = task
	runner.wake.Broadcast()
}

func (runner *TaskRunner) Stop() {
	func() {
		runner.mu.Lock()
		defer runner.mu.Unlock()
		if runner.locked.cancel != nil {
			runner.locked.cancel()
		}
		runner.locked.stop = true
		runner.wake.Broadcast()
	}()
	<-runner.stopped
}

func (runner *TaskRunner) worker() {
	defer close(runner.stopped)

	prepareTask := func() (task Task, ctx context.Context, cancel func(), done bool) {
		runner.mu.Lock()
		defer runner.mu.Unlock()

		for task == nil {
			if runner.locked.stop {
				return nil, nil, nil, true
			} else if runner.locked.next != nil {
				task = runner.locked.next
			} else {
				runner.wake.Wait()
			}
		}
		runner.locked.next, runner.locked.current = nil, task
		ctx, cancel = context.WithCancel(context.Background())
		runner.locked.ctx, runner.locked.cancel = ctx, cancel
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

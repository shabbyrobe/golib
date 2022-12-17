package contexttools

// Provides an implementation of some of the state of
// https://github.com/golang/go/issues/36503 as at 20220926.

import (
	"context"
	"reflect"
	"sync"
	"time"
)

type Cancellation interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
}

type mergedContext struct {
	parentCtx    context.Context
	parentCancel context.CancelFunc
	children     []context.Context
	cancelCtx    context.Context
	done         chan struct{}
	doneClosed   bool
	deadline     time.Time
	deadlineSet  bool

	lock sync.RWMutex
	err  error
}

var _ context.Context = (*mergedContext)(nil)

func (m *mergedContext) Err() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.err
}

func (m *mergedContext) Done() <-chan struct{} {
	return m.done
}

func (m *mergedContext) Deadline() (at time.Time, ok bool) {
	return m.deadline, m.deadlineSet
}

func (m *mergedContext) Value(key any) any {
	// XXX: Proposal https://github.com/golang/go/issues/36503 does not merge values.
	return m.parentCtx.Value(key)
}

func MergeCancel(ctx context.Context, cancelCtxs ...Cancellation) (merged context.Context, cancel context.CancelFunc) {
	mctx := &mergedContext{
		parentCtx: ctx,
		done:      make(chan struct{}),
	}

	mctx.deadline, mctx.deadlineSet = ctx.Deadline()
	for _, mergeable := range cancelCtxs {
		mergeableDeadline, mergeableHasDeadline := mergeable.Deadline()
		if mergeableHasDeadline {
			if !mctx.deadlineSet || mergeableDeadline.Before(mctx.deadline) {
				mctx.deadline = mergeableDeadline
				mctx.deadlineSet = true
			}
		}
	}

	if mctx.deadlineSet {
		mctx.cancelCtx, mctx.parentCancel = context.WithDeadline(ctx, mctx.deadline)
	} else {
		mctx.cancelCtx, mctx.parentCancel = context.WithCancel(ctx)
	}

	cases := make([]reflect.SelectCase, 0, len(cancelCtxs)+2)
	cases = append(cases,
		reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(mctx.parentCtx.Done())},
		reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(mctx.cancelCtx.Done())},
	)

	mctx.children = make([]context.Context, len(cancelCtxs))
	for idx := range cancelCtxs {
		mctx.children[idx] = cancelCtxs[idx].(context.Context)
		cases = append(cases,
			reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(cancelCtxs[idx].Done())})
	}

	go func() {
		chosenCase, _, _ := reflect.Select(cases)
		switch chosenCase {
		case 0:
			mctx.cancel(mctx.parentCtx.Err())
		case 1:
			mctx.cancel(mctx.cancelCtx.Err())
		case 2:
			mctx.cancel(mctx.children[chosenCase-2].Err())
		}
	}()

	return mctx, mctx.parentCancel
}

func (mctx *mergedContext) cancel(err error) {
	mctx.lock.Lock()
	defer mctx.lock.Unlock()
	mctx.err = err
	mctx.parentCancel()
	if !mctx.doneClosed {
		close(mctx.done)
		mctx.doneClosed = true
	}
}

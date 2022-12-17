//go:build ignore

// Alterhative implementation of MergeCancel

package contexttools

import (
	"context"
	"sync"
	"time"
)

type MergeableContext interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
}

// Merge two contexts such that the result context is Done when:
//
//   - parent.Done is closed, or
//   - any of MergeableContext from mergeables is canceled, or
//   - cancel called
//
// https://github.com/golang/go/issues/36503
//
// --
func MergeCancel(parent context.Context, mergeables ...MergeableContext) (merged context.Context, cancel context.CancelFunc) {
	deadline, hasDeadline := parent.Deadline()
	for _, mergeable := range mergeables {
		mergeableDeadline, mergeableHasDeadline := mergeable.Deadline()
		if mergeableHasDeadline {
			if !hasDeadline || mergeableDeadline.Before(deadline) {
				deadline = mergeableDeadline
				hasDeadline = true
			}
		}
	}

	var parentCancel func()
	if hasDeadline {
		merged, parentCancel = context.WithDeadline(parent, deadline)
	} else {
		merged, parentCancel = context.WithCancel(parent)
	}

	childContexts := make([]context.Context, len(mergeables))
	childCancels := make([]context.CancelFunc, len(mergeables))
	cancelOnce := sync.Once{}
	cancel = func() {
		parentCancel()
		cancelOnce.Do(func() {
			for _, childCancel := range childCancels {
				childCancel()
			}
		})
	}

	cancelWhenDone := func(ctx context.Context) {
		<-ctx.Done()
		cancel()
	}

	for idx, mergeable := range mergeables {
		mergeable := mergeable
		childContext, childCancel := context.WithCancel(mergeable.(context.Context))
		childContexts[idx], childCancels[idx] = childContext, childCancel
	}

	go cancelWhenDone(merged)
	go cancelWhenDone(parent)
	for idx := range childContexts {
		childContext := childContexts[idx]
		go cancelWhenDone(childContext)
	}

	return merged, cancel
}

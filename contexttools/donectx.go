package contexttools

import (
	"context"
	"errors"
	"sync/atomic"
)

var ErrDone = errors.New("context: done")

type doneContext struct {
	context.Context
	err  atomic.Value
	done chan struct{}
}

// Combine a context with an externally created 'done' channel such that the context is
// Done() when either the parent context, or the provided 'done' channel yield.
func WithDone(parent context.Context, done chan struct{}) context.Context {
	ch := make(chan struct{})
	doneCtx := &doneContext{
		Context: parent,
		done:    ch,
	}

	go func() {
		select {
		case <-done:
			doneCtx.err.Store(ErrDone)
		case <-parent.Done():
			doneCtx.err.Store(parent.Err())
		}
		close(ch)
	}()

	return doneCtx
}

func (d *doneContext) Done() <-chan struct{} {
	return d.done
}

func (d *doneContext) Err() error {
	if err, ok := d.err.Load().(error); ok {
		return err
	}
	return nil
}

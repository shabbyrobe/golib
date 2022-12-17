package contexttools

import (
	"context"
	"sync"
)

// This is a half-cooked mimic of this CL: https://go-review.googlesource.com/c/go/+/426215/
// If we're lucky, this comes for real in 1.20: https://github.com/golang/go/issues/51365

// A CancelCauseFunc behaves like a CancelFunc but additionally sets the cancelation cause.
// This cause can be retrieved by calling Cause on the canceled Context or any of its derived Contexts.
// If the context has already been canceled, CancelCauseFunc does not set the cause.
type CancelCauseFunc func(cause error)

var cancelCtxKey int

// Cause returns a non-nil error if a parent Context was canceled using a CancelCauseFunc that was passed that error.
// Otherwise Cause returns nil.
func Cause(c context.Context) error {
	ctx, _ := c.Value(&cancelCtxKey).(*cancelCtx)
	if ctx != nil {
		cause := ctx.resolveCause()
		if cause != nil {
			return cause
		}
	}
	return c.Err()
}

// WithCancelCause behaves like WithCancel but returns a CancelCauseFunc instead of a CancelFunc.
// Calling cancel with a non-nil error (the "cause") records that error in ctx;
// it can then be retrieved using Cause(ctx).
//
// Example use:
//
//	ctx, cancel := context.WithCancelCause(parent)
//	cancel(myError)
//	ctx.Err() // returns context.Canceled
//	context.Cause(ctx) // returns myError
func WithCancelCause(parent context.Context) (ctx context.Context, cancel CancelCauseFunc) {
	out := &cancelCtx{}
	out.Context, out.childCancel = context.WithCancel(parent)
	return out, out.cancel
}

// A cancelCtx can be canceled. When canceled, it also cancels any children
// that implement canceler.
type cancelCtx struct {
	context.Context

	mu          sync.Mutex // protects the following fields
	canceled    bool
	childCancel func()
	cause       error
	err         error
}

func (c *cancelCtx) Value(key any) any {
	if key == &cancelCtxKey {
		return c
	}
	return c.Context.Value(key)
}

func (c *cancelCtx) cancel(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// As at 20220922, the CL drops subsequent cancels with cause:
	// https://go-review.googlesource.com/c/go/+/426215/4/src/context/context.go#425
	if c.canceled {
		return
	}
	c.canceled = true
	c.childCancel()
	c.cause = err
}

func (c *cancelCtx) resolveCause() error {
	c.mu.Lock()
	if c.err == nil {
		c.err = c.Context.Err()
	}
	cause := c.cause
	if cause == nil {
		cause = c.err
	}
	c.mu.Unlock()
	return cause
}

func (c *cancelCtx) Err() error {
	c.mu.Lock()
	if c.err == nil {
		c.err = c.Context.Err()
	}
	err := c.err
	c.mu.Unlock()
	return err
}

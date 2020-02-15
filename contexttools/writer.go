package contexttools

import (
	"context"
	"io"
	"time"
)

// Writer wraps an io.Writer with one that checks ctx.Done() on each Write call.
//
// If ctx has a deadline and if w has a `SetWriteDeadline(time.Time) error` method,
// then it is called with the deadline. If SetWriteDeadline fails, the error is ignored.
//
func Writer(ctx context.Context, w io.Writer) io.Writer {
	if deadline, ok := ctx.Deadline(); ok {
		type deadliner interface {
			SetWriteDeadline(time.Time) error
		}
		if d, ok := w.(deadliner); ok {
			d.SetWriteDeadline(deadline)
		}
	}
	return writer{ctx, w}
}

type writer struct {
	ctx context.Context
	w   io.Writer
}

func (w writer) Write(p []byte) (n int, err error) {
	if err = w.ctx.Err(); err != nil { // Err returns an error if ctx is Done
		return n, err
	}
	if n, err = w.w.Write(p); err != nil {
		return n, err
	}
	// FIXME: Why are we calling this again? It won't be for no reason:
	err = w.ctx.Err()
	return n, err
}

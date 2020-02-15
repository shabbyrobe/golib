package contexttools

import (
	"context"
	"io"
	"time"
)

// Reader wraps an io.Reader with one that checks ctx.Done() on each Read call.
//
// If ctx has a deadline and if r has a `SetReadDeadline(time.Time) error` method,
// then it is called with the deadline. If SetReadDeadline fails, the error is ignored.
func Reader(ctx context.Context, r io.Reader) io.Reader {
	if deadline, ok := ctx.Deadline(); ok {
		type deadliner interface {
			SetReadDeadline(time.Time) error
		}
		if d, ok := r.(deadliner); ok {
			d.SetReadDeadline(deadline)
		}
	}
	return reader{ctx, r}
}

type reader struct {
	ctx context.Context
	r   io.Reader
}

func (r reader) Read(p []byte) (n int, err error) {
	if err = r.ctx.Err(); err != nil { // Err returns an error if ctx is Done
		return n, err
	}
	if n, err = r.r.Read(p); err != nil {
		return n, err
	}
	// FIXME: Why are we calling this again? It won't be for no reason:
	err = r.ctx.Err()
	return n, err
}

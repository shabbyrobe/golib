package contexttools

import (
	"context"
	"errors"
	"testing"
	"time"
)

var oops = errors.New("oops")

func TestCancelCause(t *testing.T) {
	ctx, cancel := WithCancelCause(context.Background())
	cancel(oops)

	select {
	case <-ctx.Done():
		if err := Cause(ctx); err != oops {
			t.Fatal(err)
		}

	case <-time.After(100 * time.Millisecond):
		panic("uh oh")
	}
}

func TestCancelCausePropagatesDone(t *testing.T) {
	ctx, cancel := WithCancelCause(context.Background())
	cancel(oops)

	rctx := context.WithValue(ctx, "nothing", "interesting")

	select {
	case <-rctx.Done():
		if err := Cause(ctx); err != oops {
			t.Fatal(err)
		}

	case <-time.After(100 * time.Millisecond):
		panic("uh oh")
	}
}

func TestCancelCauseNilReturnsErr(t *testing.T) {
	ctx, cancel := WithCancelCause(context.Background())
	cancel(nil)

	rctx := context.WithValue(ctx, "nothing", "interesting")

	select {
	case <-rctx.Done():
		if err := Cause(ctx); err != context.Canceled {
			t.Fatal(err)
		}

	case <-time.After(100 * time.Millisecond):
		panic("uh oh")
	}
}

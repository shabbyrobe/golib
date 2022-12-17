package contexttools

import (
	"context"
	"errors"
	"testing"
	"time"
)

func waitOrFail(t *testing.T, ctx context.Context, timeout time.Duration) {
	t.Helper()
	select {
	case <-ctx.Done():
		return
	case <-time.After(timeout):
		t.Fatal("timeout waiting for ctx")
	}
}

func TestMergeParentCancel(t *testing.T) {
	ctx1 := context.Background()
	ctx2 := context.Background()

	merged, cancelParent := MergeCancel(ctx1, ctx2)
	cancelParent()
	waitOrFail(t, merged, 10*time.Millisecond)
	if merged.Err() != context.Canceled {
		t.Fatal()
	}
}

func TestMergeChild1Cancel(t *testing.T) {
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2 := context.Background()

	merged, _ := MergeCancel(ctx1, ctx2)
	cancel1()
	waitOrFail(t, merged, 10*time.Millisecond)
	if merged.Err() != context.Canceled {
		t.Fatal()
	}
}

func TestMergeChild2Cancel(t *testing.T) {
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())

	merged, _ := MergeCancel(ctx1, ctx2)
	cancel1()
	cancel2()
	waitOrFail(t, merged, 10*time.Millisecond)
	if merged.Err() != context.Canceled {
		t.Fatal()
	}
}

func TestMergLotsaChildrenCancel(t *testing.T) {
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	ctx3, cancel3 := context.WithCancel(context.Background())

	merged, _ := MergeCancel(ctx1, ctx2, ctx3)
	cancel1()
	cancel2()
	cancel3()
	waitOrFail(t, merged, 10*time.Millisecond)
	if merged.Err() != context.Canceled {
		t.Fatal()
	}
}

func TestMergeParentDeadline(t *testing.T) {
	deadline1 := time.Now().Add(100 * time.Millisecond)
	deadline2 := time.Now().Add(200 * time.Millisecond)
	ctx1, cancel1 := context.WithDeadline(context.Background(), deadline1)
	ctx2, cancel2 := context.WithDeadline(context.Background(), deadline2)
	_, _ = cancel1, cancel2

	merged, _ := MergeCancel(ctx1, ctx2)
	deadline, ok := merged.Deadline()
	if !ok {
		t.Fatal()
	}
	if deadline != deadline1 {
		t.Fatal()
	}
	waitOrFail(t, merged, 120*time.Millisecond)
	if merged.Err() != context.DeadlineExceeded {
		t.Fatal(merged.Err())
	}
}

func TestMergeChildDeadline(t *testing.T) {
	deadline1 := time.Now().Add(200 * time.Millisecond)
	deadline2 := time.Now().Add(100 * time.Millisecond)
	ctx1, cancel1 := context.WithDeadline(context.Background(), deadline1)
	ctx2, cancel2 := context.WithDeadline(context.Background(), deadline2)
	_, _ = cancel1, cancel2

	merged, _ := MergeCancel(ctx1, ctx2)
	deadline, ok := merged.Deadline()
	if !ok {
		t.Fatal()
	}
	if deadline != deadline2 {
		t.Fatal()
	}
	waitOrFail(t, merged, 120*time.Millisecond)
	if merged.Err() != context.DeadlineExceeded {
		t.Fatal(merged.Err())
	}
}

func TestMergeParentCause(t *testing.T) {
	var oops = errors.New("oops")
	ctx1, cancel1 := WithCancelCause(context.Background())
	ctx2 := context.Background()

	merged, _ := MergeCancel(ctx1, ctx2)
	cancel1(oops)
	waitOrFail(t, merged, 120*time.Millisecond)

	cause := Cause(merged)
	if cause != oops {
		t.Fatal(cause)
	}
}

func TestMergeChildCause(t *testing.T) {
	var oops = errors.New("oops")
	ctx1 := context.Background()
	ctx2, cancel2 := WithCancelCause(context.Background())

	merged, _ := MergeCancel(ctx1, ctx2)
	cancel2(oops)
	waitOrFail(t, merged, 120*time.Millisecond)

	cause := Cause(merged)
	if cause != oops {
		t.Fatal(cause)
	}
}

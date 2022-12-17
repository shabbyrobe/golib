//go:build ignore

package contexttools

import (
	"context"
	"testing"
	"time"
)

func isDone(ctx context.Context) bool {
	t := time.After(1 * time.Second)
	select {
	case <-ctx.Done():
	case <-t:
		return false
	}
	return true
}

func TestMergePropagatesCancelFromParent(t *testing.T) {
	parent, parentCancel := context.WithCancel(context.Background())
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer parentCancel()
	defer cancel1()
	defer cancel2()

	merged, cancel := MergeCancel(parent, ctx1, ctx2)
	defer cancel()

	parentCancel()
	if !isDone(merged) {
		t.Fatal()
	}
}

func TestMergePropagatesCancelFromChild1(t *testing.T) {
	parent, parentCancel := context.WithCancel(context.Background())
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer parentCancel()
	defer cancel1()
	defer cancel2()

	merged, cancel := MergeCancel(parent, ctx1, ctx2)
	defer cancel()

	cancel1()
	if !isDone(merged) {
		t.Fatal()
	}
}

func TestMergePropagatesCancelFromChild2(t *testing.T) {
	parent, parentCancel := context.WithCancel(context.Background())
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer parentCancel()
	defer cancel1()
	defer cancel2()

	merged, cancel := MergeCancel(parent, ctx1, ctx2)
	defer cancel()

	cancel2()
	if !isDone(merged) {
		t.Fatal()
	}
}

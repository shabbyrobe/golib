package synctools

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestAbortGroupEmpty(t *testing.T) {
	ctx := context.Background()
	ag := NewAbortGroup(ctx)
	if err := ag.Wait(); err != nil {
		t.Fatal(err)
	}
}

func TestAbortGroupOneOK(t *testing.T) {
	ctx := context.Background()
	ag := NewAbortGroup(ctx)

	var cnt atomic.Int64
	ag.Go(func(ctx context.Context) error {
		cnt.Add(1)
		return nil
	})
	if err := ag.Wait(); err != nil {
		t.Fatal(err)
	}
	if cnt.Load() != 1 {
		t.Fatal()
	}
}

func TestAbortGroupTwoOK(t *testing.T) {
	ctx := context.Background()
	ag := NewAbortGroup(ctx)

	var cnt atomic.Int64
	ag.Go(func(ctx context.Context) error {
		cnt.Add(1)
		return nil
	})
	ag.Go(func(ctx context.Context) error {
		cnt.Add(1)
		return nil
	})
	if err := ag.Wait(); err != nil {
		t.Fatal(err)
	}
	if cnt.Load() != 2 {
		t.Fatal()
	}
}

func TestAbortGroupOneError(t *testing.T) {
	ctx := context.Background()
	ag := NewAbortGroup(ctx)
	e := fmt.Errorf("bork")
	ag.Go(func(ctx context.Context) error {
		return e
	})
	if err := ag.Wait(); err != e {
		t.Fatal(err, "!=", e)
	}
}

func TestAbortGroupTwoError(t *testing.T) {
	ctx := context.Background()
	ag := NewAbortGroup(ctx)

	var cnt atomic.Int64
	err1, err2 := fmt.Errorf("1"), fmt.Errorf("2")
	ag.Go(func(ctx context.Context) error {
		cnt.Add(1)
		return err1
	})
	ag.Go(func(ctx context.Context) error {
		cnt.Add(1)
		return err2
	})
	if err := ag.Wait(); err != err1 && err != err2 {
		t.Fatal(err)
	}
	if cnt.Load() != 2 {
		t.Fatal()
	}
}

func TestAbortGroupAbort(t *testing.T) {
	ctx := context.Background()
	ag := NewAbortGroup(ctx)

	var cnt atomic.Int64
	var bork = fmt.Errorf("bork")

	ag.Go(func(ctx context.Context) error {
		var after = time.After(1 * time.Second)
		select {
		case <-after:
			return fmt.Errorf("timeout")
		case <-ctx.Done():
			cnt.Add(1)
		}
		return nil
	})

	ag.Go(func(ctx context.Context) error {
		return bork
	})

	tm := time.Now()
	if err := ag.Wait(); err != bork {
		t.Fatal(err)
	}
	if time.Since(tm) > 100*time.Millisecond {
		t.Fatal()
	}
	if cnt.Load() != 1 {
		t.Fatal()
	}
}

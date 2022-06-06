package contexttools

import (
	"context"
	"testing"
	"time"
)

func TestDoneCtx(t *testing.T) {
	after := time.After(1 * time.Second)
	done := make(chan struct{})
	doneCtx := WithDone(context.Background(), done)
	close(done)

	select {
	case <-after:
		panic("timeout")
	case <-doneCtx.Done():
	}
	if doneCtx.Err() != Done {
		t.Fatal()
	}
}

func TestDoneCtxWithCancel(t *testing.T) {
	after := time.After(1 * time.Second)
	done := make(chan struct{})
	doneCtx := WithDone(context.Background(), done)
	cancelCtx, cancel := context.WithCancel(doneCtx)
	defer cancel()
	close(done)

	select {
	case <-after:
		panic("timeout")
	case <-cancelCtx.Done():
	}
	if cancelCtx.Err() != Done {
		t.Fatal()
	}

	select {
	case <-after:
		panic("timeout")
	case <-doneCtx.Done():
	}
	if doneCtx.Err() != Done {
		t.Fatal()
	}
}

func TestCancelCtxWithDone(t *testing.T) {
	after := time.After(1 * time.Second)
	done := make(chan struct{})
	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	doneCtx := WithDone(cancelCtx, done)
	cancel()

	select {
	case <-after:
		panic("timeout")
	case <-cancelCtx.Done():
	}
	if cancelCtx.Err() != context.Canceled {
		t.Fatal()
	}

	select {
	case <-after:
		panic("timeout")
	case <-doneCtx.Done():
	}
	if doneCtx.Err() != context.Canceled {
		t.Fatal()
	}
}

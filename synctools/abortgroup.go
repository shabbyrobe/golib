package synctools

import (
	"context"
	"sync"
)

// Alternative to sync/errgrp.ErrorGroup that cancels a context on the first
// error, rather than waiting for all goroutines to finish before failing.
//
// NOTE: This is not even necessary as there is errgroup.WithContext, though
// this may be slightly more ergonomic. Probably not worth it.
type AbortGroup struct {
	ctx     context.Context
	errc    chan error
	done    chan struct{}
	cancel  func()
	running int
	mu      sync.Mutex
}

func NewAbortGroup(ctx context.Context) *AbortGroup {
	cctx, cancel := context.WithCancel(ctx)
	ag := &AbortGroup{
		ctx:    cctx,
		done:   make(chan struct{}, 1),
		errc:   make(chan error, 1),
		cancel: cancel,
	}
	return ag
}

func (grp *AbortGroup) inc() {
	grp.mu.Lock()
	defer grp.mu.Unlock()
	grp.running++
}

func (grp *AbortGroup) dec() {
	grp.mu.Lock()
	defer grp.mu.Unlock()
	grp.running--
}

func (grp *AbortGroup) active() bool {
	grp.mu.Lock()
	defer grp.mu.Unlock()
	return grp.running > 0
}

func (grp *AbortGroup) Go(fn func(ctx context.Context) error) {
	grp.inc()

	go func() {
		defer func() {
			grp.done <- struct{}{}
		}()

		err := fn(grp.ctx)
		if err != nil {
			grp.cancel()
			select {
			case grp.errc <- err:
			default:
				// Drop, channel full.
			}
		}
	}()
}

func (grp *AbortGroup) Wait() (rerr error) {
	for grp.active() {
		select {
		case err := <-grp.errc:
			if rerr == nil {
				rerr = err
			}
		case <-grp.done:
			grp.dec()
		}
	}
	return rerr
}

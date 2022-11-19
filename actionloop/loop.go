package actionloop

import (
	"context"
	"fmt"

	"sync/atomic"
)

type Nothing struct{}

// "Public" interface to the loop. See also the top-level Query() function, which
// can not be a part of this interface due to generics.
type Doer[TState any] interface {
	loopDoer() // Tag method, does nothing.

	// Simple Queue() for Nothing tasks. Does not await. Any errors that happen within
	// the task are handled inside the loop.
	//
	// See serviceloop.Queue for a generic version that takes Task[anything].
	Queue(ctx context.Context, task Task[TState, Nothing]) error

	// Simple Do() for Nothing tasks. Awaits. Any errors that happen within the task are
	// propagated to the caller.
	//
	// See serviceloop.Do() for the complementary package function version, or
	// serviceloop.Query() for a generic version that supports the (output, error) style
	// and arbitrary returns.
	Do(ctx context.Context, task Task[TState, Nothing]) error
}

func Queue[TState any, TResponse any](ctx context.Context, doer Doer[TState], task Task[TState, TResponse]) error {
	loop := doer.(*Loop[TState])
	action := &taskAction[TState, TResponse]{
		ctx:  ctx,
		task: task,
		// No result channel
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case loop.queue <- action:
		return nil
	case <-loop.shutdownTriggered:
		return fmt.Errorf("shutting down")
	}
}

func Do[TState any](ctx context.Context, doer Doer[TState], task Task[TState, Nothing]) error {
	loop := doer.(*Loop[TState])
	action := &taskAction[TState, Nothing]{
		ctx:    ctx,
		task:   task,
		result: make(chan result[Nothing], 1),
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-loop.shutdownTriggered:
		return fmt.Errorf("shutting down")
	case loop.queue <- action:
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case result := <-action.result:
		return result.Err
	case <-loop.shutdownTriggered:
		return fmt.Errorf("shutting down")
	}
}

// Run a query with an arbitrary response type in the loop, wait for it to complete, and
// return the response and an error if one occurred.
func Query[TState any, TResponse any](ctx context.Context, doer Doer[TState], query Task[TState, TResponse]) (rs TResponse, err error) {
	loop := doer.(*Loop[TState])
	action := &taskAction[TState, TResponse]{
		ctx:    ctx,
		task:   query,
		result: make(chan result[TResponse], 1),
	}

	select {
	case <-ctx.Done():
		return rs, ctx.Err()
	case <-loop.shutdownTriggered:
		return rs, fmt.Errorf("shutting down")
	case loop.queue <- action:
	}

	select {
	case <-ctx.Done():
		return rs, ctx.Err()
	case result := <-action.result:
		return result.Value, result.Err
	case <-loop.shutdownTriggered:
		return rs, fmt.Errorf("shutting down")
	}
}

// NOTE: This will execute in the context of the loop, and thus block it. Do not
// do anything that would cause the loop to seize. You should push the error into
// a queue for handling or something like that.
type ErrorHandler func(error)

func noOpErrorHandler(error) {}

type Loop[TState any] struct {
	queue             chan action[TState]
	state             TState
	errorHandler      ErrorHandler
	shutdownCalled    atomic.Bool
	shutdownTriggered chan struct{}
	shutdownComplete  chan struct{}
}

func (loop *Loop[TState]) loopDoer() {}

func NewLoop[TState any](
	queueBuffer int,
	initState func() TState,
	errorHandler ErrorHandler,
) *Loop[TState] {
	if queueBuffer <= 0 {
		queueBuffer = 1
	}
	if errorHandler == nil {
		errorHandler = noOpErrorHandler
	}
	var state TState
	if initState != nil {
		state = initState()
	}
	loop := &Loop[TState]{
		queue:             make(chan action[TState], queueBuffer),
		shutdownTriggered: make(chan struct{}),
		shutdownComplete:  make(chan struct{}),
		errorHandler:      errorHandler,
		state:             state,
	}

	go loop.actionWorker()

	return loop
}

var _ Doer[any] = (*Loop[any])(nil)

func (loop *Loop[TState]) Shutdown(ctx context.Context) error {
	if loop.shutdownCalled.CompareAndSwap(false, true) {
		close(loop.shutdownTriggered)
	}
	select {
	case <-loop.shutdownComplete:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (loop *Loop[TState]) Queue(ctx context.Context, task Task[TState, Nothing]) error {
	return Queue[TState, Nothing](ctx, loop, task)
}

func (loop *Loop[TState]) Do(ctx context.Context, task Task[TState, Nothing]) error {
	return Do[TState](ctx, loop, task)
}

func (loop *Loop[TState]) actionWorker() {
	defer close(loop.shutdownComplete)

	ctx := context.Background()
	for {
		select {
		case action := <-loop.queue:
			if err := action.act(ctx, loop.state); err != nil {
				// An error should only ever come out of Act if we used Queue(), rather
				// than Do() or Query(). Queue() is fire-and-forget.
				loop.errorHandler(err)
			}

		case <-loop.shutdownTriggered:
			return
		}
	}
}

type result[Value any] struct {
	Value Value
	Err   error
}

package actionloop

import (
	"context"
	"fmt"

	"sync/atomic"
)

// "Public" interface to the loop. See also the top-level Query() function, which
// can not be a part of this interface due to generics.
type Doer[TState any] interface {
	// Stick a task in the loop, but don't bother waiting for it to be finished
	Queue(ctx context.Context, action Action[TState]) error

	// Run a task in the loop that only returns an error, wait for it to complete, and
	// return the error if one occurred.
	Do(ctx context.Context, task Task[TState, error]) error
}

type Loop[TState any] struct {
	queue             chan Action[TState]
	state             *TState
	errorHandler      ErrorHandler
	shutdownCalled    atomic.Bool
	shutdownTriggered chan struct{}
	shutdownComplete  chan struct{}
}

// NOTE: This will execute in the context of the loop, and thus block it. Do not
// do anything that would cause the loop to seize. You should push the error into
// a queue for handling or something like that.
type ErrorHandler func(error)

func noOpErrorHandler(error) {}

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
		queue:             make(chan Action[TState], queueBuffer),
		shutdownTriggered: make(chan struct{}),
		shutdownComplete:  make(chan struct{}),
		errorHandler:      errorHandler,
		state:             &state,
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

func (loop *Loop[TState]) Queue(ctx context.Context, action Action[TState]) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case loop.queue <- action:
		return nil
	case <-loop.shutdownTriggered:
		return fmt.Errorf("shutting down")
	}
}

func (loop *Loop[TState]) Do(ctx context.Context, task Task[TState, error]) error {
	action := &taskAction[TState, error]{
		task:   task,
		result: make(chan error, 1),
	}
	if err := loop.Queue(ctx, action); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-action.result:
		return err
	case <-loop.shutdownTriggered:
		return fmt.Errorf("shutting down")
	}
}

func (loop *Loop[TState]) actionWorker() {
	defer close(loop.shutdownComplete)

	ctx := context.Background()
	for {
		select {
		case action := <-loop.queue:
			if err := action.Act(ctx, loop.state); err != nil {
				// An error should only ever come out of Act if we used Queue(), rather
				// than Do() or Query(). Queue() is fire-and-forget.
				loop.errorHandler(err)
			}

		case <-loop.shutdownTriggered:
			return
		}
	}
}

// Run a query with an arbitrary response type in the loop, wait for it to complete, and
// return the response and an error if one occurred.
func Query[TState any, TResponse any](
	ctx context.Context,
	loop Doer[TState],
	query QueryTask[TState, TResponse],
) (rs TResponse, err error) {
	action := &queryTaskAction[TState, TResponse]{
		task:   query,
		result: make(chan result[TResponse], 1),
	}
	if err := loop.Queue(ctx, action); err != nil {
		return rs, err
	}
	select {
	case <-ctx.Done():
		return rs, ctx.Err()
	case result := <-action.result:
		return result.Value, result.Err
	}
}

type result[Value any] struct {
	Value Value
	Err   error
}

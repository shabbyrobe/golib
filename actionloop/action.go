package actionloop

import "context"

// Lowest level Queueable unit of work for the action loop. Not awaitable.
//
// You should probably implement Task if you have a request/response paradigm or
// need to await.
//
// All accesses to loopState are safe for the lifetime of the act() function.
// Extreme care should be taken when sharing anything from loopState outside
// of it.
//
// loopState MUST NOT be retained.
type Action[TState any] interface {
	Act(ctx context.Context, state *TState) error
}

// Awaitable queueable unit of work for the action loop. Can also implement
// request/response.
//
// See Action for caveats.
type Task[TState any, TResponse any] interface {
	Do(ctx context.Context, state *TState) TResponse
}

// Awaitable queueable unit of work for the action loop, with response value.
type QueryTask[TState any, TResponse any] interface {
	Query(ctx context.Context, state *TState) (TResponse, error)
}

// Adapts a Task to be an Action, such that it can signal completion to a caller outside
// the loop.
//
// This is a bit of generic jiggery-pokery. It's an implementation detail you probably
// shouldn't use yourself. There might be another way around it but it doesn't matter
// much.
type taskAction[TState any, TResponse any] struct {
	task   Task[TState, TResponse]
	result chan TResponse
}

var _ Action[any] = taskAction[any, any]{}

func (action taskAction[TState, TResponse]) Act(ctx context.Context, state *TState) error {
	// NOTE: this is reliant on the queue and do methods of the loop creating a
	// single-element buffered channel always, and discarding it when done. If you
	// have extreme performance requirements, you can recycle channels used in this
	// way if you know that all senders and receivers are done and the channel is
	// empty.
	//
	// The result channel MUST therefore always be of a type that will allow this send to
	// complete, so we don't need to select alongside ctx.Done() here.
	action.result <- action.task.Do(ctx, state)

	return nil
}

type queryTaskAction[TState any, TResponse any] struct {
	task   QueryTask[TState, TResponse]
	result chan result[TResponse]
}

var _ Action[any] = queryTaskAction[any, any]{}

func (action queryTaskAction[TState, TResponse]) Act(ctx context.Context, state *TState) error {
	out, err := action.task.Query(ctx, state)
	action.result <- result[TResponse]{out, err}
	return nil
}

package actionloop

import "context"

// Awaitable queueable unit of work for the action loop. Can also implement
// request/response.
//
// See Action for caveats.
type Task[TState any, TResponse any] interface {
	Do(ctx context.Context, state TState) (TResponse, error)
}

// Adapts a Task to be an Action, such that it can signal completion to a caller outside
// the loop.
//
// This is a bit of generic jiggery-pokery. The crux of the issue is that the loop itself
// uses a single channel that can't handle arbitrary response types, but Task implementers
// should be able to use whatever response type they want. So this type acts as a bridge
// between the part of the interface that needs to deal with arbitrary response types, and
// the loop's inner queue, which can't know about those.
// --
type taskAction[TState any, TResponse any] struct {
	// NOTE(bw): The second type parameter used to be called Response, but the more I came
	// back to it, the more I hated that I couldn't distinguish it from a real type. I
	// initially figured it might be "un-Go" to use the traditional "T-" prefix for type
	// parameters, but so far, I have only found myself wishing it was there.

	// NOTE: you're not typically supposed to retain a context.Context, but we need
	// to in order to allow it to propagate through to the loop's executor.
	ctx context.Context

	task   Task[TState, TResponse]
	result chan result[TResponse]
}

var _ action[any] = taskAction[any, any]{}

func (action taskAction[TState, TResponse]) act(ctx context.Context, state TState) error {
	v, err := action.task.Do(action.ctx, state)

	if action.result != nil {
		// NOTE: this is reliant on the queue and do methods of the loop creating a
		// single-element buffered channel always, and discarding it when done. If you
		// have extreme performance requirements, you can recycle channels used in this
		// way if you know that all senders and receivers are done and the channel is
		// empty.
		//
		// The result channel MUST therefore always be of a type that will allow this send
		// to complete, so we don't need to select alongside ctx.Done() here.
		action.result <- result[TResponse]{v, err}

		// The error gets returned to the caller, not the loop, if there's a result channel:
		return nil
	} else {
		// v is ignored if there's no result channel; return the err to the loop and
		// let the loop take care of it:
		return err
	}
}

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
type action[TState any] interface {
	act(ctx context.Context, state TState) error
}

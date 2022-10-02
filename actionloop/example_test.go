package actionloop

import (
	"context"
	"fmt"
)

type exampleLoopState struct {
	thing string
}

type queryThing struct{}

func (task queryThing) Query(ctx context.Context, state *exampleLoopState) (string, error) {
	return state.thing, nil
}

type updateThing struct {
	value string
}

func (task updateThing) Do(ctx context.Context, state *exampleLoopState) error {
	state.thing = task.value
	return nil
}

// It's potentially worth defining your own one of these, to save you having to type out
// the state type every time you call it (while we wait for Go's type inference to
// hopefully get better).
func QueryExampleLoop[TResponse any](
	ctx context.Context,
	loop Doer[exampleLoopState],
	query QueryTask[exampleLoopState, TResponse],
) (TResponse, error) {
	return Query(ctx, loop, query)
}

func ExampleLoop() {
	initial := func() exampleLoopState {
		return exampleLoopState{
			thing: "initial",
		}
	}

	loop := NewLoop(1, initial, nil)
	defer loop.Shutdown(context.Background())

	before, err := QueryExampleLoop[string](context.Background(), loop, queryThing{})
	if err != nil {
		fmt.Println("value before failed:", err)
	} else {
		fmt.Println("value before:", before)
	}

	if err := loop.Do(context.Background(), updateThing{value: "updated"}); err != nil {
		fmt.Println("value update failed:", err)
	}

	after, err := QueryExampleLoop[string](context.Background(), loop, queryThing{})
	if err != nil {
		fmt.Println("value after failed:", err)
	} else {
		fmt.Println("value after:", after)
	}

	// Output:
	// value before: initial
	// value after: updated
}

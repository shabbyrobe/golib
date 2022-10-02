package actionloop

import (
	"context"
	"testing"
)

var BenchResultString string

type benchLoopState struct {
	thing string
}

type benchLoopQuery struct{}

func (task benchLoopQuery) Query(ctx context.Context, state *benchLoopState) (string, error) {
	return state.thing, nil
}

func QueryBenchLoop[TResponse any](
	ctx context.Context,
	loop Doer[benchLoopState],
	query QueryTask[benchLoopState, TResponse],
) (TResponse, error) {
	return Query(ctx, loop, query)
}

func BenchmarkLoop(b *testing.B) {
	initial := func() benchLoopState {
		return benchLoopState{thing: "thing"}
	}

	loop := NewLoop(1, initial, nil)
	defer loop.Shutdown(context.Background())

	b.ReportAllocs()
	b.ResetTimer()

    var err error
	for i := 0; i < b.N; i++ {
		BenchResultString, err = QueryBenchLoop[string](context.Background(), loop, benchLoopQuery{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

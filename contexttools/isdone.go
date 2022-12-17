package contexttools

import "context"

type DoneContext interface {
	Done() <-chan struct{}
}

var (
	testCtx context.Context
	_       DoneContext = testCtx
)

func IsDone(ctx DoneContext) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

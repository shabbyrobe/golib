package contexttools

import "context"

type Done interface {
	Done() <-chan struct{}
}

var testCtx context.Context
var _ Done = testCtx

func IsDone(ctx Done) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

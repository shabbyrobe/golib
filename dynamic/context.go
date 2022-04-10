package dynamic

import (
	"strings"
	"testing"
)

type Context interface {
	AddError(err error)
}

// An dynamic.Context that collects errors into a buffer:
type ErrContext struct {
	errs []error
}

var _ Context = (*ErrContext)(nil)

func (ctx *ErrContext) AddError(err error) {
	ctx.errs = append(ctx.errs, err)
}

// Return all errors currently in the context. The returned slice is only valid until
// the next use of a Value that touches this context.
func (ctx *ErrContext) PeekErrors() []error {
	return ctx.errs
}

// Remove all errors from the context and return them:
func (ctx *ErrContext) PopErrors() []error {
	errs := ctx.errs
	ctx.errs = nil
	return errs
}

// Pop the most recent error from the end of the error list and return it:
func (ctx *ErrContext) PopError() error {
	if len(ctx.errs) == 0 {
		return nil
	}
	var err error
	err, ctx.errs = ctx.errs[len(ctx.errs)-1], ctx.errs[:len(ctx.errs)-1]
	return err
}

func (ctx *ErrContext) Defer(t *testing.T) {
	t.Helper()
	if len(ctx.errs) > 0 {
		var sb strings.Builder
		for idx, err := range ctx.errs {
			if idx > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(err.Error())
		}
		t.Fatal(sb.String())
	}
}

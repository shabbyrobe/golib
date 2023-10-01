package synctools

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type AllErrGroup struct {
	grp  *errgroup.Group
	errs []error
}

func WithContext(ctx context.Context) (*AllErrGroup, context.Context) {
	grp, ctx := errgroup.WithContext(ctx)
	return &AllErrGroup{grp: grp}, ctx
}

func (g *AllErrGroup) Errors() []error {
	return g.errs
}

func (g *AllErrGroup) Wait() error {
	if g.grp == nil {
		g.grp = &errgroup.Group{}
	}
	return g.grp.Wait()
}

func (g *AllErrGroup) Go(f func() error) {
	if g.grp == nil {
		g.grp = &errgroup.Group{}
	}

	erridx := len(g.errs)
	g.errs = append(g.errs, nil)

	g.grp.Go(func() error {
		err := f()
		g.errs[erridx] = err
		return err
	})
}

func (g *AllErrGroup) TryGo(f func() error) bool {
	if g.grp == nil {
		g.grp = &errgroup.Group{}
	}

	erridx := len(g.errs)
	g.errs = append(g.errs, nil)

	return g.grp.TryGo(func() error {
		err := f()
		g.errs[erridx] = err
		return err
	})
}

func (g *AllErrGroup) SetLimit(n int) {
	if g.grp == nil {
		g.grp = &errgroup.Group{}
	}
	g.grp.SetLimit(n)
}

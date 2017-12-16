// +build ignore

package service

import (
	"fmt"
	"time"
)

type Group struct {
	name         Name
	services     []Service
	haltTimeout  time.Duration
	readyTimeout time.Duration
}

func NewGroup(name Name, services []Service) *Group {
	return &Group{
		name:     name,
		services: services,
	}
}

type groupListener struct {
	errs chan Error
	ends chan Error
}

func newGroupListener(sz int) *groupListener {
	return &groupListener{
		errs: make(chan Error),
		ends: make(chan Error, sz),
	}
}

func (l *groupListener) OnServiceError(service Service, err Error)   { l.errs <- err }
func (l *groupListener) OnServiceEnd(service Service, err Error)     { l.ends <- err }
func (l *groupListener) OnServiceState(service Service, state State) {}

func (g *Group) ServiceName() Name {
	return g.name
}

func (g *Group) Run(ctx Context) error {
	listener := newGroupListener(len(g.services))
	runner := NewRunner(listener)

	for _, s := range g.services {
		if err := runner.Start(s); err != nil {
			if herr := runner.HaltAll(g.haltTimeout); herr != nil {
				// FIXME: should probably not panic
				panic(herr)
			}
		}
	}

	if err := <-runner.WhenReady(g.readyTimeout); err != nil {
		return err
	}

	if err := ctx.Ready(); err != nil {
		return err
	}

	var errRet error

	select {
	case <-ctx.Halt():
	case err := <-listener.errs:
		ctx.OnError(WrapError(err, g))
	case errRet = <-listener.ends:
	}

	herr := runner.HaltAll(g.haltTimeout)
	if herr == nil {
		return errRet
	} else if errRet == nil {
		return herr
	} else {
		for _, e := range es {
			fmt.Println(e...)
		}

		// FIXME: should probably not panic
		panic(fmt.Errorf("%v - %v", herr, errRet))
	}
}

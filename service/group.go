package service

import (
	"fmt"
	"time"
)

const (
	GroupHaltTimeout  = 10 * time.Second
	GroupReadyTimeout = 10 * time.Second
)

// Group wraps a set of services which should all start and stop at the same
// time. If a service Ends, all will be halted.
//
// If the group does not halt properly, you should panic as it is likely
// there is no way to recover the lost resources.
//
type Group struct {
	name         Name
	services     []Service
	haltTimeout  time.Duration
	readyTimeout time.Duration
}

func NewGroup(name Name, services []Service) *Group {
	return &Group{
		name:         name,
		services:     services,
		haltTimeout:  GroupHaltTimeout,
		readyTimeout: GroupReadyTimeout,
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
				return &errGroupHalt{name: g.ServiceName(), haltError: herr, cause: err}
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
		return &errGroupHalt{name: g.ServiceName(), haltError: herr, cause: errRet}
	}
}

func IsErrGroupHalt(err error) bool {
	_, ok := err.(*errGroupHalt)
	return ok
}

type errGroupHalt struct {
	name      Name
	haltError error
	cause     error
}

func (e *errGroupHalt) Error() string {
	return fmt.Sprintf("group halt error in service %s: %v, caused by: %v", e.name, e.haltError, e.cause)
}

func (e *errGroupHalt) Cause() error {
	return e.cause
}

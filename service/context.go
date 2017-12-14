package service

import "time"

const MinHaltableSleep = 50 * time.Millisecond

type Context interface {
	Halt() <-chan struct{}
	Halted() bool
	Ready(service Service) error
	OnError(service Service, err error)
}

func Sleep(ctx Context, d time.Duration) (halted bool) {
	if d < MinHaltableSleep {
		time.Sleep(d)
		select {
		case <-ctx.Halt():
			return true
		default:
			return false
		}
	}
	select {
	case <-time.After(d):
		return false
	case <-ctx.Halt():
		return true
	}
}

type ContextListener interface {
	Ready(service Service) error
	OnError(service Service, err error)
}

func NewContext(listener ContextListener, halter chan struct{}) Context {
	return &context{
		halt:            halter,
		ContextListener: listener,
	}
}

type context struct {
	ContextListener
	halt chan struct{}
}

func (c *context) Halt() <-chan struct{} { return c.halt }

func (c *context) Halted() bool {
	select {
	case <-c.halt:
		return true
	default:
		return false
	}
}

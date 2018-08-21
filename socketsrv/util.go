package socketsrv

import "time"

// chanContext allows a simple done channel to masquerade as a context.Context
type chanContext struct {
	done chan struct{}
}

func (csc *chanContext) Deadline() (deadline time.Time, ok bool) {
	return
}

func (csc *chanContext) Done() <-chan struct{} {
	return csc.done
}

func (csc *chanContext) Err() error {
	// FIXME: maybe should return an error when done is closed:
	return nil
}

func (csc *chanContext) Value(key interface{}) interface{} {
	return nil
}

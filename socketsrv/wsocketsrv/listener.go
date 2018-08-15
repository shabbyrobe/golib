package wsocketsrv

import (
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/shabbyrobe/golib/socketsrv"
)

type Listener struct {
	comms    chan socketsrv.Communicator
	upgrader websocket.Upgrader
	closed   uint32
	stop     chan struct{}
}

func NewListener(upgrader websocket.Upgrader) *Listener {
	return &Listener{
		// FIXME: buffer size
		comms: make(chan socketsrv.Communicator, 1024),

		upgrader: upgrader,
		stop:     make(chan struct{}),
	}
}

func (l *Listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws, err := l.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// FIXME: is panic appropriate here?
		panic(err)
	}
	comm := NewCommunicator(ws)
	select {
	case l.comms <- comm:
	case <-l.stop:
	}
}

func (l *Listener) Accept() (socketsrv.Communicator, error) {
	select {
	case comm := <-l.comms:
		return comm, nil
	case <-l.stop:
		return nil, errors.New("wslistener: listener closed")
	}
}

func (l *Listener) Close() error {
	if !atomic.CompareAndSwapUint32(&l.closed, 0, 1) {
		return fmt.Errorf("wsocketsrv: listener already closed")
	}
	close(l.stop)
	return nil
}

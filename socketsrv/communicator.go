package socketsrv

import (
	"time"
)

type Communicator interface {
	Close() error
	Ping(timeout time.Duration) error
	ReadMessage(into []byte, limit uint32, timeout time.Duration) (extended []byte, rerr error)

	// Pongs returns a channel that yields when a pong is received in response
	// to a Ping. Implementers can choose to either return a slice of len(0)
	// from ReadMessage, or return a channel from Pongs(). A channel buffer of
	// 1 is recommended, but sends to the Pongs channel must not block.
	Pongs() <-chan struct{}
	WriteMessage(data []byte, timeout time.Duration) (rerr error)
}

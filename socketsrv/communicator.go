package socketsrv

import (
	"time"
)

type Communicator interface {
	Close() error
	Ping(timeout time.Duration) error
	ReadMessage(into []byte, limit uint32, timeout time.Duration) (extended []byte, rerr error)
	WriteMessage(data []byte, timeout time.Duration) (rerr error)
}

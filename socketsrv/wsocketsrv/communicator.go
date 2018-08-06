package wsocketsrv

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shabbyrobe/golib/socketsrv"
)

type Communicator struct {
	ws       *websocket.Conn
	pongs    chan time.Time
	rdLenBuf [4]byte
	wrLenBuf [4]byte
}

var _ socketsrv.Communicator = &Communicator{}

func NewCommunicator(ws *websocket.Conn) *Communicator {
	comm := &Communicator{
		ws:    ws,
		pongs: make(chan time.Time, 1),
	}

	existing := ws.PongHandler()
	ws.SetPongHandler(func(s string) error {
		select {
		case comm.pongs <- time.Now():
		default:
		}
		if existing != nil {
			return existing(s)
		}
		return nil
	})

	return comm
}

func (cm *Communicator) Close() error {
	return cm.ws.Close()
}

func (cm *Communicator) ReadMessage(into []byte, limit uint32, timeout time.Duration) (extended []byte, rerr error) {
	if timeout > 0 {
		if err := cm.ws.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return into, err
		}
	}

	_, rdr, err := cm.ws.NextReader()
	if err != nil {
		return nil, err
	}

	lbuf := cm.rdLenBuf[:]
	if _, err := io.ReadFull(rdr, lbuf); err != nil {
		return into, err
	}

	// The websocket protocol makes length available as part of the header,
	// but the gorilla library does not expose the field for us to validate:
	mlen := binary.BigEndian.Uint32(lbuf)
	if mlen > limit {
		return into, fmt.Errorf("socket: message of length %d exceeded limit %d", mlen, uint32(limit))
	}

	if uint32(cap(into)) < mlen {
		into = make([]byte, mlen)
	} else {
		into = into[:mlen]
	}

	if _, err := io.ReadFull(rdr, into); err != nil {
		return into, err
	}

	return into, nil
}

func (cm *Communicator) WriteMessage(data []byte, limit uint32, timeout time.Duration) (rerr error) {
	if timeout > 0 {
		if err := cm.ws.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
			return err
		}
	}

	mlen := len(data)
	if uint32(mlen) > limit {
		return fmt.Errorf("socket: message of length %d exceeded limit %d", mlen, limit)
	}

	wr, err := cm.ws.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return err
	}

	lbuf := cm.wrLenBuf[:]
	binary.BigEndian.PutUint32(lbuf, uint32(mlen))
	if n, err := wr.Write(lbuf); err != nil {
		_ = wr.Close()
		return err

	} else if n != 4 {
		_ = wr.Close()
		return fmt.Errorf("short length write")
	}

	if n, err := wr.Write(data); err != nil {
		_ = wr.Close()
		return err
	} else if n != mlen {
		_ = wr.Close()
		return fmt.Errorf("short message write")
	}

	return wr.Close()
}

func (cm *Communicator) Ping(timeout time.Duration) (rerr error) {
	if timeout > 0 {
		if err := cm.ws.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
			return err
		}
	}

	return cm.ws.WritePreparedMessage(ping)
}

func (cm *Communicator) Pongs() <-chan time.Time {
	return cm.pongs
}

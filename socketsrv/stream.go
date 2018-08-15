package socketsrv

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

func Stream(nc net.Conn) Communicator {
	return &stream{
		conn:   nc,
		reader: nc,
		writer: nc,
	}
}

type stream struct {
	conn     net.Conn
	reader   io.Reader
	writer   io.Writer
	rdLenBuf [4]byte
	wrLenBuf [4]byte
}

var _ Communicator = &stream{}

func (nc *stream) Close() error {
	return nc.conn.Close()
}

func (nc *stream) MessageLimit() int { return 0 }

func (nc *stream) Pongs() <-chan struct{} {
	// pongs are represented as 0-length returns from ReadMessage in this
	// Communicator.
	return nil
}

func (nc *stream) ReadMessage(into []byte, limit int, timeout time.Duration) (buf []byte, rerr error) {
	if timeout > 0 {
		if err := nc.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return into, err
		}
	}

	lbuf := nc.rdLenBuf[:]
	if _, err := io.ReadFull(nc.reader, lbuf); err != nil {
		return into, err
	}

	mlen := int(binary.BigEndian.Uint32(lbuf))
	if mlen > limit {
		return into, fmt.Errorf("socket: message of length %d exceeded limit %d", mlen, uint32(limit))

	} else if mlen == 0 {
		return into[:0], nil
	}

	if cap(into) < mlen {
		into = make([]byte, mlen)
	} else {
		into = into[:mlen]
	}

	if _, err := io.ReadFull(nc.reader, into); err != nil {
		return into, err
	}

	return into, nil
}

func (nc *stream) WriteMessage(data []byte, timeout time.Duration) (rerr error) {
	if timeout > 0 {
		if err := nc.conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
			return err
		}
	}

	mlen := len(data)

	lbuf := nc.wrLenBuf[:]
	binary.BigEndian.PutUint32(lbuf, uint32(mlen))

	// FIXME: this puts pressure on the GC but it's significantly faster for smaller messages.
	// the protocol really needs to be adjusted to accept a writer.
	out := append(lbuf, data...)

	if n, err := nc.writer.Write(out); err != nil {
		return err
	} else if n != len(out) {
		return fmt.Errorf("short write")
	}

	return nil
}

func (nc *stream) Ping(timeout time.Duration) (rerr error) {
	if timeout > 0 {
		if err := nc.conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
			return err
		}
	}

	if n, err := nc.writer.Write(pingBuf); err != nil {
		return err
	} else if n != 4 {
		return fmt.Errorf("short length write")
	}

	return nil
}

var pingBuf = []byte{0, 0, 0, 0}

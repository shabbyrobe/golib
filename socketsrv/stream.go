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
		pongs:  make(chan time.Time, 1),
	}
}

type stream struct {
	conn     net.Conn
	reader   io.Reader
	writer   io.Writer
	pongs    chan time.Time
	rdLenBuf [4]byte
	wrLenBuf [4]byte
}

var _ Communicator = &stream{}

func (nc *stream) Close() error {
	return nc.conn.Close()
}

func (nc *stream) ReadMessage(into []byte, limit uint32, timeout time.Duration) (buf []byte, rerr error) {
	if timeout > 0 {
		if err := nc.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return into, err
		}
	}

	lbuf := nc.rdLenBuf[:]

	if _, err := io.ReadFull(nc.reader, lbuf); err != nil {
		return into, err
	}

	mlen := binary.BigEndian.Uint32(lbuf)
	if mlen > limit {
		return into, fmt.Errorf("socket: message of length %d exceeded limit %d", mlen, uint32(limit))

	} else if mlen == 0 {
		select {
		case nc.pongs <- time.Now():
		default:
		}
		return into[:0], nil
	}

	if uint32(cap(into)) < mlen {
		into = make([]byte, mlen)
	} else {
		into = into[:mlen]
	}

	if _, err := io.ReadFull(nc.reader, into); err != nil {
		return into, err
	}

	return into, nil
}

func (nc *stream) WriteMessage(data []byte, limit uint32, timeout time.Duration) (rerr error) {
	if timeout > 0 {
		if err := nc.conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
			return err
		}
	}

	mlen := len(data)
	if uint32(mlen) > limit {
		return fmt.Errorf("socket: message of length %d exceeded limit %d", mlen, limit)
	}

	lbuf := nc.wrLenBuf[:]
	binary.BigEndian.PutUint32(lbuf, uint32(mlen))
	if n, err := nc.writer.Write(lbuf); err != nil {
		return err
	} else if n != 4 {
		return fmt.Errorf("short length write")
	}

	if n, err := nc.writer.Write(data); err != nil {
		return err
	} else if n != mlen {
		return fmt.Errorf("short message write")
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

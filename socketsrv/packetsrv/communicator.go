package packetsrv

import (
	"fmt"
	"net"
	"time"

	"github.com/shabbyrobe/golib/socketsrv"
)

type communicator struct {
	conn  net.PacketConn
	addr  net.Addr
	pongs chan time.Time
}

var _ socketsrv.Communicator = &communicator{}

func (pc *communicator) Close() error {
	return pc.conn.Close()
}

func (pc *communicator) ReadMessage(into []byte, limit uint32, timeout time.Duration) (buf []byte, rerr error) {
	if timeout > 0 {
		if err := pc.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return into, err
		}
	}

	if cap(into) < int(limit) {
		into = make([]byte, int(limit))
	} else {
		into = into[:int(limit)]
	}

	n, addr, err := pc.conn.ReadFrom(into)
	if err != nil {
		return into, err
	}
	if addr != pc.addr {
		return into[:0], nil
	}
	if n == 1 && into[0] == 0 {
		select {
		case pc.pongs <- time.Now():
		default:
		}
		return into[:0], nil
	}

	return into[:n], nil
}

func (pc *communicator) WriteMessage(data []byte, limit uint32, timeout time.Duration) (rerr error) {
	if timeout > 0 {
		if err := pc.conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
			return err
		}
	}

	mlen := len(data)
	if uint32(mlen) > limit {
		return fmt.Errorf("socket: packet of length %d exceeded limit %d", mlen, limit)
	}

	if n, err := pc.conn.WriteTo(data, pc.addr); err != nil {
		return err
	} else if n != mlen {
		return fmt.Errorf("short message write")
	}

	return nil
}

func (pc *communicator) Ping(timeout time.Duration) (rerr error) {
	if timeout > 0 {
		if err := pc.conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
			return err
		}
	}

	if n, err := pc.conn.WriteTo(packetPingBuf, pc.addr); err != nil {
		return err
	} else if n != packetPingBufLen {
		return fmt.Errorf("short message write")
	}

	return nil
}

var packetPingBuf = []byte{0}

const packetPingBufLen = 1

package packetsrv

import (
	"fmt"
	"net"
	"time"

	"github.com/shabbyrobe/golib/socketsrv"
)

type clientCommunicator struct {
	conn net.Conn
}

var _ socketsrv.Communicator = &communicator{}

func ClientCommunicator(conn net.Conn) socketsrv.Communicator {
	return &clientCommunicator{
		conn: conn,
	}
}

func (pc *clientCommunicator) Close() error {
	return pc.conn.Close()
}

func (pc *clientCommunicator) ReadMessage(into []byte, limit uint32, timeout time.Duration) (buf []byte, rerr error) {
	if timeout > 0 {
		if err := pc.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return into, err
		}
	}

	if cap(into) < 65536 {
		into = make([]byte, 65536)
	} else {
		into = into[:65536]
	}

	n, err := pc.conn.Read(into)
	if err != nil {
		return into, err
	}

	return into[:n], nil
}

func (pc *clientCommunicator) WriteMessage(data []byte, timeout time.Duration) (rerr error) {
	if timeout > 0 {
		if err := pc.conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
			return err
		}
	}

	if n, err := pc.conn.Write(data); err != nil {
		return err
	} else if n != len(data) {
		return fmt.Errorf("short message write")
	}

	return nil
}

func (pc *clientCommunicator) Ping(timeout time.Duration) (rerr error) {
	return pc.WriteMessage(packetPingBuf, timeout)
}

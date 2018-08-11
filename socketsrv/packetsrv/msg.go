package packetsrv

import "net"

type readMsg struct {
	n    int
	addr net.Addr
	buf  []byte
}

type writeMsg struct {
	addr net.Addr
	buf  []byte
	errc chan error
}

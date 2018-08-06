package socketsrv

import "net"

type Listener interface {
	Accept() (Communicator, error)
	Close() error
}

func Listen(network, addr string) (Listener, error) {
	l, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	return &netListener{l: l}, nil
}

func ListenNet(nl net.Listener) (Listener, error) {
	return &netListener{l: nl}, nil
}

type netListener struct {
	l net.Listener
}

func (nl *netListener) Accept() (Communicator, error) {
	c, err := nl.l.Accept()
	if err != nil {
		return nil, err
	}
	raw := &netConn{conn: c}
	return raw, nil
}

func (nl *netListener) Close() error {
	return nl.l.Close()
}

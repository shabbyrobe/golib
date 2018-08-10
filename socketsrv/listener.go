package socketsrv

import (
	"net"
)

type Listener interface {
	Accept() (Communicator, error)
	Close() error
}

func ListenStream(network, addr string) (Listener, error) {
	l, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	return &streamListener{l: l}, nil
}

func ListenStreamWith(nl net.Listener) (Listener, error) {
	return &streamListener{l: nl}, nil
}

type streamListener struct {
	l net.Listener
}

func (nl *streamListener) Accept() (Communicator, error) {
	c, err := nl.l.Accept()
	if err != nil {
		return nil, err
	}
	raw := Stream(c)
	return raw, nil
}

func (nl *streamListener) Close() error {
	return nl.l.Close()
}

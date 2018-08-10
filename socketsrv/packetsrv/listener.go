package packetsrv

import (
	"errors"
	"net"

	service "github.com/shabbyrobe/go-service"
	"github.com/shabbyrobe/golib/incrementer"
	"github.com/shabbyrobe/golib/socketsrv"
)

func Listen(network, addr string) (socketsrv.Listener, error) {
	pl := &listener{
		network: network,
		addr:    addr,
		accept:  make(chan *communicator, 1024), // FIXME: buffer size
		stop:    make(chan struct{}),
	}
	return pl, nil
}

type listener struct {
	network string
	addr    string
	inc     incrementer.Inc

	accept chan *communicator
	comms  map[string]*communicator

	stop chan struct{}
}

func (pl *listener) Run(ctx service.Context) error {
	pc, err := net.ListenPacket(pl.network, pl.addr)
	if err != nil {
		return err
	}

	if err := ctx.Ready(); err != nil {
		return err
	}

	into := make([]byte, 65536) // FIXME

	for {
		n, addr, err := pc.ReadFrom(into)
		if err != nil {
			return err
		}

	}

	<-ctx.Done()

	return nil
}

func (pl *listener) Accept() (socketsrv.Communicator, error) {
	select {
	case comm := <-pl.accept:
		return comm, nil
	case <-pl.stop:
		return nil, errors.New("wslistener: listener closed")
	}
}

func (pl *listener) Close() error {
	close(pl.stop)
	return nil
}

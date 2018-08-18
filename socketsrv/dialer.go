package socketsrv

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/shabbyrobe/golib/incrementer"
)

var (
	nextID   incrementer.Inc
	nextIDMu sync.Mutex
)

func DefaultDialer(neg Negotiator) Dialer {
	return Dialer{
		ConnConfig:  DefaultConnConfig(),
		DialTimeout: 10 * time.Second,
		Negotiator:  neg,
	}
}

// dial is an accumulator for ClientOption.
type dial struct {
	onDisconnect OnClientDisconnect
}

// Dialer creates Client connections to Servers.
type Dialer struct {
	ConnConfig
	DialTimeout time.Duration
	Negotiator  Negotiator
}

type ClientOption func(dial *dial)

type OnClientDisconnect func(id ConnID, err error)

func ClientDisconnect(cb OnClientDisconnect) ClientOption {
	return func(dial *dial) { dial.onDisconnect = cb }
}

func (d Dialer) DialStream(ctx context.Context, network, host string, handler Handler, opts ...ClientOption) (Client, error) {
	if d.Negotiator == nil {
		return nil, fmt.Errorf("socketsrv: dialer missing negotiator")
	}
	nd := net.Dialer{
		Timeout: d.DialTimeout,
	}
	conn, err := nd.DialContext(ctx, network, host)
	if err != nil {
		return nil, err
	}
	raw := Stream(conn)
	return d.Client(ctx, raw, handler, opts...)
}

// Client wraps a Communicator, starts a connection and returns a client. It is
// intended for use when extending socketsrv with new Communicator
// implementations.
func (d Dialer) Client(ctx context.Context, rc Communicator, handler Handler, opts ...ClientOption) (Client, error) {
	if d.Negotiator == nil {
		return nil, fmt.Errorf("socketsrv: dialer missing negotiator")
	}

	var currentDial dial
	for _, o := range opts {
		o(&currentDial)
	}

	nextIDMu.Lock()
	id := ConnID(nextID.Next())
	nextIDMu.Unlock()

	conn := newConn(id, ClientSide, d.ConnConfig, rc, d.Negotiator, handler)

	ended := make(chan error, 1)
	if err := conn.start(ended); err != nil {
		return nil, err
	}

	if currentDial.onDisconnect != nil {
		go func() {
			currentDial.onDisconnect(id, <-ended)
		}()
	}

	return conn, nil
}

type Client interface {
	Close() error
	ID() ConnID
	Send(ctx context.Context, msg Message, recv chan<- Result) (rerr error)
	Request(ctx context.Context, msg Message) (resp Message, rerr error)
}

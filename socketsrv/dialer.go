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
	onConnects    []OnClientConnect
	onDisconnects []OnClientDisconnect
}

// Dialer creates Client connections to Servers.
type Dialer struct {
	ConnConfig
	DialTimeout time.Duration
	Negotiator  Negotiator
}

type ClientOption func(dial *dial)

type OnClientConnect func(id ConnID)
type OnClientDisconnect func(id ConnID, err error)

// ClientConnect is a ClientOption that registers a callback that happens when
// a dialer establishes a client connection. It will block the Dial/Client methods
// and is guaranteed to happen before the ClientDisconnect.
func ClientConnect(cb OnClientConnect) ClientOption {
	return func(dial *dial) { dial.onConnects = append(dial.onConnects, cb) }
}

// ClientConnect is a ClientOption that registers a callback that happens when
// a client connection is lost. It is guaranteed to happen after ClientConnect.
// It will not be called if Client() or DialStream() would return an error; you
// will only get one error or the other.
func ClientDisconnect(cb OnClientDisconnect) ClientOption {
	return func(dial *dial) { dial.onDisconnects = append(dial.onDisconnects, cb) }
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

	for _, oc := range currentDial.onConnects {
		oc(id)
	}

	if len(currentDial.onDisconnects) > 0 {
		go func() {
			err := <-ended
			for _, od := range currentDial.onDisconnects {
				od(id, err)
			}
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

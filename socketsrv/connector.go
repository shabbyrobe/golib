package socketsrv

import (
	"context"
	"net"
	"sync"

	service "github.com/shabbyrobe/go-service"
	"github.com/shabbyrobe/golib/incrementer"
)

// Connector creates Client connections to Servers.
//
// Connector is similar to the idea of net.Dialer, but Dialer is more of a
// configuration struct. Connector retains state about the connections that are
// currently running and you can shut them all down simultaneously with
// Shutdown() if you like.
//
type Connector struct {
	config     ConnectorConfig
	clients    service.Runner
	negotiator Negotiator

	nextID     incrementer.Inc
	nextIDLock sync.Mutex
}

type ClientOption func(cnct *Connector, c *client)

type OnClientDisconnect func(connector *Connector, id ConnID, err error)

func ClientDisconnect(cb OnClientDisconnect) ClientOption {
	return func(cnct *Connector, c *client) {
		c.svc.WithOnEnd(func(stage service.Stage, svc *service.Service, err error) {
			if stage == service.StageRun {
				c := svc.Runnable.(*conn)
				cb(cnct, c.ID(), err)
			}
		})
	}
}

func NewConnector(config *ConnectorConfig, negotiator Negotiator) *Connector {
	if config == nil || config.IsZero() {
		config = DefaultConnectorConfig()
	}
	dl := &Connector{
		config:     *config,
		negotiator: negotiator,
	}
	dl.clients = service.NewRunner()
	return dl
}

func (c *Connector) Shutdown(ctx context.Context) error {
	return c.clients.Shutdown(ctx)
}

func (c *Connector) StreamClient(ctx context.Context, network, host string, handler Handler, opts ...ClientOption) (Client, error) {
	d := net.Dialer{
		Timeout: c.config.DialTimeout,
	}
	conn, err := d.DialContext(ctx, network, host)
	if err != nil {
		return nil, err
	}

	raw := Stream(conn)
	return c.Client(ctx, raw, handler, opts...)
}

func (c *Connector) Client(ctx context.Context, rc Communicator, handler Handler, opts ...ClientOption) (Client, error) {
	c.nextIDLock.Lock()
	id := ConnID(c.nextID.Next())
	c.nextIDLock.Unlock()

	conn := newConn(id, ClientSide, c.config.Conn, rc, c.negotiator, handler)
	svc := service.New(service.Name(conn.ID()), conn)

	cl := &client{
		conn:      conn,
		svc:       svc,
		connector: c,
	}
	for _, o := range opts {
		o(c, cl)
	}

	if err := c.clients.Start(ctx, cl.svc); err != nil {
		_ = rc.Close()
		return nil, err
	}
	return cl, nil
}

func (c *Connector) halt(client *client) error {
	return service.HaltTimeout(c.config.HaltTimeout, c.clients, client.svc)
}

type Client interface {
	Close() error
	ID() ConnID
	Send(ctx context.Context, msg Message, recv chan<- Result) (rerr error)
	Request(ctx context.Context, msg Message) (resp Message, rerr error)
}

type client struct {
	conn      *conn
	connector *Connector
	svc       *service.Service
}

func (c *client) Close() error {
	return c.connector.halt(c)
}

func (c *client) ID() ConnID {
	return c.conn.ID()
}

func (c *client) Send(ctx context.Context, msg Message, recv chan<- Result) (rerr error) {
	return c.conn.Send(ctx, msg, recv)
}

func (c *client) Request(ctx context.Context, msg Message) (resp Message, rerr error) {
	return c.conn.Request(ctx, msg)
}

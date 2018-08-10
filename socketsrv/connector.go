package socketsrv

import (
	"context"
	"net"

	service "github.com/shabbyrobe/go-service"
	"github.com/shabbyrobe/golib/incrementer"
)

type Connector struct {
	config     ConnectorConfig
	clients    service.Runner
	negotiator Negotiator
	nextID     incrementer.Inc
}

func NewConnector(config ConnectorConfig, negotiator Negotiator) *Connector {
	if config.IsZero() {
		config = DefaultConnectorConfig()
	}
	dl := &Connector{
		config:     config,
		negotiator: negotiator,
	}
	dl.clients = service.NewRunner()
	return dl
}

func (c *Connector) NetClient(ctx context.Context, network, host string, handler Handler) (*Client, error) {
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, network, host)
	if err != nil {
		return nil, err
	}

	raw := Stream(conn)
	return c.Client(ctx, raw, handler)
}

func (c *Connector) Client(ctx context.Context, rc Communicator, handler Handler) (*Client, error) {
	id := ConnID(c.nextID.Next())
	conn := newConn(id, ClientSide, c.config.Conn, rc, c.negotiator, handler)
	cl := &Client{
		conn:      conn,
		svc:       service.New(service.Name(conn.ID()), conn),
		connector: c,
	}
	if err := c.clients.Start(ctx, cl.svc); err != nil {
		_ = rc.Close()
		return nil, err
	}
	return cl, nil
}

func (c *Connector) halt(client *Client) error {
	return service.HaltTimeout(c.config.HaltTimeout, c.clients, client.svc)
}

type Client struct {
	conn      *conn
	connector *Connector
	svc       *service.Service
}

func (c *Client) Close() error {
	return c.connector.halt(c)
}

func (c *Client) ID() ConnID {
	return c.conn.ID()
}

func (c *Client) Send(ctx context.Context, msg Message, recv chan<- Result) (rerr error) {
	return c.conn.Send(ctx, msg, recv)
}

func (c *Client) Request(ctx context.Context, msg Message) (resp Message, rerr error) {
	return c.conn.Request(ctx, msg)
}

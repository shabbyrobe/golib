package socketsrv

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	service "github.com/shabbyrobe/go-service"
	"github.com/shabbyrobe/golib/incrementer"
)

// TODO(bw): Remove go-service from this package entirely.
var serverRunner = service.NewRunner()

type (
	ServerOption       func(srv *Server)
	OnServerConnect    func(server *Server, id ConnID)
	OnServerDisconnect func(server *Server, id ConnID, err error)
)

// ServerConnect is a ServerOption that registers a callback that happens when
// a client connects to the server.
func ServerConnect(cb OnServerConnect) ServerOption {
	return func(srv *Server) { srv.onConnect = cb }
}

// ServerDisconnect is a ServerOption that registers a callback that happens when
// a client disconnects from the server.
func ServerDisconnect(cb OnServerDisconnect) ServerOption {
	return func(srv *Server) { srv.onDisconnect = cb }
}

type Server struct {
	config       ServerConfig
	runner       service.Runner
	listener     Listener
	handler      Handler
	negotiator   Negotiator
	onConnect    OnServerConnect
	onDisconnect OnServerDisconnect

	nextID   incrementer.Inc
	nextIDMu sync.Mutex

	conns   map[ConnID]*conn
	connsMu sync.Mutex
	running uint32
}

// BUG(bw): listener will be moved from NewServer to Serve() once the
// dependency on go-service is removed.
func NewServer(config *ServerConfig, listener Listener, negotiator Negotiator, handler Handler, opts ...ServerOption) *Server {
	if listener == nil {
		panic("socket: listener must not be nil")
	}
	if handler == nil {
		panic("socket: handler must not be nil")
	}
	if negotiator == nil {
		panic("socket: negotiator must not be nil")
	}
	if config == nil || config.IsZero() {
		config = DefaultServerConfig()
	}

	srv := &Server{
		config:     *config,
		listener:   listener,
		conns:      make(map[ConnID]*conn),
		handler:    handler,
		negotiator: negotiator,
	}
	for _, o := range opts {
		o(srv)
	}

	return srv
}

func (srv *Server) onEnd(id ConnID, err error) {
	srv.connsMu.Lock()
	delete(srv.conns, id)
	srv.connsMu.Unlock()
	if srv.onDisconnect != nil {
		srv.onDisconnect(srv, id, err)
	}
}

func (srv *Server) MustServe() {
	if err := srv.Serve(); err != nil {
		panic(err)
	}
}

// Serve is a shorthand for starting the Server as a service using
// github.com/shabbyrobe/go-service.
func (srv *Server) Serve() error {
	ender := service.NewEndListener(1)
	svc := service.New(service.Name("server"), srv).WithEndListener(ender)

	// Hard-coded timeout shouldn't be an issue here.
	if err := service.StartTimeout(10*time.Second, serverRunner, svc); err != nil {
		return err
	}
	return <-ender.Ends()
}

// Run implements the Runnable interface in github.com/shabbyrobe/go-service.
//
// See Serve() for an example of how to start a service in a runner.
func (srv *Server) Run(ctx service.Context) (rerr error) {
	if !atomic.CompareAndSwapUint32(&srv.running, 0, 1) {
		return fmt.Errorf("socket: server already running")
	}
	defer func() {
		atomic.StoreUint32(&srv.running, 0)
	}()

	defer func() {
		if err := srv.runner.Shutdown(context.Background()); err != nil {
			rerr = err
		}
		if cerr := srv.listener.Close(); cerr != nil && rerr == nil {
			rerr = cerr
		}
	}()

	if err := ctx.Ready(); err != nil {
		return err
	}

	errc := make(chan error, 1)
	go func() {
		for {
			raw, err := srv.listener.Accept()
			if err != nil {
				ctx.OnError(err)
				continue
			}

			// Even though we are using go-service for backgrounding, this
			// still needs to be a goroutine. The connection is not considered
			// "started" until version negotiation is complete, but version
			// negotiation may hit the network.
			go func() {
				srv.nextIDMu.Lock()
				id := ConnID(srv.nextID.Next())
				srv.nextIDMu.Unlock()

				conn := newConn(id, ServerSide, srv.config.Conn, raw, srv.negotiator, srv.handler)

				// we must start the service and raise the onConnected event
				// while the lock is acquired otherwise the "onDisconnect"
				// callback can be called before the "onConnect" callback.
				srv.connsMu.Lock()

				ended := make(chan error, 1)
				if err := conn.start(ended); err != nil {
					ctx.OnError(err)
					return
				}
				go func() {
					srv.onEnd(id, <-ended)
				}()

				if srv.onConnect != nil {
					srv.onConnect(srv, id)
				}
				srv.conns[id] = conn
				srv.connsMu.Unlock()
			}()
		}
	}()

	{
		var err error
		select {
		case <-ctx.Done():
		case err = <-errc:
		}
		return err
	}
}

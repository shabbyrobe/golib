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
	nextID       incrementer.Inc
	runner       service.Runner
	listener     Listener
	handler      Handler
	negotiator   Negotiator
	onConnect    OnServerConnect
	onDisconnect OnServerDisconnect

	conns   map[ConnID]*conn
	connsMu sync.Mutex
	running uint32
}

func NewServer(config ServerConfig, listener Listener, negotiator Negotiator, handler Handler, opts ...ServerOption) *Server {
	if listener == nil {
		panic("socket: listener must not be nil")
	}
	if handler == nil {
		panic("socket: handler must not be nil")
	}
	if negotiator == nil {
		panic("socket: negotiator must not be nil")
	}
	if config.IsZero() {
		config = DefaultServerConfig()
	}

	srv := &Server{
		config:     config,
		listener:   listener,
		conns:      make(map[ConnID]*conn),
		handler:    handler,
		negotiator: negotiator,
	}
	for _, o := range opts {
		o(srv)
	}

	srv.runner = service.NewRunner(service.RunnerOnEnd(srv.onEnd))
	return srv
}

func (srv *Server) onEnd(stage service.Stage, svc *service.Service, err error) {
	// FIXME: There are deadlock issues with connsMu. onEnd can be called
	// before runner.Start() has yielded, which would mean that connsMu is not
	// unlocked before we attempt to acquire it here. The goroutine is a bit
	// of a cheat; go-service probably needs to be less deadlock-prone for this
	// use case.
	go func() {
		id := ConnID(svc.Name)
		srv.connsMu.Lock()
		delete(srv.conns, id)
		srv.connsMu.Unlock()
		if srv.onDisconnect != nil {
			srv.onDisconnect(srv, id, err)
		}
	}()
}

func (srv *Server) MustServe(host string) {
	if err := srv.Serve(host); err != nil {
		panic(err)
	}
}

// Serve is a shorthand for starting the Server as a service using
// github.com/shabbyrobe/go-service.
func (srv *Server) Serve(host string) error {
	ender := service.NewEndListener(1)
	svc := service.New(service.Name(host), srv).WithEndListener(ender)

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
				id := ConnID(srv.nextID.Next())
				conn := newConn(id, ServerSide, srv.config.Conn, raw, srv.negotiator, srv.handler)

				// we must start the service and raise the onConnected event
				// while the lock is acquired otherwise the "onDisconnect"
				// callback can be called before the "onConnect" callback.
				srv.connsMu.Lock()
				if err := srv.runner.Start(ctx, service.New(service.Name(id), conn)); err != nil {
					srv.connsMu.Unlock()

					_ = raw.Close()
					ctx.OnError(err)
					return
				}
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

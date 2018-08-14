package socketsrv

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	service "github.com/shabbyrobe/go-service"
	"github.com/shabbyrobe/golib/incrementer"
)

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

type (
	ServerOption       func(srv *Server)
	OnServerConnect    func(server *Server, id ConnID)
	OnServerDisconnect func(server *Server, id ConnID, err error)
)

func ServerConnect(cb OnServerConnect) ServerOption {
	return func(srv *Server) { srv.onConnect = cb }
}

func ServerDisconnect(cb OnServerDisconnect) ServerOption {
	return func(srv *Server) { srv.onDisconnect = cb }
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

			} else {
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

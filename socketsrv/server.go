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
	config     ServerConfig
	nextID     incrementer.Inc
	runner     service.Runner
	listener   Listener
	handler    Handler
	negotiator Negotiator

	conns   map[ConnID]*conn
	connsMu sync.Mutex
	running uint32
}

func NewServer(config ServerConfig, listener Listener, negotiator Negotiator, handler Handler) *Server {
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
	srv.runner = service.NewRunner(service.RunnerOnEnd(srv.onEnd))
	return srv
}

func (srv *Server) onEnd(stage service.Stage, svc *service.Service, err error) {
	id := ConnID(svc.Name)
	srv.connsMu.Lock()
	delete(srv.conns, id)
	srv.connsMu.Unlock()

	// FIXME:
	if err != nil {
		fmt.Println(err)
	}
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
				id := ConnID(srv.nextID.Next())

				conn := newConn(id, ServerSide, srv.config.Conn, raw, srv.negotiator, srv.handler)
				if err := srv.runner.Start(ctx, service.New(service.Name(id), conn)); err != nil {
					_ = raw.Close()
					ctx.OnError(err)
					continue
				}

				srv.connsMu.Lock()
				srv.conns[id] = conn
				srv.connsMu.Unlock()
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

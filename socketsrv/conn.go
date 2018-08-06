package socketsrv

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	service "github.com/shabbyrobe/go-service"
)

type Side int

const (
	ClientSide Side = 1
	ServerSide Side = 1
)

type Envelope struct {
	ID      MessageID
	ReplyTo MessageID
	Kind    int
	Message Message
}

type MessageID uint32

type Message interface{}

type ConnID string

type Result struct {
	Message Message
	Err     error
}

type call struct {
	rs  chan<- Result
	env Envelope
	at  time.Time
}

const (
	connNew      uint32 = 0
	connRunning  uint32 = 1
	connComplete uint32 = 2
)

type conn struct {
	id      ConnID
	comm    Communicator
	side    Side
	proto   Protocol
	handler Handler
	config  ConnConfig

	state         uint32
	nextMessageID uint32
	lastRecv      time.Time
	lastSend      time.Time
	calls         chan call
	messageLimit  uint32

	// number of currently active calls to Send(). this is used during shutdown
	// so we know that all calls to Send have responded to the stop signal so we
	// can drain the calls channel.
	sendWait sync.WaitGroup

	// writer thread's queue
	outgoing chan Envelope

	stop chan struct{}
}

func newConn(id ConnID, side Side, config ConnConfig, comm Communicator, proto Protocol, handler Handler) *conn {
	if config.IsZero() {
		config = DefaultConnConfig()
	}

	return &conn{
		id:           id,
		side:         side,
		config:       config,
		comm:         comm,
		proto:        proto,
		handler:      handler,
		messageLimit: proto.MessageLimit(),

		nextMessageID: 1,
		outgoing:      make(chan Envelope, config.OutgoingBuffer),
		calls:         make(chan call, config.OutgoingBuffer),
		stop:          make(chan struct{}),
	}
}

func (c *conn) ID() ConnID { return c.id }

func (c *conn) Run(ctx service.Context) (rerr error) {
	if !atomic.CompareAndSwapUint32(&c.state, connNew, connRunning) {
		return fmt.Errorf("socket: cannot re-use Conn")
	}

	failer := service.NewFailureListener(1)
	incoming := make(chan Envelope, c.config.IncomingBuffer)

	// Reader thread:
	go func() {
		rdBuf := make([]byte, c.config.ReadBufferInitial)

		var err error
		for {
			rdBuf, err = c.comm.ReadMessage(rdBuf, c.messageLimit, c.config.ReadTimeout)
			if err != nil {
				failer.Send(err)
				return
			}
			if len(rdBuf) == 0 {
				continue // heartbeats can be empty
			}

			env, err := c.proto.Decode(rdBuf)
			if err != nil {
				failer.Send(err)
				return
			}
			c.lastRecv = time.Now()

			select {
			case incoming <- env:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Writer thread:
	go func() {
		wrBuf := make([]byte, 0, c.config.WriteBufferInitial)

		heartbeat := time.NewTicker(100 * time.Millisecond)
		defer heartbeat.Stop()

		for {
			select {
			case <-heartbeat.C:
				if err := c.comm.Ping(c.config.WriteTimeout); err != nil {
					failer.Send(err)
					return
				}

			case env := <-c.outgoing:
				var err error
				wrBuf, err = c.proto.Encode(env, wrBuf)
				if err != nil {
					failer.Send(err)
					return
				}
				if err := c.comm.WriteMessage(wrBuf, c.messageLimit, c.config.WriteTimeout); err != nil {
					failer.Send(err)
					return
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	calls := make(map[MessageID]call)

	// Shutdown procedure:
	defer func() {
		// This should prevent any further calls to Send() from succeeding
		atomic.StoreUint32(&c.state, connComplete)

		// Signal to Send() that it should unblock all attempts to send to the calls
		// channel
		close(c.stop)

		// Once we know that all calls to Send() have returned, we can safely drain
		// the channel.
		c.sendWait.Wait()

		if cerr := c.comm.Close(); cerr != nil && rerr == nil {
			rerr = cerr
		}

		// Drain the calls channel and add the items to the map for shutdown reporting:
		close(c.calls)
		for call := range c.calls {
			calls[call.env.ID] = call
		}

		// Report shutdown to all pending calls:
		rserr := rerr
		if rserr == nil {
			rserr = errors.New("socket: shutdown")
		}
		for _, call := range calls {
			select {
			case call.rs <- Result{Err: rserr}:
			default:
			}
		}
	}()

	// Configure cleanup channel:
	var cleanup <-chan time.Time
	if c.config.CleanupInterval > 0 {
		ct := time.NewTicker(c.config.CleanupInterval)
		defer ct.Stop()
		cleanup = ct.C
	}

	// Connection is ready to serve requests:
	if err := ctx.Ready(); err != nil {
		return err
	}

	// Connection reactor loop:
	for {
		select {

		// Process incoming message:
		case env := <-incoming:
			call, ok := calls[env.ID]
			if ok {
				delete(calls, env.ID)

				select {
				case call.rs <- Result{Message: env.Message}:
				default:
					return fmt.Errorf("call receiver would block")
				}

			} else {
				rs, err := c.handler.HandleIncoming(c.id, env.Message)
				if err != nil {
					return err
				}
				if rs != nil {
					kind, err := c.proto.MessageKind(rs)
					if err != nil {
						return err
					}
					select {
					case c.outgoing <- Envelope{ID: c.nextID(), ReplyTo: env.ID, Kind: kind, Message: rs}:
					case <-ctx.Done():
						return nil
					}
				}
			}

		// Process call:
		case out := <-c.calls:
			calls[out.env.ID] = out

			select {
			case c.outgoing <- out.env:
			case <-ctx.Done():
				return nil
			}

		// Cleanup:
		case at := <-cleanup:
			for id, call := range calls {
				if at.Sub(call.at) < c.config.ResponseTimeout {
					continue
				}
				delete(calls, id)

				select {
				case call.rs <- Result{Err: errors.New("socketsrv: response timeout")}:
				default:
					return fmt.Errorf("call receiver would block")
				}
			}

		case err := <-failer.Failures():
			return err

		case <-ctx.Done():
			return nil
		}
	}
}

func (c *conn) nextID() MessageID {
	return MessageID(atomic.AddUint32(&c.nextMessageID, 2))
}

func (c *conn) Send(ctx context.Context, msg Message, recv chan<- Result) (rerr error) {
	if atomic.LoadUint32(&c.state) != connRunning {
		return fmt.Errorf("socket: send to conn which is not running")
	}

	kind, err := c.proto.MessageKind(msg)
	if err != nil {
		return err
	}

	c.sendWait.Add(1)

	id := c.nextID()
	env := Envelope{
		ID:      id,
		Kind:    kind,
		Message: msg,
	}

	select {
	case c.calls <- call{env: env, rs: recv}:
		c.sendWait.Done()
		return nil

	case <-c.stop:
		c.sendWait.Done()
		return errors.New("socket: shutdown")

	case <-ctx.Done():
		c.sendWait.Done()
		return ctx.Err()
	}
}

func (c *conn) Request(ctx context.Context, msg Message) (resp Message, rerr error) {
	rc := make(chan Result, 1)
	if err := c.Send(ctx, msg, rc); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case result := <-rc:
		resp, rerr = result.Message, result.Err
		return resp, rerr
	}
}

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

type ConnID string

const (
	connNew      uint32 = 0
	connRunning  uint32 = 1
	connComplete uint32 = 2
)

type conn struct {
	id         ConnID
	comm       Communicator
	side       Side
	config     ConnConfig
	handler    Handler
	negotiator Negotiator

	state         uint32
	nextMessageID uint32
	lastRecv      time.Time
	lastSend      time.Time
	calls         chan call

	// number of currently active calls to Send(). this is used during shutdown
	// so we know that all calls to Send have responded to the stop signal so we
	// can drain the calls channel.
	sendWait sync.WaitGroup

	// writer thread's queue
	outgoing chan Envelope

	stop chan struct{}
}

func newConn(id ConnID, side Side, config ConnConfig, comm Communicator, neg Negotiator, handler Handler) *conn {
	if config.IsZero() {
		config = DefaultConnConfig()
	}

	return &conn{
		id:         id,
		side:       side,
		config:     config,
		comm:       comm,
		negotiator: neg,
		handler:    handler,

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

	proto, err := c.negotiator.Negotiate(c.side, c.comm)
	if err != nil {
		return err
	}
	if proto == nil {
		return fmt.Errorf("negotiator returned nil protocol")
	}

	mapper := proto.Mapper()
	if mapper == nil {
		panic(fmt.Errorf("proto %q returned nil mapper", proto.ProtocolName()))
	}

	messageLimit := proto.MessageLimit()

	// Reader thread:
	go func() {
		rdBuf := make([]byte, c.config.ReadBufferInitial)
		var encData ProtoData
		defer func() {
			if encData != nil {
				_ = encData.Close()
			}
		}()

		var err error
		for {
			rdBuf, err = c.comm.ReadMessage(rdBuf, messageLimit, c.config.ReadTimeout)
			if err != nil {
				failer.Send(err)
				return
			}
			if len(rdBuf) == 0 {
				continue // heartbeats can be empty
			}

			env, err := proto.Decode(rdBuf, &encData)
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
		var decData ProtoData
		defer func() {
			if decData != nil {
				_ = decData.Close()
			}
		}()

		var heartbeat <-chan time.Time
		if c.config.HeartbeatInterval > 0 {
			ht := time.NewTicker(c.config.HeartbeatInterval)
			defer ht.Stop()
			heartbeat = ht.C
		}

		for {
			select {
			case <-heartbeat:
				if err := c.comm.Ping(c.config.WriteTimeout); err != nil {
					failer.Send(err)
					return
				}

			case env := <-c.outgoing:
				var err error
				wrBuf, err = proto.Encode(env, wrBuf, &decData)
				if err != nil {
					failer.Send(err)
					return
				}
				mlen := uint32(len(wrBuf))
				if mlen > messageLimit {
					failer.Send(fmt.Errorf("conn: message of length %d exceeded limit %d", mlen, messageLimit))
					return
				}
				if err := c.comm.WriteMessage(wrBuf, c.config.WriteTimeout); err != nil {
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
			calls[call.id] = call
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
					kind, err := mapper.MessageKind(rs)
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
			calls[out.id] = out
			kind, err := mapper.MessageKind(out.msg)
			if err != nil {
				delete(calls, out.id)
				select {
				case out.rs <- Result{Err: err}:
				default:
					return fmt.Errorf("call receiver would block")
				}
			}

			select {
			case c.outgoing <- Envelope{ID: out.id, Message: out.msg, Kind: kind}:
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
	// This starts at 1 and increments by 2 to guarantee that we will never use
	// ID 0, which is reserved to mean "No message ID".
	return MessageID(atomic.AddUint32(&c.nextMessageID, 2))
}

func (c *conn) Send(ctx context.Context, msg Message, recv chan<- Result) (rerr error) {
	if atomic.LoadUint32(&c.state) != connRunning {
		return fmt.Errorf("socket: send to conn which is not running")
	}

	c.sendWait.Add(1)

	sendCall := call{
		id:  c.nextID(),
		at:  time.Now(),
		rs:  recv,
		msg: msg,
	}

	select {
	case c.calls <- sendCall:
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

type call struct {
	rs  chan<- Result
	id  MessageID
	msg Message
	at  time.Time
}

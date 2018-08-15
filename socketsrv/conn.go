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

func (c *conn) negotiate() (Protocol, error) {
	// Some communicators (like the websocket one) rely on the ping/pong
	// infrastructure to detect read timeouts and cannot use SetReadDeadline.
	// We need to establish our own timeout for the negotiation step because
	// the ping/pong hasn't started yet; it's outside the protocol.

	after := time.After(c.config.ReadTimeout)
	result := make(chan Protocol, 1)
	errc := make(chan error, 1)

	go func() {
		proto, err := c.negotiator.Negotiate(c.side, c.comm)
		if err != nil {
			errc <- err
			return
		}
		if proto == nil {
			errc <- errors.New("socketsrv: negotiator returned nil protocol")
			return
		}
		result <- proto
	}()

	select {
	case proto := <-result:
		return proto, nil
	case err := <-errc:
		return nil, err
	case <-after:
		return nil, errReadTimeout
	}
}

func (c *conn) Run(ctx service.Context) (rerr error) {
	if !atomic.CompareAndSwapUint32(&c.state, connNew, connRunning) {
		return errors.New("socketsrv: cannot re-use Conn")
	}

	failer := service.NewFailureListener(1)
	incoming := make(chan Envelope, c.config.IncomingBuffer)

	// Negotiate may access the network. The connection is not ready until
	// negotiation has succeeded:
	proto, err := c.negotiate()
	if err != nil {
		return err
	}

	mapper := proto.Mapper()
	if mapper == nil {
		return fmt.Errorf("socketsrv: proto %q returned nil mapper", proto.ProtocolName())
	}

	// The Communicator's view of the maximum message size is a harder limit than
	// the protocol's. If it is non-zero, it takes precedence.
	messageLimit := proto.MessageLimit()
	commMessageLimit := c.comm.MessageLimit()
	if commMessageLimit > 0 && commMessageLimit < messageLimit {
		messageLimit = commMessageLimit
	}

	// Reader thread:
	go func() {
		rdBuf := make([]byte, c.config.ReadBufferInitial)

		// encData contains reader-local shared memory that the Protocol may use
		// to store connection-scoped arbitrary data:
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

		// decData contains writer-local shared memory that the Protocol may use
		// to store connection-scoped arbitrary data:
		var decData ProtoData

		defer func() {
			if decData != nil {
				_ = decData.Close()
			}
		}()

		var heartbeat <-chan time.Time
		if c.config.HeartbeatSendInterval > 0 {
			ht := time.NewTicker(c.config.HeartbeatSendInterval)
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
				mlen := len(wrBuf)
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

	// calls contains all requests that are currently awaiting a response.
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
			rserr = errConnShutdown
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

	// Heartbeat setup. Pong channel may be nil:
	lastRecv := time.Now()
	pongs := c.comm.Pongs()
	var heartbeatCheck <-chan time.Time
	if c.config.HeartbeatCheckInterval > 0 {
		ct := time.NewTicker(c.config.HeartbeatCheckInterval)
		defer ct.Stop()
		heartbeatCheck = ct.C
	}

	// Connection is ready to serve requests:
	if err := ctx.Ready(); err != nil {
		return err
	}

	// Connection reactor loop:
	for {
		select {
		case <-pongs:
			lastRecv = time.Now()

		// Connection processes incoming message:
		case env := <-incoming:
			lastRecv = time.Now()

			call, ok := calls[env.ID]
			if ok {
				// Incoming message is a response to an existing call:
				delete(calls, env.ID)
				select {
				case call.rs <- Result{Message: env.Message}:
				default:
					return fmt.Errorf("call receiver would block")
				}

			} else {
				// Incoming message is not a response, it's either a remote
				// originated request or a late response to a local request so
				// it should be handled by a Handler:
				irq := IncomingRequest{
					conn:      c,
					ConnID:    c.id,
					MessageID: env.ID,
					Message:   env.Message,
					Deadline:  lastRecv.Add(c.config.ResponseTimeout),
				}
				rs, err := c.handler.HandleRequest(ctx, irq)
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

		// Connection handles call (conn.Request, conn.Send, conn.Reply):
		case out := <-c.calls:
			if out.env.ReplyTo == MessageNone {
				calls[out.env.ID] = out
			}

			// If the call itself carries an error, this has come from a Handler. These
			// errors need to terminate the connection and be passed through to the
			// disconnection handler as there is no way to handle errors of this kind
			// inside a handler.
			if out.err != nil {
				return out.err
			}

			var err error
			out.env.Kind, err = mapper.MessageKind(out.env.Message)
			if err != nil {
				delete(calls, out.env.ID)
				select {
				case out.rs <- Result{Err: err}:
				default:
					return fmt.Errorf("call receiver would block")
				}
			}

			select {
			case c.outgoing <- out.env:
			case <-ctx.Done():
				return nil
			}

		// Connection check heartbeat:
		case at := <-heartbeatCheck:
			if at.Sub(lastRecv) > c.config.ReadTimeout {
				return errReadTimeout
			}

		// Connection cleans up expired calls:
		case at := <-cleanup:
			for id, call := range calls {
				if call.deadline.Sub(at) >= 0 {
					continue
				}
				delete(calls, id)

				select {
				case call.rs <- Result{Err: errResponseTimeout}:
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

func (c *conn) send(ctx context.Context, sendCall call) (rerr error) {
	if c.config.ResponseTimeout > 0 {
		sendCall.deadline = time.Now().Add(c.config.ResponseTimeout)
	}

	c.sendWait.Add(1)
	select {
	case c.calls <- sendCall:
		c.sendWait.Done()
		return nil

	case <-c.stop:
		c.sendWait.Done()
		return errConnShutdown

	case <-ctx.Done():
		c.sendWait.Done()
		return ctx.Err()
	}
}

var resultPool = sync.Pool{
	New: func() interface{} {
		return make(chan Result, 1)
	},
}

func (c *conn) sendResult(ctx context.Context, sendCall call) (rs Message, rerr error) {
	recv := resultPool.Get().(chan Result)
	sendCall.rs = recv
	if err := c.send(ctx, sendCall); err != nil {
		return nil, rerr
	}

	// Don't wait on c.stop() in this select block. The stop channel may yield before the
	// result channel does even when a real result is available.
	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case result := <-recv:
		// We can only easily guarantee the channel is in a state fit to return
		// to the pool if we have received a value from it. It may be possible
		// in other branches but it's harder to verify.
		resultPool.Put(recv)
		return result.Message, result.Err
	}
}

func (c *conn) Send(ctx context.Context, msg Message, recv chan<- Result) (rerr error) {
	if atomic.LoadUint32(&c.state) != connRunning {
		return errConnSendNotRunning
	}
	sendCall := call{
		env: Envelope{
			ID:      c.nextID(),
			Message: msg,
		},
		rs: recv,
	}
	if err := c.send(ctx, sendCall); err != nil {
		return err
	}
	return nil
}

func (c *conn) Reply(ctx context.Context, to MessageID, msg Message, replyError error) (rerr error) {
	if atomic.LoadUint32(&c.state) != connRunning {
		return errConnSendNotRunning
	}
	sendCall := call{
		env: Envelope{
			ID:      c.nextID(),
			ReplyTo: to,
			Message: msg,
		},
		err: replyError,
	}
	_, rerr = c.sendResult(ctx, sendCall)
	return rerr
}

func (c *conn) Request(ctx context.Context, msg Message) (resp Message, rerr error) {
	if atomic.LoadUint32(&c.state) != connRunning {
		return nil, errConnSendNotRunning
	}
	sendCall := call{
		env: Envelope{
			ID:      c.nextID(),
			Message: msg,
		},
	}
	resp, rerr = c.sendResult(ctx, sendCall)
	return resp, rerr
}

type call struct {
	env      Envelope
	rs       chan<- Result
	err      error
	deadline time.Time
}

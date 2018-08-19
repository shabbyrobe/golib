package socketsrv

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Side int

const (
	ClientSide Side = 1
	ServerSide Side = 2
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
	calls         chan *call

	// number of currently active calls to Send(). this is used during shutdown
	// so we know that all calls to Send have responded to the stop signal so we
	// can drain the calls channel.
	sendWait sync.WaitGroup

	// reader thread's queue. must not be used outside readerThread or reactor.
	incoming chan *Envelope

	// writer thread's queue. may be used anywhere as long as the cconnection state
	// is running.
	outgoing chan *Envelope

	// stop is closed when the connection is instructed to stop:
	stop        chan struct{}
	stopCloseMu sync.Mutex
	stopClosed  bool

	// stopped is closed when the connection is fully stopped:
	stopped chan struct{}
}

func newConn(id ConnID, side Side, config ConnConfig, comm Communicator, neg Negotiator, handler Handler) *conn {
	if config.IsZero() {
		config = DefaultConnConfig()
	}

	nmID := uint32(1)

	// FIXME: this is a temporary cheat to try to make sure the IDs don't line up
	// exactly between the local and the remote by default. It helps to flush out
	// bugs with ID vs ReplyTo:
	if side == ClientSide {
		nmID = 10001
	}

	return &conn{
		id:         id,
		side:       side,
		config:     config,
		comm:       comm,
		negotiator: neg,
		handler:    handler,

		nextMessageID: nmID,
		incoming:      make(chan *Envelope, config.IncomingBuffer),
		outgoing:      make(chan *Envelope, config.OutgoingBuffer),
		calls:         make(chan *call, config.OutgoingBuffer),
		stop:          make(chan struct{}),
		stopped:       make(chan struct{}),
	}
}

func (c *conn) ID() ConnID { return c.id }

func (c *conn) Close() error {
	c.stopCloseMu.Lock()
	defer c.stopCloseMu.Unlock()
	if !c.stopClosed {
		close(c.stop)
		c.stopClosed = true

	} else {
		return fmt.Errorf("socketsrv: conn already closed")
	}

	return nil
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

func (c *conn) start(ended chan error) (rerr error) {
	if cap(ended) < 1 {
		panic("socketsrv: ended must have cap >= 1")
	}

	if !atomic.CompareAndSwapUint32(&c.state, connNew, connRunning) {
		return errors.New("socketsrv: cannot re-use Conn")
	}

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

	codec := proto.Codec()
	if codec == nil {
		return fmt.Errorf("socketsrv: proto %q returned nil codec", proto.ProtocolName())
	}

	// The Communicator's view of the maximum message size is a harder limit than
	// the protocol's. If it is non-zero, it takes precedence.
	messageLimit := proto.MessageLimit()
	commMessageLimit := c.comm.MessageLimit()
	if commMessageLimit > 0 && commMessageLimit < messageLimit {
		messageLimit = commMessageLimit
	}

	go func() {
		var wg sync.WaitGroup
		wg.Add(3)

		errc := make(chan error, 1)

		go func() {
			select {
			case errc <- c.writerThread(codec, messageLimit):
			default:
			}
			wg.Done()
		}()

		go func() {
			select {
			case errc <- c.readerThread(codec, messageLimit, mapper):
			default:
			}
			wg.Done()
		}()

		go func() {
			select {
			case errc <- c.reactor(proto, mapper, codec):
			default:
			}
			wg.Done()
		}()

		err := <-errc
		_ = c.Close()
		wg.Wait()

		select {
		case ended <- err:
		default:
			panic("socketsrv: ended would block!")
		}
	}()

	return nil
}

func (c *conn) negotiate() (Protocol, error) {
	// Some communicators (like the websocket one) rely on the ping/pong
	// facility of socketsrv.Conn to detect read timeouts and cannot use
	// SetReadDeadline. We need to establish our own timeout for the
	// negotiation step because the ping/pong hasn't started yet; it's outside
	// the protocol.

	after := time.After(c.config.ReadTimeout)
	result := make(chan Protocol, 1)
	errc := make(chan error, 1)

	go func() {
		proto, err := c.negotiator.Negotiate(c.side, c.comm, c.config)
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

func (c *conn) writerThread(
	codec Codec,
	messageLimit int,
) error {
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
				return err
			}

		case env := <-c.outgoing:
			var err error
			wrBuf, err = codec.Encode(*env, wrBuf, &decData)
			if err != nil {
				return err
			}
			mlen := len(wrBuf)
			if mlen > messageLimit {
				return fmt.Errorf("conn: message of length %d exceeded limit %d", mlen, messageLimit)
			}
			if err := c.comm.WriteMessage(wrBuf, c.config.WriteTimeout); err != nil {
				return err
			}

		case <-c.stop:
			return nil
		}
	}
}

func (c *conn) readerThread(
	codec Codec,
	messageLimit int,
	mapper Mapper,
) error {
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
			return err
		}

		var env *Envelope

		// heartbeats can be represented as empty buffers
		if len(rdBuf) != 0 {
			denv, err := codec.Decode(rdBuf, mapper, &encData)
			if err != nil {
				return err
			}
			env = &denv
		}

		select {
		case c.incoming <- env:
		case <-c.stop:
			return nil
		}
	}
}

// reactorShutdown is expected to be called in a defer from within run().
func (c *conn) reactorShutdown(calls map[MessageID]*call, rerr *error) {
	// This should prevent any further calls to Send(), Reply() or
	// Request() from succeeding, leaving only those currently in-flight:
	atomic.StoreUint32(&c.state, connComplete)

	// Close will signal to Send() via the 'stop' channel that it should
	// unblock all attempts to send to the calls channel (if this has not
	// already happened):
	_ = c.Close()

	// Once we know that all calls to Send() have returned, we can safely drain
	// the channel.
	c.sendWait.Wait()

	if cerr := c.comm.Close(); cerr != nil && *rerr == nil {
		*rerr = cerr
	}

	// Drain the calls channel and add the items to the map for shutdown reporting:
	close(c.calls)
	for call := range c.calls {
		calls[call.env.ID] = call
	}

	// Report shutdown to all pending calls:
	rserr := *rerr
	if rserr == nil {
		rserr = errConnShutdown
	}
	for _, call := range calls {
		select {
		case call.rs <- Result{ID: call.env.ID, Err: rserr}:
		default:
		}
	}

	// Now we can signal to all async handlers that the connection's Context
	// is complete. This will cause connShutdownContext to become Done():
	close(c.stopped)
}

func (c *conn) reactor(proto Protocol, mapper Mapper, codec Codec) (rerr error) {
	// calls contains all requests that are currently awaiting a response.
	calls := make(map[MessageID]*call)

	defer c.reactorShutdown(calls, &rerr)

	shutdownCtx := &connShutdownContext{done: c.stopped}

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

	// Connection reactor loop:
	for {
		select {
		case <-pongs:
			lastRecv = time.Now()

		// Connection processes incoming message:
		case env := <-c.incoming:
			lastRecv = time.Now()

			if env == nil {
				// It's a heartbeat. Do nothing, but we still need to send it to the reactor
				// so it can update the lastRecv time.

			} else if env.ReplyTo != 0 {
				// Incoming message is a response to an existing call:
				call, ok := calls[env.ReplyTo]
				if !ok {
					return fmt.Errorf("socketsrv: unexpected incoming %d in reply to message %d", env.ID, env.ReplyTo)
				}

				delete(calls, env.ReplyTo)
				select {
				case call.rs <- Result{ID: env.ID, Message: env.Message}:
				default:
					return fmt.Errorf("call receiver would block")
				}

			} else if c.handler != nil {
				// Incoming message is not a response so it should be handled
				// by a Handler:
				irq := IncomingRequest{
					conn:      c,
					ConnID:    c.id,
					MessageID: env.ID,
					Message:   env.Message,
					Deadline:  lastRecv.Add(c.config.ResponseTimeout),
				}

				rs, err := c.handler.HandleRequest(shutdownCtx, irq)
				if err != nil {
					return err
				}
				if rs != nil {
					kind, err := mapper.MessageKind(rs)
					if err != nil {
						return err
					}

					select {
					case c.outgoing <- &Envelope{ID: c.nextID(), ReplyTo: env.ID, Kind: kind, Message: rs}:
					case <-c.stop:
						return nil
					}
				}

			} else {
				// Incoming message is unhandled; terminate the connection:
				return fmt.Errorf("socketsrv: unexpected incoming message %T(%d)", env.Message, env.Kind)
			}

		// Connection handles call (conn.Request, conn.Send, conn.Reply):
		case out := <-c.calls:
			if out.env.ReplyTo == MessageNone {
				// Request or Send, so we need to retain the call so we can look it up when
				// we get the corresponding incoming message from the remote:
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
			}

			// If the call is a conn.Reply or if an error occurred, we need to unblock the caller.
			// If the call is a Send or a Request, we don't unblock until we receive an incoming
			// message:
			if err != nil || out.env.ReplyTo != MessageNone {
				select {
				case out.rs <- Result{ID: out.env.ID, Err: err}:
				default:
					return fmt.Errorf("call receiver would block")
				}
			}

			select {
			case c.outgoing <- &out.env:
			case <-c.stop:
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
				case call.rs <- Result{ID: call.env.ID, Err: errResponseTimeout}:
				default:
					return fmt.Errorf("call receiver would block")
				}
			}

		case <-c.stop:
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
	case c.calls <- &sendCall:
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

	// Don't wait on c.stop in this select block. The stop channel may yield
	// before the result channel does even when a real result is available.
	// The receiver channel will yield if the connection shuts down properly.
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

type call struct {
	env      Envelope
	rs       chan<- Result
	err      error
	deadline time.Time
}

// connShutdownContext is used when a context is needed that becomes Done()
// after the conn's Run method has completely shut down, rather than a
// context that becomes Done() when the shutdown process begins.
type connShutdownContext struct {
	done chan struct{}
}

func (csc *connShutdownContext) Deadline() (deadline time.Time, ok bool) {
	return
}

func (csc *connShutdownContext) Done() <-chan struct{} {
	return csc.done
}

func (csc *connShutdownContext) Err() error {
	// FIXME: maybe should return an error when done is closed:
	return nil
}

func (csc *connShutdownContext) Value(key interface{}) interface{} {
	return nil
}

package socketsrv

import (
	"context"
	"fmt"
	"io"
	"sync/atomic"
	"time"
)

type Protocol interface {
	MessageLimit() uint32
	ProtocolName() string

	// Mapper is cached by Conn on creation.
	Mapper() Mapper

	Decode(in []byte, decdata *ProtoData) (Envelope, error)
	Encode(env Envelope, into []byte, encdata *ProtoData) (extended []byte, rerr error)
}

type Mapper interface {
	Message(kind int) (Message, error)
	MessageKind(msg Message) (int, error)
}

type Negotiator interface {
	Negotiate(Side, Communicator) (Protocol, error)
}

// ProtoData is used as a type-unsafe way for a Protocol to store shared memory
// against a Conn object for reuse.
type ProtoData interface {
	io.Closer
}

type IncomingRequest struct {
	conn      *conn
	ConnID    ConnID
	MessageID MessageID
	Message   Message
	Deadline  time.Time
	done      uint32
}

// Done marks the incoming request as complete, writing the response to the
// socket if one is provided or reporting the error to your application through
// the Conn if one is provided instead. If both a response and an error are
// provided, the error takes precedence.
//
// Done Allows you to use the IncomingRequest as a handle to defer responding
// until outside the Handler. This allows a non-blocking response pattern. For
// a blocking response, use the 'rs Message' return value of
// Handler.HandleRequest instead.
//
// Calling Done more than once is an error.
//
func (i IncomingRequest) Done(ctx context.Context, rs Message, replyError error) error {
	if !atomic.CompareAndSwapUint32(&i.done, 0, 1) {
		panic(fmt.Errorf("socketsrv: request done"))
	}
	return i.conn.Reply(ctx, i.MessageID, rs, replyError)
}

type Handler interface {
	// HandleRequest is called when a Request is received from the remote end
	// of a connection
	//
	// HandleRequest blocks the connection's reactor loop until it returns,
	// so keep any processing inside HandleRequest to an absolute minimum.
	//
	// If you wish to defer processing the response (for example, in a queue),
	// the IncomingRequest may be retained, but this will have interactions
	// with ConnConfig.ResponseTimeout. The request will time out after
	// IncomingRequest.Deadline if Deadline is not the zero time.
	//
	// If the connection is lost, any retained IncomingRequest will become
	// invalid and attempts to send on it will fail.
	//
	HandleRequest(context.Context, IncomingRequest) (rs Message, rerr error)
}

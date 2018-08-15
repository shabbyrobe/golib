package socketsrv

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

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
// Hopefully temporary caveat:
//
// Avoid using the Context that HandleRequest receives for the first argument;
// the context will be cancelled as soon as the connection is lost. This leads
// to the following race:
//
//  - conn.Reply() is called in thread A
//  - conn.Reply() tells the connection's reactor to write the response
//  - connection reactor writes reply to remote in thread B
//  - connection reactor sends result back to conn.Reply() through a channel
//  - remote closes, terminating connection
//  - both ctx.Done() and the connection reactor's reply will be available to
//    the select block in conn.Reply(), which gives you a 50/50 chance of the
//    correct 'nil error' being returned or a context.Canceled being returned.
//
// If this is not a concern, you can use the context from the Handler.
//
func (i IncomingRequest) Done(ctx context.Context, rs Message, replyError error) error {
	if !atomic.CompareAndSwapUint32(&i.done, 0, 1) {
		return fmt.Errorf("socketsrv: request done")
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
	// the IncomingRequest may be retained. You should be careful not to retain
	// forever as this will prevent the underlying connection resource from
	// being garbage collected. This will also have interactions with
	// ConnConfig.ResponseTimeout. The request will time out after
	// IncomingRequest.Deadline if Deadline is not the zero time.
	//
	// If the connection is lost, any retained IncomingRequest will become
	// invalid and attempts to send on it will fail.
	//
	HandleRequest(context.Context, IncomingRequest) (rs Message, rerr error)
}

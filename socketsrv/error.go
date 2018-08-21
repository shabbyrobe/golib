package socketsrv

import (
	"errors"
	"net"
)

var (
	errNoProtocol = errors.New("no protocol negotiated")

	errResponseTimeout error = &errTimeout{"socketsrv: response timeout"}
	errReadTimeout     error = &errTimeout{"socketsrv: read timeout"}

	errConnShutdown   error = &errTemporary{"socketsrv: conn shutdown"}
	errAlreadyClosed  error = &errTemporary{"socketsrv: already closed"}
	errNotRunning     error = &errTemporary{"socketsrv: not running"}
	errUnavailable    error = &errTemporary{"socketsrv: resource temporarily unavailable"}
	errAlreadyRunning error = &errTemporary{"socketsrv: already running"}

	// This is an error rather than a panic because callers of Send() are free
	// to pass their own receiver channel in. If this receiver blocks, the
	// connection is terminated.
	errReceiverBlocked error = &errTemporary{"socketsrv: call receiver would block"}
)

type ConnError struct {
	ID ConnID

	// Op is the operation which caused the error, such as
	// "read" or "write".
	Op ConnOp

	// Err is the error that occurred during the operation.
	Err error
}

var _ net.Error = &ConnError{}

func (e *ConnError) Error() string {
	if e == nil {
		return "<nil>"
	}
	s := "socketsrv: " + string(e.Op) + " - " + e.Err.Error()
	return s
}

func (e *ConnError) Cause() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (e *ConnError) Temporary() bool {
	if e.Err == errReadTimeout ||
		e.Err == errResponseTimeout ||
		e.Err == errConnShutdown {
		return true
	}

	if ne, ok := e.Err.(interface{ Temporary() bool }); ok {
		return ne.Temporary()
	}

	return false
}

func (e *ConnError) Timeout() bool {
	if e == nil {
		return false
	}
	return errIsTimeout(e.Err)
}

type protocolError struct {
	ID  ConnID
	Err error
}

func (e *protocolError) Temporary() bool { return true }

func (e *protocolError) Timeout() bool {
	if e == nil {
		return false
	}
	return errIsTimeout(e.Err)
}

func (e *protocolError) Error() string {
	if e == nil {
		return "<nil>"
	}
	s := "socketsrv: protocol error: " + e.Err.Error()
	return s
}

func (e *protocolError) Cause() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func errIsTimeout(err error) bool {
	if err == nil {
		return false
	}
	if err == errReadTimeout {
		return true
	}
	if ne, ok := err.(interface{ Timeout() bool }); ok {
		return ne.Timeout()
	}
	return false
}

type errPermanent struct {
	msg string
}

func (e *errPermanent) Error() string { return e.msg }

type errTemporary struct {
	msg string
}

func (e *errTemporary) Error() string   { return e.msg }
func (e *errTemporary) Temporary() bool { return true }

type errTimeout struct {
	msg string
}

func (e *errTimeout) Error() string   { return e.msg }
func (e *errTimeout) Temporary() bool { return true }
func (e *errTimeout) Timeout() bool   { return true }

type ConnOp string

const (
	OpRead      ConnOp = "read"
	OpWrite     ConnOp = "write"
	OpClose     ConnOp = "close"
	OpNegotiate ConnOp = "negotiate"
	OpPing      ConnOp = "ping"
	OpConnect   ConnOp = "connect"
	OpHandle    ConnOp = "handle"
	OpSend      ConnOp = "send"
	OpReply     ConnOp = "reply"
	OpRequest   ConnOp = "request"
)

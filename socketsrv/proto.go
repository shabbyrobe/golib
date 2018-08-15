package socketsrv

import (
	"io"
)

type Mapper interface {
	Message(kind int) (Message, error)
	MessageKind(msg Message) (int, error)
}

type Negotiator interface {
	Negotiate(Side, Communicator) (Protocol, error)
}

type Protocol interface {
	MessageLimit() int
	ProtocolName() string

	// Mapper is cached by Conn on creation.
	Mapper() Mapper

	Decode(in []byte, decdata *ProtoData) (Envelope, error)
	Encode(env Envelope, into []byte, encdata *ProtoData) (extended []byte, rerr error)
}

// ProtoData is used as a type-unsafe way for a Protocol to store shared memory
// against a Conn object for reuse.
type ProtoData interface {
	io.Closer
}

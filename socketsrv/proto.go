package socketsrv

import "io"

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

type Handler interface {
	HandleIncoming(id ConnID, msg Message) (rs Message, rerr error)
}

type Negotiator interface {
	Negotiate(Side, Communicator) (Protocol, error)
}

// ProtoData is used as a type-unsafe way for a Protocol to store shared memory
// against a Conn object for reuse.
type ProtoData interface {
	io.Closer
}
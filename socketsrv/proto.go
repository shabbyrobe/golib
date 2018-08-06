package socketsrv

type Protocol interface {
	Version() int
	MessageLimit() uint32
	Message(kind int) (Message, error)
	MessageKind(msg Message) (int, error)
	Decode(in []byte) (Envelope, error)
	Encode(env Envelope, into []byte) (extended []byte, rerr error)
}

type Handler interface {
	HandleIncoming(id ConnID, msg Message) (rs Message, rerr error)
}

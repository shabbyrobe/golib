package socketsrv

type Envelope struct {
	ID      MessageID
	ReplyTo MessageID
	Kind    int
	Message Message
}

type MessageID uint32

type Message interface{}

type Result struct {
	Message Message
	Err     error
}

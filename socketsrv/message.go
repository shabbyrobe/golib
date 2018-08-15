package socketsrv

type Envelope struct {
	ID      MessageID
	ReplyTo MessageID
	Kind    int
	Message Message
}

const MessageNone MessageID = 0

type MessageID uint32

func (m MessageID) Empty() bool { return m == MessageNone }

type Message interface{}

type Result struct {
	ID      MessageID
	Message Message
	Err     error
}

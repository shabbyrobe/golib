package socketsrv

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

type Negotiator interface {
	Negotiate(Side, Communicator) (Protocol, error)
}

type Mapper interface {
	Message(kind int) (Message, error)
	MessageKind(msg Message) (int, error)
}

type Protocol interface {
	Mapper() Mapper

	MessageLimit() int
	ProtocolName() string

	// Codec is cached by Conn
	Codec() Codec
}

type Codec interface {
	Decode(in []byte, mapper Mapper, decdata *ProtoData) (Envelope, error)
	Encode(env Envelope, into []byte, encdata *ProtoData) (extended []byte, rerr error)
}

// ProtoData is used as a type-unsafe way for a Protocol to store shared memory
// against a Conn object for reuse.
type ProtoData interface {
	io.Closer
}

type VersionedProtocol interface {
	Protocol
	Version() int
}

type VersionNegotiator struct {
	protocols map[uint32]VersionedProtocol
	timeout   time.Duration
	encoding  binary.ByteOrder
	ours      []byte
}

func NewVersionNegotiator(timeout time.Duration, protos ...VersionedProtocol) *VersionNegotiator {
	if len(protos) == 0 {
		panic("socketsrv: no procols specified")
	}
	vn := &VersionNegotiator{
		protocols: make(map[uint32]VersionedProtocol),
		encoding:  binary.BigEndian,
		timeout:   timeout,
	}

	var ours = make([]byte, len(protos)*4)
	for i, p := range protos {
		if p.Version() < 0 || int64(p.Version()) > int64(1<<32-1) {
			panic("socketsrv: proto version must fit inside uint32")
		}
		pv := uint32(p.Version())
		vn.encoding.PutUint32(ours[i*4:], pv)
		vn.protocols[pv] = p
	}
	vn.ours = ours

	return vn
}

func (v *VersionNegotiator) Negotiate(side Side, c Communicator) (Protocol, error) {
	timeout := v.timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	if err := c.WriteMessage(v.ours, timeout); err != nil {
		return nil, err
	}

	msg, err := c.ReadMessage(nil, 1024, timeout)
	if err != nil {
		return nil, err
	}
	if len(msg)%4 != 0 {
		return nil, fmt.Errorf("unexpected remote versions")
	}

	var max uint32
	var found bool
	for i := 0; i < len(msg); i += 4 {
		cur := v.encoding.Uint32(msg[i:])
		if _, ok := v.protocols[cur]; ok {
			found = true
			if cur > max {
				max = cur
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("could not negotiate protocol")
	}

	return v.protocols[max], nil
}

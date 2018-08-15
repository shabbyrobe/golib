package socketsrv

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
)

type JSONProtocol struct {
	name     string
	mapper   Mapper
	limit    int
	Encoding binary.ByteOrder
}

func NewJSONProtocol(name string, mapper Mapper, limit int) *JSONProtocol {
	if limit <= 0 {
		limit = 1 << 18
	}
	return &JSONProtocol{
		name:     name,
		mapper:   mapper,
		limit:    limit,
		Encoding: binary.BigEndian,
	}
}

func (p JSONProtocol) ProtocolName() string { return p.name }
func (p JSONProtocol) Mapper() Mapper       { return p.mapper }
func (p JSONProtocol) MessageLimit() int    { return p.limit }

func (p JSONProtocol) Decode(in []byte, decdata *ProtoData) (env Envelope, err error) {
	if len(in) < 12 {
		return env, fmt.Errorf("socketsrv: short message")
	}

	env.ID = MessageID(p.Encoding.Uint32(in))
	env.ReplyTo = MessageID(p.Encoding.Uint32(in[4:]))
	env.Kind = int(p.Encoding.Uint32(in[8:]))
	env.Message, err = p.mapper.Message(env.Kind)
	if err != nil {
		return env, err
	}

	if err := json.Unmarshal(in[12:], &env.Message); err != nil {
		return env, err
	}

	return env, nil
}

func (p JSONProtocol) Encode(env Envelope, into []byte, encData *ProtoData) (extended []byte, rerr error) {
	var hdr [12]byte
	p.Encoding.PutUint32(hdr[0:], uint32(env.ID))
	p.Encoding.PutUint32(hdr[4:], uint32(env.ReplyTo))
	p.Encoding.PutUint32(hdr[8:], uint32(env.Kind))

	buf := bytes.NewBuffer(into[:0])
	buf.Write(hdr[:])

	enc := json.NewEncoder(buf)
	if err := enc.Encode(env.Message); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

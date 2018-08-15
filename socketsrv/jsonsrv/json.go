package jsonsrv

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/shabbyrobe/golib/socketsrv"
)

type Codec struct {
	encoding binary.ByteOrder
}

var _ socketsrv.Codec = &Codec{}

type Option func(jp *Codec)

func Encoding(bo binary.ByteOrder) Option {
	return func(jp *Codec) { jp.encoding = bo }
}

func NewCodec(opts ...Option) *Codec {
	jp := &Codec{}
	for _, o := range opts {
		o(jp)
	}
	if jp.encoding == nil {
		jp.encoding = binary.BigEndian
	}
	return jp
}

func (p *Codec) Decode(in []byte, mapper socketsrv.Mapper, decdata *socketsrv.ProtoData) (env socketsrv.Envelope, err error) {
	if len(in) < 12 {
		return env, fmt.Errorf("socketsrv: short message")
	}

	env.ID = socketsrv.MessageID(p.encoding.Uint32(in))
	env.ReplyTo = socketsrv.MessageID(p.encoding.Uint32(in[4:]))
	env.Kind = int(p.encoding.Uint32(in[8:]))
	env.Message, err = mapper.Message(env.Kind)
	if err != nil {
		return env, err
	}

	if err := json.Unmarshal(in[12:], &env.Message); err != nil {
		return env, err
	}

	return env, nil
}

func (p *Codec) Encode(env socketsrv.Envelope, into []byte, encData *socketsrv.ProtoData) (extended []byte, rerr error) {
	var hdr [12]byte
	p.encoding.PutUint32(hdr[0:], uint32(env.ID))
	p.encoding.PutUint32(hdr[4:], uint32(env.ReplyTo))
	p.encoding.PutUint32(hdr[8:], uint32(env.Kind))

	buf := bytes.NewBuffer(into[:0])
	buf.Write(hdr[:])

	enc := json.NewEncoder(buf)
	if err := enc.Encode(env.Message); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

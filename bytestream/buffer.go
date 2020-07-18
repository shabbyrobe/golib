package bytestream

import "io"

type Buffer struct {
	buf []byte
	rem []byte
	err error
}

var _ ByteStream = &Buffer{}

func NewBuffer(buf []byte) *Buffer {
	bs := &Buffer{}
	bs.Reset(buf)
	return bs
}

func (b *Buffer) Reset(buf []byte) {
	b.buf = buf
	b.rem = buf
	b.err = nil
}

func (b *Buffer) Err() error {
	return b.err
}

func (b *Buffer) Tell() int64 {
	return int64(len(b.buf) - len(b.rem))
}

func (b *Buffer) DiscardExactly(n int) (err error) {
	if n > len(b.rem) {
		return io.ErrUnexpectedEOF
	}
	b.rem = b.rem[n:]
	return
}

func (b *Buffer) DiscardUpTo(n int) error {
	if n > len(b.rem) {
		b.rem = nil
	} else {
		b.rem = b.rem[n:]
	}
	return nil
}

func (b *Buffer) PeekExactly(n int) (o []byte, err error) {
	if n > len(b.rem) {
		return nil, io.ErrShortBuffer
	}
	o = b.rem[:n]
	return
}

func (b *Buffer) PeekUpTo(n int) (o []byte, err error) {
	if n > len(b.rem) {
		return b.rem, nil
	}
	return b.rem[:n], nil
}

func (b *Buffer) TakeExactly(n int) (o []byte, err error) {
	if n > len(b.rem) {
		return nil, io.ErrUnexpectedEOF
	}
	o, b.rem = b.rem[:n], b.rem[n:]
	return o, nil
}

func (b *Buffer) TakeUpTo(n int) (o []byte, err error) {
	if n > len(b.rem) {
		return b.rem, nil
	}
	o, b.rem = b.rem[:n], b.rem[n:]
	return
}

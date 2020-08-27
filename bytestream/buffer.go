package bytestream

import (
	"errors"
	"io"
)

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

func (b *Buffer) Limit() int64 {
	return int64(len(b.buf))
}

func (b *Buffer) Tell() int64 {
	return int64(len(b.buf) - len(b.rem))
}

func (b *Buffer) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, errors.New("bytetools: BufferAt.ReadAt: negative offset")
	}
	if off >= int64(len(b.buf)) {
		return 0, io.EOF
	}
	n = copy(p, b.buf[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
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

func (b *Buffer) ReadByte() (o byte, err error) {
	if len(b.rem) == 0 {
		return 0, io.ErrUnexpectedEOF
	}
	o, b.rem = b.rem[0], b.rem[1:]
	return o, nil
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

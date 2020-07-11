package iotools

import (
	"io"
)

type ReaderAtByteStream struct {
	rdr io.ReaderAt
	off int64
	rem []byte
	buf []byte
	err error
}

func NewReaderAtByteStream(rdr io.ReaderAt, buf []byte) *ReaderAtByteStream {
	if len(buf) == 0 {
		buf = make([]byte, 8192)
	}
	return &ReaderAtByteStream{
		rdr: rdr,
		rem: buf[:0],
		buf: buf[:len(buf)],
	}
}

func (b *ReaderAtByteStream) Avail() int {
	return len(b.rem)
}

func (b *ReaderAtByteStream) TakeExactly(n int) (o []byte, err error) {
	// This is extremely unfortunate. Nothing I do can get this below 80.
	// ./readeratbytestream.go:30:6: cannot inline (*ReaderAtByteStream).Exactly:
	//		function too complex: cost 83 exceeds budget 80
	if n > len(b.rem) {
		return b.takeExactlySlow(n), b.err
	}
	o, b.rem = b.rem[:n], b.rem[n:]
	return
}

func (b *ReaderAtByteStream) takeExactlySlow(n int) (out []byte) {
	if b.err != nil {
		return nil
	}
	if err := b.fill(); err != nil {
		b.err = err
		return nil
	}
	if n > len(b.rem) {
		b.err = io.ErrShortBuffer
		return nil
	}
	out, b.rem = b.rem[:n], b.rem[n:]
	return
}

func (b *ReaderAtByteStream) DiscardExactly(n int) (err error) {
	if n > len(b.rem) {
		return b.discardExactlySlow(n)
	}
	b.rem = b.rem[n:]
	return
}

func (b *ReaderAtByteStream) discardExactlySlow(sz int) error {
	if b.err != nil {
		return b.err
	}
	if sz <= 0 {
		return nil
	}

	// We must attempt a read so we can know if we've successfully discarded
	// the right number of bytes. If we are discarding the exact number of bytes
	// left in the underlying reader, we will get an EOF just the same as if
	// we run past the end, but if we back up one byte, we can distinguish the
	// two conditions.
	discarded := int64(len(b.rem))
	endingOffset := b.off + (int64(sz) - discarded)
	readAt := endingOffset - 1

	b.rem = nil

	rd, err := b.rdr.ReadAt(b.buf, readAt)
	if err != nil && err != io.EOF {
		b.err = err
		return err
	}

	if rd < 1 {
		b.err = io.EOF
		return b.err
	}

	b.off = readAt + int64(rd)
	b.rem = b.buf[1:rd] // Dispose of the leading byte
	return nil
}

func (b *ReaderAtByteStream) DiscardUpTo(n int) error {
	if n > len(b.rem) {
		b.off += int64(n - len(b.rem))
		b.rem = nil
	} else {
		b.rem = b.rem[n:]
	}
	return nil
}

func (b *ReaderAtByteStream) PeekExactly(n int) (o []byte, err error) {
	if n > len(b.rem) {
		return b.peekExactlySlow(n), b.err
	}
	o = b.rem[:n]
	return
}

func (b *ReaderAtByteStream) peekExactlySlow(n int) (out []byte) {
	if b.err != nil {
		return nil
	}
	if err := b.fill(); err != nil {
		b.err = err
		return nil
	}
	if n > len(b.rem) {
		b.err = io.ErrShortBuffer
		return nil
	}
	out = b.rem[:n]
	return
}

func (b *ReaderAtByteStream) PeekUpTo(n int) (o []byte, err error) {
	if n > len(b.rem) {
		return b.peekUpToSlow(n), b.err
	}
	o = b.rem[:n]
	return
}

func (b *ReaderAtByteStream) peekUpToSlow(n int) (out []byte) {
	if len(b.rem) > 0 && b.err == io.EOF {
		out, b.rem = b.rem, nil
		return
	}
	if n > len(b.buf) {
		b.err = io.ErrShortBuffer
		return nil
	}
	if b.err != nil {
		return nil
	}
	if err := b.fill(); err != nil {
		b.err = err
		return nil
	}
	if n > len(b.rem) {
		return b.rem
	}
	return b.rem[:n]
}

func (b *ReaderAtByteStream) TakeUpTo(n int) (o []byte, err error) {
	if n > len(b.rem) {
		return b.takeUpToSlow(n), b.err
	}
	o, b.rem = b.rem[:n], b.rem[n:]
	return
}

func (b *ReaderAtByteStream) takeUpToSlow(n int) (out []byte) {
	if b.err != nil {
		return nil
	}
	if err := b.fill(); err != nil {
		if len(b.rem) > 0 && b.err == io.EOF {
			out, b.rem = b.rem, nil
			b.err = nil
			return
		}
		b.err = err
		return nil
	}
	if n > len(b.rem) {
		n = len(b.rem)
	}
	out, b.rem = b.rem[:n], b.rem[n:]
	return
}

func (b *ReaderAtByteStream) fill() error {
	left := copy(b.buf, b.rem)
	n, err := b.rdr.ReadAt(b.buf[left:], b.off)
	if err != nil && (err != io.EOF || n == 0) {
		b.err = err
		return err
	}
	b.off += int64(n)
	b.rem = b.buf[:left+n]
	return nil
}

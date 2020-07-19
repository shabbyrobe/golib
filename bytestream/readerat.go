package bytestream

import (
	"io"
)

type ReaderAt struct {
	rdr io.ReaderAt
	off int64
	rem []byte
	buf []byte
	err error
}

var _ ByteStream = &ReaderAt{}

func NewReaderAt(rdr io.ReaderAt, buf []byte) *ReaderAt {
	if len(buf) == 0 {
		buf = make([]byte, 8192)
	}
	return &ReaderAt{
		rdr: rdr,
		rem: buf[:0],
		buf: buf[:len(buf)],
	}
}

func (b *ReaderAt) Err() error {
	return b.err
}

func (b *ReaderAt) Avail() int64 {
	return int64(len(b.rem))
}

func (b *ReaderAt) Limit() int64 {
	return int64(len(b.buf))
}

func (b *ReaderAt) Tell() int64 {
	return b.off - int64(len(b.rem))
}

func (b *ReaderAt) ReadByte() (o byte, err error) {
	if len(b.rem) == 0 {
		return b.readByteSlow(), b.err
	}
	o, b.rem = b.rem[0], b.rem[1:]
	return
}

func (b *ReaderAt) readByteSlow() (out byte) {
	if b.err != nil {
		return 0
	}
	if err := b.fill(); err != nil {
		b.err = err
		return 0
	}
	out, b.rem = b.rem[0], b.rem[1:]
	return
}

func (b *ReaderAt) TakeExactly(n int) (o []byte, err error) {
	// This is extremely unfortunate. Nothing I do can get this below 80.
	// ./ReaderAt.go:30:6: cannot inline (*ReaderAt).Exactly:
	//		function too complex: cost 83 exceeds budget 80
	if n > len(b.rem) {
		return b.takeExactlySlow(n), b.err
	}
	o, b.rem = b.rem[:n], b.rem[n:]
	return
}

func (b *ReaderAt) takeExactlySlow(n int) (out []byte) {
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

func (b *ReaderAt) DiscardExactly(n int) (err error) {
	if n > len(b.rem) {
		return b.discardExactlySlow(n)
	}
	b.rem = b.rem[n:]
	return
}

func (b *ReaderAt) discardExactlySlow(sz int) error {
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

func (b *ReaderAt) DiscardUpTo(n int) error {
	if n > len(b.rem) {
		b.off += int64(n - len(b.rem))
		b.rem = nil
	} else {
		b.rem = b.rem[n:]
	}
	return nil
}

func (b *ReaderAt) PeekExactly(n int) (o []byte, err error) {
	if n > len(b.rem) {
		return b.peekExactlySlow(n), b.err
	}
	o = b.rem[:n]
	return
}

func (b *ReaderAt) peekExactlySlow(n int) (out []byte) {
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

func (b *ReaderAt) PeekUpTo(n int) (o []byte, err error) {
	if n > len(b.rem) {
		return b.peekUpToSlow(n), b.err
	}
	o = b.rem[:n]
	return
}

func (b *ReaderAt) peekUpToSlow(n int) (out []byte) {
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
		if len(b.rem) > 0 && err == io.EOF {
			err = nil
		} else {
			b.err = err
			return nil
		}
	}
	if n > len(b.rem) {
		return b.rem
	}
	return b.rem[:n]
}

func (b *ReaderAt) TakeUpTo(n int) (o []byte, err error) {
	if n > len(b.rem) {
		return b.takeUpToSlow(n), b.err
	}
	o, b.rem = b.rem[:n], b.rem[n:]
	return
}

func (b *ReaderAt) takeUpToSlow(n int) (out []byte) {
	if b.err != nil {
		return nil
	}
	if err := b.fill(); err != nil {
		if len(b.rem) > 0 && err == io.EOF {
			out, b.rem = b.rem, nil
			err = nil
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

func (b *ReaderAt) fill() error {
	left := copy(b.buf, b.rem)
	n, err := b.rdr.ReadAt(b.buf[left:], b.off)
	if err != nil && (err != io.EOF || n == 0) {
		return err
	}
	b.off += int64(n)
	b.rem = b.buf[:left+n]
	return nil
}

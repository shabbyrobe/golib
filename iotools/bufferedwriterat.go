package iotools

import (
	"io"
)

// SequentialBufferedWriterAt buffers calls to WriteAt if they are adjacent.
// It is optimised for the case that writes using WriteAt probably occur
// like calls to Write, but may occasionally need to WriteAt a specific random
// offset (albeit more slowly).
//
type SequentialBufferedWriterAt struct {
	w      io.WriterAt
	cls    io.Closer
	buffer []byte
	max    int64

	start int64
	left  int64
	len   int64
}

var _ io.WriterAt = &SequentialBufferedWriterAt{}

func NewSequentialBufferedWriterAt(w io.WriterAt, size int64) *SequentialBufferedWriterAt {
	cls, _ := w.(io.Closer)
	if size <= 0 {
		size = 8192
	}

	bw := &SequentialBufferedWriterAt{
		w:      w,
		cls:    cls,
		buffer: make([]byte, size),
		max:    size,
		start:  -1,
	}

	return bw
}

func (wr *SequentialBufferedWriterAt) WriteAt(in []byte, off int64) (n int, err error) {
	plen := len(in)
	plen64 := int64(plen)
	if plen64 > wr.max {
		if err := wr.Flush(); err != nil {
			return 0, err
		}
		return wr.w.WriteAt(in, off)
	}

	if off != wr.start+wr.len {
		// We can only append to the existing buffer if the offset is identical to the
		// existing buffer's end byte.
		if err := wr.Flush(); err != nil {
			return 0, err
		}
	}

	if plen64 > wr.left {
		copied := int64(copy(wr.buffer[wr.len:], in))
		wr.len += copied
		off += copied
		if err := wr.Flush(); err != nil {
			return 0, err
		}

		in = in[copied:]
	}

	if wr.start < 0 {
		wr.start = off
	}

	copied := int64(copy(wr.buffer[wr.len:], in))
	wr.len += copied
	wr.left -= copied
	return plen, nil
}

func (wr *SequentialBufferedWriterAt) Close() error {
	if wr.cls != nil {
		return wr.cls.Close()
	}
	return nil
}

func (wr *SequentialBufferedWriterAt) Flush() error {
	if wr.len > 0 {
		if _, err := wr.w.WriteAt(wr.buffer[:wr.len], wr.start); err != nil {
			return err
		}
	}
	wr.len = 0
	wr.start = -1
	wr.left = wr.max
	return nil
}

package iotools

import "io"

type WriterAtOffset struct {
	Offset   int64
	WriterAt io.WriterAt
}

func NewWriterAtOffset(w io.WriterAt, offset int64) *WriterAtOffset {
	if offset < 0 {
		panic("iotools: offset must be >= 0")
	}
	return &WriterAtOffset{WriterAt: w, Offset: offset}
}

func (wr *WriterAtOffset) WriteAt(p []byte, off int64) (n int, err error) {
	return wr.WriterAt.WriteAt(p, off+wr.Offset)
}

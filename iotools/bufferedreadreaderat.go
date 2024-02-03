package iotools

import (
	"bufio"
	"fmt"
	"io"
)

// BufferedReadReaderAt wraps a buffered io.Reader so that it supports io.ReaderAt.
//
// This will perform optimally if calls to ReadAt produce sequential calls to Read on the
// underlying reader. If calls to ReadAt are random, it will perform disastrously.
type BufferedReadReaderAt struct {
	bufr *bufio.Reader
	pos  int64
}

var _ io.Reader = &BufferedReadReaderAt{}
var _ io.ReaderAt = &BufferedReadReaderAt{}

func NewBufferedReadReaderAt(rdr *bufio.Reader) *BufferedReadReaderAt {
	return &BufferedReadReaderAt{bufr: rdr}
}

func (r *BufferedReadReaderAt) Read(b []byte) (n int, err error) {
	n, err = io.ReadFull(r.bufr, b)
	r.pos += int64(n)
	return n, err
}

func (r *BufferedReadReaderAt) ReadAt(b []byte, off int64) (n int, err error) {
	if r.pos < off {
		chuck := off - r.pos
		if _, err := r.bufr.Discard(int(chuck)); err != nil {
			return 0, err
		}
		r.pos = off
	}
	if off != r.pos {
		return 0, fmt.Errorf("unexpected read position %d, expected %d", off, r.pos)
	}
	return r.Read(b)
}

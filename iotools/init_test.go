package iotools

import (
	"bytes"
	"io"

	"github.com/shabbyrobe/golib/assert"
)

type splitReader struct {
	input  *bytes.Reader
	lens   []int
	curLen int
	done   bool
}

func newSplitReader(bts []byte, lens ...int) *splitReader {
	return &splitReader{input: bytes.NewReader(bts), lens: lens}
}

func (rdr *splitReader) Read(buf []byte) (n int, err error) {
	if rdr.curLen < len(rdr.lens) {
		curLen := rdr.lens[rdr.curLen]
		rdr.curLen++
		if curLen < len(buf) {
			return rdr.input.Read(buf[:curLen])
		} else {
			return rdr.input.Read(buf)
		}
	}

	return rdr.input.Read(buf)
}

type writeEvent struct {
	op string
	p  []byte
	at int64
}

type flushingWriterAt interface {
	io.WriterAt
	Flush() error
}

type loggingWriterAt struct {
	writes []writeEvent
}

func (lwa *loggingWriterAt) WriteAt(p []byte, offset int64) (n int, err error) {
	pc := make([]byte, len(p))
	copy(pc, p)
	lwa.writes = append(lwa.writes, writeEvent{
		p: pc, at: offset,
	})
	return len(pc), nil
}

func (lwa *loggingWriterAt) assertWritesFlush(tt assert.T, via flushingWriterAt, data []byte, at int64, result ...writeEvent) {
	lwa.assertWrites(tt, via, data, at)
	lwa.assertFlush(tt, via, result...)
}

func (lwa *loggingWriterAt) assertWrites(tt assert.T, via io.WriterAt, data []byte, at int64, result ...writeEvent) {
	tt.Helper()
	n, err := via.WriteAt(data, at)
	tt.MustOK(err)
	tt.MustEqual(len(data), n)
	tt.MustEqual(result, lwa.writes)
	lwa.writes = nil
}

func (lwa *loggingWriterAt) assertFlush(tt assert.T, via flushingWriterAt, result ...writeEvent) {
	tt.Helper()
	tt.MustOK(via.Flush())
	tt.MustEqual(result, lwa.writes)
	lwa.writes = nil
}

func assertWriteAt(tt assert.T, to io.WriterAt, data []byte, at int64) {
	tt.Helper()
	n, err := to.WriteAt(data, at)
	tt.MustOK(err)
	tt.MustEqual(len(data), n)
}

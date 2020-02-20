package iotools

import (
	"bytes"
	"io"
	"reflect"
	"testing"
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

func (lwa *loggingWriterAt) assertWritesFlush(t testing.TB, via flushingWriterAt, data []byte, at int64, result ...writeEvent) {
	lwa.assertWrites(t, via, data, at)
	lwa.assertFlush(t, via, result...)
}

func (lwa *loggingWriterAt) assertWrites(t testing.TB, via io.WriterAt, data []byte, at int64, result ...writeEvent) {
	t.Helper()
	n, err := via.WriteAt(data, at)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != n {
		t.Fatal(len(data), n)
	}
	if !reflect.DeepEqual(result, lwa.writes) {
		t.Fatal(result, lwa.writes)
	}
	lwa.writes = nil
}

func (lwa *loggingWriterAt) assertFlush(t testing.TB, via flushingWriterAt, result ...writeEvent) {
	t.Helper()
	if err := via.Flush(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result, lwa.writes) {
		t.Fatal(result, lwa.writes)
	}
	lwa.writes = nil
}

func assertWriteAt(t testing.TB, to io.WriterAt, data []byte, at int64) {
	t.Helper()
	n, err := to.WriteAt(data, at)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != n {
		t.Fatal(len(data), n)
	}
}

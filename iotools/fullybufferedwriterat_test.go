package iotools

import (
	"bytes"
	"testing"

	"github.com/shabbyrobe/golib/assert"
)

type testBufferedWriterDestination struct {
	events []writeEvent
}

func (tb *testBufferedWriterDestination) Write(buf []byte) (n int, err error) {
	tb.events = append(tb.events, writeEvent{op: "write", p: buf})
	return 0, nil
}

func (tb *testBufferedWriterDestination) Close() error {
	tb.events = append(tb.events, writeEvent{op: "close"})
	return nil
}

func (tb *testBufferedWriterDestination) Truncate(sz int64) error {
	tb.events = append(tb.events, writeEvent{op: "truncate"})
	return nil
}

func TestFullyBufferedWriterAt(t *testing.T) {
	t.Run("simple-write-flush", func(t *testing.T) {
		tt := assert.WrapTB(t)

		var tbw testBufferedWriterDestination
		bwa := NewFullyBufferedWriterAt(bytes.NewReader([]byte{}), &tbw)
		assertWriteAt(tt, bwa, []byte{1, 2}, 0)
		tt.MustEqual(0, len(tbw.events))

		tt.MustOK(bwa.Flush())
		tt.MustEqual([]writeEvent{
			{op: "truncate"},
			{op: "write", p: []byte{1, 2}},
		}, tbw.events)
	})

	t.Run("extends", func(t *testing.T) {
		tt := assert.WrapTB(t)

		var tbw testBufferedWriterDestination
		bwa := NewFullyBufferedWriterAt(bytes.NewReader([]byte{}), &tbw)
		assertWriteAt(tt, bwa, []byte{1, 2}, 0)
		tt.MustEqual(0, len(tbw.events))

		assertWriteAt(tt, bwa, []byte{4, 5}, 3)
		tt.MustEqual(0, len(tbw.events))

		tt.MustOK(bwa.Flush())
		tt.MustEqual([]writeEvent{
			{op: "truncate"},
			{op: "write", p: []byte{1, 2, 0, 4, 5}},
		}, tbw.events)
	})

	t.Run("refresh", func(t *testing.T) {
		tt := assert.WrapTB(t)

		buf := []byte{1, 2, 3, 4}
		rdr := bytes.NewReader(buf)
		var tbw testBufferedWriterDestination
		bwa := NewFullyBufferedWriterAt(rdr, &tbw)
		assertWriteAt(tt, bwa, []byte{9, 9}, 0)

		tt.MustOK(bwa.Flush())
		tt.MustEqual([]writeEvent{
			{op: "truncate"},
			{op: "write", p: []byte{9, 9, 3, 4}},
		}, tbw.events)

		buf[2] = 9
		into := make([]byte, 4)
		n, err := bwa.ReadAt(into, 0)
		tt.MustOK(err)
		tt.MustEqual(4, n)
		tt.MustEqual([]byte{9, 9, 3, 4}, into)
		tt.MustOK(bwa.Refresh())

		n, err = bwa.ReadAt(into, 0)
		tt.MustOK(err)
		tt.MustEqual(4, n)
		tt.MustEqual([]byte{1, 2, 9, 4}, into)
	})
}

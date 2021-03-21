package iotools

import (
	"bytes"
	"reflect"
	"testing"
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
		var tbw testBufferedWriterDestination
		bwa := NewFullyBufferedWriterAt(bytes.NewReader([]byte{}), &tbw)
		assertWriteAt(t, bwa, []byte{1, 2}, 0)
		if len(tbw.events) != 0 {
			t.Fatal()
		}
		if err := bwa.Flush(); err != nil {
			t.Fatal()
		}
		if !reflect.DeepEqual(tbw.events, []writeEvent{
			{op: "truncate"},
			{op: "write", p: []byte{1, 2}}},
		) {
			t.Fatal(tbw.events)
		}
	})

	t.Run("extends", func(t *testing.T) {
		var tbw testBufferedWriterDestination
		bwa := NewFullyBufferedWriterAt(bytes.NewReader([]byte{}), &tbw)
		assertWriteAt(t, bwa, []byte{1, 2}, 0)
		if len(tbw.events) != 0 {
			t.Fatal()
		}

		assertWriteAt(t, bwa, []byte{4, 5}, 3)
		if len(tbw.events) != 0 {
			t.Fatal()
		}
		if err := bwa.Flush(); err != nil {
			t.Fatal()
		}
		if !reflect.DeepEqual(tbw.events, []writeEvent{
			{op: "truncate"},
			{op: "write", p: []byte{1, 2, 0, 4, 5}},
		}) {
			t.Fatal()
		}
	})

	t.Run("refresh", func(t *testing.T) {
		buf := []byte{1, 2, 3, 4}
		rdr := bytes.NewReader(buf)
		var tbw testBufferedWriterDestination
		bwa := NewFullyBufferedWriterAt(rdr, &tbw)
		assertWriteAt(t, bwa, []byte{9, 9}, 0)
		if err := bwa.Flush(); err != nil {
			t.Fatal()
		}
		if !reflect.DeepEqual(tbw.events, []writeEvent{
			{op: "truncate"},
			{op: "write", p: []byte{9, 9, 3, 4}},
		}) {
			t.Fatal()
		}

		buf[2] = 9
		into := make([]byte, 4)
		n, err := bwa.ReadAt(into, 0)
		if err != nil {
			t.Fatal(err)
		}
		if n != 4 {
			t.Fatal()
		}
		if !reflect.DeepEqual([]byte{9, 9, 3, 4}, into) {
			t.Fatal()
		}
		if err := bwa.Refresh(); err != nil {
			t.Fatal()
		}

		n, err = bwa.ReadAt(into, 0)
		if err != nil {
			t.Fatal(err)
		}
		if n != 4 {
			t.Fatal()
		}
		if !reflect.DeepEqual([]byte{1, 2, 9, 4}, into) {
			t.Fatal()
		}
	})
}

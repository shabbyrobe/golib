package iotools

import (
	"crypto/rand"
	"io/ioutil"
	"os"
	"testing"

	"github.com/shabbyrobe/golib/assert"
)

func TestSequentialBufferedWriterAt(t *testing.T) {
	t.Run("simple-write-flush", func(t *testing.T) {
		tt := assert.WrapTB(t)

		var lwa loggingWriterAt
		bwa := NewSequentialBufferedWriterAt(&lwa, 4)
		lwa.assertWritesFlush(tt, bwa, []byte{1, 2}, 1, writeEvent{p: []byte{1, 2}, at: 1})
		lwa.assertWritesFlush(tt, bwa, []byte{3, 4}, 4, writeEvent{p: []byte{3, 4}, at: 4})
	})

	t.Run("simple-write-buffer", func(t *testing.T) {
		tt := assert.WrapTB(t)

		var lwa loggingWriterAt
		bwa := NewSequentialBufferedWriterAt(&lwa, 4)
		lwa.assertWrites(tt, bwa, []byte{1, 2}, 1)
		lwa.assertWrites(tt, bwa, []byte{3, 4}, 3)
		lwa.assertFlush(tt, bwa, writeEvent{p: []byte{1, 2, 3, 4}, at: 1})
	})

	t.Run("write-overrun-flushes", func(t *testing.T) {
		tt := assert.WrapTB(t)

		var lwa loggingWriterAt
		bwa := NewSequentialBufferedWriterAt(&lwa, 4)
		lwa.assertWrites(tt, bwa, []byte{1, 2, 3}, 1)
		lwa.assertWrites(tt, bwa, []byte{4, 5}, 4, writeEvent{p: []byte{1, 2, 3, 4}, at: 1})
		lwa.assertFlush(tt, bwa, writeEvent{p: []byte{5}, at: 5})
	})

	t.Run("write-too-big-with-empty-buffer", func(t *testing.T) {
		tt := assert.WrapTB(t)

		var lwa loggingWriterAt
		bwa := NewSequentialBufferedWriterAt(&lwa, 4)
		lwa.assertWrites(tt, bwa, []byte{1, 2, 3, 4, 5}, 1, writeEvent{p: []byte{1, 2, 3, 4, 5}, at: 1})
		lwa.assertFlush(tt, bwa)
	})

	t.Run("write-too-big-with-flushes-buffer", func(t *testing.T) {
		tt := assert.WrapTB(t)

		var lwa loggingWriterAt
		bwa := NewSequentialBufferedWriterAt(&lwa, 4)
		lwa.assertWrites(tt, bwa, []byte{1, 2}, 1)
		lwa.assertWrites(tt, bwa, []byte{3, 4, 5, 6, 7}, 3,
			writeEvent{p: []byte{1, 2}, at: 1},
			writeEvent{p: []byte{3, 4, 5, 6, 7}, at: 3})
		lwa.assertFlush(tt, bwa)
	})

	t.Run("random-write-after-flushes", func(t *testing.T) {
		tt := assert.WrapTB(t)

		var lwa loggingWriterAt
		bwa := NewSequentialBufferedWriterAt(&lwa, 4)
		lwa.assertWrites(tt, bwa, []byte{1, 2}, 1)
		lwa.assertWrites(tt, bwa, []byte{5, 6}, 5, writeEvent{p: []byte{1, 2}, at: 1})
		lwa.assertFlush(tt, bwa, writeEvent{p: []byte{5, 6}, at: 5})
	})

	t.Run("random-write-before-flushes", func(t *testing.T) {
		tt := assert.WrapTB(t)

		var lwa loggingWriterAt
		bwa := NewSequentialBufferedWriterAt(&lwa, 4)
		lwa.assertWrites(tt, bwa, []byte{5, 6}, 5)
		lwa.assertWrites(tt, bwa, []byte{1, 2}, 1, writeEvent{p: []byte{5, 6}, at: 5})
		lwa.assertFlush(tt, bwa, writeEvent{p: []byte{1, 2}, at: 1})
	})

	t.Run("random-overlap-flushes", func(t *testing.T) {
		tt := assert.WrapTB(t)

		var lwa loggingWriterAt
		bwa := NewSequentialBufferedWriterAt(&lwa, 4)
		lwa.assertWrites(tt, bwa, []byte{5, 6}, 5)
		lwa.assertWrites(tt, bwa, []byte{6, 7}, 6, writeEvent{p: []byte{5, 6}, at: 5})
		lwa.assertFlush(tt, bwa, writeEvent{p: []byte{6, 7}, at: 6})
	})
}

func BenchmarkSequentialBufferedWriterAt(b *testing.B) {
	b.Run("baseline", func(b *testing.B) {
		f, err := ioutil.TempFile("", "")
		if err != nil {
			panic(err)
		}
		defer os.Remove(f.Name())

		input := make([]byte, 64)
		rand.Read(input)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			offset := int64((i % 1000) * 64)
			if _, err := f.WriteAt(input, offset); err != nil {
				panic(err)
			}
		}
	})

	b.Run("writerat", func(b *testing.B) {
		f, err := ioutil.TempFile("", "")
		if err != nil {
			panic(err)
		}
		defer os.Remove(f.Name())

		input := make([]byte, 64)
		rand.Read(input)
		b.ResetTimer()

		batchSize := 1000

		var bw = NewSequentialBufferedWriterAt(f, 8192)
		cur := 0
		for i := 0; i < b.N; i++ {
			offset := int64(cur * 64)
			if _, err := bw.WriteAt(input, offset); err != nil {
				panic(err)
			}
			cur++
			if cur >= batchSize {
				if err := bw.Flush(); err != nil {
					panic(err)
				}
				cur = 0
			}
		}
	})
}

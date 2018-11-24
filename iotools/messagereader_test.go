package iotools

import (
	"bytes"
	"io"
	"testing"

	"github.com/shabbyrobe/golib/assert"
)

func TestBytePrefixMessageReader(t *testing.T) {
	tt := assert.WrapTB(t)

	buf := []byte{}
	expected := [][]byte{}

	for i := 0; i < 256; i++ {
		buf = append(buf, byte(i))
		var cur []byte
		for j := 0; j < i; j++ {
			buf = append(buf, byte(j))
			cur = append(cur, byte(j))
		}
		if i > 0 {
			expected = append(expected, cur)
		}
	}

	// This attempts to chop the last message up. Probably worth making a separate test.
	maxScratch := len(buf) + 2
	scratch := make([]byte, maxScratch)
	for scratchSize := maxScratch - 256 - 10; scratchSize <= maxScratch; scratchSize++ {
		pr := NewBytePrefixMessageReader(bytes.NewReader(buf), scratch[:scratchSize])
		i := 0
		for {
			out, n, err := pr.ReadNext()
			if err == io.EOF {
				break
			}
			if i < len(expected) {
				tt.MustOK(err)
				tt.MustEqual(len(out), n)
				tt.MustEqual(expected[i], out, "failed at index %d", i)
			}
			i++
		}
		tt.MustEqual(i, 255) // 255, not 256: the '0' case doesn't yield a message.
	}
}

func TestBytePrefixMessageReaderReadEmpty(t *testing.T) {
	tt := assert.WrapTB(t)

	pr := NewBytePrefixMessageReader(bytes.NewReader([]byte{}), nil)
	i := 0
	for {
		out, n, err := pr.ReadNext()
		if err == io.EOF {
			break
		}
		tt.MustEqual([]byte{}, out)
		tt.MustEqual(0, n)
		i++
	}
	tt.MustEqual(i, 0)
}

var BenchBytePrefixMessageReaderResult int

func BenchmarkBytePrefixMessageReader(b *testing.B) {
	buf := []byte{}

	for k := 0; k < 16; k++ {
		for i := 0; i < 256; i++ {
			buf = append(buf, byte(i))
			for j := 0; j < i; j++ {
				buf = append(buf, byte(j))
			}
		}
	}

	b.SetBytes(int64(len(buf)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pr := NewBytePrefixMessageReader(bytes.NewReader(buf), nil)
		for {
			_, n, err := pr.ReadNext()
			if err == io.EOF {
				break
			}
			BenchBytePrefixMessageReaderResult += n
		}
	}
}

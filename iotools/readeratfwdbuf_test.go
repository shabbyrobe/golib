package iotools

import (
	"bytes"
	"io"
	"testing"
)

type readCountingReaderAt struct {
	inner io.ReaderAt
	reads int
}

func (r *readCountingReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	r.reads++
	return r.inner.ReadAt(p, off)
}

func readAll(inner io.ReaderAt, scratch []byte) ([]byte, error) {
	var pos int64
	var out []byte
	for {
		n, err := inner.ReadAt(scratch, pos)
		out = append(out, scratch[:n]...)
		if err == io.EOF {
			if n == 0 {
				return out, nil
			}
		} else if err != nil {
			return out, err
		}
		pos += int64(n)
	}
}

func TestReaderAtFwd(t *testing.T) {
	src := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	inner := &readCountingReaderAt{inner: bytes.NewReader(src)}
	rdrScratchSize := 3
	readAllScratchSize := 3
	expectedReads := 4 // 3 data reads and an EOF

	rdrScratch := make([]byte, rdrScratchSize)
	rdr := NewReaderAtFwdBuffer(inner, rdrScratch)

	readAllScratch := make([]byte, readAllScratchSize)
	b, err := readAll(rdr, readAllScratch)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, src) {
		t.Fatal(b, "!=", src)
	}
	if inner.reads != expectedReads {
		t.Fatal("reads", inner.reads, "!=", expectedReads)
	}
}

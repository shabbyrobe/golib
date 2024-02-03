package iotools

import (
	"bufio"
	"io"
)

type LineChunker struct {
	rdr  io.Reader
	brdr *bufio.Reader
	kept []byte
}

func NewLineChunker(rdr io.Reader, sz int) *LineChunker {
	if sz < 0 {
		sz = 4096
	}

	brdr, ok := rdr.(*bufio.Reader)
	if !ok {
		brdr = bufio.NewReaderSize(rdr, sz)
	}

	return &LineChunker{
		rdr:  rdr,
		brdr: brdr,
	}
}

func (lc *LineChunker) NextChunk(into []byte) (n int, err error) {
	if lc.kept != nil {
		if len(lc.kept) > len(into) {
			return n, bufio.ErrBufferFull
		}
		n += copy(into, lc.kept)
		lc.kept = nil
	}

	for {
		line, err := lc.brdr.ReadSlice('\n')
		if err == io.EOF {
			if n == 0 && len(line) == 0 {
				return n, err
			} else if len(line) == 0 {
				return n, nil
			}
		} else if err != nil {
			return n, err
		}
		if n+len(line) > len(into) {
			if n == 0 {
				return n, bufio.ErrBufferFull
			}
			lc.kept = line
			break
		}
		n += copy(into[n:], line)

	}
	return n, nil
}

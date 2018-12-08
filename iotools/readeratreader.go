package iotools

import "io"

// ReaderAtReader turns an io.ReaderAt into an io.Reader.
type ReaderAtReader struct {
	at  io.ReaderAt
	pos int
}

func NewReaderAtReader(at io.ReaderAt, startPos int) *ReaderAtReader {
	return &ReaderAtReader{at, startPos}
}

func (rdr *ReaderAtReader) Read(buf []byte) (n int, err error) {
	n, err = rdr.at.ReadAt(buf, int64(rdr.pos))
	if n > 0 && err == io.EOF {
		err = nil
	}
	rdr.pos += n
	return n, err
}

package iotools

import "io"

// io.ReaderAt implementation which expects the calls to ReadAt will be sequential.
//
// Each read will always be exactly the size of buf. Reads must not be larger than buf.
type ReaderAtFwdBuffer struct {
	inner io.ReaderAt
	buf   []byte
	cur   []byte
	off   int64
	pos   int64
	sz    int64
}

func NewReaderAtFwdBuffer(inner io.ReaderAt, buf []byte) *ReaderAtFwdBuffer {
	if len(buf) == 0 {
		buf = make([]byte, 8192)
	}
	return &ReaderAtFwdBuffer{
		inner: inner,
		buf:   buf,
	}
}

func (rdr *ReaderAtFwdBuffer) ReadAt(p []byte, off int64) (n int, err error) {
	if len(p) > len(rdr.buf) {
		return 0, io.ErrShortBuffer
	}

	if off < rdr.off || off >= rdr.off+rdr.sz {
		rdr.cur = nil
		rdr.pos = 0
		rdr.sz = 0
	}

	if len(rdr.cur[rdr.pos:]) < len(p) {
		c := copy(rdr.buf, rdr.cur[rdr.pos:rdr.sz])
		n, err := rdr.inner.ReadAt(rdr.buf[c:], off)
		if n+c == 0 && err == io.EOF {
			return 0, err
		} else if err != nil && err != io.EOF {
			return 0, err
		}
		rdr.cur = rdr.buf[:n+c]
		rdr.sz = int64(n + c)
		rdr.off = off
		rdr.pos = 0
	}

	n = copy(p, rdr.cur[rdr.pos:])
	rdr.pos += int64(n)
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

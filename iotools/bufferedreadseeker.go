package iotools

import (
	"io"
)

// BufferedReadSeeker implements an io.ReadSeeker that maintains a buffer
// at the site of the last unbuffered read.
//
// The buffer is invalidated if Seek jumps to a position outside the buffered
// section.
//
// It is not intended for use with an io.ReadSeeker that will change size
// during the lifetime of this object.
//
// WARNING: I think there's a gremlin in here somewhere. Don't use it until
// there are more tests.
//
// Also, this didn't really improve performance at all in the application
// it was made for. Tragic waste of time, really.
//
type BufferedReadSeeker struct {
	sz    int64
	inner io.ReadSeeker

	buf     []byte
	bufLen  int64 // Current length of buffered data
	bufPos  int64 // Position relative to start of inner of the start of buf
	posReal int64 // Current position of inner reader
	posVirt int64 // Position of outer reader; buffered reads advance this but not posReal
}

func NewBufferedReadSeeker(in io.ReadSeeker, readerSize int64, buf []byte) (io.ReadSeeker, error) {
	pos, err := in.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	if len(buf) == 0 {
		buf = make([]byte, 65536)
	}

	bt := &BufferedReadSeeker{
		sz:      readerSize,
		inner:   in,
		buf:     buf,
		posReal: pos,
		posVirt: pos,
	}

	return bt, nil
}

func (btr *BufferedReadSeeker) resolvePos() error {
	if btr.posVirt != btr.posReal {
		pos, err := btr.inner.Seek(btr.posVirt, io.SeekStart)
		if err != nil {
			return err
		}
		btr.posReal = pos
		btr.posVirt = pos
	}
	return nil
}

func (btr *BufferedReadSeeker) clearBuf() {
	btr.bufLen = 0
}

func (btr *BufferedReadSeeker) Read(p []byte) (n int, err error) {
	bufEnd := btr.bufPos + btr.bufLen

	if btr.bufLen > 0 && btr.posVirt < bufEnd {
		n = copy(p, btr.buf[btr.posVirt-btr.bufPos:])
		btr.posVirt += int64(n)

	} else {
		if err := btr.resolvePos(); err != nil {
			return 0, err
		}

		bufLen, err := io.ReadFull(btr.inner, btr.buf)
		if err == io.ErrUnexpectedEOF {
			err = nil
		}
		if err != nil {
			return 0, err
		}

		btr.bufPos = btr.posReal
		btr.bufLen = int64(bufLen)
		btr.posReal += btr.bufLen

		n = copy(p, btr.buf[:btr.bufLen])
		btr.posVirt += int64(n)
	}

	return n, err
}

func (btr *BufferedReadSeeker) Seek(offset int64, whence int) (int64, error) {
	pos := ComputeSeekStart(btr.posVirt, btr.sz, offset, whence)

	if btr.bufLen > 0 && pos >= btr.bufPos && pos < btr.bufPos+btr.bufLen {
		btr.posVirt = pos
		return pos, nil

	} else {
		pos, err := btr.inner.Seek(pos, io.SeekStart)
		if err != nil {
			return pos, err
		}

		btr.clearBuf()
		btr.posReal = pos
		btr.posVirt = pos
		return pos, nil
	}
}

package iotools

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/shabbyrobe/golib/assert"
)

func newFile(tt assert.T, data []byte) (*os.File, int64, func()) {
	f, err := ioutil.TempFile("", "")
	tt.MustOK(err)
	tt.MustOKAll(f.Write(data))
	tt.MustOKAll(f.Seek(0, io.SeekStart))

	st, _ := f.Stat()

	return f, st.Size(), func() {
		f.Close()
		os.Remove(f.Name())
	}
}

func TestBufferedReadSeeker(t *testing.T) {
	assertRead := func(tt assert.T, from io.Reader, into []byte, expect []byte) {
		tt.Helper()
		n, err := from.Read(into)
		tt.MustOK(err)
		tt.MustEqual(string(expect), string(into[:n]))
		tt.MustEqual(len(expect), n)
	}

	assertSeek := func(tt assert.T, from io.Seeker, offset int64, whence int, expected int64) {
		tt.Helper()
		pos, err := from.Seek(offset, whence)
		tt.MustOK(err)
		tt.MustEqual(expected, pos)
	}

	t.Run("plain-read", func(t *testing.T) {
		tt := assert.WrapTB(t)

		f, fsz, done := newFile(tt, []byte("1234567890"))
		defer done()

		brs, err := NewBufferedReadSeeker(f, fsz, make([]byte, 4))
		tt.MustOK(err)

		into := make([]byte, 16)
		assertRead(tt, brs, into[:3], []byte("123"))  // Exact read
		assertRead(tt, brs, into[:3], []byte("4"))    // Buffered boundary - only returns the remaining buffered portion
		assertRead(tt, brs, into[:5], []byte("5678")) // Destination buffer larger than inner buffer
		assertRead(tt, brs, into[:4], []byte("90"))   // Last bit
		assertRead(tt, brs, into[:4], nil)            // Last bit
	})

	t.Run("seek-within-buffer", func(t *testing.T) {
		tt := assert.WrapTB(t)

		f, fsz, done := newFile(tt, []byte("12345678901234567890"))
		defer done()

		brs, err := NewBufferedReadSeeker(f, fsz, make([]byte, 5))
		tt.MustOK(err)

		into := make([]byte, 16)
		assertSeek(tt, brs, 3, io.SeekStart, 3)
		assertRead(tt, brs, into[:2], []byte("45")) // Initial buffering read
		assertRead(tt, brs, into[:2], []byte("67")) // Fully buffered read
		assertRead(tt, brs, into[:2], []byte("8"))  // Buffer dregs

		assertSeek(tt, brs, -1, io.SeekCurrent, 7)
		assertRead(tt, brs, into[:2], []byte("8")) // Buffer dregs again

		// Repeat the whole first series of reads, should still be buffered:
		assertSeek(tt, brs, 3, io.SeekStart, 3)
		assertRead(tt, brs, into[:2], []byte("45")) // Initial buffering read
		assertRead(tt, brs, into[:2], []byte("67")) // Fully buffered read
		assertRead(tt, brs, into[:2], []byte("8"))  // Buffer dregs

		// Seek one byte before the buffer, should invalidate:
		assertSeek(tt, brs, 2, io.SeekStart, 2)
		assertRead(tt, brs, into[:4], []byte("3456")) // Initial buffering read
		assertRead(tt, brs, into[:2], []byte("7"))    // Buffer dregs

		// Seek one byte after the buffer, should invalidate:
		assertSeek(tt, brs, 1, io.SeekCurrent, 8)
		assertRead(tt, brs, into[:4], []byte("9012")) // Initial buffering read
		assertRead(tt, brs, into[:2], []byte("3"))    // Buffer dregs
	})
}

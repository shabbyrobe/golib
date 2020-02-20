package iotools

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func newFile(t testing.TB, data []byte) (*os.File, int64, func()) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(data); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}

	st, _ := f.Stat()

	return f, st.Size(), func() {
		f.Close()
		os.Remove(f.Name())
	}
}

func TestBufferedReadSeeker(t *testing.T) {
	assertRead := func(t testing.TB, from io.Reader, into []byte, expect []byte) {
		t.Helper()
		n, err := from.Read(into)
		if err != nil {
			t.Fatal(err)
		}
		if string(expect) != string(into[:n]) {
			t.Fatal(string(expect), "!=", string(into[:n]))
		}
		if len(expect) != n {
			t.Fatal(expect, n)
		}
	}

	assertSeek := func(t testing.TB, from io.Seeker, offset int64, whence int, expected int64) {
		t.Helper()
		pos, err := from.Seek(offset, whence)
		if err != nil {
			t.Fatal(err)
		}
		if expected != pos {
			t.Fatal(expected, pos)
		}
	}

	t.Run("plain-read", func(t *testing.T) {
		f, fsz, done := newFile(t, []byte("1234567890"))
		defer done()

		brs, err := NewBufferedReadSeeker(f, fsz, make([]byte, 4))
		if err != nil {
			t.Fatal(err)
		}

		into := make([]byte, 16)
		assertRead(t, brs, into[:3], []byte("123"))  // Exact read
		assertRead(t, brs, into[:0], []byte{})       // Empty read shouldn't affect anything
		assertRead(t, brs, into[:3], []byte("4"))    // Buffered boundary - only returns the remaining buffered portion
		assertRead(t, brs, into[:0], []byte{})       // Empty read shouldn't affect anything
		assertRead(t, brs, into[:5], []byte("5678")) // Destination buffer larger than inner buffer
		assertRead(t, brs, into[:4], []byte("90"))   // Last bit

		n, err := brs.Read(into[:4])
		if err != io.EOF || n != 0 {
			t.Fatal(n, err)
		}
	})

	t.Run("seek-within-buffer", func(t *testing.T) {
		f, fsz, done := newFile(t, []byte("12345678901234567890"))
		defer done()

		brs, err := NewBufferedReadSeeker(f, fsz, make([]byte, 5))
		if err != nil {
			t.Fatal(err)
		}

		into := make([]byte, 16)
		assertSeek(t, brs, 3, io.SeekStart, 3)
		assertRead(t, brs, into[:2], []byte("45")) // Initial buffering read
		assertRead(t, brs, into[:2], []byte("67")) // Fully buffered read
		assertRead(t, brs, into[:2], []byte("8"))  // Buffer dregs

		assertSeek(t, brs, -1, io.SeekCurrent, 7)
		assertRead(t, brs, into[:2], []byte("8")) // Buffer dregs again

		// Repeat the whole first series of reads, should still be buffered:
		assertSeek(t, brs, 3, io.SeekStart, 3)
		assertRead(t, brs, into[:2], []byte("45")) // Initial buffering read
		assertRead(t, brs, into[:2], []byte("67")) // Fully buffered read
		assertRead(t, brs, into[:2], []byte("8"))  // Buffer dregs

		// Seek one byte before the buffer, should invalidate:
		assertSeek(t, brs, 2, io.SeekStart, 2)
		assertRead(t, brs, into[:4], []byte("3456")) // Initial buffering read
		assertRead(t, brs, into[:2], []byte("7"))    // Buffer dregs

		// Seek one byte after the buffer, should invalidate:
		assertSeek(t, brs, 1, io.SeekCurrent, 8)
		assertRead(t, brs, into[:4], []byte("9012")) // Initial buffering read
		assertRead(t, brs, into[:2], []byte("3"))    // Buffer dregs
	})
}

package iotools

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

/*
ReadAt reads len(p) bytes into p starting at offset off in the underlying input source. It
returns the number of bytes read (0 <= n <= len(p)) and any error encountered.

When ReadAt returns n < len(p), it returns a non-nil error explaining why more bytes were
not returned. In this respect, ReadAt is stricter than Read.

Even if ReadAt returns n < len(p), it may use all of p as scratch space during the call.
If some data is available but not len(p) bytes, ReadAt blocks until either all the data is
available or an error occurs. In this respect ReadAt is different from Read.

If the n = len(p) bytes returned by ReadAt are at the end of the input source, ReadAt may
return either err == EOF or err == nil.

If ReadAt is reading from an input source with a seek offset, ReadAt should not affect nor
be affected by the underlying seek offset.

Clients of ReadAt can execute parallel ReadAt calls on the same input source.

Implementations must not retain p.
*/

type readerAtEOFLast struct {
	buf []byte
	pos int
}

func (r *readerAtEOFLast) ReadAt(p []byte, off int64) (n int, err error) {
	n = copy(p, r.buf[r.pos:])
	r.pos += n
	if n < len(p) {
		return n, io.EOF
	}
	if r.pos == len(r.buf) {
		return n, io.EOF
	}
	return n, nil
}

type readerAtEOFAfter struct {
	buf []byte
	pos int
}

func (r *readerAtEOFAfter) ReadAt(p []byte, off int64) (n int, err error) {
	n = copy(p, r.buf[r.pos:])
	r.pos += n
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

func assertTakeExactly(t *testing.T, bs *ReaderAtByteStream, n int, left int, exp []byte) {
	t.Helper()
	v, err := bs.TakeExactly(n)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, exp) {
		t.Fatal(v, exp)
	}
	if left >= 0 && bs.Avail() != left {
		t.Fatal("avail", bs.Avail(), left)
	}
}

func assertTakeUpTo(t *testing.T, bs *ReaderAtByteStream, n int, left int, exp []byte) {
	t.Helper()
	v, err := bs.TakeUpTo(n)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, exp) {
		t.Fatal(v, exp)
	}
	if left >= 0 && bs.Avail() != left {
		t.Fatal("avail", bs.Avail(), left)
	}
}

func assertPeekUpTo(t *testing.T, bs *ReaderAtByteStream, n int, exp ...byte) {
	t.Helper()
	v, err := bs.PeekUpTo(n)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, exp) {
		t.Fatal(v, exp)
	}
}

func assertTakeExactlyEOF(t *testing.T, bs *ReaderAtByteStream, n int) {
	t.Helper()
	v, err := bs.TakeExactly(n)
	if err != io.EOF {
		t.Fatal(err)
	}
	if !bytes.Equal(v, []byte{}) {
		t.Fatal(v, []byte{})
	}
	if bs.Avail() != 0 {
		t.Fatal("avail", bs.Avail(), 0)
	}
}

func TestReaderAtByteStreamTakeExactly(t *testing.T) {
	bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
	bs := NewReaderAtByteStream(bra, make([]byte, 3))
	assertTakeExactly(t, bs, 2, 1, []byte{0, 1})
	assertTakeExactly(t, bs, 2, 1, []byte{2, 3})
	assertTakeExactly(t, bs, 2, 1, []byte{4, 5})
	assertTakeExactly(t, bs, 2, 1, []byte{6, 7})
	assertTakeExactly(t, bs, 1, 0, []byte{8})
	if _, err := bs.TakeExactly(1); err != io.EOF {
		t.Fatal(err)
	}
}

func TestReaderAtByteStreamTakeUpToExact(t *testing.T) {
	bra := bytes.NewReader([]byte{0, 1, 2, 3})
	bs := NewReaderAtByteStream(bra, nil)
	assertTakeUpTo(t, bs, 3, -1, []byte{0, 1, 2})
	assertTakeUpTo(t, bs, 1, -1, []byte{3})
	if _, err := bs.TakeUpTo(1); err != io.EOF {
		t.Fatal(err)
	}
}

func TestReaderAtByteStreamTakeUpToTooMany(t *testing.T) {
	bra := bytes.NewReader([]byte{0, 1, 2, 3})
	bs := NewReaderAtByteStream(bra, nil)
	assertTakeUpTo(t, bs, 3, -1, []byte{0, 1, 2})
	assertTakeUpTo(t, bs, 2, -1, []byte{3})
	if _, err := bs.TakeUpTo(1); err != io.EOF {
		t.Fatal(err)
	}
}

func TestReaderAtByteStreamTakeExactlyOneUntilEnd(t *testing.T) {
	var inputs = [][]byte{
		{},
		{0},
		{0, 1},
		{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}

	var bufSizes = []int{1, 2, 10, 64}

	for _, input := range inputs {
		for _, bufSize := range bufSizes {
			for idx, tc := range []struct {
				rdr io.ReaderAt
			}{
				{&readerAtEOFLast{buf: input}},
				{&readerAtEOFAfter{buf: input}},
			} {
				t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
					bs := NewReaderAtByteStream(tc.rdr, make([]byte, bufSize))

					for i := 0; i < len(input); i++ {
						if v, err := bs.PeekExactly(1); err != nil {
							t.Fatal(len(input), i, err)
						} else if !bytes.Equal(v, input[i:i+1]) {
							t.Fatal(len(input), i)
						}

						assertTakeExactly(t, bs, 1, -1, []byte{input[i]})
					}

					if _, err := bs.TakeExactly(1); err != io.EOF {
						t.Fatal(err)
					}
				})
			}
		}
	}
}

func TestReaderAtByteStreamTakeUpToOneUntilEnd(t *testing.T) {
	var inputs = [][]byte{
		{},
		{0},
		{0, 1},
		{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}

	var bufSizes = []int{1, 2, 10, 64}

	for _, input := range inputs {
		for _, bufSize := range bufSizes {
			for idx, tc := range []struct {
				rdr io.ReaderAt
			}{
				{&readerAtEOFLast{buf: input}},
				{&readerAtEOFAfter{buf: input}},
			} {
				t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
					bs := NewReaderAtByteStream(tc.rdr, make([]byte, bufSize))

					for i := 0; i < len(input); i++ {
						assertTakeUpTo(t, bs, 1, -1, []byte{input[i]})
					}

					if _, err := bs.TakeUpTo(1); err != io.EOF {
						t.Fatal(err)
					}
				})
			}
		}
	}
}

func TestReaderAtByteStreamDiscardExactlyOneUntilEnd(t *testing.T) {
	var inputs = [][]byte{
		{},
		{0},
		{0, 1},
		{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}

	var bufSizes = []int{1, 2, 10, 64}

	for _, input := range inputs {
		for _, bufSize := range bufSizes {
			for idx, tc := range []struct {
				rdr io.ReaderAt
			}{
				{&readerAtEOFLast{buf: input}},
				{&readerAtEOFAfter{buf: input}},
			} {
				t.Run(fmt.Sprintf("insz=%d/bsz=%d/%d", len(input), bufSize, idx), func(t *testing.T) {
					bs := NewReaderAtByteStream(tc.rdr, make([]byte, bufSize))
					for i := 0; i < len(input); i++ {
						if v, err := bs.PeekExactly(1); err != nil {
							t.Fatal(len(input), i, err)
						} else if !bytes.Equal(v, input[i:i+1]) {
							t.Fatal(len(input), i)
						}

						if err := bs.DiscardExactly(1); err != nil {
							t.Fatal(len(input), i, err)
						}
					}

					if _, err := bs.PeekExactly(1); err != io.EOF {
						t.Fatal(err)
					}
					if err := bs.DiscardExactly(1); err != io.EOF {
						t.Fatal(err)
					}
				})
			}
		}
	}
}

func TestReaderAtByteStreamTakeExactlyFailsWhenTakingTooMuch(t *testing.T) {
	bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
	bs := NewReaderAtByteStream(bra, make([]byte, 3))
	if _, err := bs.TakeExactly(5); err != io.ErrShortBuffer {
		t.Fatal(err)
	}
}

func TestReaderAtByteStreamTakeExactlySucceedsAfterTakeExactlyFailsWithShortBuffer(t *testing.T) {
	bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
	bs := NewReaderAtByteStream(bra, make([]byte, 3))
	assertTakeExactly(t, bs, 1, 2, []byte{0})
	if _, err := bs.TakeExactly(5); err != io.ErrShortBuffer {
		t.Fatal(err)
	}
	assertTakeExactly(t, bs, 1, 2, []byte{1})
}

func TestReaderAtByteStreamDiscardExactly(t *testing.T) {
	t.Run("some", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAtByteStream(bra, make([]byte, 3))
		if err := bs.DiscardExactly(5); err != nil {
			t.Fatal(err)
		}
		assertTakeExactly(t, bs, 2, 0, []byte{5, 6})
	})

	t.Run("discard-take-discard", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4})
		bs := NewReaderAtByteStream(bra, make([]byte, 3))
		assertTakeExactly(t, bs, 2, -1, []byte{0, 1})
		if err := bs.DiscardExactly(2); err != nil {
			t.Fatal(err)
		}
		assertTakeExactly(t, bs, 1, -1, []byte{4})
	})

	t.Run("all-but-last", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAtByteStream(bra, make([]byte, 3))
		if err := bs.DiscardExactly(8); err != nil {
			t.Fatal()
		}
		assertTakeExactly(t, bs, 1, 0, []byte{8})
	})

	t.Run("all-aligned", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAtByteStream(bra, make([]byte, 3))
		if err := bs.DiscardExactly(9); err != nil {
			t.Fatal(err)
		}
		assertTakeExactlyEOF(t, bs, 1)
	})

	t.Run("all-misaligned", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAtByteStream(bra, make([]byte, 2))
		if err := bs.DiscardExactly(9); err != nil {
			t.Fatal(err)
		}
		assertTakeExactlyEOF(t, bs, 1)
	})

	t.Run("too-many", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAtByteStream(bra, make([]byte, 3))
		if err := bs.DiscardExactly(10); err != io.EOF {
			t.Fatal(err)
		}
	})

	t.Run("too-many-after-take", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAtByteStream(bra, make([]byte, 5))

		assertTakeExactly(t, bs, 5, -1, []byte{0, 1, 2, 3, 4})
		if err := bs.DiscardExactly(5); err != io.EOF {
			t.Fatal(err)
		}
	})
}

func TestReaderAtByteStreamDiscardUpTo(t *testing.T) {
	t.Run("some", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAtByteStream(bra, make([]byte, 3))
		if err := bs.DiscardUpTo(5); err != nil {
			t.Fatal(err)
		}
		assertTakeExactly(t, bs, 2, 1, []byte{5, 6})
	})

	t.Run("all-but-last", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAtByteStream(bra, make([]byte, 3))
		if err := bs.DiscardUpTo(8); err != nil {
			t.Fatal()
		}
		assertTakeExactly(t, bs, 1, 0, []byte{8})
	})

	t.Run("all-aligned", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAtByteStream(bra, make([]byte, 3))
		if err := bs.DiscardUpTo(9); err != nil {
			t.Fatal(err)
		}
		assertTakeExactlyEOF(t, bs, 1)
	})

	t.Run("all-misaligned", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAtByteStream(bra, make([]byte, 2))
		if err := bs.DiscardUpTo(9); err != nil {
			t.Fatal(err)
		}
		assertTakeExactlyEOF(t, bs, 1)
	})

	t.Run("too-many", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAtByteStream(bra, make([]byte, 3))
		if err := bs.DiscardUpTo(10); err != nil {
			t.Fatal()
		}
		assertTakeExactlyEOF(t, bs, 1)
	})
}

var BenchResultBytes []byte

func BenchmarkReaderAtByteStream(b *testing.B) {
	b.Run("base", func(b *testing.B) {
		buf := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8}
		bra := bytes.NewReader(buf)
		bs := NewReaderAtByteStream(bra, make([]byte, 3))

		for i := 0; i < b.N; i++ {
			bs.rem = bs.buf[:2]
			BenchResultBytes = buf[:2]
		}
	})

	b.Run("fast", func(b *testing.B) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAtByteStream(bra, make([]byte, 3))
		bs.TakeExactly(9) // seed

		for i := 0; i < b.N; i++ {
			bs.rem = bs.buf[:2]
			bs.TakeExactly(2)
		}
	})

	b.Run("slow", func(b *testing.B) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAtByteStream(bra, make([]byte, 3))

		for i := 0; i < b.N; i++ {
			bs.rem = bs.buf[:0]
			bs.TakeExactly(2)
		}
	})
}

func TestReaderAtByteStreamPeekUpTo(t *testing.T) {
	t.Run("within/bsz=1", func(t *testing.T) {
		buf := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8}
		bra := bytes.NewReader(buf)
		bs := NewReaderAtByteStream(bra, make([]byte, 1))
		assertPeekUpTo(t, bs, 1, 0)
		if _, err := bs.PeekUpTo(2); err != io.ErrShortBuffer {
			t.Fatal(err)
		}
	})

	t.Run("within/bsz=3", func(t *testing.T) {
		buf := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8}
		bra := bytes.NewReader(buf)
		bs := NewReaderAtByteStream(bra, make([]byte, 3))
		assertPeekUpTo(t, bs, 1, 0)
		assertPeekUpTo(t, bs, 2, 0, 1)
		assertPeekUpTo(t, bs, 3, 0, 1, 2)
		if _, err := bs.PeekUpTo(4); err != io.ErrShortBuffer {
			t.Fatal(err)
		}
	})

	t.Run("all", func(t *testing.T) {
		buf := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8}
		bra := bytes.NewReader(buf)
		bs := NewReaderAtByteStream(bra, make([]byte, len(buf)+1))
		assertPeekUpTo(t, bs, len(buf), buf...)
	})

	t.Run("more-than-buf", func(t *testing.T) {
		buf := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8}
		bra := bytes.NewReader(buf)
		bs := NewReaderAtByteStream(bra, make([]byte, len(buf)+1))
		assertPeekUpTo(t, bs, len(buf)+1, buf...)
	})
}

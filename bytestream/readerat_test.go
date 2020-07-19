package bytestream

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

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

func assertTakeExactly(t *testing.T, bs *ReaderAt, n int, left int64, exp []byte) {
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

func assertTakeUpTo(t *testing.T, bs *ReaderAt, n int, left int64, exp []byte) {
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

func assertPeekUpTo(t *testing.T, bs *ReaderAt, n int, exp ...byte) {
	t.Helper()
	v, err := bs.PeekUpTo(n)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, exp) {
		t.Fatal(v, exp)
	}
}

func assertTakeExactlyEOF(t *testing.T, bs *ReaderAt, n int) {
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

func TestReaderAtTakeExactly(t *testing.T) {
	bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
	bs := NewReaderAt(bra, make([]byte, 3))
	assertTakeExactly(t, bs, 2, 1, []byte{0, 1})
	assertTakeExactly(t, bs, 2, 1, []byte{2, 3})
	assertTakeExactly(t, bs, 2, 1, []byte{4, 5})
	assertTakeExactly(t, bs, 2, 1, []byte{6, 7})
	assertTakeExactly(t, bs, 1, 0, []byte{8})
	if _, err := bs.TakeExactly(1); err != io.EOF {
		t.Fatal(err)
	}
}

func TestReaderAtTakeUpToExact(t *testing.T) {
	bra := bytes.NewReader([]byte{0, 1, 2, 3})
	bs := NewReaderAt(bra, nil)
	assertTakeUpTo(t, bs, 3, -1, []byte{0, 1, 2})
	assertTakeUpTo(t, bs, 1, -1, []byte{3})
	if _, err := bs.TakeUpTo(1); err != io.EOF {
		t.Fatal(err)
	}
}

func TestReaderAtTakeUpToTooMany(t *testing.T) {
	bra := bytes.NewReader([]byte{0, 1, 2, 3})
	bs := NewReaderAt(bra, nil)
	assertTakeUpTo(t, bs, 3, -1, []byte{0, 1, 2})
	assertTakeUpTo(t, bs, 2, -1, []byte{3})
	if _, err := bs.TakeUpTo(1); err != io.EOF {
		t.Fatal(err)
	}
}

func TestReaderAtTakeExactlyOneUntilEnd(t *testing.T) {
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
					bs := NewReaderAt(tc.rdr, make([]byte, bufSize))

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

func TestReaderAtTakeUpToOneUntilEnd(t *testing.T) {
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
					bs := NewReaderAt(tc.rdr, make([]byte, bufSize))

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

func TestReaderAtDiscardExactlyOneUntilEnd(t *testing.T) {
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
					bs := NewReaderAt(tc.rdr, make([]byte, bufSize))
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

func TestReaderAtTakeExactlyFailsWhenTakingTooMuch(t *testing.T) {
	bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
	bs := NewReaderAt(bra, make([]byte, 3))
	if _, err := bs.TakeExactly(5); err != io.ErrShortBuffer {
		t.Fatal(err)
	}
}

func TestReaderAtTakeExactlySucceedsAfterTakeExactlyFailsWithShortBuffer(t *testing.T) {
	bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
	bs := NewReaderAt(bra, make([]byte, 3))
	assertTakeExactly(t, bs, 1, 2, []byte{0})
	if _, err := bs.TakeExactly(5); err != io.ErrShortBuffer {
		t.Fatal(err)
	}
	assertTakeExactly(t, bs, 1, 2, []byte{1})
}

func TestReaderAtDiscardExactly(t *testing.T) {
	t.Run("some", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAt(bra, make([]byte, 3))
		if err := bs.DiscardExactly(5); err != nil {
			t.Fatal(err)
		}
		assertTakeExactly(t, bs, 2, 0, []byte{5, 6})
	})

	t.Run("discard-take-discard", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4})
		bs := NewReaderAt(bra, make([]byte, 3))
		assertTakeExactly(t, bs, 2, -1, []byte{0, 1})
		if err := bs.DiscardExactly(2); err != nil {
			t.Fatal(err)
		}
		assertTakeExactly(t, bs, 1, -1, []byte{4})
	})

	t.Run("all-but-last", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAt(bra, make([]byte, 3))
		if err := bs.DiscardExactly(8); err != nil {
			t.Fatal()
		}
		assertTakeExactly(t, bs, 1, 0, []byte{8})
	})

	t.Run("all-aligned", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAt(bra, make([]byte, 3))
		if err := bs.DiscardExactly(9); err != nil {
			t.Fatal(err)
		}
		assertTakeExactlyEOF(t, bs, 1)
	})

	t.Run("all-misaligned", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAt(bra, make([]byte, 2))
		if err := bs.DiscardExactly(9); err != nil {
			t.Fatal(err)
		}
		assertTakeExactlyEOF(t, bs, 1)
	})

	t.Run("too-many", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAt(bra, make([]byte, 3))
		if err := bs.DiscardExactly(10); err != io.EOF {
			t.Fatal(err)
		}
	})

	t.Run("too-many-after-take", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAt(bra, make([]byte, 5))

		assertTakeExactly(t, bs, 5, -1, []byte{0, 1, 2, 3, 4})
		if err := bs.DiscardExactly(5); err != io.EOF {
			t.Fatal(err)
		}
	})
}

func TestReaderAtDiscardUpTo(t *testing.T) {
	t.Run("some", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAt(bra, make([]byte, 3))
		if err := bs.DiscardUpTo(5); err != nil {
			t.Fatal(err)
		}
		assertTakeExactly(t, bs, 2, 1, []byte{5, 6})
	})

	t.Run("all-but-last", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAt(bra, make([]byte, 3))
		if err := bs.DiscardUpTo(8); err != nil {
			t.Fatal()
		}
		assertTakeExactly(t, bs, 1, 0, []byte{8})
	})

	t.Run("all-aligned", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAt(bra, make([]byte, 3))
		if err := bs.DiscardUpTo(9); err != nil {
			t.Fatal(err)
		}
		assertTakeExactlyEOF(t, bs, 1)
	})

	t.Run("all-misaligned", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAt(bra, make([]byte, 2))
		if err := bs.DiscardUpTo(9); err != nil {
			t.Fatal(err)
		}
		assertTakeExactlyEOF(t, bs, 1)
	})

	t.Run("too-many", func(t *testing.T) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAt(bra, make([]byte, 3))
		if err := bs.DiscardUpTo(10); err != nil {
			t.Fatal()
		}
		assertTakeExactlyEOF(t, bs, 1)
	})
}

var BenchResultBytes []byte

func BenchmarkReaderAt(b *testing.B) {
	b.Run("base", func(b *testing.B) {
		buf := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8}
		bra := bytes.NewReader(buf)
		bs := NewReaderAt(bra, make([]byte, 3))

		for i := 0; i < b.N; i++ {
			bs.rem = bs.buf[:2]
			BenchResultBytes = buf[:2]
		}
	})

	b.Run("fast", func(b *testing.B) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAt(bra, make([]byte, 3))
		bs.TakeExactly(9) // seed

		for i := 0; i < b.N; i++ {
			bs.rem = bs.buf[:2]
			bs.TakeExactly(2)
		}
	})

	b.Run("slow", func(b *testing.B) {
		bra := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
		bs := NewReaderAt(bra, make([]byte, 3))

		for i := 0; i < b.N; i++ {
			bs.rem = bs.buf[:0]
			bs.TakeExactly(2)
		}
	})
}

func TestReaderAtPeekUpTo(t *testing.T) {
	t.Run("within/bsz=1", func(t *testing.T) {
		buf := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8}
		bra := bytes.NewReader(buf)
		bs := NewReaderAt(bra, make([]byte, 1))
		assertPeekUpTo(t, bs, 1, 0)
		if _, err := bs.PeekUpTo(2); err != io.ErrShortBuffer {
			t.Fatal(err)
		}
	})

	t.Run("within/bsz=3", func(t *testing.T) {
		buf := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8}
		bra := bytes.NewReader(buf)
		bs := NewReaderAt(bra, make([]byte, 3))
		assertPeekUpTo(t, bs, 1, 0)
		assertPeekUpTo(t, bs, 2, 0, 1)
		assertPeekUpTo(t, bs, 3, 0, 1, 2)
		if _, err := bs.PeekUpTo(4); err != io.ErrShortBuffer {
			t.Fatal(err)
		}
	})

	t.Run("second-peek-spans-eof", func(t *testing.T) {
		buf := []byte{0, 1, 2, 3}
		bra := bytes.NewReader(buf)
		bs := NewReaderAt(bra, make([]byte, 4))
		assertPeekUpTo(t, bs, 2, 0, 1)
		bs.DiscardExactly(2)
		assertPeekUpTo(t, bs, 3, 2, 3)
	})

	t.Run("all", func(t *testing.T) {
		buf := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8}
		bra := bytes.NewReader(buf)
		bs := NewReaderAt(bra, make([]byte, len(buf)+1))
		assertPeekUpTo(t, bs, len(buf), buf...)
	})

	t.Run("more-than-buf", func(t *testing.T) {
		buf := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8}
		bra := bytes.NewReader(buf)
		bs := NewReaderAt(bra, make([]byte, len(buf)+1))
		assertPeekUpTo(t, bs, len(buf)+1, buf...)
	})
}

package iotools

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/shabbyrobe/golib/assert"
)

func TestCommitReader(t *testing.T) {
	var (
		n    int
		err  error
		into = make([]byte, 2)
	)

	tt := assert.WrapTB(t)

	in := "1234567890"
	rdr := strings.NewReader(in)
	cr := NewCommitReader(rdr)

	{ // Read and rewind
		for i := 0; i < 3; i++ {
			n, err = cr.Read(into)
			tt.MustOK(err)
			tt.MustEqual(2, n)
			tt.MustEqual("12", string(into))
			cr.Rewind()
		}
	}

	{ // Advance
		cr.Advance(1)
		for i := 0; i < 3; i++ {
			mustRead(tt, cr, into, 2)
			tt.MustEqual("23", string(into))
			cr.Rewind()
		}
		cr.Advance(1)
		for i := 0; i < 3; i++ {
			mustRead(tt, cr, into, 2)
			tt.MustEqual("34", string(into))
			cr.Rewind()
		}
	}
}

func TestCommitReaderRest(t *testing.T) {
	tt := assert.WrapTB(t)

	{
		ir := bytes.NewReader([]byte{'a', 'b', 'c'})
		cr := NewCommitReaderSize(ir, 2)
		x := make([]byte, 1)
		cr.Read(x)
		rest := cr.Rest()
		out, err := ioutil.ReadAll(rest)
		tt.MustOK(err)
		tt.MustEqual([]byte{'b', 'c'}, out)
	}

	{
		ir := bytes.NewReader([]byte{'a', 'b', 'c'})
		cr := NewCommitReaderSize(ir, 3)
		x := make([]byte, 1)
		cr.Read(x)
		rest := cr.Rest()
		out, err := ioutil.ReadAll(rest)
		tt.MustOK(err)
		tt.MustEqual([]byte{'b', 'c'}, out)
	}

	{
		ir := bytes.NewReader([]byte{'a', 'b', 'c'})
		cr := NewCommitReaderSize(ir, 3)
		rest := cr.Rest()
		out, err := ioutil.ReadAll(rest)
		tt.MustOK(err)
		tt.MustEqual([]byte{'a', 'b', 'c'}, out)
	}

	{
		ir := bytes.NewReader([]byte{'a', 'b', 'c'})
		cr := NewCommitReaderSize(ir, 2)
		x := make([]byte, 3)
		cr.Read(x)
		rest := cr.Rest()
		out, err := ioutil.ReadAll(rest)
		tt.MustOK(err)
		tt.MustEqual([]byte{'c'}, out)
	}
}

func TestCommitReaderCommit(t *testing.T) { // Commit
	tt := assert.WrapTB(t)

	in := "1234567890"
	rdr := strings.NewReader(in)
	cr := NewCommitReader(rdr)
	into := make([]byte, 2)

	for i := 0; i < len(in); i += 2 {
		mustRead(tt, cr, into, 2)
		tt.MustEqual(in[i:i+2], string(into))
		cr.Commit()
	}
}

func mustRead(tt assert.T, rdr io.Reader, into []byte, n int) {
	tt.Helper()
	n, err := rdr.Read(into)
	tt.MustOK(err)
	tt.MustEqual(n, n)
}

type testingReader struct {
	bts []byte
	pos int
	max int
}

func (t *testingReader) Read(b []byte) (n int, err error) {
	left := len(t.bts) - t.pos
	if left == 0 {
		err = io.EOF
		return
	}
	n = len(b)
	if n > t.max {
		n = t.max
	}
	if left < n {
		n = left
	}
	copy(b, t.bts[t.pos:t.pos+n])
	t.pos += n
	return
}

func TestCommitReaderMultipleReads(t *testing.T) { // Commit
	tt := assert.WrapTB(t)

	in := "1234567890"
	rdr := &testingReader{bts: []byte(in), max: 3}

	cr := NewCommitReader(rdr)
	into := make([]byte, 16)

	mustRead(tt, cr, into, 3)
	tt.MustEqual(in[0:3], string(into[:3]))

	mustRead(tt, cr, into, 3)
	tt.MustEqual(in[3:6], string(into[:3]))

	mustRead(tt, cr, into, 3)
	tt.MustEqual(in[6:9], string(into[:3]))

	mustRead(tt, cr, into, 1)
	tt.MustEqual(in[9:10], string(into[:1]))

	n, err := cr.Read(into)
	tt.MustEqual(io.EOF, err)
	tt.MustEqual(0, n)

	cr.Rewind()

	mustRead(tt, cr, into, 3)
	tt.MustEqual(in[0:3], string(into[:3]))
}

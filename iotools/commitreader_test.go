package iotools

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

func TestCommitReader(t *testing.T) {
	var (
		n    int
		err  error
		into = make([]byte, 2)
	)

	in := "1234567890"
	rdr := strings.NewReader(in)
	cr := NewCommitReader(rdr)

	{ // Read and rewind
		for i := 0; i < 3; i++ {
			n, err = cr.Read(into)
			if err != nil {
				t.Fatal(err)
			}
			if n != 2 {
				t.Fatal(n)
			}
			if string(into) != "12" {
				t.Fatal(string(into))
			}
			cr.Rewind()
		}
	}

	{ // Advance
		cr.Advance(1)
		for i := 0; i < 3; i++ {
			mustRead(t, cr, into, 2)
			if string(into) != "23" {
				t.Fatal(string(into))
			}
			cr.Rewind()
		}
		cr.Advance(1)
		for i := 0; i < 3; i++ {
			mustRead(t, cr, into, 2)
			if string(into) != "34" {
				t.Fatal(string(into))
			}
			cr.Rewind()
		}
	}
}

func TestCommitReaderRest(t *testing.T) {
	{
		ir := bytes.NewReader([]byte{'a', 'b', 'c'})
		cr := NewCommitReaderSize(ir, 2)
		x := make([]byte, 1)
		cr.Read(x)
		rest := cr.Rest()
		out, err := ioutil.ReadAll(rest)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal([]byte{'b', 'c'}, out) {
			t.Fatal(out)
		}
	}

	{
		ir := bytes.NewReader([]byte{'a', 'b', 'c'})
		cr := NewCommitReaderSize(ir, 3)
		x := make([]byte, 1)
		cr.Read(x)
		rest := cr.Rest()
		out, err := ioutil.ReadAll(rest)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal([]byte{'b', 'c'}, out) {
			t.Fatal(out)
		}
	}

	{
		ir := bytes.NewReader([]byte{'a', 'b', 'c'})
		cr := NewCommitReaderSize(ir, 3)
		rest := cr.Rest()
		out, err := ioutil.ReadAll(rest)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal([]byte{'a', 'b', 'c'}, out) {
			t.Fatal(out)
		}
	}

	{
		ir := bytes.NewReader([]byte{'a', 'b', 'c'})
		cr := NewCommitReaderSize(ir, 2)
		x := make([]byte, 3)
		cr.Read(x)
		rest := cr.Rest()
		out, err := ioutil.ReadAll(rest)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal([]byte{'c'}, out) {
			t.Fatal(out)
		}
	}
}

func TestCommitReaderCommit(t *testing.T) { // Commit
	in := "1234567890"
	rdr := strings.NewReader(in)
	cr := NewCommitReader(rdr)
	into := make([]byte, 2)

	for i := 0; i < len(in); i += 2 {
		mustRead(t, cr, into, 2)
		part := in[i : i+2]
		if part != string(into) {
			t.Fatal(part, "!=", string(into))
		}
		cr.Commit()
	}
}

func mustRead(t testing.TB, rdr io.Reader, into []byte, n int) {
	t.Helper()
	rn, err := rdr.Read(into)
	if err != nil {
		t.Fatal(err)
	}
	if rn != n {
		t.Fatal(rn, "!=", n)
	}
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
	in := "1234567890"
	rdr := &testingReader{bts: []byte(in), max: 3}

	cr := NewCommitReader(rdr)
	into := make([]byte, 16)

	mustRead(t, cr, into, 3)
	if in[0:3] != string(into[:3]) {
		t.Fatal()
	}

	mustRead(t, cr, into, 3)
	if in[3:6] != string(into[:3]) {
		t.Fatal()
	}

	mustRead(t, cr, into, 3)
	if in[6:9] != string(into[:3]) {
		t.Fatal()
	}

	mustRead(t, cr, into, 1)
	if in[9:10] != string(into[:1]) {
		t.Fatal()
	}

	n, err := cr.Read(into)
	if err != io.EOF || n != 0 {
		t.Fatal()
	}

	cr.Rewind()

	mustRead(t, cr, into, 10)
	if in[0:3] != string(into[:3]) {
		t.Fatal()
	}
}

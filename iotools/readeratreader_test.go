package iotools

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/shabbyrobe/golib/assert"
)

func TestReaderAtReaderEOFAssumption(t *testing.T) {
	tt := assert.WrapTB(t)
	tmp, err := ioutil.TempFile("", "")
	tt.MustOK(err)
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	tt.MustOKAll(tmp.WriteAt([]byte{1, 2}, 0))

	read := make([]byte, 2)
	tt.MustOKAll(tmp.ReadAt(read, 0))
	tt.MustEqual([]byte{1, 2}, read)

	n, err := tmp.ReadAt(read, 1)
	tt.MustEqual(1, n)
	tt.MustEqual(io.EOF, err)
}

func TestReaderAtReader(t *testing.T) {
	tt := assert.WrapTB(t)
	tmp, err := ioutil.TempFile("", "")
	tt.MustOK(err)
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	data := []byte{1, 2, 3, 4, 5}
	tt.MustOKAll(tmp.WriteAt(data, 0))

	result, err := ioutil.ReadAll(NewReaderAtReader(tmp, 0))
	tt.MustOK(err)
	tt.MustEqual(data, result)
}

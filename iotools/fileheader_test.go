package iotools

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/shabbyrobe/golib/assert"
	"github.com/shabbyrobe/golib/iotools/bytewriter"
)

func TestFileHeaderWrite(t *testing.T) {
	tt := assert.WrapTB(t)
	_ = tt

	var hdr = []byte{'h', 'd', 'r', '!'}
	var buf = make([]byte, 0, 1024)
	var bw bytewriter.Writer
	bw.Give(buf)

	n, err := FileHeaderWrite("abcd", &bw, hdr)
	tt.MustOK(err)

	exp := []byte{
		'a', 'b', 'c', 'd', // magic
		0x4, 0x0, 0x0, 0x0, // uint32le hdr length
		'h', 'd', 'r', '!', // hdr bytes
	}
	tt.MustEqual(exp, buf[:n])
}

func TestFileHeaderRead(t *testing.T) {
	tt := assert.WrapTB(t)

	exp := []byte{
		'a', 'b', 'c', 'd', // magic
		0x4, 0x0, 0x0, 0x0, // uint32le hdr length
		'h', 'd', 'r', '!', // hdr bytes
		'b', 'o', 'd', 'y', // file body
	}

	rdr := bytes.NewReader(exp)

	hdr, n, err := FileHeaderRead("abcd", rdr)
	tt.MustOK(err)

	tt.MustEqual(12, n)
	tt.MustEqual(exp[8:12], hdr)

	body, err := ioutil.ReadAll(rdr)
	tt.MustOK(err)

	tt.MustEqual(exp[12:], body)
}

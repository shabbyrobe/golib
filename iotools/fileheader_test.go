package iotools

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/shabbyrobe/golib/iotools/bytewriter"
)

func TestFileHeaderWrite(t *testing.T) {
	var hdr = []byte{'h', 'd', 'r', '!'}
	var buf = make([]byte, 0, 1024)
	var bw bytewriter.Writer
	bw.Give(buf)

	n, err := FileHeaderWrite("abcd", &bw, hdr)
	if err != nil {
		t.Fatal(err)
	}

	exp := []byte{
		'a', 'b', 'c', 'd', // magic
		0x4, 0x0, 0x0, 0x0, // uint32le hdr length
		'h', 'd', 'r', '!', // hdr bytes
	}
	if !reflect.DeepEqual(exp, buf[:n]) {
		t.Fatal(exp, "!=", buf[:n])
	}
}

func TestFileHeaderRead(t *testing.T) {
	exp := []byte{
		'a', 'b', 'c', 'd', // magic
		0x4, 0x0, 0x0, 0x0, // uint32le hdr length
		'h', 'd', 'r', '!', // hdr bytes
		'b', 'o', 'd', 'y', // file body
	}

	rdr := bytes.NewReader(exp)

	hdr, n, err := FileHeaderRead("abcd", rdr)
	if err != nil {
		t.Fatal(err)
	}
	if 12 != n {
		t.Fatal()
	}
	if !reflect.DeepEqual(exp[8:12], hdr) {
		t.Fatal(exp[8:12], "!=", hdr)
	}

	body, err := ioutil.ReadAll(rdr)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(exp[12:], body) {
		t.Fatal(exp[12:], "!=", body)
	}
}

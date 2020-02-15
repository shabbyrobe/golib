package iotools

import (
	"bytes"
	"crypto/rand"
	"io"
	"io/ioutil"
	"testing"

	"github.com/shabbyrobe/golib/assert"
)

func TestLeadingReader(t *testing.T) {
	tt := assert.WrapTB(t)

	lr := NewLeadingReader([]byte{'a'}, bytes.NewReader([]byte{'b'}))
	all, err := ioutil.ReadAll(lr)
	tt.MustOK(err)
	tt.MustEqual([]byte{'a', 'b'}, all)
}

func TestLeadingReaderEmptyLeading(t *testing.T) {
	tt := assert.WrapTB(t)

	lr := NewLeadingReader([]byte{}, bytes.NewReader([]byte{'b'}))
	all, err := ioutil.ReadAll(lr)
	tt.MustOK(err)
	tt.MustEqual([]byte{'b'}, all)
}

func TestLeadingReaderEmptyReader(t *testing.T) {
	tt := assert.WrapTB(t)

	lr := NewLeadingReader([]byte{'a'}, bytes.NewReader([]byte{}))
	all, err := ioutil.ReadAll(lr)
	tt.MustOK(err)
	tt.MustEqual([]byte{'a'}, all)
}

var BenchBytesResult []byte

func BenchmarkLeadingReader(b *testing.B) {
	const total = 32768
	d1 := make([]byte, 16384)
	d2 := make([]byte, 16384)
	rand.Read(d1)
	rand.Read(d2)

	var err error
	for i := 0; i < b.N; i++ {
		lr := NewLeadingReader(d1, bytes.NewReader(d2))
		BenchBytesResult, err = ioutil.ReadAll(lr)
		if err != nil {
			panic(err)
		}
		if len(BenchBytesResult) != total {
			panic(len(BenchBytesResult))
		}
	}
}

func BenchmarkMultiReader(b *testing.B) {
	const total = 32768
	d1 := make([]byte, 16384)
	d2 := make([]byte, 16384)
	rand.Read(d1)
	rand.Read(d2)

	var err error
	for i := 0; i < b.N; i++ {
		lr := io.MultiReader(bytes.NewReader(d1), bytes.NewReader(d2))
		BenchBytesResult, err = ioutil.ReadAll(lr)
		if err != nil {
			panic(err)
		}
		if len(BenchBytesResult) != total {
			panic(len(BenchBytesResult))
		}
	}
}

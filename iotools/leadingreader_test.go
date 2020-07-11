package iotools

import (
	"bytes"
	"crypto/rand"
	"io"
	"io/ioutil"
	"testing"
)

func TestLeadingReader(t *testing.T) {
	lr := NewLeadingReader([]byte{'a'}, bytes.NewReader([]byte{'b'}))
	all, err := ioutil.ReadAll(lr)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal([]byte{'a', 'b'}, all) {
		t.Fatal()
	}
}

func TestLeadingReaderEmptyLeading(t *testing.T) {
	lr := NewLeadingReader([]byte{}, bytes.NewReader([]byte{'b'}))
	all, err := ioutil.ReadAll(lr)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal([]byte{'b'}, all) {
		t.Fatal()
	}
}

func TestLeadingReaderEmptyReader(t *testing.T) {
	lr := NewLeadingReader([]byte{'a'}, bytes.NewReader([]byte{}))
	all, err := ioutil.ReadAll(lr)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal([]byte{'a'}, all) {
		t.Fatal()
	}
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

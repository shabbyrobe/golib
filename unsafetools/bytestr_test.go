package unsafetools

import (
	"reflect"
	"testing"
)

func TestString(t *testing.T) {
	if !reflect.DeepEqual(String([]byte("foo")), "foo") {
		t.Fatal()
	}
}

func TestStringEmpty(t *testing.T) {
	if !reflect.DeepEqual(String([]byte("")), "") {
		t.Fatal()
	}
}

func TestStringNil(t *testing.T) {
	if !reflect.DeepEqual(String(nil), "") {
		t.Fatal()
	}
}

func TestBytes(t *testing.T) {
	if !reflect.DeepEqual(Bytes("foo"), []byte("foo")) {
		t.Fatal()
	}
}

func TestBytesEmpty(t *testing.T) {
	var result []byte = nil
	if !reflect.DeepEqual(Bytes(""), result) {
		t.Fatal()
	}
	if len(Bytes("")) != 0 {
		t.Fatal()
	}
}

var BenchBytes []byte

func BenchmarkBytes(b *testing.B) {
	b.ReportAllocs()
	in := "foo bar baz"
	for i := 0; i < b.N; i++ {
		BenchBytes = Bytes(in)
	}
}

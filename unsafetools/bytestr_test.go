package unsafetools

import (
	"reflect"
	"testing"
)

func TestBytes(t *testing.T) {
	if !reflect.DeepEqual(Bytes("foo"), []byte("foo")) {
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

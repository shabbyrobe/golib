package bytewriter

import (
	"testing"
)

func TestByteWriterAppend(t *testing.T) {
	var bw Writer
	{
		bw.Write([]byte("abc"))
		bw.Write([]byte("def"))
		bw.Write([]byte("ghi"))
		buf, n := bw.Take()
		if n != 9 {
			t.Fatal()
		}
		if string(buf) != "abcdefghi" {
			t.Fatal()
		}
	}

	{ // the writer should be reset if you take the buffer
		bw.Write([]byte("abc"))
		buf, n := bw.Take()
		if n != 3 {
			t.Fatal()
		}
		if string(buf) != "abc" {
			t.Fatal()
		}
	}
}

func TestByteWriterSet(t *testing.T) {
	var bw Writer
	bw.Give(make([]byte, 0, 1024))
	bw.Write([]byte("abc"))
	bw.Write([]byte("def"))
	bw.Write([]byte("ghi"))
	buf, n := bw.Take()
	if n != 9 {
		t.Fatal()
	}
	if string(buf) != "abcdefghi" {
		t.Fatal()
	}
	if cap(buf) != 1024 {
		t.Fatal()
	}
	if len(buf) != 9 {
		t.Fatal()
	}
}

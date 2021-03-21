package iotools

import (
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestReaderAtReaderEOFAssumption(t *testing.T) {
	tmp, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := tmp.WriteAt([]byte{1, 2}, 0); err != nil {
		t.Fatal(err)
	}

	read := make([]byte, 2)
	if _, err := tmp.ReadAt(read, 0); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual([]byte{1, 2}, read) {
		t.Fatal()
	}

	n, err := tmp.ReadAt(read, 1)
	if n != 1 {
		t.Fatal()
	}
	if err != io.EOF {
		t.Fatal()
	}
}

func TestReaderAtReader(t *testing.T) {
	tmp, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	data := []byte{1, 2, 3, 4, 5}
	if _, err := tmp.WriteAt(data, 0); err != nil {
		t.Fatal(err)
	}

	result, err := ioutil.ReadAll(NewReaderAtReader(tmp, 0))
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(data, result) {
		t.Fatal()
	}
}

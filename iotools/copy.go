package iotools

import (
	"fmt"
	"io"
	"os"

	"github.com/shabbyrobe/golib/errtools"
)

func CopyFile(from, to string) (rerr error) {
	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer in.Close()

	st, err := in.Stat()
	if err != nil {
		return err
	}

	expected := st.Size()

	out, err := os.Create(to)
	if err != nil {
		return err
	}
	defer errtools.DeferClose(&rerr, out)

	n, err := io.Copy(out, in)
	if err != nil {
		return err
	}

	if n != expected {
		return fmt.Errorf("iotools: copy expected %d bytes, but only copied %d", expected, n)
	}

	return nil
}

// MoveFile copies a file from one location to another, then removes the
// original if the copy succeeded.
//
// It is useful for situations where it is known that os.Rename would produce
// an "invalid cross-device link" error
func MoveFile(from, to string) (rerr error) {
	if err := CopyFile(from, to); err != nil {
		return err
	}
	if err := os.Remove(from); err != nil {
		return err
	}
	return nil
}

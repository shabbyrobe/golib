package iotools

import (
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

	out, err := os.Create(to)
	if err != nil {
		return err
	}
	defer errtools.DeferClose(&rerr, out)

	_, err = io.Copy(out, in)
	return err
}

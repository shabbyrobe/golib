package iotools

import (
	"fmt"
	"io"
	"os"
)

type WriterAtPretender struct {
	writer io.Writer
	pos    int64
}

var _ io.Writer = &WriterAtPretender{}
var _ io.WriterAt = &WriterAtPretender{}

var writeNull8k = make([]byte, 8192)

func PretendWriterAt(w io.Writer) *WriterAtPretender {
	return &WriterAtPretender{writer: w}
}

func (w *WriterAtPretender) WriteAt(p []byte, off int64) (n int, err error) {
	fmt.Fprintln(os.Stderr, w.pos, off)
	if off < w.pos {
		return 0, fmt.Errorf("iotools: expected write offset >=%d, found %d", w.pos, off)

	} else if off > w.pos {
		// Write zeroes until we get to offset
		gap := off - w.pos
		fmt.Fprintln(os.Stderr, "GAP", gap)

		for gap > 0 {
			end := int64(8192)
			if gap < end {
				end = gap
			}
			_, err := w.Write(writeNull8k[:end])
			if err != nil {
				return 0, err
			}
			gap -= end
		}
	}

	n, err = w.writer.Write(p)
	w.pos += int64(n)
	return n, err
}

func (w *WriterAtPretender) Write(p []byte) (n int, err error) {
	n, err = w.writer.Write(p)
	w.pos += int64(n)
	return n, err
}

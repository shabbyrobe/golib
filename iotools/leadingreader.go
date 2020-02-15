package iotools

import (
	"bytes"
	"io"
)

// LeadingReader is an io.Reader that proxies an existing reader after first
// reading from a fixed byte slice.
//
// Deprecated: you'd probably be better off using io.MultiReader instead:
//	io.MultiReader(bytes.NewReader(leading), rdr)
//
type LeadingReader struct {
	leading     io.Reader
	leadingDone bool
	reader      io.Reader
}

func NewLeadingReader(leading []byte, rdr io.Reader) *LeadingReader {
	return &LeadingReader{
		leading: bytes.NewReader(leading),
		reader:  rdr,
	}
}

func (l *LeadingReader) Read(p []byte) (n int, err error) {
	if l.leadingDone {
		return l.reader.Read(p)
	} else {
		n, err = l.leading.Read(p)
		if err == io.EOF {
			err = nil
			l.leadingDone = true
			n, err = l.reader.Read(p)
		}
		return
	}
}

func (l *LeadingReader) Close() error {
	if c, ok := l.reader.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

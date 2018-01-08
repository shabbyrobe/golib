package iotools

import "io"

// DummyReadCloser wraps an io.Reader with a closer that doesn't do anything, so
// your io.Reader can satisfy io.ReadCloser.
type DummyReadCloser struct {
	io.Reader
}

func (d DummyReadCloser) Close() error { return nil }

// DummyWriteCloser wraps an io.Writer with a closer that doesn't do anything, so
// your io.Writer can satisfy io.WriteCloser.
type DummyWriteCloser struct {
	io.Writer
}

func (d DummyWriteCloser) Close() error { return nil }

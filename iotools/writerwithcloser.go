package iotools

import "io"

func writerNullCloser() error { return nil }

type WriterWithCloser struct {
	writer io.Writer
	closer func() error
}

func NewWriterWithCloser(wr io.Writer, closer func() error) *WriterWithCloser {
	if closer == nil {
		wc, ok := wr.(io.Closer)
		if ok {
			closer = wc.Close
		} else {
			closer = writerNullCloser
		}
	}

	return &WriterWithCloser{wr, closer}
}

func (wwc *WriterWithCloser) Write(b []byte) (n int, err error) {
	return wwc.writer.Write(b)
}

func (wwc *WriterWithCloser) Close() error {
	return wwc.closer()
}

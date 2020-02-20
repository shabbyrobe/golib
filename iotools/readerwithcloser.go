package iotools

import "io"

func readerNullCloser() error { return nil }

type ReaderWithCloser struct {
	reader io.Reader
	closer func() error
}

func NewReaderWithCloser(rdr io.Reader, closer func() error) *ReaderWithCloser {
	if closer == nil {
		rc, ok := rdr.(io.Closer)
		if ok {
			closer = rc.Close
		} else {
			closer = readerNullCloser
		}
	}

	return &ReaderWithCloser{rdr, closer}
}

func (rwc *ReaderWithCloser) Read(b []byte) (n int, err error) {
	return rwc.reader.Read(b)
}

func (rwc *ReaderWithCloser) Close() error {
	return rwc.closer()
}

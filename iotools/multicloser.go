package iotools

import "io"

type MultiCloser struct {
	closers []io.Closer
}

func NewMultiCloser(closers ...io.Closer) *MultiCloser {
	return &MultiCloser{
		closers: closers,
	}
}

func (mc *MultiCloser) Close() (rerr error) {
	for _, cls := range mc.closers {
		if cerr := cls.Close(); cerr != nil && rerr == nil {
			rerr = cerr
		}
	}
	return rerr
}

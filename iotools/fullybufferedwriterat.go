package iotools

import (
	"fmt"
	"io"
	"io/ioutil"
)

type FullyBufferedWriterAt struct {
	closed   bool
	buffer   []byte
	buffered bool
	len      int64
	from     io.ReadSeeker
	to       FullyBufferedWriterDestination
}

type FullyBufferedWriterDestination interface {
	io.Writer
	io.Closer
	Truncate(sz int64) error
}

func NewFullyBufferedWriterAt(from io.ReadSeeker, to FullyBufferedWriterDestination) *FullyBufferedWriterAt {
	return &FullyBufferedWriterAt{
		from: from,
		to:   to,
	}
}

func (w *FullyBufferedWriterAt) WriteAt(p []byte, offset int64) (n int, err error) {
	if !w.buffered {
		if err := w.Refresh(); err != nil {
			return 0, err
		}
	}

	plen64 := int64(len(p))
	if offset+plen64 > w.len {
		w.buffer = append(w.buffer, make([]byte, offset+plen64-w.len)...)
		w.len = offset + plen64
	}

	copy(w.buffer[offset:], p)
	return int(plen64), nil
}

func (w *FullyBufferedWriterAt) ReadAt(p []byte, off int64) (n int, err error) {
	if !w.buffered {
		if err := w.Refresh(); err != nil {
			return 0, err
		}
	}

	if off >= w.len {
		return 0, io.EOF
	}
	n = copy(p, w.buffer[off:])
	return n, nil
}

func (w *FullyBufferedWriterAt) Close() (err error) {
	if w.closed {
		return errAlreadyClosed(1)
	}
	w.closed = true
	err = w.Flush()
	cerr := w.to.Close()
	if err == nil {
		err = cerr
	}
	return err
}

func (w *FullyBufferedWriterAt) Truncate(sz int64) (err error) {
	if !w.buffered {
		if err := w.Refresh(); err != nil {
			return err
		}
	}
	if sz > w.len {
		return fmt.Errorf("iotools: truncate %d greater than buffer length %d", sz, w.len)
	}

	w.buffer = w.buffer[:sz]
	w.len = sz
	return nil
}

func (w *FullyBufferedWriterAt) Flush() (err error) {
	if w.buffered {
		if err := w.to.Truncate(0); err != nil {
			return err
		}
		_, err = w.to.Write(w.buffer)
	}
	return err
}

func (w *FullyBufferedWriterAt) Refresh() (err error) {
	if _, err := w.from.Seek(0, io.SeekStart); err != nil {
		return err
	}

	w.buffer, err = ioutil.ReadAll(w.from)
	if err != nil {
		return err
	}

	w.len = int64(len(w.buffer))
	w.buffered = true
	return nil
}

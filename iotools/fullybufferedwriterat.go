package iotools

import (
	"io"
	"io/ioutil"
)

type FullyBufferedWriterAt struct {
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

func (w *FullyBufferedWriterAt) Close() (err error) {
	return w.to.Close()
}

func (w *FullyBufferedWriterAt) Flush() (err error) {
	if err := w.to.Truncate(0); err != nil {
		return err
	}
	_, err = w.to.Write(w.buffer)
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

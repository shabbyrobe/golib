package bytetools

import (
	"bytes"
	"fmt"
	"io"
)

// BufferAt is a buffer that supports ReadAt and WriteAt.
type BufferAt struct {
	buffer []byte
	len    int64
}

func NewBufferAt(initial []byte) *BufferAt {
	return &BufferAt{
		buffer: initial,
		len:    int64(len(initial)),
	}
}

// Bytes returns a slice of length b.Len() holding the contents of the buffer.
// The slice is valid for use only until the next buffer modification (that is,
// only until the next call to a method like WriteAt, or Truncate).
// The slice aliases the buffer content at least until the next buffer modification,
// so immediate changes to the slice will affect the result of future reads.
func (w *BufferAt) Bytes() []byte {
	return w.buffer
}

// AsReader returns a reader for the underlying byte slice. You should not
// write to the buffer while a Reader returned by AsReader is active.
func (w *BufferAt) AsReader() io.Reader {
	return bytes.NewReader(w.buffer)
}

func (w *BufferAt) WriteAt(p []byte, offset int64) (n int, err error) {
	plen64 := int64(len(p))
	if offset+plen64 > w.len {
		w.buffer = append(w.buffer, make([]byte, offset+plen64-w.len)...)
		w.len = offset + plen64
	}

	copy(w.buffer[offset:], p)
	return int(plen64), nil
}

func (w *BufferAt) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= w.len {
		return 0, io.EOF
	}
	n = copy(p, w.buffer[off:])
	return n, nil
}

func (w *BufferAt) Truncate(sz int64) (err error) {
	if sz > w.len {
		return fmt.Errorf("iotools: truncate %d greater than buffer length %d", sz, w.len)
	}

	w.buffer = w.buffer[:sz]
	w.len = sz
	return nil
}

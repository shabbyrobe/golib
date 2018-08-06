package bytewriter

type Writer struct {
	buf []byte
	len int
}

func (w *Writer) Clear() {
	w.buf = w.buf[:0]
	w.len = 0
}

func (w *Writer) Give(buf []byte) {
	w.buf = buf
	w.len = len(buf)
}

func (w *Writer) Take() (b []byte, n int) {
	n, b = w.len, w.buf
	w.buf, w.len = nil, 0
	return
}

func (w *Writer) Write(p []byte) (n int, err error) {
	n = len(p)
	w.buf = append(w.buf, p...)
	w.len += n
	return n, nil
}

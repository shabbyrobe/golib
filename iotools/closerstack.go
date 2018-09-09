package iotools

import "io"

// ReadCloserStack wraps a group of io.Readers with a closer that closes
// things in the reverse order to the order they were added.
// The main use case is wrapping a GzipReader.
// The Read function will call the last added ReadCloser.
type ReadCloserStack struct {
	io.Reader
	readers []io.ReadCloser
}

var _ io.ReadCloser = &ReadCloserStack{}

func NewReadCloserStack(rc ...io.ReadCloser) *ReadCloserStack {
	rcs := &ReadCloserStack{}
	rcs.AddCloser(rc...)
	return rcs
}

func (d *ReadCloserStack) AddCloser(rc ...io.ReadCloser) {
	rcl := len(rc)
	if rcl > 0 {
		d.Reader = rc[rcl-1]
		d.readers = append(d.readers, rc...)
	}
}

func (d *ReadCloserStack) Close() error {
	var errs []error
	for i := len(d.readers) - 1; i >= 0; i-- {
		err := d.readers[i].Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return &closerStackError{errs}
	}
	return nil
}

// WriteCloserStack wraps a group of io.Writers with a closer that closes
// things in the reverse order to the order they were added.
// The main use case is wrapping a GzipWriter.
// The Write function will call the last added WriteCloser.
type WriteCloserStack struct {
	io.Writer
	writers []io.WriteCloser
}

var _ io.WriteCloser = &WriteCloserStack{}

func NewWriteCloserStack(wc ...io.WriteCloser) *WriteCloserStack {
	wcs := &WriteCloserStack{}
	wcs.AddCloser(wc...)
	return wcs
}

func (d *WriteCloserStack) AddCloser(wc ...io.WriteCloser) {
	wcl := len(wc)
	if wcl > 0 {
		d.Writer = wc[wcl-1]
		d.writers = append(d.writers, wc...)
	}
}

func (d *WriteCloserStack) Close() error {
	var errs []error
	for i := len(d.writers) - 1; i >= 0; i-- {
		err := d.writers[i].Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return &closerStackError{errs}
	}
	return nil
}

type closerStackError struct {
	errors []error
}

func (w *closerStackError) Errors() []error {
	return w.errors
}

func (w *closerStackError) Error() string {
	out := ""
	for i, e := range w.errors {
		if i == 0 {
			out = e.Error()
		} else {
			out += ", " + e.Error()
		}
	}
	return out
}

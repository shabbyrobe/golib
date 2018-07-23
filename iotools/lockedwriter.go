package iotools

import (
	"io"
	"sync"
)

type LockedWriter struct {
	w  io.Writer
	mu sync.Mutex
}

func NewLockedWriter(w io.Writer) *LockedWriter {
	return &LockedWriter{w: w}
}

func (l *LockedWriter) Write(b []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Write(b)
}

func (l *LockedWriter) Unwrap() io.Writer {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w
}

type LockedWriteCloser struct {
	w  io.WriteCloser
	mu sync.Mutex
}

func NewLockedWriteCloser(w io.WriteCloser) *LockedWriteCloser {
	return &LockedWriteCloser{w: w}
}

func (l *LockedWriteCloser) Write(b []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Write(b)
}

func (l *LockedWriteCloser) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Close()
}

func (l *LockedWriteCloser) Unwrap() io.WriteCloser {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w
}

type LockedWriteSeeker struct {
	w  io.WriteSeeker
	mu sync.Mutex
}

func NewLockedWriteSeeker(w io.WriteSeeker) *LockedWriteSeeker {
	return &LockedWriteSeeker{w: w}
}

func (l *LockedWriteSeeker) Write(b []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Write(b)
}

func (l *LockedWriteSeeker) Seek(offset int64, whence int) (int64, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Seek(offset, whence)
}

func (l *LockedWriteSeeker) Unwrap() io.WriteSeeker {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w
}

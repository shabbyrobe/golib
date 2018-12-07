package editfile

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

	"github.com/gofrs/flock"
	"github.com/shabbyrobe/golib/iotools"
)

// Edit creates a copy of a file with a mangled name in the same directory,
// opens that for writing, then replaces the original file with an atomic rename
// on Close.
//
// If the file does not exist, an empty file will be created, which will be locked.
//
// The original file will be locked using go-flock.
//
func Edit(name string) (file *File, rerr error) {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	defer func() {
		if rerr != nil {
			f.Close()
		}
	}()

	lock := flock.New(name)

	if locked, err := lock.TryLock(); err != nil {
		return nil, fmt.Errorf("editfile: lock failed: %v", err)
	} else if !locked {
		return nil, errFileLocked(fmt.Sprintf("editfile: file %q could not be locked", name))
	}

	tmpName := fmt.Sprintf("%s-%x-%x.edit", name, time.Now().UnixNano(), rand.Int31())

	tmp, err := os.OpenFile(tmpName, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer func() {
		if rerr != nil {
			tmp.Close()
		}
	}()

	buf := iotools.NewFullyBufferedWriterAt(f, tmp)

	defer func() {
		if rerr != nil {
			buf.Close()
		}
	}()

	if err := buf.Refresh(); err != nil {
		f.Close()
		buf.Close()
		tmp.Close()
		return nil, err
	}

	return &File{name: name, tmp: tmp, buf: buf, lock: lock}, nil
}

type bufferedWriterAt interface {
	Close() error
	Truncate(sz int64) error
	Flush() error
	io.ReaderAt
	io.WriterAt
}

var _ bufferedWriterAt = &iotools.FullyBufferedWriterAt{}

type File struct {
	closed bool
	name   string
	tmp    *os.File
	buf    *iotools.FullyBufferedWriterAt
	lock   *flock.Flock
	err    error
}

func (f *File) Close() (rerr error) {
	if f.closed {
		return errAlreadyClosed(1)
	}

	f.closed = true

	defer func() {
		if lerr := f.lock.Unlock(); rerr == nil && lerr != nil {
			rerr = lerr
		}
	}()
	defer f.tmp.Close()

	rerr = f.buf.Close()
	if rerr != nil && f.err == nil {
		f.err = rerr
	}

	if f.err == nil {
		rnerr := os.Rename(f.tmp.Name(), f.name)
		if rnerr != nil && rerr == nil {
			rerr = rnerr
		}
		if rerr != nil && f.err == nil {
			f.err = rnerr
		}
	} else {
		rmerr := os.Remove(f.tmp.Name())
		if rmerr != nil && rerr == nil {
			rerr = rmerr
		}
		if rerr != nil && f.err == nil {
			f.err = rmerr
		}
	}

	return rerr
}

func (f *File) Err() error { return f.err }

func (f *File) Truncate(sz int64) error {
	if f.err != nil {
		return errDisabled(1)
	}

	err := f.buf.Truncate(sz)
	if err != nil {
		f.err = err
	}
	return err
}

func (f *File) WriteAt(p []byte, offset int64) (n int, err error) {
	if f.err != nil {
		return 0, errDisabled(1)
	}

	n, err = f.buf.WriteAt(p, offset)
	if err != nil {
		f.err = err
	}
	return n, err
}

func (f *File) ReadAt(p []byte, off int64) (n int, err error) {
	if f.err != nil {
		return 0, errDisabled(1)
	}

	n, err = f.buf.ReadAt(p, off)
	if err != nil && err != io.EOF {
		f.err = err
	}
	return n, err
}

func (f *File) Flush() (err error) {
	if f.err != nil {
		return fmt.Errorf("editfile: error state")
	}

	err = f.buf.Flush()
	if err != nil {
		f.err = err
	}
	return err
}

func IsLocked(err error) bool {
	_, ok := err.(errFileLocked)
	return ok
}

func IsDisabled(err error) bool {
	_, ok := err.(errDisabled)
	return ok
}

type errFileLocked string

func (err errFileLocked) Error() string { return string(err) }

type errDisabled int

func (err errDisabled) Error() string { return "editfile: disabled due to previous error" }

type errAlreadyClosed int

func (err errAlreadyClosed) Error() string { return "iotools: already closed" }

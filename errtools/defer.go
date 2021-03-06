package errtools

import (
	"errors"
	"io"
	"os"
)

func DeferCall(err *error, call func() error) {
	cerr := call()
	if *err == nil && cerr != nil {
		*err = cerr
	}
}

// DeferClose closes an io.Closer and sets the error into err if one occurs and the
// value of err is nil.
func DeferClose(err *error, closer io.Closer) {
	DeferCall(err, closer.Close)
}

// DeferEnsureClose closes an io.Closer and sets the error into err if one
// occurs and the value of err is nil or an error that is known to be safe
// to ignore.
//
// Ignored errors:
//
//	os.ErrClosed
//
func DeferEnsureClose(err *error, closer io.Closer) {
	cerr := closer.Close()
	if cerr == nil || *err != nil {
		return
	}
	if errors.Is(cerr, os.ErrClosed) {
		return
	}

	*err = cerr
}

// DeferSet sets next into err if the value of err and next is both nil. err
// itself must not be nil.
//
// It is intended to be used like so:
//
//  func Pants() (err *error) {
//      defer func() { errtools.DeferSet(&err, file.Close()) }
//  }
//
func DeferSet(err *error, next error) {
	if *err == nil && next != nil {
		*err = next
	}
}

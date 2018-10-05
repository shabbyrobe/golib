package errtools

import (
	"io"
)

// DeferClose closes an io.Closer and sets the error into err if one occurs and the
// value of err is nil.
func DeferClose(err *error, closer io.Closer) {
	cerr := closer.Close()
	if *err == nil && cerr != nil {
		*err = cerr
	}
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

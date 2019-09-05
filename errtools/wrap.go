package errtools

import "fmt"

type wrapError struct {
	msg string
	err error
}

func (e *wrapError) Error() string {
	return e.msg
}

func (e *wrapError) Unwrap() error {
	return e.err
}

func Wrap(err error, msg string) error {
	return &wrapError{msg, err}
}

func Wrapf(err error, format string, a ...interface{}) error {
	msg := fmt.Sprintf(format, a...)
	return &wrapError{msg, err}
}

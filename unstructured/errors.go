package unstructured

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	InvalidTypeErrorKind  = errors.New("invalid type")
	InvalidValueErrorKind = errors.New("invalid value")
	KeyNotFoundErrorKind  = errors.New("key not found")
	IdxNotFoundErrorKind  = errors.New("idx not found")
)

type InvalidTypeError struct {
	Path     string
	Expected Kind
	Found    reflect.Type
	Msg      string
}

func (err *InvalidTypeError) Is(check error) bool { return check == InvalidTypeErrorKind }

func (err *InvalidTypeError) Error() string {
	msg := fmt.Sprintf("unstructured: %q: expected type %q, found %q", err.Path, err.Expected, err.Found)
	if err.Msg != "" {
		msg = fmt.Sprintf("%s: %s", msg, err.Msg)
	}
	return msg
}

type InvalidValueError struct {
	Path string
	Msg  string
}

func (err *InvalidValueError) Is(check error) bool { return check == InvalidValueErrorKind }

func (err *InvalidValueError) Error() string {
	msg := fmt.Sprintf("unstructured: %q: tried to create unstructured.Value from invalid value", err.Path)
	if err.Msg != "" {
		msg = fmt.Sprintf("%s: %s", msg, err.Msg)
	}
	return msg
}

type KeyNotFoundError struct {
	Path string
	Key  string
	Msg  string
}

func (err *KeyNotFoundError) descendNotFound() {}

func (err *KeyNotFoundError) Is(check error) bool {
	return check == KeyNotFoundErrorKind
}

func (err *KeyNotFoundError) Error() string {
	return fmt.Sprintf("unstructured: %q: key %q not found in map", err.Path, err.Key)
}

type IdxNotFoundError struct {
	Path string
	Idx  int
	Msg  string
}

func (err *IdxNotFoundError) descendNotFound() {}

func (err *IdxNotFoundError) Is(check error) bool {
	return check == IdxNotFoundErrorKind
}

func (err *IdxNotFoundError) Error() string {
	return fmt.Sprintf("unstructured: %q: idx %d not found in slice", err.Path, err.Idx)
}

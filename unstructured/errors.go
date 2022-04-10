package unstructured

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	TypeInvalidKind              = errors.New("invalid type")
	NumericConversionInvalidKind = errors.New("invalid numeric conversion")
	NumericConversionFailedKind  = errors.New("numeric conversion failed")
	NumericOverflowKind          = errors.New("numeric oveflow")
	ValueInvalidKind             = errors.New("invalid value")
	KeyNotFoundKind              = errors.New("key not found")
	KeyInvalidKind               = errors.New("key invalid")
	IndexNotFoundKind            = errors.New("idx not found")
	ValueErrorKind               = errors.New("value error")
)

type TypeInvalid struct {
	Path     string
	Expected Kind
	Found    reflect.Type
	Msg      string
}

func (err *TypeInvalid) Is(check error) bool { return check == TypeInvalidKind }

func (err *TypeInvalid) Error() string {
	msg := fmt.Sprintf("unstructured: %q: expected type %q, found %q", err.Path, err.Expected, err.Found)
	if err.Msg != "" {
		msg = fmt.Sprintf("%s: %s", msg, err.Msg)
	}
	return msg
}

type ValueInvalid struct {
	Path string
	Msg  string
}

func (err *ValueInvalid) Is(check error) bool { return check == ValueInvalidKind }

func (err *ValueInvalid) Error() string {
	msg := fmt.Sprintf("unstructured: %q: tried to create unstructured.Value from invalid value", err.Path)
	if err.Msg != "" {
		msg = fmt.Sprintf("%s: %s", msg, err.Msg)
	}
	return msg
}

type NumericConversionInvalid struct {
	Path string
	To   Kind
	Msg  string
}

func (err *NumericConversionInvalid) Is(check error) bool {
	return check == NumericConversionInvalidKind
}

func (err *NumericConversionInvalid) Error() string {
	msg := fmt.Sprintf("unstructured: %q: number cannot be converted to kind %q", err.Path, err.To)
	if err.Msg != "" {
		msg = fmt.Sprintf("%s: %s", msg, err.Msg)
	}
	return msg
}

type NumericConversionFailed struct {
	Path string
	To   Kind
	Msg  string
	err  error
}

func (err *NumericConversionFailed) Unwrap() error {
	return err.err
}

func (err *NumericConversionFailed) Is(check error) bool {
	return check == NumericConversionFailedKind
}

func (err *NumericConversionFailed) Error() string {
	msg := fmt.Sprintf("unstructured: %q: number cannot be converted to kind %q: %s", err.Path, err.To, err.err)
	if err.Msg != "" {
		msg = fmt.Sprintf("%s: %s", msg, err.Msg)
	}
	return msg
}

type NumericOverflow struct {
	Path string
	From Kind
	To   Kind
	Msg  string
}

func (err *NumericOverflow) Is(check error) bool {
	return check == NumericOverflowKind
}

func (err *NumericOverflow) Error() string {
	msg := fmt.Sprintf("unstructured: %q: overflow when converting %q to %q", err.Path, err.From, err.To)
	if err.Msg != "" {
		msg = fmt.Sprintf("%s: %s", msg, err.Msg)
	}
	return msg
}

type KeyNotFound struct {
	Path string
	Key  string
}

func (err *KeyNotFound) Is(check error) bool {
	return check == KeyNotFoundKind
}

func (err *KeyNotFound) Error() string {
	return fmt.Sprintf("unstructured: %q: key %q not found in map", err.Path, err.Key)
}

type IndexNotFound struct {
	Path string
	Idx  int
}

func (err *IndexNotFound) Is(check error) bool {
	return check == IndexNotFoundKind
}

func (err *IndexNotFound) Error() string {
	return fmt.Sprintf("unstructured: %q: idx %d not found in slice", err.Path, err.Idx)
}

// Error indicating that a key found in a map was not allowed. Note that this is never
// raised directly by 'unstructured', it must be raised yourself via MapIter:
//
//   v := unstructured.ValueOf(yourMap)
//   iter := v.Map().Iterate()
//   for iter.Next() {
//      switch iter.Key() {
//      case ...:
//      default:
//          iter.FailKey("")
//      }
//   }
//
type KeyInvalid struct {
	Path string
	Key  string
	err  error
}

func (err *KeyInvalid) Unwrap() error {
	return err.err
}

func (err *KeyInvalid) Is(check error) bool {
	return check == KeyInvalidKind
}

func (err *KeyInvalid) Error() string {
	return fmt.Sprintf("unstructured: %q: invalid key %q found in map: %s", err.Path, err.Key, err.err)
}

type ValueError struct {
	Path string
	Kind Kind
	err  error
}

func (err *ValueError) Unwrap() error {
	return err.err
}

func (err *ValueError) Is(check error) bool {
	return check == ValueErrorKind
}

func (err *ValueError) Error() string {
	return fmt.Sprintf("unstructured: %q: value error for kind %q: %s", err.Path, err.Kind, err.err)
}

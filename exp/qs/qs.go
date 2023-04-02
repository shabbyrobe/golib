// qs is a simple, experimental single-file copypasta library for query string handling:
// https://github.com/shabbyrobe/golib/blob/master/exp/qs
package qs

import (
	"encoding"
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

type Encoder interface {
	EncodeQueryValues() (url.Values, error)
}

type Decoder interface {
	DecodeQueryValues(url.Values) error
}

type Chain struct {
	loader *Loader
	Key    string
}

func (c *Chain) AddErr(err error) error {
	c.loader.AddErr(err)
	return err
}

type Loader struct {
	Values url.Values
	Errs   []error
}

func NewLoader(values url.Values) *Loader {
	return &Loader{Values: values}
}

func (loader *Loader) Err() error {
	if len(loader.Errs) == 0 {
		return nil
	}
	return &ErrQueryInvalid{Inner: errors.Join(loader.Errs...)}
}

func (loader *Loader) AddErr(err error) error {
	loader.Errs = append(loader.Errs, err)
	return err
}

func (loader *Loader) RequireFirst(key string) (chain *Chain, v string, ok bool, err error) {
	chain = &Chain{loader, key}
	vs, ok := loader.Values[key]
	if !ok || len(vs) == 0 {
		return chain, "", false, chain.AddErr(&ErrNotFound{Key: key})
	}
	return chain, vs[0], true, nil
}

func (loader *Loader) First(key string) (chain *Chain, v string, ok bool, err error) {
	chain = &Chain{loader, key}
	vs, ok := loader.Values[key]
	if !ok || len(vs) == 0 {
		return chain, "", false, nil
	}
	return chain, vs[0], true, nil
}

func (loader *Loader) Get(key string) (chain *Chain, v []string, ok bool, err error) {
	chain = &Chain{loader, key}
	vs, ok := loader.Values[key]
	return chain, vs, ok, nil
}

func (loader *Loader) Require(key string) (chain *Chain, v []string, ok bool, err error) {
	chain = &Chain{loader, key}
	vs, ok := loader.Values[key]
	if !ok {
		return chain, nil, false, chain.AddErr(&ErrNotFound{Key: key})
	}
	return chain, vs, ok, nil
}

type Chainable[I any, O any] func(
	loader *Loader,
	in I,
	lastOk bool,
	lastErr error,
) (
	chain *Loader,
	out O,
	ok bool,
	err error,
)

func Val[T any](chain *Chain, in T, lastOk bool, lastErr error) (out T) {
	if !lastOk || lastErr != nil {
		return out
	}
	return in
}

func Ptr[T any](chain *Chain, in T, lastOk bool, lastErr error) (out *T) {
	if !lastOk || lastErr != nil {
		return nil
	}
	v := in
	return &v
}

func Text[T any](chain *Chain, in string, lastOk bool, lastErr error) (next *Chain, out T, ok bool, err error) {
	if !lastOk || lastErr != nil {
		return chain, out, lastOk, lastErr
	}
	var dest T
	var u any = &dest
	unmarshaler, ok := (u).(encoding.TextUnmarshaler)
	if !ok {
		panic(fmt.Errorf("pointer to destination type for qs.Text() is not an encoding.TextUnmarshaler"))
	}
	if err := unmarshaler.UnmarshalText([]byte(in)); err != nil {
		return chain, out, false, chain.AddErr(&ErrValueInvalid{Path: chain.Key, Value: in, Inner: err})
	}
	return chain, dest, true, nil
}

func Texts[T any](chain *Chain, ins []string, lastOk bool, lastErr error) (next *Chain, out []T, ok bool, err error) {
	if !lastOk || lastErr != nil {
		return chain, out, lastOk, lastErr
	}
	out = make([]T, len(ins))
	for idx, in := range ins {
		var dest T
		var u any = &dest
		unmarshaler, ok := (u).(encoding.TextUnmarshaler)
		if !ok {
			panic(fmt.Errorf("pointer to destination type for qs.Text() is not an encoding.TextUnmarshaler"))
		}
		if err := unmarshaler.UnmarshalText([]byte(in)); err != nil {
			return chain, out, false, chain.AddErr(&ErrValueInvalid{
				Path: fmt.Sprintf("%s/%d", chain.Key, idx), Value: in, Inner: err})
		}
		out[idx] = dest
	}
	return chain, out, true, nil
}

func AnyInt[T ~int](chain *Chain, in string, lastOk bool, lastErr error) (next *Chain, out T, ok bool, err error) {
	return anyInt[T](chain, in, 10, 0, lastOk, lastErr)
}

func Int(chain *Chain, in string, lastOk bool, lastErr error) (next *Chain, out int, ok bool, err error) {
	return anyInt[int](chain, in, 10, 0, lastOk, lastErr)
}

func AnyInts[T ~int](chain *Chain, ins []string, lastOk bool, lastErr error) (next *Chain, out []T, ok bool, err error) {
	return anyInts[T](chain, ins, 10, 0, lastOk, lastErr)
}

func Ints(chain *Chain, ins []string, lastOk bool, lastErr error) (next *Chain, out []int, ok bool, err error) {
	return anyInts[int](chain, ins, 10, 0, lastOk, lastErr)
}

func AnyInt64[T ~int64](chain *Chain, in string, lastOk bool, lastErr error) (next *Chain, out T, ok bool, err error) {
	return anyInt[T](chain, in, 10, 64, lastOk, lastErr)
}

func Int64(chain *Chain, in string, lastOk bool, lastErr error) (next *Chain, out int64, ok bool, err error) {
	return anyInt[int64](chain, in, 10, 64, lastOk, lastErr)
}

func AnyInt64s[T ~int64](chain *Chain, ins []string, lastOk bool, lastErr error) (next *Chain, out []T, ok bool, err error) {
	return anyInts[T](chain, ins, 10, 64, lastOk, lastErr)
}

func Int64s(chain *Chain, ins []string, lastOk bool, lastErr error) (next *Chain, out []int64, ok bool, err error) {
	return anyInts[int64](chain, ins, 10, 64, lastOk, lastErr)
}

func AnyUint64[T ~uint64](chain *Chain, in string, lastOk bool, lastErr error) (next *Chain, out T, ok bool, err error) {
	return anyUint[T](chain, in, 10, 64, lastOk, lastErr)
}

func Uint64(chain *Chain, in string, lastOk bool, lastErr error) (next *Chain, out uint64, ok bool, err error) {
	return anyUint[uint64](chain, in, 10, 64, lastOk, lastErr)
}

func AnyUint64s[T ~uint64](chain *Chain, ins []string, lastOk bool, lastErr error) (next *Chain, out []T, ok bool, err error) {
	return anyUints[T](chain, ins, 10, 64, lastOk, lastErr)
}

func Uint64s(chain *Chain, ins []string, lastOk bool, lastErr error) (next *Chain, out []uint64, ok bool, err error) {
	return anyUints[uint64](chain, ins, 10, 64, lastOk, lastErr)
}

func AnyBool[T ~bool](chain *Chain, in string, lastOk bool, lastErr error) (next *Chain, out T, ok bool, err error) {
	if !lastOk || lastErr != nil {
		return chain, out, lastOk, lastErr
	}
	parsed, err := strconv.ParseBool(in)
	if err != nil {
		return chain, out, false, chain.AddErr(&ErrValueInvalid{Path: chain.Key, Value: in, Inner: err})
	}
	return chain, T(parsed), true, nil
}

func Bool(chain *Chain, in string, lastOk bool, lastErr error) (next *Chain, out bool, ok bool, err error) {
	return AnyBool[bool](chain, in, lastOk, lastErr)
}

func AnyFloat64[T ~float64](chain *Chain, in string, lastOk bool, lastErr error) (next *Chain, out T, ok bool, err error) {
	return anyFloat[T](chain, in, 64, lastOk, lastErr)
}

func Float64(chain *Chain, in string, lastOk bool, lastErr error) (next *Chain, out float64, ok bool, err error) {
	return anyFloat[float64](chain, in, 64, lastOk, lastErr)
}

func AnyFloat64s[T ~float64](chain *Chain, ins []string, lastOk bool, lastErr error) (next *Chain, out []T, ok bool, err error) {
	return anyFloats[T](chain, ins, 64, lastOk, lastErr)
}

func Float64s(chain *Chain, ins []string, lastOk bool, lastErr error) (next *Chain, out []float64, ok bool, err error) {
	return anyFloats[float64](chain, ins, 64, lastOk, lastErr)
}

func anyInt[T ~int | ~int8 | ~int16 | ~int32 | ~int64](
	chain *Chain, in string, base int, bitSize int, lastOk bool, lastErr error,
) (
	next *Chain, out T, ok bool, err error,
) {
	if !lastOk || lastErr != nil {
		return chain, out, lastOk, lastErr
	}
	parsed, err := strconv.ParseInt(in, base, bitSize)
	if err != nil {
		return chain, out, false, chain.AddErr(&ErrValueInvalid{Path: chain.Key, Value: in, Inner: err})
	}
	return chain, T(parsed), true, nil
}

func anyUint[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](
	chain *Chain, in string, base int, bitSize int, lastOk bool, lastErr error,
) (
	next *Chain, out T, ok bool, err error,
) {
	if !lastOk || lastErr != nil {
		return chain, out, lastOk, lastErr
	}
	parsed, err := strconv.ParseUint(in, base, bitSize)
	if err != nil {
		return chain, out, false, chain.AddErr(&ErrValueInvalid{Path: chain.Key, Value: in, Inner: err})
	}
	return chain, T(parsed), true, nil
}

func anyFloat[T ~float32 | ~float64](
	chain *Chain, in string, bitSize int, lastOk bool, lastErr error,
) (
	next *Chain, out T, ok bool, err error,
) {
	if !lastOk || lastErr != nil {
		return chain, out, lastOk, lastErr
	}
	parsed, err := strconv.ParseFloat(in, bitSize)
	if err != nil {
		return chain, out, false, chain.AddErr(&ErrValueInvalid{Path: chain.Key, Value: in, Inner: err})
	}
	return chain, T(parsed), true, nil
}

func anyInts[T ~int | ~int8 | ~int16 | ~int32 | ~int64](
	chain *Chain, ins []string, base int, bitSize int, lastOk bool, lastErr error,
) (
	next *Chain, out []T, ok bool, err error,
) {
	if !lastOk || lastErr != nil {
		return chain, out, lastOk, lastErr
	}
	out = make([]T, len(ins))
	for idx, in := range ins {
		parsed, err := strconv.ParseInt(in, base, bitSize)
		if err != nil {
			return chain, out, false, chain.AddErr(
				&ErrValueInvalid{Path: fmt.Sprintf("%s/%d", chain.Key, idx), Value: in, Inner: err})
		}
		out[idx] = T(parsed)
	}
	return chain, out, true, nil
}

func anyUints[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](
	chain *Chain, ins []string, base int, bitSize int, lastOk bool, lastErr error,
) (
	next *Chain, out []T, ok bool, err error,
) {
	if !lastOk || lastErr != nil {
		return chain, out, lastOk, lastErr
	}
	out = make([]T, len(ins))
	for idx, in := range ins {
		parsed, err := strconv.ParseUint(in, base, bitSize)
		if err != nil {
			return chain, out, false, chain.AddErr(
				&ErrValueInvalid{Path: fmt.Sprintf("%s/%d", chain.Key, idx), Value: in, Inner: err})
		}
		out[idx] = T(parsed)
	}
	return chain, out, true, nil
}

func anyFloats[T ~float32 | ~float64](
	chain *Chain, ins []string, bitSize int, lastOk bool, lastErr error,
) (
	next *Chain, out []T, ok bool, err error,
) {
	if !lastOk || lastErr != nil {
		return chain, out, lastOk, lastErr
	}
	out = make([]T, len(ins))
	for idx, in := range ins {
		parsed, err := strconv.ParseFloat(in, bitSize)
		if err != nil {
			return chain, out, false, chain.AddErr(
				&ErrValueInvalid{Path: fmt.Sprintf("%s/%d", chain.Key, idx), Value: in, Inner: err})
		}
		out[idx] = T(parsed)
	}
	return chain, out, true, nil
}

type ErrNotFound struct {
	Key string
}

func (err *ErrNotFound) Error() string {
	return fmt.Sprintf("query parameter %q is required", err.Key)
}

type ErrValueInvalid struct {
	Path  string
	Value string
	Inner error
}

func (err *ErrValueInvalid) Error() string {
	return fmt.Sprintf("query parameter %q, value %q could not be parsed: %s", err.Path, err.Value, err.Inner)
}

func (err *ErrValueInvalid) Unwrap() error {
	return err.Inner
}

type ErrQueryInvalid struct {
	Inner error
}

func (err *ErrQueryInvalid) Error() string {
	return fmt.Sprintf("query string could not be parsed: %s", err.Inner)
}

func (err *ErrQueryInvalid) Unwrap() error {
	return err.Inner
}

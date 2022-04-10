package unstructured

import (
	"reflect"
)

const (
	intSize = 32 << (^uint(0) >> 63) // 32 or 64

	maxInt  = 1<<(intSize-1) - 1
	minInt  = -1 << (intSize - 1)
	maxUint = 1<<intSize - 1
)

// Should support json.Number
type Int64Number interface {
	Int64() (int64, error)
}

// Should support json.Number
type Float64Number interface {
	Float64() (float64, error)
}

var (
	int64Number       Int64Number
	int64NumberType   = reflect.TypeOf(&int64Number).Elem()
	float64Number     Float64Number
	float64NumberType = reflect.TypeOf(&float64Number).Elem()
	numberType        = reflect.TypeOf(number{})
)

type number struct {
	inner      reflect.Value
	canInt64   bool
	canFloat64 bool
}

func numberFromInterface(v reflect.Value) (n number, ok bool) {
	n = number{
		inner:      v,
		canInt64:   v.CanConvert(int64NumberType),
		canFloat64: v.CanConvert(float64NumberType),
	}
	return n, n.canInt64 || n.canFloat64
}

func (n number) canBe(kind Kind) bool {
	if n.canInt64 && (kind == Int64Kind ||
		kind == IntKind ||
		kind == Uint64Kind ||
		kind == UintKind) {
		return true
	} else if n.canFloat64 && kind == Float64Kind {
		return true
	}
	return false
}

func (n number) asInt64(v Value) (i int64, ok bool) {
	if !n.canInt64 {
		v.ctx.AddError(&NumericConversionInvalid{
			Path: string(v.path),
			To:   Int64Kind,
		})
		return 0, false
	}

	iv, err := n.inner.Interface().(Int64Number).Int64()
	if err != nil {
		v.ctx.AddError(&NumericConversionFailed{
			Path: string(v.path),
			To:   Int64Kind,
			err:  err,
		})
		return 0, false
	}

	return iv, true
}

func (n number) asInt(v Value) (i int, ok bool) {
	i64, ok := n.asInt64(v)
	if !ok {
		return 0, false
	}

	if i64 < minInt || i64 > maxInt {
		v.ctx.AddError(&NumericOverflow{
			Path: string(v.path),
			From: Int64Kind,
			To:   IntKind,
		})
		return 0, false
	}
	return int(i64), true
}

func (n number) asUint64(v Value) (u uint64, ok bool) {
	i64, ok := n.asInt64(v)
	if !ok {
		return 0, false
	}
	if i64 < 0 {
		v.ctx.AddError(&NumericOverflow{
			Path: string(v.path),
			From: Int64Kind,
			To:   Uint64Kind,
		})
		return 0, false
	}
	return uint64(i64), true
}

func (n number) asUint(v Value) (u uint, ok bool) {
	i64, ok := n.asInt64(v)
	if !ok {
		return 0, false
	}

	if i64 < 0 || (i64 > 0 && uint64(i64) > maxUint) {
		v.ctx.AddError(&NumericOverflow{
			Path: string(v.path),
			From: Int64Kind,
			To:   UintKind,
		})
		return 0, false
	}

	return uint(i64), true
}

func (n number) asFloat64(v Value) (f float64, ok bool) {
	if !n.canInt64 {
		v.ctx.AddError(&NumericConversionInvalid{
			Path: string(v.path),
			To:   Float64Kind,
		})
		return 0, false
	}

	iv, err := n.inner.Interface().(Float64Number).Float64()
	if err != nil {
		v.ctx.AddError(&NumericConversionFailed{
			Path: string(v.path),
			To:   Float64Kind,
			err:  err,
		})
		return 0, false
	}

	return iv, true
}

package unstructured

import (
	"strings"
	"testing"
)

func ptrOf[V any](v V) *V {
	var x *V
	x = &v
	return x
}

func nilOf[V any]() *V {
	var x *V
	return x
}

type testingContext struct {
	ErrContext
}

var _ Context = (*testingContext)(nil)

func (tctx *testingContext) Defer(t *testing.T) {
	t.Helper()
	if len(tctx.errs) > 0 {
		var sb strings.Builder
		for idx, err := range tctx.errs {
			if idx > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(err.Error())
		}
		t.Fatal(sb.String())
	}
}

func ensureValueOfKind[T any](t *testing.T, v T, kind Kind) Value {
	t.Helper()
	ctx := &testingContext{}
	defer ctx.Defer(t)
	result := ValueOf(ctx, "", v)
	if result.kind != kind {
		t.Fatal("expected", "kind", kind, "== actual", result.kind)
	}
	return result
}

func assertNull(t *testing.T, v Value) {
	t.Helper()
	if !v.IsNull() {
		t.Fatal("value is not null, found", v.Kind())
	}
	if v.Kind() != NullKind {
		t.Fatal()
	}
	if c := v.BoolOptional(); c != false {
		t.Fatal("optional bool was not zero:", c)
	}
	if c := v.Float64Optional(); c != 0 {
		t.Fatal("optional float64 was not zero:", c)
	}
	if c := v.IntOptional(); c != 0 {
		t.Fatal("optional int was not zero:", c)
	}
	if c := v.Int64Optional(); c != 0 {
		t.Fatal("optional int64 was not zero:", c)
	}
	if c := v.MapOptional(); !c.IsNull() {
		t.Fatal("optional map was not null:", c)
	}
	if c := v.SliceOptional(); !c.IsNull() {
		t.Fatal("optional slice was not null:", c)
	}
	if c := v.StrOptional(); c != "" {
		t.Fatal("optional str was not zero:", c)
	}
	if c := v.UintOptional(); c != 0 {
		t.Fatal("optional uint was not zero:", c)
	}
	if c := v.Uint64Optional(); c != 0 {
		t.Fatal("optional uint64 was not zero:", c)
	}
}

func assertBool(t *testing.T, v Value, s bool) {
	t.Helper()
	if v.IsNull() {
		t.Fatal("bool was null")
	}
	if v.Kind() != BoolKind {
		t.Fatalf("kind is not bool, found %s", v.Kind())
	}
	if v.Bool() != s {
		t.Fatal("expected", "s", s, "== actual", v.Str())
	}
	if r := v.BoolOptional(); r != s {
		t.Fatal("expected", "s", s, "== actual", r)
	}
}

func assertBoolOptional(t *testing.T, v Value, s *bool) {
	t.Helper()
	if s == nil {
		assertNull(t, v)
	} else {
		assertBool(t, v, *s)
	}
}

func assertFloat64(t *testing.T, v Value, s float64) {
	t.Helper()
	if v.IsNull() {
		t.Fatal("float64 was null")
	}
	if v.Kind() != Float64Kind {
		t.Fatalf("kind is not float64, found %s", v.Kind())
	}
	if v.Float64() != s {
		t.Fatal("expected", "s", s, "== actual", v.Str())
	}
	if r := v.Float64Optional(); r != s {
		t.Fatal("expected", "s", s, "== actual", r)
	}
}

func assertFloat64Optional(t *testing.T, v Value, s *float64) {
	t.Helper()
	if s == nil {
		assertNull(t, v)
	} else {
		assertFloat64(t, v, *s)
	}
}

func assertInt(t *testing.T, v Value, s int) {
	t.Helper()
	if v.IsNull() {
		t.Fatal("int was null")
	}
	if v.Kind() != IntKind {
		t.Fatalf("kind is not int, found %s", v.Kind())
	}
	if v.Int() != s {
		t.Fatal("expected", "s", s, "== actual", v.Str())
	}
	if r := v.IntOptional(); r != s {
		t.Fatal("expected", "s", s, "== actual", r)
	}

	assertInt64(t, v, int64(s))
}

func assertIntOptional(t *testing.T, v Value, s *int) {
	t.Helper()
	if s == nil {
		assertNull(t, v)
	} else {
		assertInt(t, v, *s)
	}
}

func assertInt64(t *testing.T, v Value, s int64) {
	t.Helper()
	if v.IsNull() {
		t.Fatal("int64 was null")
	}
	if v.Kind() != Int64Kind && v.Kind() != IntKind {
		t.Fatalf("kind is not int or int64, found %s", v.Kind())
	}
	if v.Int64() != s {
		t.Fatal("expected", "s", s, "== actual", v.Str())
	}
	if r := v.Int64Optional(); r != s {
		t.Fatal("expected", "s", s, "== actual", r)
	}
}

func assertInt64Optional(t *testing.T, v Value, s *int64) {
	t.Helper()
	if s == nil {
		assertNull(t, v)
	} else {
		assertInt64(t, v, *s)
	}
}

func assertStr(t *testing.T, v Value, s string) {
	t.Helper()
	if v.IsNull() {
		t.Fatal("str was null")
	}
	if v.Kind() != StrKind {
		t.Fatalf("kind is not str, found %s", v.Kind())
	}
	if v.Str() != s {
		t.Fatal("expected", "s", s, "== actual", v.Str())
	}
	if r := v.StrOptional(); r != s {
		t.Fatal("expected", "s", s, "== actual", r)
	}
}

func assertStrOptional(t *testing.T, v Value, s *string) {
	t.Helper()
	if s == nil {
		assertNull(t, v)
	} else {
		assertStr(t, v, *s)
	}
}

func assertUint64(t *testing.T, v Value, s uint64) {
	t.Helper()
	if v.IsNull() {
		t.Fatal("uint64 was null")
	}
	if v.Kind() != Uint64Kind && v.Kind() != UintKind {
		t.Fatalf("kind is not uint or uint64, found %s", v.Kind())
	}
	if v.Uint64() != s {
		t.Fatal("expected", "s", s, "== actual", v.Str())
	}
	if r := v.Uint64Optional(); r != s {
		t.Fatal("expected", "s", s, "== actual", r)
	}
}

func assertUint64Optional(t *testing.T, v Value, s *uint64) {
	t.Helper()
	if s == nil {
		assertNull(t, v)
	} else {
		assertUint64(t, v, *s)
	}
}

func assertUint(t *testing.T, v Value, s uint) {
	t.Helper()
	if v.IsNull() {
		t.Fatal("uint was null")
	}
	if v.Kind() != UintKind {
		t.Fatalf("kind is not uint, found %s", v.Kind())
	}
	if v.Uint() != s {
		t.Fatal("expected", "s", s, "== actual", v.Str())
	}
	if r := v.UintOptional(); r != s {
		t.Fatal("expected", "s", s, "== actual", r)
	}
}

func assertUintOptional(t *testing.T, v Value, s *uint) {
	t.Helper()
	if s == nil {
		assertNull(t, v)
	} else {
		assertUint(t, v, *s)
	}
}

func assertSlice[T any](t *testing.T, v Value, s []T, get func(v Value) T) {
	t.Helper()

	if v.IsNull() {
		t.Fatal("slice was null")
	}
	if v.Kind() != SliceKind {
		t.Fatalf("kind is not slice, found %s", v.Kind())
	}

	r := v.Slice()
	if err := v.ctx.(*testingContext).PopError(); err != nil {
		t.Fatal(err)
	}

	iter := r.Iterate()
	result := make([]uint64, iter.Len())
	i := 0
	for iter.Next() {
		result[i] = iter.Value().Uint64()
		if iter.Idx() != i {
			t.Fatalf("unexpected idx %d", i)
		}
		if r.At(i).Unwrap() != iter.Value().Unwrap() {
			t.Fatalf("value mismatch at idx %d", i)
		}
		i++
	}

	if ro := v.SliceOptional(); ro.v.inner != r.v.inner {
		t.Fatal("expected", "s", s, "== actual", r)
	}
}

func assertSliceOptional[T any](t *testing.T, v Value, s *[]T, get func(v Value) T) {
	t.Helper()
	if s == nil {
		assertNull(t, v)
	} else {
		assertSlice(t, v, *s, get)
	}
}

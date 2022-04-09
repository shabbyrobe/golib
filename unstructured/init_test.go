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
	errs []error
}

var _ Context = (*testingContext)(nil)

func PopAllErrors(ctx Context) (errs []error) {
	for {
		err := ctx.PopError()
		if err == nil {
			break
		}
		errs = append(errs, err)
	}
	return errs
}

func (tctx *testingContext) PopError() error {
	if len(tctx.errs) == 0 {
		return nil
	}
	var err error
	err, tctx.errs = tctx.errs[len(tctx.errs)-1], tctx.errs[:len(tctx.errs)-1]
	return err
}

func (tctx *testingContext) ShiftError() error {
	if len(tctx.errs) == 0 {
		return nil
	}
	var err error
	err, tctx.errs = tctx.errs[0], tctx.errs[1:]
	return err
}

func (t *testingContext) Defer() {
	if len(t.errs) > 0 {
		var sb strings.Builder
		for idx, err := range t.errs {
			if idx > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(err.Error())
		}
		panic(sb.String())
	}
}

func (t *testingContext) AddError(err error) error {
	t.errs = append(t.errs, err)
	return nil
}

func ensureValueOfKind[T any](t *testing.T, v T, kind Kind) Value {
	t.Helper()
	ctx := &testingContext{}
	defer ctx.Defer()
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
	if _, set := v.BoolOptional(); set {
		t.Fatal("optional bool set for null")
	}
	if _, set := v.Float64Optional(); set {
		t.Fatal("optional float64 set for null")
	}
	if _, set := v.IntOptional(); set {
		t.Fatal("optional int set for null")
	}
	if _, set := v.Int64Optional(); set {
		t.Fatal("optional int64 set for null")
	}
	if _, set := v.MapOptional(); set {
		t.Fatal("optional map set for null")
	}
	if _, set := v.SliceOptional(); set {
		t.Fatal("optional slice set for null")
	}
	if _, set := v.StrOptional(); set {
		t.Fatal("optional str set for null")
	}
	if _, set := v.UintOptional(); set {
		t.Fatal("optional uint set for null")
	}
	if _, set := v.Uint64Optional(); set {
		t.Fatal("optional uint64 set for null")
	}
}

func assertBool(t *testing.T, v Value, s bool) {
	t.Helper()
	if v.IsNull() {
		t.Fatal()
	}
	if v.Kind() != BoolKind {
		t.Fatal()
	}
	if v.Bool() != s {
		t.Fatal("expected", "s", s, "== actual", v.Str())
	}
	if r, set := v.BoolOptional(); r != s {
		t.Fatal("expected", "s", s, "== actual", r)
	} else if !set {
		t.Fatal()
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
		t.Fatal()
	}
	if v.Kind() != Float64Kind {
		t.Fatal()
	}
	if v.Float64() != s {
		t.Fatal("expected", "s", s, "== actual", v.Str())
	}
	if r, set := v.Float64Optional(); r != s {
		t.Fatal("expected", "s", s, "== actual", r)
	} else if !set {
		t.Fatal()
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
		t.Fatal()
	}
	if v.Kind() != IntKind {
		t.Fatal()
	}
	if v.Int() != s {
		t.Fatal("expected", "s", s, "== actual", v.Str())
	}
	if r, set := v.IntOptional(); r != s {
		t.Fatal("expected", "s", s, "== actual", r)
	} else if !set {
		t.Fatal()
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
		t.Fatal()
	}
	if v.Kind() != Int64Kind && v.Kind() != IntKind {
		t.Fatal()
	}
	if v.Int64() != s {
		t.Fatal("expected", "s", s, "== actual", v.Str())
	}
	if r, set := v.Int64Optional(); r != s {
		t.Fatal("expected", "s", s, "== actual", r)
	} else if !set {
		t.Fatal()
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

func assertUint64(t *testing.T, v Value, s uint64) {
	t.Helper()
	if v.IsNull() {
		t.Fatal()
	}
	if v.Kind() != Uint64Kind && v.Kind() != UintKind {
		t.Fatal()
	}
	if v.Uint64() != s {
		t.Fatal("expected", "s", s, "== actual", v.Str())
	}
	if r, set := v.Uint64Optional(); r != s {
		t.Fatal("expected", "s", s, "== actual", r)
	} else if !set {
		t.Fatal()
	}
}

func assertStr(t *testing.T, v Value, s string) {
	t.Helper()
	if v.IsNull() {
		t.Fatal()
	}
	if v.Kind() != StrKind {
		t.Fatal()
	}
	if v.Str() != s {
		t.Fatal("expected", "s", s, "== actual", v.Str())
	}
	if r, set := v.StrOptional(); r != s {
		t.Fatal("expected", "s", s, "== actual", r)
	} else if !set {
		t.Fatal()
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
		t.Fatal()
	}
	if v.Kind() != UintKind {
		t.Fatal()
	}
	if v.Uint() != s {
		t.Fatal("expected", "s", s, "== actual", v.Str())
	}
	if r, set := v.UintOptional(); r != s {
		t.Fatal("expected", "s", s, "== actual", r)
	} else if !set {
		t.Fatal()
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
	if err := v.ctx.PopError(); err != nil {
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

	if ro, set := v.SliceOptional(); ro.inner != r.inner {
		t.Fatal("expected", "s", s, "== actual", r)
	} else if !set {
		t.Fatal()
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

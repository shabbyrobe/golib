package dynamic

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestValueOf(t *testing.T) {
	var seen []string
	run := func(n string, fn func(t *testing.T)) {
		t.Helper()
		seen = append(seen, n)
		t.Run(n, fn)
	}

	// Values in 'any' vars at the top scope
	run("any-direct-nil", func(t *testing.T) {
		var v any
		var result = ensureValueOfKind(t, v, NullKind)
		assertNull(t, result)
	})

	run("any-direct-bool-val", func(t *testing.T) {
		var v any = bool(true)
		var result = ensureValueOfKind(t, v, BoolKind)
		assertBool(t, result, true)
	})

	run("any-direct-bool-zero", func(t *testing.T) {
		var v any = bool(false)
		var result = ensureValueOfKind(t, v, BoolKind)
		assertBool(t, result, false)
	})

	run("any-direct-float64-val", func(t *testing.T) {
		var v any = float64(1.1)
		var result = ensureValueOfKind(t, v, Float64Kind)
		assertFloat64(t, result, 1.1)
	})

	run("any-direct-float64-zero", func(t *testing.T) {
		var v any = float64(0.0)
		var result = ensureValueOfKind(t, v, Float64Kind)
		assertFloat64(t, result, 0.0)
	})

	run("any-direct-int-val", func(t *testing.T) {
		var v any = int(1)
		var result = ensureValueOfKind(t, v, IntKind)
		assertInt(t, result, 1)
	})

	run("any-direct-int-zero", func(t *testing.T) {
		var v any = int(0)
		var result = ensureValueOfKind(t, v, IntKind)
		assertInt(t, result, 0)
	})

	run("any-direct-int64-val", func(t *testing.T) {
		var v any = int64(1)
		var result = ensureValueOfKind(t, v, Int64Kind)
		assertInt64(t, result, 1)
	})

	run("any-direct-int64-zero", func(t *testing.T) {
		var v any = int64(0)
		var result = ensureValueOfKind(t, v, Int64Kind)
		assertInt64(t, result, 0)
	})

	run("any-direct-slice-val", func(t *testing.T) {
		var v any = []int64{1, 2}
		var result = ensureValueOfKind(t, v, SliceKind)
		assertSlice(t, result, []int64{1, 2}, func(v Value) int64 { return v.Int64() })
	})

	run("any-direct-slice-nil", func(t *testing.T) {
		var v []int64 = nil
		var i any = v
		var result = ensureValueOfKind(t, i, NullKind)
		assertNull(t, result)
	})

	run("any-direct-str-val", func(t *testing.T) {
		var v any = "yep"
		var result = ensureValueOfKind(t, v, StrKind)
		assertStr(t, result, "yep")
	})

	run("any-direct-str-zero", func(t *testing.T) {
		var v any = ""
		var result = ensureValueOfKind(t, v, StrKind)
		assertStr(t, result, "")
	})

	run("any-direct-uint-val", func(t *testing.T) {
		var v any = uint(1)
		var result = ensureValueOfKind(t, v, UintKind)
		assertUint(t, result, 1)
	})

	run("any-direct-uint-zero", func(t *testing.T) {
		var v any = uint(0)
		var result = ensureValueOfKind(t, v, UintKind)
		assertUint(t, result, 0)
	})

	run("any-direct-uint64-val", func(t *testing.T) {
		var v any = uint64(1)
		var result = ensureValueOfKind(t, v, Uint64Kind)
		assertUint64(t, result, 1)
	})

	run("any-direct-uint64-zero", func(t *testing.T) {
		var v any = uint64(0)
		var result = ensureValueOfKind(t, v, Uint64Kind)
		assertUint64(t, result, 0)
	})

	// Values in 'any' vars nested in an []any
	run("any-inslice-nil", func(t *testing.T) {
		var v = []any{nil}
		var result = ensureValueOfKind(t, v[0], NullKind)
		assertNull(t, result)
	})

	run("any-inslice-slice-val", func(t *testing.T) {
		var v = []any{[]int64{1, 2}}
		var result = ensureValueOfKind(t, v[0], SliceKind)
		assertSlice(t, result, []int64{1, 2}, func(v Value) int64 { return v.Int64() })
	})

	run("any-inslice-slice-nil", func(t *testing.T) {
		var v = []any{([]int64)(nil)}
		var result = ensureValueOfKind(t, v[0], NullKind)
		assertNull(t, result)
	})

	run("json-float64", func(t *testing.T) {
		v := json.Number("1234")
		var result = ensureValueOfKind(t, v, NumberKind)
		assertFloat64Number(t, result, 1234.0)
	})

	run("json-int", func(t *testing.T) {
		v := json.Number("1234")
		var result = ensureValueOfKind(t, v, NumberKind)
		assertIntNumber(t, result, 1234)
	})

	run("json-int64", func(t *testing.T) {
		v := json.Number("1234")
		var result = ensureValueOfKind(t, v, NumberKind)
		assertInt64Number(t, result, 1234)
	})

	run("json-uint", func(t *testing.T) {
		v := json.Number("1234")
		var result = ensureValueOfKind(t, v, NumberKind)
		assertUintNumber(t, result, 1234)
	})

	run("json-uint64", func(t *testing.T) {
		v := json.Number("1234")
		var result = ensureValueOfKind(t, v, NumberKind)
		assertUint64Number(t, result, 1234)
	})

	run("ptr-bool-nil", func(t *testing.T) {
		var v *bool
		var result = ensureValueOfKind(t, v, NullKind)
		assertBoolOptional(t, result, nil)
	})

	run("ptr-bool-val", func(t *testing.T) {
		var v bool = true
		var result = ensureValueOfKind(t, &v, BoolKind)
		assertBoolOptional(t, result, &v)
	})

	run("ptr-bool-typed-nil", func(t *testing.T) {
		type pants bool
		var v *pants
		var result = ensureValueOfKind(t, v, NullKind)
		assertBoolOptional(t, result, nil)
	})

	run("ptr-bool-typed-val", func(t *testing.T) {
		type pants bool
		var v pants = true
		var result = ensureValueOfKind(t, &v, BoolKind)
		var r = bool(v)
		assertBoolOptional(t, result, &r)
	})

	run("ptr-float64-nil", func(t *testing.T) {
		var v *float64
		var result = ensureValueOfKind(t, v, NullKind)
		assertFloat64Optional(t, result, nil)
	})

	run("ptr-float64-val", func(t *testing.T) {
		var v float64 = 3
		var result = ensureValueOfKind(t, &v, Float64Kind)
		assertFloat64Optional(t, result, &v)
	})

	run("ptr-float64-typed-nil", func(t *testing.T) {
		type pants float64
		var v *pants
		var result = ensureValueOfKind(t, v, NullKind)
		assertFloat64Optional(t, result, nil)
	})

	run("ptr-float64-typed-val", func(t *testing.T) {
		type pants float64
		var v pants = 3
		var result = ensureValueOfKind(t, &v, Float64Kind)
		var r = float64(v)
		assertFloat64Optional(t, result, &r)
	})

	run("ptr-int-nil", func(t *testing.T) {
		var v *int
		var result = ensureValueOfKind(t, v, NullKind)
		assertIntOptional(t, result, nil)
	})

	run("ptr-int-val", func(t *testing.T) {
		var v int = 3
		var result = ensureValueOfKind(t, &v, IntKind)
		assertIntOptional(t, result, &v)
	})

	run("ptr-int-typed-nil", func(t *testing.T) {
		type pants int
		var v *pants
		var result = ensureValueOfKind(t, v, NullKind)
		assertIntOptional(t, result, nil)
	})

	run("ptr-int-typed-val", func(t *testing.T) {
		type pants int
		var v pants = 3
		var result = ensureValueOfKind(t, &v, IntKind)
		var r = int(v)
		assertIntOptional(t, result, &r)
	})

	run("ptr-int64-nil", func(t *testing.T) {
		var v *int64
		var result = ensureValueOfKind(t, v, NullKind)
		assertNull(t, result)
	})

	run("ptr-int64-val", func(t *testing.T) {
		var v int64 = 3
		var result = ensureValueOfKind(t, &v, Int64Kind)
		assertInt64(t, result, int64(v))
	})

	run("ptr-int64-typed-nil", func(t *testing.T) {
		type pants int64
		var v *pants
		var result = ensureValueOfKind(t, v, NullKind)
		assertInt64Optional(t, result, nil)
	})

	run("ptr-int64-typed-val", func(t *testing.T) {
		type pants int64
		var v pants = 3
		var result = ensureValueOfKind(t, &v, Int64Kind)
		var r = int64(v)
		assertInt64Optional(t, result, &r)
	})

	run("ptr-str-nil", func(t *testing.T) {
		var v *string
		var result = ensureValueOfKind(t, v, NullKind)
		assertNull(t, result)
	})

	run("ptr-str-val", func(t *testing.T) {
		var v = "string"
		var result = ensureValueOfKind(t, &v, StrKind)
		assertStr(t, result, v)
	})

	run("ptr-str-typed-nil", func(t *testing.T) {
		type pants string
		var v *pants
		ensureValueOfKind(t, v, NullKind)
	})

	run("ptr-str-typed-val", func(t *testing.T) {
		type pants string
		var v pants = "pants"
		var result = ensureValueOfKind(t, &v, StrKind)
		var r = string(v)
		assertStrOptional(t, result, &r)
	})

	run("ptr-uint-nil", func(t *testing.T) {
		var v *uint
		var result = ensureValueOfKind(t, v, NullKind)
		assertNull(t, result)
	})

	run("ptr-uint-val", func(t *testing.T) {
		var v uint = 3
		var result = ensureValueOfKind(t, &v, UintKind)
		assertUint(t, result, uint(v))
	})

	run("ptr-uint-typed-nil", func(t *testing.T) {
		type pants uint
		var v *pants
		var result = ensureValueOfKind(t, v, NullKind)
		assertUintOptional(t, result, nil)
	})

	run("ptr-uint-typed-val", func(t *testing.T) {
		type pants uint
		var v pants = 3
		var result = ensureValueOfKind(t, &v, UintKind)
		var r = uint(v)
		assertUintOptional(t, result, &r)
	})

	run("ptr-uint64-nil", func(t *testing.T) {
		var v *uint64
		var result = ensureValueOfKind(t, v, NullKind)
		assertNull(t, result)
	})

	run("ptr-uint64-val", func(t *testing.T) {
		var v uint64 = 3
		var result = ensureValueOfKind(t, &v, Uint64Kind)
		assertUint64(t, result, uint64(v))
	})

	run("ptr-uint64-typed-nil", func(t *testing.T) {
		type pants uint64
		var v *pants
		var result = ensureValueOfKind(t, v, NullKind)
		assertUint64Optional(t, result, nil)
	})

	run("ptr-uint64-typed-val", func(t *testing.T) {
		type pants uint64
		var v pants = 3
		var result = ensureValueOfKind(t, &v, Uint64Kind)
		var r = uint64(v)
		assertUint64Optional(t, result, &r)
	})

	run("bool-zero", func(t *testing.T) {
		var v = false
		var result = ensureValueOfKind(t, v, BoolKind)
		assertBool(t, result, v)
	})

	run("bool-val", func(t *testing.T) {
		var v = true
		var result = ensureValueOfKind(t, v, BoolKind)
		assertBool(t, result, v)
	})

	run("bool-typed-zero", func(t *testing.T) {
		type pants bool
		var v pants = false
		var result = ensureValueOfKind(t, v, BoolKind)
		assertBool(t, result, bool(v))
	})

	run("bool-typed-val", func(t *testing.T) {
		type pants bool
		var v pants = true
		var result = ensureValueOfKind(t, v, BoolKind)
		assertBool(t, result, bool(v))
	})

	run("float64-zero", func(t *testing.T) {
		var v float64 = 0.0
		var result = ensureValueOfKind(t, v, Float64Kind)
		assertFloat64(t, result, v)
	})

	run("float64-val", func(t *testing.T) {
		var v float64 = 1.1
		var result = ensureValueOfKind(t, v, Float64Kind)
		assertFloat64(t, result, v)
	})

	run("float64-typed-zero", func(t *testing.T) {
		type pants float64
		var v pants = 0.0
		var result = ensureValueOfKind(t, v, Float64Kind)
		assertFloat64(t, result, float64(v))
	})

	run("float64-typed-val", func(t *testing.T) {
		type pants float64
		var v pants = 1.1
		var result = ensureValueOfKind(t, v, Float64Kind)
		assertFloat64(t, result, float64(v))
	})

	run("int-zero", func(t *testing.T) {
		var v int = 0
		var result = ensureValueOfKind(t, v, IntKind)
		assertInt(t, result, v)
	})

	run("int-val", func(t *testing.T) {
		var v int = 1
		var result = ensureValueOfKind(t, v, IntKind)
		assertInt(t, result, v)
	})

	run("int-typed-zero", func(t *testing.T) {
		type pants int
		var v pants = 0
		var result = ensureValueOfKind(t, v, IntKind)
		assertInt(t, result, int(v))
	})

	run("int-typed-val", func(t *testing.T) {
		type pants int
		var v pants = 1
		var result = ensureValueOfKind(t, v, IntKind)
		assertInt(t, result, int(v))
	})

	run("int64-zero", func(t *testing.T) {
		var v int64 = 0
		var result = ensureValueOfKind(t, v, Int64Kind)
		assertInt64(t, result, v)
	})

	run("int64-val", func(t *testing.T) {
		var v int64 = 1
		var result = ensureValueOfKind(t, v, Int64Kind)
		assertInt64(t, result, v)
	})

	run("int64-typed-zero", func(t *testing.T) {
		type pants int64
		var v pants = 0
		var result = ensureValueOfKind(t, v, Int64Kind)
		assertInt64(t, result, int64(v))
	})

	run("int64-typed-val", func(t *testing.T) {
		type pants int64
		var v pants = 1
		var result = ensureValueOfKind(t, v, Int64Kind)
		assertInt64(t, result, int64(v))
	})

	run("str-zero", func(t *testing.T) {
		var v = ""
		var result = ensureValueOfKind(t, v, StrKind)
		assertStr(t, result, v)
	})

	run("str-val", func(t *testing.T) {
		var v = "yep"
		var result = ensureValueOfKind(t, v, StrKind)
		assertStr(t, result, v)
	})

	run("str-typed-zero", func(t *testing.T) {
		type pants string
		var v pants = ""
		var result = ensureValueOfKind(t, v, StrKind)
		assertStr(t, result, string(v))
	})

	run("str-typed-val", func(t *testing.T) {
		type pants string
		var v pants = "yep"
		var result = ensureValueOfKind(t, v, StrKind)
		assertStr(t, result, string(v))
	})

	run("uint-zero", func(t *testing.T) {
		var v uint = 0
		var result = ensureValueOfKind(t, v, UintKind)
		assertUint(t, result, v)
	})

	run("uint-val", func(t *testing.T) {
		var v uint = 1
		var result = ensureValueOfKind(t, v, UintKind)
		assertUint(t, result, v)
	})

	run("uint-typed-zero", func(t *testing.T) {
		type pants uint
		var v pants = 0
		var result = ensureValueOfKind(t, v, UintKind)
		assertUint(t, result, uint(v))
	})

	run("uint-typed-val", func(t *testing.T) {
		type pants uint
		var v pants = 1
		var result = ensureValueOfKind(t, v, UintKind)
		assertUint(t, result, uint(v))
	})

	run("uint64-zero", func(t *testing.T) {
		var v uint64 = 0
		var result = ensureValueOfKind(t, v, Uint64Kind)
		assertUint64(t, result, v)
	})

	run("uint64-val", func(t *testing.T) {
		var v uint64 = 1
		var result = ensureValueOfKind(t, v, Uint64Kind)
		assertUint64(t, result, v)
	})

	run("uint64-typed-zero", func(t *testing.T) {
		type pants uint64
		var v pants = 0
		var result = ensureValueOfKind(t, v, Uint64Kind)
		assertUint64(t, result, uint64(v))
	})

	run("uint64-typed-val", func(t *testing.T) {
		type pants uint64
		var v pants = 1
		var result = ensureValueOfKind(t, v, Uint64Kind)
		assertUint64(t, result, uint64(v))
	})

	run("slice-uint64-nil", func(t *testing.T) {
		var v []uint64 = nil
		var result = ensureValueOfKind(t, v, NullKind)
		assertSliceOptional(t, result, nil, func(v Value) uint64 { return v.Uint64() })
	})

	run("slice-uint64-empty", func(t *testing.T) {
		var v = []uint64{}
		var result = ensureValueOfKind(t, v, SliceKind)
		assertSlice(t, result, v, func(v Value) uint64 { return v.Uint64() })
	})

	run("slice-uint64-vals", func(t *testing.T) {
		var v = []uint64{1, 2, 3}
		var result = ensureValueOfKind(t, v, SliceKind)
		assertSlice(t, result, v, func(v Value) uint64 { return v.Uint64() })
	})

	run("reflect-interface", func(t *testing.T) {
		// Covers the case where we pass a reflect.Value in directly that represents
		// an interface:
		var v any = []any{map[string]any{}}
		var rv = reflect.ValueOf(v)
		var iface = rv.Index(0)
		ensureValueOfKind(t, iface, MapKind)
	})

	if os.Getenv("UNSTRUCTURED_TEST_DEBUG") != "" {
		sort.Strings(seen)
		fmt.Println(strings.Join(seen, "\n"))
	}
}

func TestMapDescentTryProducesNoError(t *testing.T) {
	ctx := &testingContext{}
	defer ctx.Defer(t)
	v := ValueOf(ctx, "", map[string]any{
		"pants": map[string]any{
			"foo": "yep",
		},
	})

	result := v.IfMap().Try("pants").
		IfMap().Try("foo").
		IfMap().Try("wat").
		IfMap().Try("bork").
		IfMap().Try("foo")

	if err := ctx.PopError(); err != nil {
		t.Fatal(err)
	}
	assertNull(t, result)
}

func TestMapDescentStopsCollectingAfterError(t *testing.T) {
	ctx := &testingContext{}
	defer ctx.Defer(t)
	v := ValueOf(ctx, "", map[string]any{
		"pants": map[string]any{
			"foo": "yep",
		},
	})
	result := v.
		Map().Key("pants").
		Map().Key("foo").
		Map().Key("wat").
		Map().Key("bork").
		Map().Key("foo")

	exp := &TypeInvalid{Path: "/pants/foo", Expected: MapKind, Found: reflect.TypeOf("")}
	if err := ctx.PopError(); !reflect.DeepEqual(err, exp) {
		t.Fatalf("%+v != %+v", err, exp)
	}
	assertNull(t, result)
}

func TestMapTryDescend(t *testing.T) {
	ctx := &testingContext{}
	defer ctx.Defer(t)
	v := ValueOf(ctx, "", map[string]any{
		"pants": map[string]any{
			"foo": map[string]any{
				"bar": "yep",
				"qux": []any{
					0: 1, 1: 2, 2: 3, 3: map[string]any{
						"deep": "very",
					},
				},
			},
		},
	})

	t.Run("", func(t *testing.T) {
		result := v.TryDescend("pants", "foo", "qux", 3, "deep")
		if err := ctx.PopError(); err != nil {
			t.Fatal(err)
		}
		assertStr(t, result, "very")
	})

	t.Run("", func(t *testing.T) {
		result := v.TryDescend("pants", "foo", "wat", "bork", "foo")
		if err := ctx.PopError(); err != nil {
			t.Fatal(err)
		}
		assertNull(t, result)
	})

	t.Run("", func(t *testing.T) {
		result := v.TryDescend("pants", "foo", "wat", "bork")
		if err := ctx.PopError(); err != nil {
			t.Fatal(err)
		}
		assertNull(t, result)
	})

	t.Run("", func(t *testing.T) {
		result := v.Map().TryDescend("pants", "foo", "wat")
		if err := ctx.PopError(); err != nil {
			t.Fatal(err)
		}
		assertNull(t, result)
	})

	t.Run("", func(t *testing.T) {
		result := v.TryDescend("pants", "foo", "bar")
		if err := ctx.PopError(); err != nil {
			t.Fatal(err)
		}
		assertStr(t, result, "yep")
	})
}

func TestJSONNumber(t *testing.T) {
}

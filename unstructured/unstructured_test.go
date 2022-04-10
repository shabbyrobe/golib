package unstructured

import (
	"reflect"
	"sort"
	"testing"
)

func TestValueOf(t *testing.T) {
	seen := []string{}
	all := []string{}
	for k := range valueOfCases {
		all = append(all, k)
	}
	run := func(k string) {
		t.Helper()
		tcase, ok := valueOfCases[k]
		if !ok {
			tcase = func(t *testing.T) {
				t.Fatalf("unknown valueOf case %s", k)
			}
		}
		t.Run(k, tcase)
		seen = append(seen, k)
	}

	// Values in 'any' vars at the top scope
	run("any-direct-nil")
	run("any-direct-bool-val")
	run("any-direct-bool-zero")
	run("any-direct-float64-val")
	run("any-direct-float64-zero")
	run("any-direct-int-val")
	run("any-direct-int-zero")
	run("any-direct-int64-val")
	run("any-direct-int64-zero")
	run("any-direct-slice-val")
	run("any-direct-slice-nil")
	run("any-direct-str-val")
	run("any-direct-str-zero")
	run("any-direct-uint-val")
	run("any-direct-uint-zero")
	run("any-direct-uint64-val")
	run("any-direct-uint64-zero")

	// Values in 'any' vars nested in an []any
	run("any-inslice-nil")
	run("any-inslice-slice-val")
	run("any-inslice-slice-nil")

	run("ptr-bool-nil")
	run("ptr-bool-val")
	run("ptr-bool-typed-nil")
	run("ptr-bool-typed-val")

	run("ptr-float64-nil")
	run("ptr-float64-val")
	run("ptr-float64-typed-nil")
	run("ptr-float64-typed-val")

	run("ptr-int-nil")
	run("ptr-int-val")
	run("ptr-int-typed-nil")
	run("ptr-int-typed-val")

	run("ptr-int64-nil")
	run("ptr-int64-val")
	run("ptr-int64-typed-nil")
	run("ptr-int64-typed-val")

	run("ptr-str-nil")
	run("ptr-str-val")
	run("ptr-str-typed-nil")
	run("ptr-str-typed-val")

	run("ptr-uint-nil")
	run("ptr-uint-val")
	run("ptr-uint-typed-nil")
	run("ptr-uint-typed-val")

	run("ptr-uint64-nil")
	run("ptr-uint64-val")
	run("ptr-uint64-typed-nil")
	run("ptr-uint64-typed-val")

	run("bool-zero")
	run("bool-val")
	run("bool-typed-zero")
	run("bool-typed-val")

	run("float64-zero")
	run("float64-val")
	run("float64-typed-zero")
	run("float64-typed-val")

	run("int-zero")
	run("int-val")
	run("int-typed-zero")
	run("int-typed-val")

	run("int64-zero")
	run("int64-val")
	run("int64-typed-zero")
	run("int64-typed-val")

	run("str-zero")
	run("str-val")
	run("str-typed-zero")
	run("str-typed-val")

	run("uint-zero")
	run("uint-val")
	run("uint-typed-zero")
	run("uint-typed-val")

	run("uint64-zero")
	run("uint64-val")
	run("uint64-typed-zero")
	run("uint64-typed-val")

	run("slice-uint64-nil")
	run("slice-uint64-empty")
	run("slice-uint64-vals")

	run("reflect-interface")

	sort.Strings(seen)
	sort.Strings(all)
	if !reflect.DeepEqual(seen, all) {
		t.Fatalf("run list not up to date:\nseen: %v\nall:  %v", seen, all)
	}
}

var valueOfCases = map[string]func(t *testing.T){
	"any-direct-nil": func(t *testing.T) {
		var v any
		var result = ensureValueOfKind(t, v, NullKind)
		assertNull(t, result)
	},

	"any-direct-bool-val": func(t *testing.T) {
		var v any = bool(true)
		var result = ensureValueOfKind(t, v, BoolKind)
		assertBool(t, result, true)
	},

	"any-direct-bool-zero": func(t *testing.T) {
		var v any = bool(false)
		var result = ensureValueOfKind(t, v, BoolKind)
		assertBool(t, result, false)
	},

	"any-direct-float64-val": func(t *testing.T) {
		var v any = float64(1.1)
		var result = ensureValueOfKind(t, v, Float64Kind)
		assertFloat64(t, result, 1.1)
	},

	"any-direct-float64-zero": func(t *testing.T) {
		var v any = float64(0.0)
		var result = ensureValueOfKind(t, v, Float64Kind)
		assertFloat64(t, result, 0.0)
	},

	"any-direct-int-val": func(t *testing.T) {
		var v any = int(1)
		var result = ensureValueOfKind(t, v, IntKind)
		assertInt(t, result, 1)
	},

	"any-direct-int-zero": func(t *testing.T) {
		var v any = int(0)
		var result = ensureValueOfKind(t, v, IntKind)
		assertInt(t, result, 0)
	},

	"any-direct-int64-val": func(t *testing.T) {
		var v any = int64(1)
		var result = ensureValueOfKind(t, v, Int64Kind)
		assertInt64(t, result, 1)
	},

	"any-direct-int64-zero": func(t *testing.T) {
		var v any = int64(0)
		var result = ensureValueOfKind(t, v, Int64Kind)
		assertInt64(t, result, 0)
	},

	"any-direct-slice-val": func(t *testing.T) {
		var v any = []int64{1, 2}
		var result = ensureValueOfKind(t, v, SliceKind)
		assertSlice(t, result, []int64{1, 2}, func(v Value) int64 { return v.Int64() })
	},

	"any-direct-slice-nil": func(t *testing.T) {
		var v []int64 = nil
		var i any = v
		var result = ensureValueOfKind(t, i, NullKind)
		assertNull(t, result)
	},

	"any-direct-str-val": func(t *testing.T) {
		var v any = "yep"
		var result = ensureValueOfKind(t, v, StrKind)
		assertStr(t, result, "yep")
	},

	"any-direct-str-zero": func(t *testing.T) {
		var v any = ""
		var result = ensureValueOfKind(t, v, StrKind)
		assertStr(t, result, "")
	},

	"any-direct-uint-val": func(t *testing.T) {
		var v any = uint(1)
		var result = ensureValueOfKind(t, v, UintKind)
		assertUint(t, result, 1)
	},

	"any-direct-uint-zero": func(t *testing.T) {
		var v any = uint(0)
		var result = ensureValueOfKind(t, v, UintKind)
		assertUint(t, result, 0)
	},

	"any-direct-uint64-val": func(t *testing.T) {
		var v any = uint64(1)
		var result = ensureValueOfKind(t, v, Uint64Kind)
		assertUint64(t, result, 1)
	},

	"any-direct-uint64-zero": func(t *testing.T) {
		var v any = uint64(0)
		var result = ensureValueOfKind(t, v, Uint64Kind)
		assertUint64(t, result, 0)
	},

	"any-inslice-nil": func(t *testing.T) {
		var v = []any{nil}
		var result = ensureValueOfKind(t, v[0], NullKind)
		assertNull(t, result)
	},

	"any-inslice-slice-val": func(t *testing.T) {
		var v = []any{[]int64{1, 2}}
		var result = ensureValueOfKind(t, v[0], SliceKind)
		assertSlice(t, result, []int64{1, 2}, func(v Value) int64 { return v.Int64() })
	},

	"any-inslice-slice-nil": func(t *testing.T) {
		var v = []any{([]int64)(nil)}
		var result = ensureValueOfKind(t, v[0], NullKind)
		assertNull(t, result)
	},

	"ptr-bool-nil": func(t *testing.T) {
		var v *bool
		var result = ensureValueOfKind(t, v, NullKind)
		assertBoolOptional(t, result, nil)
	},

	"ptr-bool-val": func(t *testing.T) {
		var v bool = true
		var result = ensureValueOfKind(t, &v, BoolKind)
		assertBoolOptional(t, result, &v)
	},

	"ptr-bool-typed-nil": func(t *testing.T) {
		type pants bool
		var v *pants
		var result = ensureValueOfKind(t, v, NullKind)
		assertBoolOptional(t, result, nil)
	},

	"ptr-bool-typed-val": func(t *testing.T) {
		type pants bool
		var v pants = true
		var result = ensureValueOfKind(t, &v, BoolKind)
		var r = bool(v)
		assertBoolOptional(t, result, &r)
	},

	"ptr-float64-nil": func(t *testing.T) {
		var v *float64
		var result = ensureValueOfKind(t, v, NullKind)
		assertFloat64Optional(t, result, nil)
	},

	"ptr-float64-val": func(t *testing.T) {
		var v float64 = 3
		var result = ensureValueOfKind(t, &v, Float64Kind)
		assertFloat64Optional(t, result, &v)
	},

	"ptr-float64-typed-nil": func(t *testing.T) {
		type pants float64
		var v *pants
		var result = ensureValueOfKind(t, v, NullKind)
		assertFloat64Optional(t, result, nil)
	},

	"ptr-float64-typed-val": func(t *testing.T) {
		type pants float64
		var v pants = 3
		var result = ensureValueOfKind(t, &v, Float64Kind)
		var r = float64(v)
		assertFloat64Optional(t, result, &r)
	},

	"ptr-int-nil": func(t *testing.T) {
		var v *int
		var result = ensureValueOfKind(t, v, NullKind)
		assertIntOptional(t, result, nil)
	},

	"ptr-int-val": func(t *testing.T) {
		var v int = 3
		var result = ensureValueOfKind(t, &v, IntKind)
		assertIntOptional(t, result, &v)
	},

	"ptr-int-typed-nil": func(t *testing.T) {
		type pants int
		var v *pants
		var result = ensureValueOfKind(t, v, NullKind)
		assertIntOptional(t, result, nil)
	},

	"ptr-int-typed-val": func(t *testing.T) {
		type pants int
		var v pants = 3
		var result = ensureValueOfKind(t, &v, IntKind)
		var r = int(v)
		assertIntOptional(t, result, &r)
	},

	"ptr-int64-nil": func(t *testing.T) {
		var v *int64
		var result = ensureValueOfKind(t, v, NullKind)
		assertNull(t, result)
	},

	"ptr-int64-val": func(t *testing.T) {
		var v int64 = 3
		var result = ensureValueOfKind(t, &v, Int64Kind)
		assertInt64(t, result, int64(v))
	},

	"ptr-int64-typed-nil": func(t *testing.T) {
		type pants int64
		var v *pants
		var result = ensureValueOfKind(t, v, NullKind)
		assertInt64Optional(t, result, nil)
	},

	"ptr-int64-typed-val": func(t *testing.T) {
		type pants int64
		var v pants = 3
		var result = ensureValueOfKind(t, &v, Int64Kind)
		var r = int64(v)
		assertInt64Optional(t, result, &r)
	},

	"ptr-str-nil": func(t *testing.T) {
		var v *string
		var result = ensureValueOfKind(t, v, NullKind)
		assertNull(t, result)
	},

	"ptr-str-val": func(t *testing.T) {
		var v = "string"
		var result = ensureValueOfKind(t, &v, StrKind)
		assertStr(t, result, v)
	},

	"ptr-str-typed-nil": func(t *testing.T) {
		type pants string
		var v *pants
		ensureValueOfKind(t, v, NullKind)
	},

	"ptr-str-typed-val": func(t *testing.T) {
		type pants string
		var v pants = "pants"
		var result = ensureValueOfKind(t, &v, StrKind)
		var r = string(v)
		assertStrOptional(t, result, &r)
	},

	"ptr-uint-nil": func(t *testing.T) {
		var v *uint
		var result = ensureValueOfKind(t, v, NullKind)
		assertNull(t, result)
	},

	"ptr-uint-val": func(t *testing.T) {
		var v uint = 3
		var result = ensureValueOfKind(t, &v, UintKind)
		assertUint(t, result, uint(v))
	},

	"ptr-uint-typed-nil": func(t *testing.T) {
		type pants uint
		var v *pants
		var result = ensureValueOfKind(t, v, NullKind)
		assertUintOptional(t, result, nil)
	},

	"ptr-uint-typed-val": func(t *testing.T) {
		type pants uint
		var v pants = 3
		var result = ensureValueOfKind(t, &v, UintKind)
		var r = uint(v)
		assertUintOptional(t, result, &r)
	},

	"ptr-uint64-nil": func(t *testing.T) {
		var v *uint64
		var result = ensureValueOfKind(t, v, NullKind)
		assertNull(t, result)
	},

	"ptr-uint64-val": func(t *testing.T) {
		var v uint64 = 3
		var result = ensureValueOfKind(t, &v, Uint64Kind)
		assertUint64(t, result, uint64(v))
	},

	"ptr-uint64-typed-nil": func(t *testing.T) {
		type pants uint64
		var v *pants
		var result = ensureValueOfKind(t, v, NullKind)
		assertUint64Optional(t, result, nil)
	},

	"ptr-uint64-typed-val": func(t *testing.T) {
		type pants uint64
		var v pants = 3
		var result = ensureValueOfKind(t, &v, Uint64Kind)
		var r = uint64(v)
		assertUint64Optional(t, result, &r)
	},

	"bool-val": func(t *testing.T) {
		var v = true
		var result = ensureValueOfKind(t, v, BoolKind)
		assertBool(t, result, v)
	},

	"bool-zero": func(t *testing.T) {
		var v = false
		var result = ensureValueOfKind(t, v, BoolKind)
		assertBool(t, result, v)
	},

	"bool-typed-val": func(t *testing.T) {
		type pants bool
		var v pants = true
		var result = ensureValueOfKind(t, v, BoolKind)
		assertBool(t, result, bool(v))
	},

	"bool-typed-zero": func(t *testing.T) {
		type pants bool
		var v pants = false
		var result = ensureValueOfKind(t, v, BoolKind)
		assertBool(t, result, bool(v))
	},

	"float64-val": func(t *testing.T) {
		var v float64 = 1.1
		var result = ensureValueOfKind(t, v, Float64Kind)
		assertFloat64(t, result, v)
	},

	"float64-zero": func(t *testing.T) {
		var v float64 = 0.0
		var result = ensureValueOfKind(t, v, Float64Kind)
		assertFloat64(t, result, v)
	},

	"float64-typed-val": func(t *testing.T) {
		type pants float64
		var v pants = 1.1
		var result = ensureValueOfKind(t, v, Float64Kind)
		assertFloat64(t, result, float64(v))
	},

	"float64-typed-zero": func(t *testing.T) {
		type pants float64
		var v pants = 0.0
		var result = ensureValueOfKind(t, v, Float64Kind)
		assertFloat64(t, result, float64(v))
	},

	"int-val": func(t *testing.T) {
		var v int = 1
		var result = ensureValueOfKind(t, v, IntKind)
		assertInt(t, result, v)
	},

	"int-zero": func(t *testing.T) {
		var v int = 0
		var result = ensureValueOfKind(t, v, IntKind)
		assertInt(t, result, v)
	},

	"int-typed-val": func(t *testing.T) {
		type pants int
		var v pants = 1
		var result = ensureValueOfKind(t, v, IntKind)
		assertInt(t, result, int(v))
	},

	"int-typed-zero": func(t *testing.T) {
		type pants int
		var v pants = 0
		var result = ensureValueOfKind(t, v, IntKind)
		assertInt(t, result, int(v))
	},

	"int64-val": func(t *testing.T) {
		var v int64 = 1
		var result = ensureValueOfKind(t, v, Int64Kind)
		assertInt64(t, result, v)
	},

	"int64-zero": func(t *testing.T) {
		var v int64 = 0
		var result = ensureValueOfKind(t, v, Int64Kind)
		assertInt64(t, result, v)
	},

	"int64-typed-val": func(t *testing.T) {
		type pants int64
		var v pants = 1
		var result = ensureValueOfKind(t, v, Int64Kind)
		assertInt64(t, result, int64(v))
	},

	"int64-typed-zero": func(t *testing.T) {
		type pants int64
		var v pants = 0
		var result = ensureValueOfKind(t, v, Int64Kind)
		assertInt64(t, result, int64(v))
	},

	"str-val": func(t *testing.T) {
		var v = "yep"
		var result = ensureValueOfKind(t, v, StrKind)
		assertStr(t, result, v)
	},

	"str-zero": func(t *testing.T) {
		var v = ""
		var result = ensureValueOfKind(t, v, StrKind)
		assertStr(t, result, v)
	},

	"str-typed-val": func(t *testing.T) {
		type pants string
		var v pants = "yep"
		var result = ensureValueOfKind(t, v, StrKind)
		assertStr(t, result, string(v))
	},

	"str-typed-zero": func(t *testing.T) {
		type pants string
		var v pants = ""
		var result = ensureValueOfKind(t, v, StrKind)
		assertStr(t, result, string(v))
	},

	"uint-val": func(t *testing.T) {
		var v uint = 1
		var result = ensureValueOfKind(t, v, UintKind)
		assertUint(t, result, v)
	},

	"uint-zero": func(t *testing.T) {
		var v uint = 0
		var result = ensureValueOfKind(t, v, UintKind)
		assertUint(t, result, v)
	},

	"uint-typed-val": func(t *testing.T) {
		type pants uint
		var v pants = 1
		var result = ensureValueOfKind(t, v, UintKind)
		assertUint(t, result, uint(v))
	},

	"uint-typed-zero": func(t *testing.T) {
		type pants uint
		var v pants = 0
		var result = ensureValueOfKind(t, v, UintKind)
		assertUint(t, result, uint(v))
	},

	"uint64-val": func(t *testing.T) {
		var v uint64 = 1
		var result = ensureValueOfKind(t, v, Uint64Kind)
		assertUint64(t, result, v)
	},

	"uint64-zero": func(t *testing.T) {
		var v uint64 = 0
		var result = ensureValueOfKind(t, v, Uint64Kind)
		assertUint64(t, result, v)
	},

	"uint64-typed-val": func(t *testing.T) {
		type pants uint64
		var v pants = 1
		var result = ensureValueOfKind(t, v, Uint64Kind)
		assertUint64(t, result, uint64(v))
	},

	"uint64-typed-zero": func(t *testing.T) {
		type pants uint64
		var v pants = 0
		var result = ensureValueOfKind(t, v, Uint64Kind)
		assertUint64(t, result, uint64(v))
	},

	"slice-uint64-nil": func(t *testing.T) {
		var v []uint64 = nil
		var result = ensureValueOfKind(t, v, NullKind)
		assertSliceOptional(t, result, nil, func(v Value) uint64 { return v.Uint64() })
	},

	"slice-uint64-empty": func(t *testing.T) {
		var v = []uint64{}
		var result = ensureValueOfKind(t, v, SliceKind)
		assertSlice(t, result, v, func(v Value) uint64 { return v.Uint64() })
	},

	"slice-uint64-vals": func(t *testing.T) {
		var v = []uint64{1, 2, 3}
		var result = ensureValueOfKind(t, v, SliceKind)
		assertSlice(t, result, v, func(v Value) uint64 { return v.Uint64() })
	},

	"reflect-interface": func(t *testing.T) {
		// Covers the case where we pass a reflect.Value in directly that represents
		// an interface:
		var v any = []any{map[string]any{}}
		var rv = reflect.ValueOf(v)
		var iface = rv.Index(0)
		ensureValueOfKind(t, iface, MapKind)
	},
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

	exp := &InvalidTypeError{Path: "/pants/foo", Expected: MapKind, Found: reflect.TypeOf("")}
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
		result := v.TryDescend("pants", "foo", "qux", 3)
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

/*
func TestMapPaths(t *testing.T) {
	ctx := &testingContext{}
	v := ValueOf(ctx, "", map[string]any{
		"pants": map[string]any{
			"foo": "yep",
		},
	})

	child, ok := Descend(v, "pants", "foo")
	fmt.Println(child, ok)
}
*/

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

	run("any-nil")
	run("any-bool-val")
	run("any-bool-zero")
	run("any-float64-val")
	run("any-float64-zero")
	run("any-int-val")
	run("any-int-zero")
	run("any-int64-val")
	run("any-int64-zero")
	run("any-str-val")
	run("any-str-zero")
	run("any-uint-val")
	run("any-uint-zero")
	run("any-uint64-val")
	run("any-uint64-zero")

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

	sort.Strings(seen)
	sort.Strings(all)
	if !reflect.DeepEqual(seen, all) {
		t.Fatalf("run list not up to date:\nseen: %v\nall:  %v", seen, all)
	}
}

var valueOfCases = map[string]func(t *testing.T){
	"any-nil": func(t *testing.T) {
		var v any
		var result = ensureValueOfKind(t, v, NullKind)
		assertNull(t, result)
	},

	"any-bool-val": func(t *testing.T) {
		var v any = bool(true)
		var result = ensureValueOfKind(t, v, BoolKind)
		assertBool(t, result, true)
	},

	"any-bool-zero": func(t *testing.T) {
		var v any = bool(false)
		var result = ensureValueOfKind(t, v, BoolKind)
		assertBool(t, result, false)
	},

	"any-float64-val": func(t *testing.T) {
		var v any = float64(1.1)
		var result = ensureValueOfKind(t, v, Float64Kind)
		assertFloat64(t, result, 1.1)
	},

	"any-float64-zero": func(t *testing.T) {
		var v any = float64(0.0)
		var result = ensureValueOfKind(t, v, Float64Kind)
		assertFloat64(t, result, 0.0)
	},

	"any-int-val": func(t *testing.T) {
		var v any = int(1)
		var result = ensureValueOfKind(t, v, IntKind)
		assertInt(t, result, 1)
	},

	"any-int-zero": func(t *testing.T) {
		var v any = int(0)
		var result = ensureValueOfKind(t, v, IntKind)
		assertInt(t, result, 0)
	},

	"any-int64-val": func(t *testing.T) {
		var v any = int64(1)
		var result = ensureValueOfKind(t, v, Int64Kind)
		assertInt64(t, result, 1)
	},

	"any-int64-zero": func(t *testing.T) {
		var v any = int64(0)
		var result = ensureValueOfKind(t, v, Int64Kind)
		assertInt64(t, result, 0)
	},

	"any-str-val": func(t *testing.T) {
		var v any = "yep"
		var result = ensureValueOfKind(t, v, StrKind)
		assertStr(t, result, "yep")
	},

	"any-str-zero": func(t *testing.T) {
		var v any = ""
		var result = ensureValueOfKind(t, v, StrKind)
		assertStr(t, result, "")
	},

	"any-uint-val": func(t *testing.T) {
		var v any = uint(1)
		var result = ensureValueOfKind(t, v, UintKind)
		assertUint(t, result, 1)
	},

	"any-uint-zero": func(t *testing.T) {
		var v any = uint(0)
		var result = ensureValueOfKind(t, v, UintKind)
		assertUint(t, result, 0)
	},

	"any-uint64-val": func(t *testing.T) {
		var v any = uint64(1)
		var result = ensureValueOfKind(t, v, Uint64Kind)
		assertUint64(t, result, 1)
	},

	"any-uint64-zero": func(t *testing.T) {
		var v any = uint64(0)
		var result = ensureValueOfKind(t, v, Uint64Kind)
		assertUint64(t, result, 0)
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
}

func TestMapDescentStopsCollectingAfterError(t *testing.T) {
	ctx := &testingContext{}
	defer ctx.Defer()
	v := ValueOf(ctx, "", map[string]any{
		"pants": map[string]any{
			"foo": "yep",
		},
	})
	result := v.Key("pants").Key("foo").Key("wat").Key("bork").Key("foo")

	exp := &InvalidTypeError{Path: "/pants/foo", Expected: MapKind, Found: reflect.TypeOf("")}
	if err := ctx.PopError(); !reflect.DeepEqual(err, exp) {
		t.Fatalf("%+v != %+v", err, exp)
	}
	assertNull(t, result)
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

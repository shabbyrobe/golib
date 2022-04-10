package unstructured

import (
	"fmt"
	"reflect"
)

type Value struct {
	ctx    Context
	inner  reflect.Value
	kind   Kind
	path   string
	named  bool
	dead   bool
	number number
}

var _ value = Value{}

func ValueOf(ctx Context, path string, v any) Value {
	var rv reflect.Value
	var ok bool
	var kind Kind
	var named bool
	var number number

	if rv, ok = v.(reflect.Value); !ok {
		rv = reflect.ValueOf(v)
	}

	if !rv.IsValid() {
		return Value{ctx: ctx, kind: NullKind, path: path, dead: true}
	}

	if rv.Kind() == reflect.Interface {
		// If a value is passed in that is wrapped in an interface, unwrap its
		// element. This can happen if you retrieve a reflect.Value via
		// reflect.Value.Index() and pass it in directly:
		rv = rv.Elem()
	}

	if num, ok := numberFromInterface(rv); ok {
		kind = NumberKind
		number = num
		rv = reflect.ValueOf(num)
	} else if isNullable(rv.Kind()) && rv.IsNil() {
		kind = NullKind
	} else {
		typ := rv.Type()
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
			rv = rv.Elem()
		}
		kind = kindOf(typ)
		named = typ.Name() != ""
	}

	return Value{
		ctx:    ctx,
		inner:  rv,
		kind:   kind,
		path:   path,
		named:  named,
		number: number,
	}
}

// Optional chaining: returns a clone of v with future descent disabled.
func (v Value) kill() Value {
	return Value{
		ctx:  v.ctx,
		kind: NullKind,
		dead: true,
		path: v.path,
	}
}

func (v Value) unnamed() reflect.Value {
	if v.named {
		return v.inner.Convert(basic[v.kind])
	}
	return v.inner
}

func (v Value) IsNull() bool  { return v.dead || v.kind == NullKind }
func (v Value) IsValid() bool { return !v.dead }
func (v Value) Path() string  { return string(v.path) }
func (v Value) Unwrap() any   { return unwrap(v) }

func (v Value) CanBe(kind Kind) bool {
	if v.kind == kind {
		return true
	}
	if v.kind == NumberKind {
		return v.number.canBe(kind)
	}
	return false
}

// Returns the underlying Kind in this Value. Note that instead of switching
// on the return of this, you should instead use Value.CanBe(...) to test
// as this supports the interface-based types (i.e. json.Number).
func (v Value) Kind() Kind { return v.kind }

func (v Value) Str() string {
	// XXX: don't call this String(), it makes a mess with fmt.Stringer.
	if v.dead {
		return ""
	}
	if v.inner.Kind() != reflect.String {
		v.ctx.AddError(&TypeInvalid{Path: string(v.path), Expected: StrKind, Found: v.inner.Type()})
		return ""
	}
	return v.unnamed().Interface().(string)
}

func (v Value) Reject(err error) {
	// XXX: deliberately not dropping these if 'dead'; this seems prudent as this
	// has to be explicitly called?
	v.ctx.AddError(&ValueError{Path: v.path, Kind: v.kind, err: err})
}

func (v Value) StrOptional() string {
	if v.kind == NullKind {
		return ""
	}
	return v.Str()
}

func (v Value) Int() int {
	if v.dead {
		return 0
	}

	if v.kind == NumberKind {
		i, _ := v.number.asInt(v)
		return i

	} else if v.kind != IntKind {
		v.ctx.AddError(&TypeInvalid{Path: string(v.path), Expected: IntKind, Found: v.inner.Type()})
		return 0
	}
	return v.unnamed().Interface().(int)
}

func (v Value) IntOptional() (value int) {
	if v.kind == NullKind {
		return 0
	}
	return v.Int()
}

func (v Value) Int64() int64 {
	if v.dead {
		return 0
	}

	if v.kind == NumberKind {
		i, _ := v.number.asInt64(v)
		return i

	} else if v.kind == IntKind {
		i := v.unnamed().Interface().(int)
		return int64(i)

	} else if v.kind == Int64Kind {
		return v.unnamed().Interface().(int64)
	}

	v.ctx.AddError(&TypeInvalid{Path: string(v.path), Expected: Int64Kind, Found: v.inner.Type()})
	return 0
}

func (v Value) Int64Optional() (value int64) {
	if v.kind == NullKind {
		return 0
	}
	return v.Int64()
}

func (v Value) Uint() uint {
	if v.dead {
		return 0
	}

	if v.kind == NumberKind {
		u, _ := v.number.asUint(v)
		return u

	} else if v.kind != UintKind {
		v.ctx.AddError(&TypeInvalid{Path: string(v.path), Expected: UintKind, Found: v.inner.Type()})
		return 0
	}

	return v.unnamed().Interface().(uint)
}

func (v Value) UintOptional() (value uint) {
	if v.kind == NullKind {
		return 0
	}
	return v.Uint()
}

func (v Value) Uint64() uint64 {
	if v.dead {
		return 0
	}

	if v.kind == NumberKind {
		u, _ := v.number.asUint64(v)
		return u

	} else if v.kind == UintKind {
		u := v.unnamed().Interface().(uint)
		return uint64(u)

	} else if v.kind == Uint64Kind {
		return v.unnamed().Interface().(uint64)
	}

	v.ctx.AddError(&TypeInvalid{Path: string(v.path), Expected: Uint64Kind, Found: v.inner.Type()})
	return 0
}

func (v Value) Uint64Optional() (value uint64) {
	if v.kind == NullKind {
		return 0
	}
	return v.Uint64()
}

func (v Value) Float64() float64 {
	if v.dead {
		return 0
	}

	if v.kind == NumberKind {
		f, _ := v.number.asFloat64(v)
		return f

	} else if v.kind != Float64Kind {
		v.ctx.AddError(&TypeInvalid{Path: string(v.path), Expected: Float64Kind, Found: v.inner.Type()})
		return 0
	}

	return v.unnamed().Interface().(float64)
}

func (v Value) Float64Optional() (value float64) {
	if v.kind == NullKind {
		return 0
	}
	return v.Float64()
}

func (v Value) Bool() bool {
	if v.dead {
		return false
	}
	if v.kind != BoolKind {
		v.ctx.AddError(&TypeInvalid{Path: string(v.path), Expected: BoolKind, Found: v.inner.Type()})
		return false
	}
	return v.unnamed().Interface().(bool)
}

func (v Value) BoolOptional() (value bool) {
	if v.kind == NullKind {
		return false
	}
	return v.Bool()
}

func (v Value) Map() (m MapValue) {
	if v.dead {
		return MapValue{v: v.kill()}
	}

	if v.kind != MapKind {
		v.ctx.AddError(&TypeInvalid{Path: string(v.path), Expected: MapKind, Found: v.inner.Type()})
		return MapValue{v: v.kill()}
	}

	if v.inner.Type().Key().Kind() != reflect.String {
		v.ctx.AddError(&TypeInvalid{
			Path:     string(v.path),
			Expected: StrKind,
			Found:    v.inner.Type().Key(),
			Msg:      "map must have string keys",
		})
		return MapValue{v: v.kill()}
	}

	return MapValue{v: v}
}

func (v Value) MapOptional() (m MapValue) {
	if v.kind == NullKind {
		return MapValue{v: v.kill()}
	}
	return v.Map()
}

// Returns a value if its kind is the same as the passed kind, otherwise
// returns a dead value. No error is raised.
func (v Value) If(k Kind) Value {
	if v.dead {
		return v
	}
	if v.kind != k {
		return v.kill()
	}
	return v
}

// Returns a map if the value is a map, otherwise returns a dead value. No error is
// raised.
func (v Value) IfMap() MapValue {
	if v.kind != MapKind {
		return MapValue{v: v.kill()}
	}
	return v.Map()
}

// Returns a slice if the value is a slice, otherwise returns a dead value. No error is
// raised.
func (v Value) IfSlice() SliceValue {
	if v.kind != SliceKind {
		return SliceValue{v: v.kill()}
	}
	return v.Slice()
}

func (v Value) Slice() (m SliceValue) {
	if v.dead {
		return SliceValue{v: v.kill()}
	}
	if v.kind != SliceKind {
		v.ctx.AddError(&TypeInvalid{Path: string(v.path), Expected: SliceKind, Found: v.inner.Type()})
		return SliceValue{v: v.kill()}
	}
	return SliceValue{v: v}
}

func (v Value) SliceOptional() (m SliceValue) {
	if v.kind == NullKind {
		return SliceValue{v: v.kill()}
	}
	return v.Slice()
}

func (v Value) Descend(part ...any) Value    { return descend(v, part...) }
func (v Value) TryDescend(part ...any) Value { return tryDescend(v, part...) }

type SliceValue struct {
	v Value
}

var _ value = SliceValue{}

func (s SliceValue) IsNull() bool                 { return s.v.dead || s.v.kind == NullKind }
func (s SliceValue) IsValid() bool                { return !s.v.dead }
func (s SliceValue) Kind() Kind                   { return s.v.kind }
func (s SliceValue) Path() string                 { return string(s.v.path) }
func (s SliceValue) Descend(part ...any) Value    { return descend(s.v, part...) }
func (s SliceValue) TryDescend(part ...any) Value { return tryDescend(s.v, part...) }
func (s SliceValue) Unwrap() any                  { return unwrap(s.v) }
func (s SliceValue) Reject(err error)             { s.v.Reject(err) }

func (s SliceValue) Len() int {
	if s.v.dead || !s.v.inner.IsValid() {
		return 0
	}
	return s.v.inner.Len()
}

func (s SliceValue) Has(idx int) bool {
	return !s.v.dead && idx < s.Len()
}

func (s SliceValue) At(idx int) Value {
	if s.v.dead {
		return s.v
	}

	ln := s.Len()
	if idx < 0 {
		idx = s.Len() - idx
	}
	if idx >= ln {
		s.v.ctx.AddError(&IndexNotFound{
			Path: string(s.v.path),
			Idx:  idx,
		})
		return s.v.kill()
	}
	return s.Try(idx)
}

func (s SliceValue) Try(idx int) (v Value) {
	if s.v.dead {
		return s.v
	}
	if idx >= s.v.inner.Len() {
		return s.v.kill()
	}
	iv := s.v.inner.Index(idx)
	return ValueOf(s.v.ctx, pathWithIdx(s.v.path, idx), iv)
}

func (s SliceValue) Iterate() *SliceIter {
	if s.v.kind == NullKind {
		return &SliceIter{SliceValue: SliceValue{v: s.v.kill()}}
	}
	return &SliceIter{SliceValue: s}
}

type SliceIter struct {
	SliceValue
	idx int
	len int
}

func (iter *SliceIter) Next() bool {
	if !iter.v.inner.IsValid() {
		return false
	}
	iter.idx++
	if iter.idx >= iter.len {
		iter.idx = iter.len
		return false
	}
	return true
}

func (iter *SliceIter) Idx() int { return iter.idx }
func (iter *SliceIter) Len() int { return iter.len }

func (iter *SliceIter) Value() Value {
	return ValueOf(iter.v.ctx,
		pathWithIdx(iter.v.path, iter.idx),
		iter.v.inner.Index(iter.idx))
}

type MapValue struct {
	v Value
}

var _ value = MapValue{}

func (m MapValue) IsNull() bool                 { return m.v.dead || m.v.kind == NullKind }
func (m MapValue) IsValid() bool                { return !m.v.dead }
func (m MapValue) Kind() Kind                   { return m.v.kind }
func (m MapValue) Path() string                 { return string(m.v.path) }
func (m MapValue) Descend(part ...any) Value    { return descend(m.v, part...) }
func (m MapValue) TryDescend(part ...any) Value { return tryDescend(m.v, part...) }
func (m MapValue) Unwrap() any                  { return unwrap(m.v) }
func (m MapValue) Reject(err error)             { m.v.Reject(err) }

func (m MapValue) Len() int {
	if m.v.dead || !m.v.inner.IsValid() {
		return 0
	}
	return m.v.inner.Len()
}

func (m MapValue) Key(key string) Value {
	if m.v.dead {
		return m.v
	}
	v, ok := m.tryInner(key)
	if !ok {
		m.v.ctx.AddError(&KeyNotFound{
			Path: string(m.v.path),
			Key:  key,
		})
		return m.v.kill()
	}
	return v
}

func (m MapValue) Try(key string) (v Value) {
	v, _ = m.tryInner(key)
	return v
}

func (m MapValue) tryInner(key string) (v Value, ok bool) {
	if m.v.dead {
		return m.v, false
	}
	if !m.v.inner.IsValid() {
		return m.v.kill(), false
	}
	kv := m.v.inner.MapIndex(reflect.ValueOf(key))
	if !kv.IsValid() {
		return m.v.kill(), false
	}
	iv := kv.Elem()
	if !iv.IsValid() {
		return m.v.kill(), false
	}
	return ValueOf(
		m.v.ctx,
		pathWithKey(m.v.path, key),
		iv,
	), true
}

// Iterate over a Map.
//
// If the value is null, an iterator with zero items is returned and no error is raised.
// If the iteration encounters a key that is not a string, an error is raised.
//
// Example:
//   for iter := value.Map().Iterate(); iter.Next(); {
//      key := iter.Key()
//      val := iter.Value()
//   }
//
func (m MapValue) Iterate() *MapIter {
	if m.v.kind == NullKind {
		return &MapIter{MapValue: MapValue{v: m.v.kill()}}
	}
	return m.Iterate()
}

type MapIter struct {
	MapValue
	inner *reflect.MapIter
	key   string
	valid bool
}

func (iter *MapIter) Len() int {
	return iter.v.inner.Len()
}

func (iter *MapIter) Next() bool {
	for {
		if ok := iter.inner.Next(); !ok {
			iter.valid = false
			return false
		}

		key := iter.inner.Key()
		if key.Kind() != reflect.String {
			iter.v.ctx.AddError(&TypeInvalid{
				Path:     string(iter.v.path),
				Expected: StrKind,
				Found:    key.Type(),
				Msg:      "map must have string keys",
			})
			continue
		}

		iter.key = key.Interface().(string)

		return true
	}
}

func (iter *MapIter) Key() string {
	if !iter.valid {
		panic(fmt.Errorf("unstructured: %q: attempt to access key in invalid map iterator", iter.v.path))
	}
	return iter.key
}

func (iter *MapIter) Value() Value {
	if !iter.valid {
		panic(fmt.Errorf("unstructured: %q: attempt to access value in invalid map iterator", iter.v.path))
	}
	return ValueOf(
		iter.v.ctx,
		pathWithKey(iter.v.path, iter.key),
		iter.inner.Value(),
	)
}

func (iter *MapIter) RejectKey(err error) {
	if !iter.valid {
		panic(fmt.Errorf("unstructured: %q: attempt to reject key in invalid map iterator", iter.v.path))
	}
	iter.v.ctx.AddError(&KeyInvalid{
		Path: string(iter.v.path),
		Key:  iter.key,
		err:  err,
	})
}

type MapEachFn[C Context, K ~string, V any] func(ctx C, k K, v Value) V

func MapEach[C Context, K ~string, V any](ctx C, value Value, fn MapEachFn[C, K, V]) map[K]V {
	iter := value.Map().Iterate()

	out := make(map[K]V, iter.Len())
	for iter.Next() {
		key := K(iter.Key())
		value := iter.Value()
		out[key] = fn(ctx, key, value)
	}

	return out
}

type SliceEachFn[C Context, V any] func(ctx C, idx int, v Value) V

func SliceEach[C Context, V any](ctx C, value Value, fn SliceEachFn[C, V]) []V {
	iter := value.Slice().Iterate()

	out := make([]V, iter.Len())
	for iter.Next() {
		idx := iter.Idx()
		value := iter.Value()
		out[idx] = fn(ctx, idx, value)
	}

	return out
}

func SliceFloat64s[C Context](ctx C, value Value) []float64 {
	return SliceEach(ctx, value, func(ctx C, idx int, v Value) float64 { return v.Float64() })
}

func SliceInts[C Context](ctx C, value Value) []int {
	return SliceEach(ctx, value, func(ctx C, idx int, v Value) int { return v.Int() })
}

func SliceInt64s[C Context](ctx C, value Value) []int64 {
	return SliceEach(ctx, value, func(ctx C, idx int, v Value) int64 { return v.Int64() })
}

func SliceStrings[C Context](ctx C, value Value) []string {
	return SliceEach(ctx, value, func(ctx C, idx int, v Value) string { return v.Str() })
}

func SliceUints[C Context](ctx C, value Value) []uint {
	return SliceEach(ctx, value, func(ctx C, idx int, v Value) uint { return v.Uint() })
}

func SliceUint64s[C Context](ctx C, value Value) []uint64 {
	return SliceEach(ctx, value, func(ctx C, idx int, v Value) uint64 { return v.Uint64() })
}

func tryDescend(v Value, part ...any) Value {
	var cur = v
	for idx, part := range part {
		switch part := part.(type) {
		case string:
			cur = cur.IfMap().Try(part)
		case int:
			cur = cur.IfSlice().Try(part)
		case int64:
			cur = cur.IfSlice().Try(int(part))
		case uint:
			cur = cur.IfSlice().Try(int(part))
		case uint64:
			cur = cur.IfSlice().Try(int(part))
		default:
			panic(fmt.Errorf("unexpected segment %[1]v (%[1]T) in path at index %[2]d", part, idx))
		}
	}
	return cur
}

func descend(v Value, part ...any) Value {
	var cur = v
	for idx, part := range part {
		switch part := part.(type) {
		case string:
			cur = cur.Map().Key(part)
		case int:
			cur = cur.Slice().At(part)
		case int64:
			cur = cur.Slice().At(int(part))
		case uint:
			cur = cur.Slice().At(int(part))
		case uint64:
			cur = cur.Slice().At(int(part))
		default:
			panic(fmt.Errorf("unexpected segment %[1]v (%[1]T) in path at index %[2]d", part, idx))
		}
	}
	return cur
}

func unwrap(v Value) any {
	if v.dead || !v.inner.IsValid() || !v.inner.CanInterface() {
		return nil
	}
	return v.inner.Interface()
}

package unstructured

import (
	"fmt"
	"reflect"
)

type Path string

func (p Path) withIdx(idx int) Path {
	return Path(fmt.Sprintf("%s/%d", p, idx))
}

func (p Path) withKey(key string) Path {
	return Path(fmt.Sprintf("%s/%s", p, key))
}

type value interface {
	IsNull() bool
	IsValid() bool
	Kind() Kind
	Path() string
	Unwrap() any
	Descend(part ...any) Value
	TryDescend(part ...any) Value
}

type Value struct {
	ctx   Context
	inner reflect.Value
	kind  Kind
	path  Path
	named bool
	dead  bool
}

var _ value = Value{}

func ValueOf(ctx Context, path string, v any) Value {
	var rv reflect.Value
	var ok bool
	var kind Kind
	var named bool

	if rv, ok = v.(reflect.Value); !ok {
		rv = reflect.ValueOf(v)
	}

	if !rv.IsValid() {
		return Value{ctx: ctx, kind: NullKind, path: Path(path), dead: true}
	}

	// If a type comes in wrapped in an interface, unwrap its element. This can happen if
	// you retrieve a reflect.Value via reflect.Value.Index() and pass it in directly:
	if rv.Kind() == reflect.Interface {
		rv = rv.Elem()
	}

	if isNullable(rv.Kind()) && rv.IsNil() {
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
		ctx:   ctx,
		inner: rv,
		kind:  kind,
		path:  Path(path),
		named: named,
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
func (v Value) Kind() Kind    { return v.kind }
func (v Value) Path() string  { return string(v.path) }
func (v Value) Unwrap() any   { return unwrap(v) }

func (v Value) Str() string {
	// XXX: don't call this String(), it makes a mess with fmt.Stringer.
	if v.dead {
		return ""
	}
	if v.inner.Kind() != reflect.String {
		v.ctx.AddError(&InvalidTypeError{Path: string(v.path), Expected: StrKind, Found: v.inner.Type()})
		return ""
	}
	return v.unnamed().Interface().(string)
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
	if v.kind != IntKind {
		v.ctx.AddError(&InvalidTypeError{Path: string(v.path), Expected: IntKind, Found: v.inner.Type()})
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
	if v.kind == IntKind {
		i := v.unnamed().Interface().(int)
		return int64(i)
	}
	if v.kind == Int64Kind {
		return v.unnamed().Interface().(int64)
	}
	v.ctx.AddError(&InvalidTypeError{Path: string(v.path), Expected: Int64Kind, Found: v.inner.Type()})
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
	if v.kind != UintKind {
		v.ctx.AddError(&InvalidTypeError{Path: string(v.path), Expected: UintKind, Found: v.inner.Type()})
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
	if v.kind == UintKind {
		u := v.unnamed().Interface().(uint)
		return uint64(u)
	}
	if v.kind == Uint64Kind {
		return v.unnamed().Interface().(uint64)
	}
	v.ctx.AddError(&InvalidTypeError{Path: string(v.path), Expected: Uint64Kind, Found: v.inner.Type()})
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
	if v.kind != Float64Kind {
		v.ctx.AddError(&InvalidTypeError{Path: string(v.path), Expected: Float64Kind, Found: v.inner.Type()})
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
		v.ctx.AddError(&InvalidTypeError{Path: string(v.path), Expected: BoolKind, Found: v.inner.Type()})
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
		v.ctx.AddError(&InvalidTypeError{Path: string(v.path), Expected: MapKind, Found: v.inner.Type()})
		return MapValue{v: v.kill()}
	}

	if v.inner.Type().Key().Kind() != reflect.String {
		v.ctx.AddError(&InvalidTypeError{
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
		v.ctx.AddError(&InvalidTypeError{Path: string(v.path), Expected: SliceKind, Found: v.inner.Type()})
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

func (s SliceValue) IsNull() bool                 { return s.v.dead || s.v.kind == NullKind }
func (s SliceValue) IsValid() bool                { return !s.v.dead }
func (s SliceValue) Kind() Kind                   { return s.v.kind }
func (s SliceValue) Path() string                 { return string(s.v.path) }
func (s SliceValue) Descend(part ...any) Value    { return descend(s.v, part...) }
func (s SliceValue) TryDescend(part ...any) Value { return tryDescend(s.v, part...) }
func (s SliceValue) Unwrap() any                  { return unwrap(s.v) }

var _ value = SliceValue{}

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
		s.v.ctx.AddError(&IndexNotFoundError{
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
	return ValueOf(s.v.ctx, string(s.v.path.withIdx(idx)), iv)
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
		string(iter.v.path.withIdx(iter.idx)),
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
		m.v.ctx.AddError(&KeyNotFoundError{
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
		string(m.v.path.withKey(key)),
		iv,
	), true
}

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
}

func (iter *MapIter) Len() int {
	return iter.v.inner.Len()
}

func (iter *MapIter) Next() bool {
	for {
		if ok := iter.inner.Next(); !ok {
			return false
		}

		key := iter.inner.Key()
		if key.Kind() != reflect.String {
			iter.v.ctx.AddError(&InvalidTypeError{
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
	return iter.key
}

func (iter *MapIter) Value() Value {
	return ValueOf(
		iter.v.ctx,
		string(iter.v.path.withKey(iter.key)),
		iter.inner.Value(),
	)
}

func MapEach[C any, K ~string, V any](ctx C, value Value, fn func(ctx C, k K, v Value) V) map[K]V {
	iter := value.Map().Iterate()

	out := make(map[K]V, iter.Len())
	for iter.Next() {
		key := K(iter.Key())
		value := iter.Value()
		out[key] = fn(ctx, key, value)
	}

	return out
}

func SliceEach[C any, V any](ctx C, value Value, fn func(ctx C, idx int, v Value) V) []V {
	iter := value.Slice().Iterate()

	out := make([]V, iter.Len())
	for iter.Next() {
		idx := iter.Idx()
		value := iter.Value()
		out[idx] = fn(ctx, idx, value)
	}

	return out
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

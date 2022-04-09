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

type Value struct {
	ctx   Context
	inner reflect.Value
	kind  Kind
	path  Path
	named bool
	dead  bool
}

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

func (v Value) IsNull() bool { return v.kind == NullKind }
func (v Value) Kind() Kind   { return v.kind }
func (v Value) Path() string { return string(v.path) }

func (v Value) Unwrap() any {
	if !v.inner.CanInterface() {
		return nil
	}
	return v.inner.Interface()
}

// XXX: don't call this String(), it makes a mess with fmt.Stringer.
func (v Value) Str() string {
	if v.dead {
		return ""
	}
	if v.inner.Kind() != reflect.String {
		v.ctx.AddError(&InvalidTypeError{Path: string(v.path), Expected: StrKind, Found: v.inner.Type()})
		return ""
	}
	return v.unnamed().Interface().(string)
}

func (v Value) StrOptional() (value string, set bool) {
	if v.dead {
		return "", false
	}
	if v.kind == NullKind {
		return "", false
	}
	value = v.Str()
	return value, true
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

func (v Value) IntOptional() (value int, set bool) {
	if v.dead {
		return 0, false
	}
	if v.kind == NullKind {
		return 0, false
	}
	value = v.Int()
	return value, true
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

func (v Value) Int64Optional() (value int64, set bool) {
	if v.dead {
		return 0, false
	}
	if v.kind == NullKind {
		return 0, false
	}
	value = v.Int64()
	return value, true
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

func (v Value) UintOptional() (value uint, set bool) {
	if v.dead {
		return 0, false
	}
	if v.kind == NullKind {
		return 0, false
	}
	value = v.Uint()
	return value, true
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

func (v Value) Uint64Optional() (value uint64, set bool) {
	if v.dead {
		return 0, false
	}
	if v.kind == NullKind {
		return 0, false
	}
	value = v.Uint64()
	return value, true
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

func (v Value) Float64Optional() (value float64, set bool) {
	if v.dead {
		return 0, false
	}
	if v.kind == NullKind {
		return 0, false
	}
	value = v.Float64()
	return value, true
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

func (v Value) BoolOptional() (value bool, set bool) {
	if v.dead {
		return false, false
	}
	if v.kind == NullKind {
		return false, false
	}
	value = v.Bool()
	return value, true
}

func (v Value) Map() (m MapValue) {
	if v.dead {
		return MapValue{Value: v.kill()}
	}

	if v.kind != MapKind {
		v.ctx.AddError(&InvalidTypeError{Path: string(v.path), Expected: MapKind, Found: v.inner.Type()})
		return MapValue{Value: v.kill()}
	}

	if v.inner.Type().Key().Kind() != reflect.String {
		v.ctx.AddError(&InvalidTypeError{
			Path:     string(v.path),
			Expected: StrKind,
			Found:    v.inner.Type().Key(),
			Msg:      "map must have string keys",
		})
		return MapValue{Value: v.kill()}
	}

	return MapValue{Value: v}
}

func (v Value) MapOptional() (m MapValue, set bool) {
	if v.kind == NullKind {
		return MapValue{Value: v.kill()}, false
	}
	m = v.Map()
	return m, true
}

func (v Value) Key(key string) Value {
	return v.Map().Key(key)
}

func (v Value) KeyOptional(key string) (m Value, set bool) {
	return v.Map().KeyOptional(key)
}

func (v Value) Slice() (m SliceValue) {
	if v.dead {
		return SliceValue{Value: v.kill()}
	}
	if v.kind != SliceKind {
		v.ctx.AddError(&InvalidTypeError{Path: string(v.path), Expected: SliceKind, Found: v.inner.Type()})
		return SliceValue{Value: v.kill()}
	}
	return SliceValue{Value: v}
}

func (v Value) SliceOptional() (m SliceValue, set bool) {
	if v.kind == NullKind {
		return SliceValue{Value: v.kill()}, false
	}
	m = v.Slice()
	return m, true
}

type SliceValue struct {
	Value
}

func (s SliceValue) Len() int {
	if s.dead || !s.inner.IsValid() {
		return 0
	}
	return s.inner.Len()
}

func (s SliceValue) At(idx int) Value {
	v, ok := s.AtOptional(idx)
	if !ok {
		// FIXME error
		return s.kill()
	}
	return v
}

func (s SliceValue) AtOptional(idx int) (v Value, ok bool) {
	if idx >= s.inner.Len() {
		return v, false
	}
	iv := s.inner.Index(idx)
	return ValueOf(s.ctx, string(s.path.withIdx(idx)), iv), true
}

func (s SliceValue) Iterate() *SliceIter {
	return &SliceIter{v: s}
}

func (s SliceValue) IterateOptional() *SliceIter {
	if s.kind == NullKind {
		return &SliceIter{v: SliceValue{Value: s.kill()}}
	}
	return &SliceIter{v: s}
}

type SliceIter struct {
	v   SliceValue
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
	Value
}

func (m MapValue) Len() int {
	if m.dead || !m.inner.IsValid() {
		return 0
	}
	return m.inner.Len()
}

func (m MapValue) Key(key string) Value {
	if m.dead {
		return m.Value
	}
	v, ok := m.KeyOptional(key)
	if !ok {
		m.ctx.AddError(&KeyNotFoundError{
			Path: string(m.path),
			Key:  key,
		})
		return m.kill()
	}
	return v
}

func (m MapValue) KeyOptional(key string) (v Value, set bool) {
	if m.dead {
		return m.Value, false
	}
	if !m.inner.IsValid() {
		return m.kill(), false
	}
	iv := m.inner.MapIndex(reflect.ValueOf(key)).Elem()
	if iv.IsZero() {
		return v, false
	}
	return ValueOf(
		m.ctx,
		string(m.path.withKey(key)),
		iv,
	), true
}

func (m MapValue) Iterate() *MapIter {
	return &MapIter{MapValue: m, inner: m.inner.MapRange()}
}

func (m MapValue) IterateOptional() *MapIter {
	if m.kind == NullKind {
		return &MapIter{MapValue: MapValue{Value: Value{dead: true}}}
	}
	return m.Iterate()
}

type MapIter struct {
	MapValue
	inner *reflect.MapIter
	key   string
}

func (iter *MapIter) Len() int {
	return iter.MapValue.inner.Len()
}

func (iter *MapIter) Next() bool {
	for {
		if ok := iter.inner.Next(); !ok {
			return false
		}

		key := iter.inner.Key()
		if key.Kind() != reflect.String {
			iter.ctx.AddError(&InvalidTypeError{
				Path:     string(iter.path),
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
		iter.ctx,
		string(iter.path.withKey(iter.key)),
		iter.inner.Value(),
	)
}

func MapEach[C any, K ~string, V any](ctx C, value Value, fn func(ctx C, k K, v Value) V) map[K]V {
	iter := value.Map().IterateOptional()

	out := make(map[K]V, iter.Len())
	for iter.Next() {
		key := K(iter.Key())
		value := iter.Value()
		out[key] = fn(ctx, key, value)
	}

	return out
}

func SliceEach[C any, V any](ctx C, value Value, fn func(ctx C, idx int, v Value) V) []V {
	iter := value.Slice().IterateOptional()

	out := make([]V, iter.Len())
	for iter.Next() {
		idx := iter.Idx()
		value := iter.Value()
		out[idx] = fn(ctx, idx, value)
	}

	return out
}

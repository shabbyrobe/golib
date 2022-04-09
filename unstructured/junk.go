package unstructured

// type FromUnstructured interface {
//     FromUnstructured(dctx any, value Value) error
// }
//
// func BuildMap[K ~string, V FromUnstructured](value Value) (map[K]V, error) {
//     iter, err := IterateMap(value)
//     if err != nil {
//         return nil, err
//     }
//
//     out := make(map[K]V, iter.Len())
//     for iter.Next() {
//         value := iter.Value()
//         rawKey, err := String(value)
//         if err != nil {
//             return nil, err
//         }
//
//         key := K(rawKey)
//         var m V
//         if err := m.FromUnstructured(nil, value); err != nil {
//             return nil, err
//         }
//         out[key] = m
//     }
//
//     return out, nil
// }
//
// func Str[T ~string](v Value) T {
//     if v.inner.Kind() != reflect.String {
//         v.ctx.AddError(&InvalidTypeError{Path: string(v.path), Expected: reflect.String, Found: v.inner.Kind()})
//         return ""
//     }
//     return T(v.inner.Interface().(string))
// }
//
// func StrOptional[T ~string](v Value) (value T, set bool) {
//     if v.inner.IsNil() {
//         return "", false
//     }
//     value = Str[T](v)
//     return value, true
// }
//
// func Int[T ~int](v Value) T {
//     if v.inner.Kind() != reflect.Int {
//         v.ctx.AddError(&InvalidTypeError{Path: string(v.path), Expected: reflect.Int, Found: v.inner.Kind()})
//         return 0
//     }
//     return T(v.inner.Interface().(int))
// }
//
// func IntOptional[T ~int](v Value) (value T, set bool) {
//     if v.inner.IsNil() {
//         return 0, false
//     }
//     value = Int[T](v)
//     return value, true
// }
//
// func Int64[T ~int64](v Value) T {
//     if v.inner.Kind() == reflect.Int {
//         i := v.inner.Interface().(int)
//         return T(i)
//     }
//     if v.inner.Kind() == reflect.Int64 {
//         return T(v.inner.Interface().(int64))
//     }
//     v.ctx.AddError(&InvalidTypeError{Path: string(v.path), Expected: reflect.Int64, Found: v.inner.Kind()})
//     return 0
// }
//
// func Int64Optional[T ~int64](v Value) (value T, set bool) {
//     if v.inner.IsNil() {
//         return 0, false
//     }
//     value = Int64[T](v)
//     return value, true
// }
//
// func Uint[T ~uint](v Value) T {
//     if v.inner.Kind() != reflect.Uint {
//         v.ctx.AddError(&InvalidTypeError{Path: string(v.path), Expected: reflect.Uint, Found: v.inner.Kind()})
//         return 0
//     }
//     return T(v.inner.Interface().(uint))
// }
//
// func UintOptional[T ~uint](v Value) (value T, set bool) {
//     if v.inner.IsNil() {
//         return 0, false
//     }
//     value = Uint[T](v)
//     return value, true
// }
//
// func Uint64[T ~uint64](v Value) T {
//     if v.inner.Kind() == reflect.Uint {
//         u := v.inner.Interface().(uint)
//         return T(u)
//     }
//     if v.inner.Kind() == reflect.Uint64 {
//         return T(v.inner.Interface().(uint64))
//     }
//     v.ctx.AddError(&InvalidTypeError{Path: string(v.path), Expected: reflect.Uint64, Found: v.inner.Kind()})
//     return 0
// }
//
// func Uint64Optional[T ~uint64](v Value) (value T, set bool) {
//     if v.inner.IsNil() {
//         return 0, false
//     }
//     value = Uint64[T](v)
//     return value, true
// }
//
// func Float64[T ~float64](v Value) T {
//     if v.inner.Kind() != reflect.Float64 {
//         v.ctx.AddError(&InvalidTypeError{Path: string(v.path), Expected: reflect.Float64, Found: v.inner.Kind()})
//         return 0
//     }
//     return T(v.inner.Interface().(float64))
// }
//
// func Float64Optional[T ~float64](v Value) (value T, set bool) {
//     if v.inner.IsNil() {
//         return 0, false
//     }
//     value = Float64[T](v)
//     return value, true
// }
//
// func Bool[T ~bool](v Value) T {
//     if v.inner.Kind() != reflect.Bool {
//         v.ctx.AddError(&InvalidTypeError{Path: string(v.path), Expected: reflect.Bool, Found: v.inner.Kind()})
//         return false
//     }
//     return T(v.inner.Interface().(bool))
// }
//
// func BoolOptional[T ~bool](v Value) (value T, set bool) {
//     if v.inner.IsNil() {
//         return false, false
//     }
//     value = Bool[T](v)
//     return value, true
// }
//

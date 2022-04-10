package unstructured

import "reflect"

type Kind int

const (
	InvalidKind Kind = iota
	IntKind
	Int64Kind
	UintKind
	Uint64Kind
	Float64Kind
	BoolKind
	NullKind
	MapKind
	SliceKind
	StrKind
	NumberKind
)

func (k Kind) String() string {
	switch k {
	case BoolKind:
		return "bool"
	case IntKind:
		return "int"
	case Int64Kind:
		return "int64"
	case UintKind:
		return "uint"
	case Uint64Kind:
		return "uint64"
	case Float64Kind:
		return "float64"
	case MapKind:
		return "map"
	case SliceKind:
		return "slice"
	case StrKind:
		return "str"
	case NullKind:
		return "null"
	case NumberKind:
		return "number"
	default:
		return "<unknown>"
	}
}

func kindOf(v reflect.Type) Kind {
	if v == numberType {
		return NumberKind
	}
	switch v.Kind() {
	case reflect.Bool:
		return BoolKind
	case reflect.Int:
		return IntKind
	case reflect.Int64:
		return Int64Kind
	case reflect.Uint:
		return UintKind
	case reflect.Uint64:
		return Uint64Kind
	case reflect.Float64:
		return Float64Kind
	case reflect.Map:
		return MapKind
	case reflect.Slice:
		return SliceKind
	case reflect.String:
		return StrKind
	default:
		return InvalidKind
	}
}

func isValidKind(kind reflect.Kind) bool {
	return kind == reflect.Bool ||
		kind == reflect.Int ||
		kind == reflect.Int64 ||
		kind == reflect.Uint ||
		kind == reflect.Uint64 ||
		kind == reflect.Float64 ||
		kind == reflect.Map ||
		kind == reflect.Slice ||
		kind == reflect.String
}

func isNullable(kind reflect.Kind) bool {
	return kind == reflect.Ptr ||
		kind == reflect.Map ||
		kind == reflect.Slice
}

func isValid(v reflect.Value) bool {
	kind := v.Kind()
	if isValidKind(kind) {
		return true
	}
	if kind == reflect.Ptr {
		v = v.Elem()
		return isValidKind(v.Kind())
	}

	return false
}

var basic = []reflect.Type{
	BoolKind:    reflect.TypeOf(false),
	IntKind:     reflect.TypeOf(int(0)),
	Int64Kind:   reflect.TypeOf(int64(0)),
	UintKind:    reflect.TypeOf(uint(0)),
	Uint64Kind:  reflect.TypeOf(uint64(0)),
	Float64Kind: reflect.TypeOf(float64(0)),
	StrKind:     reflect.TypeOf(""),
}

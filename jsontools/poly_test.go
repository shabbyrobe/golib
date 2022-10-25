package jsontools

import (
	"reflect"
	"testing"
)

type Foo struct {
	Foo string
}

func (f Foo) Kind() string { return "foo" }
func (f Foo) Direct()      {}
func (f *Foo) Ptr()        {}

type Bar struct {
	Bar string
}

func (b Bar) Kind() string { return "bar" }
func (b Bar) Direct()      {}
func (b *Bar) Ptr()        {}

type Ptr interface {
	Kind() string
	Ptr()
}

type Direct interface {
	Kind() string
	Direct()
}

func DirectKind(d Direct) string { return d.Kind() }

func DirectFactory(kind string) (Direct, error) {
	switch kind {
	case "foo":
		return Foo{}, nil
	case "bar":
		return Bar{}, nil
	default:
		panic("nah")
	}
}

func PtrKind(p Ptr) string { return p.Kind() }

func PtrFactory(kind string) (Ptr, error) {
	switch kind {
	case "foo":
		return &Foo{}, nil
	case "bar":
		return &Bar{}, nil
	default:
		panic("nah")
	}
}

func TestPolymorphicJSONDirect(t *testing.T) {
	var v Foo
	var d Direct = v
	bts, err := MarshalPolymorphic(d, "kind", DirectKind)
	if err != nil {
		t.Fatal(err)
	}

	var result Direct
	if err := UnmarshalPolymorphic(bts, "kind", DirectFactory, &result); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(d, result) {
		t.Fatal()
	}
}

func TestPolymorphicJSONPtr(t *testing.T) {
	var v Foo
	var p Ptr = &v
	bts, err := MarshalPolymorphic(p, "kind", PtrKind)
	if err != nil {
		t.Fatal(err)
	}

	var result Ptr
	if err := UnmarshalPolymorphic(bts, "kind", PtrFactory, &result); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(p, result) {
		t.Fatal()
	}
}

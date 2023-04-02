package qs

import (
	"net"
	"net/url"
	"testing"
)

type TestInt int
type TestInt64 int64
type TestUint64 uint64

func mustParseQuery(t *testing.T, q string) url.Values {
	t.Helper()
	values, err := url.ParseQuery(q)
	if err != nil {
		t.Fatal(err)
	}
	return values
}

func assertLoaderOK(t *testing.T, loader *Loader) {
	t.Helper()
	if err := loader.Err(); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeFirstInt(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		loader := NewLoader(mustParseQuery(t, "foo=1"))
		if Val(Int(loader.First("foo"))) != int(1) {
			t.Fatal()
		}
		assertLoaderOK(t, loader)
	})

	t.Run("anyint", func(t *testing.T) {
		loader := NewLoader(mustParseQuery(t, "foo=1"))
		if Val(AnyInt[TestInt](loader.First("foo"))) != TestInt(1) {
			t.Fatal()
		}
		assertLoaderOK(t, loader)
	})

	t.Run("int64", func(t *testing.T) {
		loader := NewLoader(mustParseQuery(t, "foo=1"))
		if Val(Int64(loader.First("foo"))) != int64(1) {
			t.Fatal()
		}
		assertLoaderOK(t, loader)
	})

	t.Run("anyint64", func(t *testing.T) {
		loader := NewLoader(mustParseQuery(t, "foo=1"))
		if Val(AnyInt64[TestInt64](loader.First("foo"))) != TestInt64(1) {
			t.Fatal()
		}
		assertLoaderOK(t, loader)
	})

	t.Run("uint64", func(t *testing.T) {
		loader := NewLoader(mustParseQuery(t, "foo=1"))
		if Val(Uint64(loader.First("foo"))) != uint64(1) {
			t.Fatal()
		}
		assertLoaderOK(t, loader)
	})

	t.Run("anyuint64", func(t *testing.T) {
		loader := NewLoader(mustParseQuery(t, "foo=1"))
		if Val(AnyUint64[TestUint64](loader.First("foo"))) != TestUint64(1) {
			t.Fatal()
		}
		assertLoaderOK(t, loader)
	})
}

func TestDecodeFirstTextUnmarshaler(t *testing.T) {
	values := mustParseQuery(t, "foo=10.0.0.1")
	loader := NewLoader(values)
	ip := Val(Text[net.IP](loader.First("foo")))
	if ip.String() != "10.0.0.1" {
		t.Fatal(ip.String())
	}
	assertLoaderOK(t, loader)
}

func TestDecodePtrPresent(t *testing.T) {
	loader := NewLoader(mustParseQuery(t, "foo=1"))
	var v *int = Ptr(Int(loader.First("foo")))
	if v == nil || *v != 1 {
		t.Fatal(v)
	}
	assertLoaderOK(t, loader)
}

func TestDecodePtrMissing(t *testing.T) {
	loader := NewLoader(mustParseQuery(t, "foo=1"))
	var v *int = Ptr(Int(loader.First("bar")))
	if v != nil {
		t.Fatal(v)
	}
	assertLoaderOK(t, loader)
}

package jsontools

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func ToMap(v any) (out map[string]any, err error) {
	bts, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bts, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func UnmarshalMap(m map[string]any, into any) (err error) {
	bts, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bts, into); err != nil {
		return err
	}
	return nil
}

func MarshalPolymorphic[K ~string, V any](
	v V,
	kindKey string,
	keyFn func(v any) (K, error),
) (raw []byte, err error) {
	m, err := ToMap(v)
	if err != nil {
		return nil, err
	}
	if m != nil {
		key, err := keyFn(v)
		if err != nil {
			return nil, err
		}
		m[kindKey] = string(key)
	}
	return json.Marshal(m)
}

func MarshalIndentPolymorphic[K ~string](
	v any,
	kindKey string,
	keyFn func(v any) (K, error),
	indent string,
) (raw []byte, err error) {
	m, err := ToMap(v)
	if err != nil {
		return nil, err
	}
	if m != nil {
		key, err := keyFn(v)
		if err != nil {
			return nil, err
		}
		m[kindKey] = string(key)
	}
	return json.MarshalIndent(m, "", indent)
}

func UnmarshalPolymorphic[K ~string, V any](
	raw []byte,
	kindKey string,
	factory func(K) (V, error),
	into *V,
) (err error) {
	var tmp map[string]any
	if err := json.Unmarshal(raw, &tmp); err != nil {
		return err
	}
	if tmp == nil {
		return nil
	}

	kindAny, ok := tmp[string(kindKey)]
	if !ok {
		return fmt.Errorf("key %q not found in object", kindKey)
	}

	kind, ok := kindAny.(string)
	if !ok {
		return fmt.Errorf("key %q is not a string, found %T", kindKey, kindAny)
	}

	v, err := factory(K(kind))
	if err != nil {
		return err
	}

	// XXX: We need to do some gymnastics here otherwise json.Unmarshal can't see the
	// concrete type emitted by our factory (which will have an interface as its return
	// type). We also need to take special care if we are using non-pointer types as the
	// values inside the interface.
	vv := reflect.ValueOf(v)
	ptr := false
	if vv.Kind() != reflect.Pointer {
		vp := reflect.New(vv.Type())
		vp.Elem().Set(vv)
		vv = vp
		ptr = true
	}

	delete(tmp, kindKey)
	vi := vv.Interface()
	if err := UnmarshalMap(tmp, vi); err != nil {
		return err
	}

	// If we nested a pointer before unmarshalling, we need to unwrap it again:
	if ptr {
		vv := reflect.ValueOf(vi).Elem()
		*into = vv.Interface().(V)
	} else {
		*into = vi.(V)
	}

	return nil
}

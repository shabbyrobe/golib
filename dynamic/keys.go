package dynamic

import (
	"fmt"
	"strconv"
)

// Traverses a dynamic value and converts any maps with non-string primitive keys into
// maps with string keys, as dynamic.Value only supports maps with string keys.
//
// This primarily exists to ease treating the results of decoding
// https://github.com/go-yaml/yaml, which allows non-string keys.
//
// If a key would clobber another after stringification, an error is raised.
func StringifyKeys(value any, path string) (out any, err error) {
	switch value := value.(type) {
	case map[any]any:
		return stringifyAnyMapKeys(value, path)
	case map[string]any:
		return stringifyMapKeys(value, path)
	case []any:
		for i := range value {
			value[i], err = StringifyKeys(value[i], fmt.Sprintf("%s/%d", path, i))
			if err != nil {
				return nil, err // FIXME: include path
			}
		}
		return value, nil
	default:
		return value, nil
	}
}

func stringifyMapKeys(in map[string]any, path string) (out map[string]any, err error) {
	out = make(map[string]any, len(in))
	for key := range in {
		out[key], err = StringifyKeys(in[key], path+"/"+key)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func stringifyAnyMapKeys(in map[any]any, path string) (out map[string]any, err error) {
	out = make(map[string]any, len(in))
	for rawKey, value := range in {
		var key string
		switch typeKey := rawKey.(type) {
		case string:
			key = typeKey
		case int:
			key = strconv.FormatInt(int64(typeKey), 10)
		case int64:
			key = strconv.FormatInt(typeKey, 10)
		case uint:
			key = strconv.FormatUint(uint64(typeKey), 10)
		case uint64:
			key = strconv.FormatUint(typeKey, 10)
		default:
			return nil, err // FIXME: include path
		}

		if _, ok := out[key]; ok {
			return nil, fmt.Errorf("key doesn't exist")
		}

		out[key], err = StringifyKeys(value, path+"/"+key)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

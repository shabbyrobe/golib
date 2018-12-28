package jsontools

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// StringFromScalar forces any scalar value (numeric, bool, string, null)
// to be a string. Useful for murky APIs where you are building a struct but
// are unsure what the exact scalar type of a value might turn out to be, but a
// string will do as a simple representation.
type StringFromScalar string

func (s *StringFromScalar) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch v := v.(type) {
	case nil:
		*s = ""
	case float64:
		*s = StringFromScalar(strconv.FormatFloat(v, 'f', -1, 64))
	case bool:
		*s = StringFromScalar(strconv.FormatBool(v))
	case string:
		*s = StringFromScalar(v)
	default:
		return fmt.Errorf("cannot convert type %T to string", v)
	}

	return nil
}

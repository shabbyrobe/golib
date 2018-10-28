package jsontools

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/shabbyrobe/golib/times"
)

type FloatTimeSecs time.Time

func (t FloatTimeSecs) MarshalJSON() ([]byte, error) {
	return json.Marshal(times.ToFloat64Secs(time.Time(t)))
}

func (t *FloatTimeSecs) UnmarshalJSON(b []byte) error {
	var f float64
	if err := json.Unmarshal(b, &f); err != nil {
		return err
	}
	*t = FloatTimeSecs(times.FromFloat64Secs(f))
	return nil
}

type IntTimeSecs time.Time

func (t IntTimeSecs) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(t).Unix())
}

func (t *IntTimeSecs) UnmarshalJSON(b []byte) error {
	var f float64
	if err := json.Unmarshal(b, &f); err != nil {
		return err
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return fmt.Errorf("input %q is an invalid unix time", string(b))
	}
	*t = IntTimeSecs(time.Unix(int64(math.RoundToEven(f)), 0))
	return nil
}

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

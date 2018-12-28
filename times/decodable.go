package times

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"
)

// DurationString provides a time.Duration that marshals to/from a string
// using time.Duration.String()/time.ParseDuration().
type DurationString time.Duration

func (d DurationString) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Duration(d).String() + `"`), nil
}

func (d *DurationString) UnmarshalJSON(b []byte) (err error) {
	s, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	td, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = DurationString(td)
	return nil
}

type UnixSecInt time.Time

func (u UnixSecInt) Time() time.Time {
	return time.Time(u)
}

func (u UnixSecInt) MarshalJSON() ([]byte, error) {
	ut := time.Time(u).Unix()
	return json.Marshal(ut)
}

func (u *UnixSecInt) UnmarshalJSON(bts []byte) error {
	var uv float64
	if err := json.Unmarshal(bts, &uv); err != nil {
		return err
	}
	if math.IsNaN(uv) || math.IsInf(uv, 0) {
		return fmt.Errorf("input %q is an invalid unix time", string(bts))
	}
	var ui = int64(uv)
	*u = UnixSecInt(time.Unix(ui, 0).In(time.UTC))
	return nil
}

type UnixSecFloat time.Time

func (u UnixSecFloat) Time() time.Time {
	return time.Time(u)
}

func (u UnixSecFloat) MarshalJSON() ([]byte, error) {
	ut := ToFloat64Secs(time.Time(u))
	return json.Marshal(ut)
}

func (u *UnixSecFloat) UnmarshalJSON(bts []byte) error {
	var uv float64
	if err := json.Unmarshal(bts, &uv); err != nil {
		return err
	}
	if math.IsNaN(uv) || math.IsInf(uv, 0) {
		return fmt.Errorf("input %q is an invalid unix time", string(bts))
	}
	*u = UnixSecFloat(FromFloat64SecsLocation(uv, time.UTC))
	return nil
}

type UnixMsecInt time.Time

func (u UnixMsecInt) Time() time.Time {
	return time.Time(u)
}

func (u UnixMsecInt) MarshalJSON() ([]byte, error) {
	ut := time.Time(u).UnixNano() / int64(time.Millisecond)
	return json.Marshal(ut)
}

func (u *UnixMsecInt) UnmarshalJSON(bts []byte) error {
	var uv float64
	if err := json.Unmarshal(bts, &uv); err != nil {
		return err
	}
	if math.IsNaN(uv) || math.IsInf(uv, 0) {
		return fmt.Errorf("input %q is an invalid unix time", string(bts))
	}
	var nsec = int64(uv) * 1000000
	*u = UnixMsecInt(time.Unix(0, nsec).In(time.UTC))
	return nil
}

type UnixMsecFloat time.Time

func (u UnixMsecFloat) Time() time.Time {
	return time.Time(u)
}

func (u UnixMsecFloat) MarshalJSON() ([]byte, error) {
	ut := float64(time.Time(u).Unix()) + (float64(time.Time(u).Nanosecond()) / 1000000.0)
	return json.Marshal(ut)
}

func (u *UnixMsecFloat) UnmarshalJSON(bts []byte) error {
	var uv float64
	if err := json.Unmarshal(bts, &uv); err != nil {
		return err
	}
	if math.IsNaN(uv) || math.IsInf(uv, 0) {
		return fmt.Errorf("input %q is an invalid unix time", string(bts))
	}
	t := time.Unix(int64(uv), int64(math.Mod(uv, 1)*1000000)).In(time.UTC)
	*u = UnixMsecFloat(t)
	return nil
}

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
type DurationString struct {
	// Turns out wrapping the type in a struct is ever so slightly faster than
	// deriving ('type DurationString time.Duration').
	time.Duration
}

func (d DurationString) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Duration.String() + `"`), nil
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
	*d = DurationString{td}
	return nil
}

// DurationMsecFloat provides a time.Duration that marshals to/from a float
// representing milliseconds.
type DurationMsecFloat time.Duration

func (d DurationMsecFloat) Duration() time.Duration {
	return time.Duration(d)
}

func (d DurationMsecFloat) String() string {
	return time.Duration(d).String()
}

func (d DurationMsecFloat) MarshalJSON() ([]byte, error) {
	fv := float64(d) / float64(time.Millisecond)
	return json.Marshal(fv)
}

func (d *DurationMsecFloat) UnmarshalJSON(b []byte) (err error) {
	fv, err := strconv.ParseFloat(string(b), 64)
	if err != nil {
		return err
	}
	fd := time.Duration(fv * float64(time.Millisecond))
	*d = DurationMsecFloat(fd)
	return nil
}

// DurationSecFloat provides a time.Duration that marshals to/from a float
// representing seconds.
type DurationSecFloat time.Duration

func (d DurationSecFloat) Duration() time.Duration {
	return time.Duration(d)
}

func (d DurationSecFloat) String() string {
	return time.Duration(d).String()
}

func (d DurationSecFloat) MarshalJSON() ([]byte, error) {
	fv := float64(d) / float64(time.Second)
	return json.Marshal(fv)
}

func (d *DurationSecFloat) UnmarshalJSON(b []byte) (err error) {
	fv, err := strconv.ParseFloat(string(b), 64)
	if err != nil {
		return err
	}
	fd := time.Duration(fv * float64(time.Second))
	*d = DurationSecFloat(fd)
	return nil
}

// DurationSecInt64 provides a time.Duration that marshals to/from an int64
// representing seconds.
type DurationSecInt64 time.Duration

func (d DurationSecInt64) Duration() time.Duration {
	return time.Duration(d)
}

func (d DurationSecInt64) String() string {
	return time.Duration(d).String()
}

func (d DurationSecInt64) MarshalJSON() ([]byte, error) {
	iv := int64(time.Duration(d) / time.Second)
	is := strconv.FormatInt(iv, 10)
	return []byte(is), nil
}

func (d *DurationSecInt64) UnmarshalJSON(b []byte) (err error) {
	iv, err := strconv.ParseInt(string(b), 0, 64)
	if err != nil {
		return err
	}
	fd := time.Duration(iv) * time.Second
	*d = DurationSecInt64(fd)
	return nil
}

// DurationMsecInt64 provides a time.Duration that marshals to/from an int64
// representing milliseconds.
type DurationMsecInt64 time.Duration

func (d DurationMsecInt64) Duration() time.Duration {
	return time.Duration(d)
}

func (d DurationMsecInt64) String() string {
	return time.Duration(d).String()
}

func (d DurationMsecInt64) MarshalJSON() ([]byte, error) {
	iv := int64(time.Duration(d) / time.Millisecond)
	is := strconv.FormatInt(iv, 10)
	return []byte(is), nil
}

func (d *DurationMsecInt64) UnmarshalJSON(b []byte) (err error) {
	iv, err := strconv.ParseInt(string(b), 0, 64)
	if err != nil {
		return err
	}
	fd := time.Duration(iv) * time.Millisecond
	*d = DurationMsecInt64(fd)
	return nil
}

// UnixSecInt provides a time.Time that marshals to/from an int representing
// the number of whole seconds since the Unix epoch.
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

// UnixSecInt provides a time.Time that marshals to/from a float representing
// the number of whole or partial seconds since the Unix epoch.
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

// UnixMsecInt provides a time.Time that marshals to/from an int representing
// the number of whole milliseconds since the Unix epoch.
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

// UnixMsecFloat provides a time.Time that marshals to/from a float representing
// the number of whole or partial milliseconds since the Unix epoch.
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

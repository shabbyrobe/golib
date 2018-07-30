package jsontools

import (
	"encoding/json"
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

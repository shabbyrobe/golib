package times

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shabbyrobe/golib/assert"
)

func TestUnixSecInt(t *testing.T) {
	for _, tc := range []struct {
		in  time.Time
		out time.Time
	}{
		{time.Date(2018, 1, 1, 12, 0, 0, 0, time.UTC), time.Date(2018, 1, 1, 12, 0, 0, 0, time.UTC)},
		{time.Date(2018, 1, 1, 12, 0, 1, 0, time.UTC), time.Date(2018, 1, 1, 12, 0, 1, 0, time.UTC)},
		{time.Date(2018, 1, 1, 12, 0, 1, 1, time.UTC), time.Date(2018, 1, 1, 12, 0, 1, 0, time.UTC)},
	} {
		t.Run("", func(t *testing.T) {
			tt := assert.WrapTB(t)
			var ut = UnixSecInt(tc.in)
			bts, err := json.Marshal(ut)
			tt.MustOK(err)

			var ot UnixSecInt
			tt.MustOK(json.Unmarshal(bts, &ot))
			tt.MustEqual(tc.out, ot.Time())
		})
	}
}

func TestUnixMsecInt(t *testing.T) {
	for _, tc := range []struct {
		in  time.Time
		out time.Time
	}{
		{time.Date(2018, 1, 1, 12, 0, 0, 0, time.UTC), time.Date(2018, 1, 1, 12, 0, 0, 0, time.UTC)},
		{time.Date(2018, 1, 1, 12, 0, 1, 0, time.UTC), time.Date(2018, 1, 1, 12, 0, 1, 0, time.UTC)},
		{time.Date(2018, 1, 1, 12, 0, 1, 1, time.UTC), time.Date(2018, 1, 1, 12, 0, 1, 0, time.UTC)},
		{time.Date(2018, 1, 1, 12, 0, 1, 1000000, time.UTC), time.Date(2018, 1, 1, 12, 0, 1, 1000000, time.UTC)},
		{time.Date(2018, 1, 1, 12, 0, 1, 999999999, time.UTC), time.Date(2018, 1, 1, 12, 0, 1, 999000000, time.UTC)},
	} {
		t.Run("", func(t *testing.T) {
			tt := assert.WrapTB(t)
			var ut = UnixMsecInt(tc.in)
			bts, err := json.Marshal(ut)
			tt.MustOK(err)

			var ot UnixMsecInt
			tt.MustOK(json.Unmarshal(bts, &ot))
			tt.MustEqual(tc.out, ot.Time())
		})
	}
}

package times

import (
	"encoding/json"
	"testing"
	"time"
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
			var ut = UnixSecInt(tc.in)
			bts, err := json.Marshal(ut)
			if err != nil {
				t.Fatal(err)
			}

			var ot UnixSecInt
			if err := json.Unmarshal(bts, &ot); err != nil {
				t.Fatal(err)
			}
			if !tc.out.Equal(ot.Time()) {
				t.Fatal()
			}
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
			var ut = UnixMsecInt(tc.in)
			bts, err := json.Marshal(ut)
			if err != nil {
				t.Fatal(err)
			}

			var ot UnixMsecInt
			if err := json.Unmarshal(bts, &ot); err != nil {
				t.Fatal(err)
			}
			if !tc.out.Equal(ot.Time()) {
				t.Fatal()
			}
		})
	}
}

func BenchmarkDurationString(b *testing.B) {
	var ds DurationString
	var in = []byte(`"1m"`)

	for i := 0; i < b.N; i++ {
		if err := json.Unmarshal(in, &ds); err != nil {
			panic(err)
		}
	}
}

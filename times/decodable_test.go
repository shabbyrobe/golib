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

func TestDurationMsecFloat(t *testing.T) {
	var x = DurationMsecFloat(time.Microsecond * 123456789)
	bts, err := json.Marshal(x)
	if err != nil {
		t.Fatal(err)
	}
	if string(bts) != "123456.789" {
		t.Fatal(string(bts))
	}

	var y DurationMsecFloat
	if err := json.Unmarshal(bts, &y); err != nil {
		t.Fatal(err)
	}
	if y.Duration() != 123456789*time.Microsecond {
		t.Fatal(y)
	}
}

func TestDurationSecFloat(t *testing.T) {
	var x = DurationSecFloat(time.Millisecond * 12345)
	bts, err := json.Marshal(x)
	if err != nil {
		t.Fatal(err)
	}
	if string(bts) != "12.345" {
		t.Fatal(string(bts))
	}

	var y DurationSecFloat
	if err := json.Unmarshal(bts, &y); err != nil {
		t.Fatal(err)
	}
	if y.Duration() != 12345*time.Millisecond {
		t.Fatal(y)
	}
}

func TestDurationMsecInt64(t *testing.T) {
	var x = DurationMsecInt64(time.Second * 10)
	bts, err := json.Marshal(x)
	if err != nil {
		t.Fatal(err)
	}
	if string(bts) != "10000" {
		t.Fatal(string(bts))
	}

	var y DurationMsecInt64
	if err := json.Unmarshal(bts, &y); err != nil {
		t.Fatal(err)
	}
	if y.Duration() != 10*time.Second {
		t.Fatal(y)
	}
}

func TestDurationSecInt64(t *testing.T) {
	var x = DurationSecInt64(time.Second * 10)
	bts, err := json.Marshal(x)
	if err != nil {
		t.Fatal(err)
	}
	if string(bts) != "10" {
		t.Fatal(string(bts))
	}

	var y DurationSecInt64
	if err := json.Unmarshal(bts, &y); err != nil {
		t.Fatal(err)
	}
	if y.Duration() != 10*time.Second {
		t.Fatal(y)
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

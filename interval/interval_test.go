package interval

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestInterval(t *testing.T) {
	tz := time.FixedZone("+10:00", 36000)

	for idx, tc := range []struct {
		in       Interval
		at       time.Time
		expected time.Time
	}{
		{Raw(4, Month), time.Date(2017, 7, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2017, 5, 1, 0, 0, 0, 0, time.UTC)},

		{Raw(3, Week), time.Date(2017, 7, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2017, 6, 26, 0, 0, 0, 0, time.UTC)},

		{Raw(4, Hour), time.Date(2017, 6, 26, 5, 30, 0, 0, time.UTC),
			time.Date(2017, 6, 26, 4, 0, 0, 0, time.UTC)},

		{Raw(4, Month), time.Date(2017, 7, 1, 0, 0, 0, 0, tz),
			time.Date(2017, 5, 1, 0, 0, 0, 0, tz)},

		{Raw(3, Week), time.Date(2017, 7, 1, 0, 0, 0, 0, tz),
			time.Date(2017, 6, 26, 0, 0, 0, 0, tz)},

		// This is 02:00 instead of 04:00 because the timezone offset is 10 hours
		// and the buckets start from UTC.
		{Raw(4, Hour), time.Date(2017, 6, 26, 5, 30, 0, 0, tz),
			time.Date(2017, 6, 26, 2, 0, 0, 0, tz)},
	} {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			if !tc.in.Start(tc.at).Equal(tc.expected) {
				t.Fatal(tc.in, "at", tc.at, "!=", tc.expected)
			}
		})
	}
}

var stringCases = []struct {
	in       Interval
	expected string
}{
	{Raw(1, Second), "1sec"},
	{Raw(2, Second), "2sec"},
	{Raw(1, Minute), "1min"},
	{Raw(2, Minute), "2min"},
	{Raw(1, Hour), "1hr"},
	{Raw(2, Hour), "2hr"},
	{Raw(1, Week), "1wk"},
	{Raw(2, Week), "2wk"},
	{Raw(1, Day), "1d"},
	{Raw(2, Day), "2d"},
	{Raw(1, Month), "1mo"},
	{Raw(2, Month), "2mo"},
	{Raw(1, Year), "1yr"},
	{Raw(2, Year), "2yr"},
}

func TestString(t *testing.T) {
	for idx, tc := range stringCases {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			result := tc.in.String()
			if result != tc.expected {
				t.Fatal(result)
			}
		})
	}
}

func TestMarshal(t *testing.T) {
	for idx, tc := range stringCases {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			inEnc := IntervalEncoded(tc.in)
			mt, err := inEnc.MarshalText()
			if err != nil {
				t.Fatal(err)
			}
			if string(mt) != tc.expected {
				t.Fatal(string(mt), "!=", tc.expected)
			}

			out, err := json.Marshal(inEnc)
			if err != nil {
				t.Fatal(err)
			}
			expectedJSON := `"` + tc.expected + `"`
			if string(out) != expectedJSON {
				t.Fatal(string(out), "!=", expectedJSON)
			}

			var r IntervalEncoded
			if err := json.Unmarshal(out, &r); err != nil {
				t.Fatal(err)
			}

			if tc.in != r.Interval() {
				t.Fatal(r.Interval(), "!=", tc.in)
			}
		})
	}
}

/*
func TestConvertFuzz(t *testing.T) {
	for i := 0; i < 100; i++ {
		from, to := randomDivisibleIntervals(nil)
		period := randomPeriod(nil, from)
		if from != to {
			t.Fail()
		}

		there := from.ConvertPeriodTo(period, to)
		back := to.ConvertPeriodTo(there, from)
		if period != back {
			t.Fail()
		}
	}
}
*/

func TestSort(t *testing.T) {
	for idx, tc := range []struct {
		in []Interval
		ex []Interval
	}{
		{[]Interval{Raw(61, Minute), Raw(1, Hour), Raw(59, Minute)},
			[]Interval{Raw(59, Minute), Raw(1, Hour), Raw(61, Minute)}},
		{[]Interval{Raw(61, Second), Raw(1, Minute), Raw(59, Second)},
			[]Interval{Raw(59, Second), Raw(1, Minute), Raw(61, Second)}},
	} {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			sort.Slice(tc.in, func(i, j int) bool { return tc.in[i].Less(tc.in[j]) })
			if !reflect.DeepEqual(tc.in, tc.ex) {
				t.Fatal(tc.in, "!=", tc.ex)
			}
		})
	}
}

func TestPeriodTime(t *testing.T) {
	for i, c := range []struct {
		Interval   Interval
		Period     Period
		TestTime   time.Time
		PeriodTime time.Time
	}{
		// 1 second
		{Raw(1, Second), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Second), 0, time.Date(1970, 1, 1, 0, 0, 0, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Second), 1, time.Date(1970, 1, 1, 0, 0, 1, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 1, 0, time.UTC)},
		{Raw(1, Second), 1, time.Date(1970, 1, 1, 0, 0, 1, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 1, 0, time.UTC)},
		{Raw(1, Second), 2, time.Date(1970, 1, 1, 0, 0, 2, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 2, 0, time.UTC)},
		{Raw(1, Second), -1, time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 59, 59, 0, time.UTC)},
		{Raw(1, Second), -1, time.Date(1969, 12, 31, 23, 59, 59, 0, time.UTC), time.Date(1969, 12, 31, 23, 59, 59, 0, time.UTC)},
		{Raw(1, Second), -2, time.Date(1969, 12, 31, 23, 59, 58, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 59, 58, 0, time.UTC)},
		{Raw(1, Second), -2, time.Date(1969, 12, 31, 23, 59, 58, 0, time.UTC), time.Date(1969, 12, 31, 23, 59, 58, 0, time.UTC)},
		{Raw(1, Second), -3, time.Date(1969, 12, 31, 23, 59, 57, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 59, 57, 0, time.UTC)},

		// 4 second
		{Raw(4, Second), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Second), 0, time.Date(1970, 1, 1, 0, 0, 3, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Second), 1, time.Date(1970, 1, 1, 0, 0, 4, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 4, 0, time.UTC)},
		{Raw(4, Second), 1, time.Date(1970, 1, 1, 0, 0, 7, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 4, 0, time.UTC)},
		{Raw(4, Second), 2, time.Date(1970, 1, 1, 0, 0, 8, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 8, 0, time.UTC)},
		{Raw(4, Second), -1, time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 59, 56, 0, time.UTC)},
		{Raw(4, Second), -1, time.Date(1969, 12, 31, 23, 59, 57, 0, time.UTC), time.Date(1969, 12, 31, 23, 59, 56, 0, time.UTC)},
		{Raw(4, Second), -1, time.Date(1969, 12, 31, 23, 59, 56, 0, time.UTC), time.Date(1969, 12, 31, 23, 59, 56, 0, time.UTC)},
		{Raw(4, Second), -2, time.Date(1969, 12, 31, 23, 59, 55, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 59, 52, 0, time.UTC)},
		{Raw(4, Second), -2, time.Date(1969, 12, 31, 23, 59, 55, 0, time.UTC), time.Date(1969, 12, 31, 23, 59, 52, 0, time.UTC)},
		{Raw(4, Second), -2, time.Date(1969, 12, 31, 23, 59, 52, 0, time.UTC), time.Date(1969, 12, 31, 23, 59, 52, 0, time.UTC)},
		{Raw(4, Second), -3, time.Date(1969, 12, 31, 23, 59, 51, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 59, 48, 0, time.UTC)},

		// 1 minute
		{Raw(1, Minute), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Minute), 0, time.Date(1970, 1, 1, 0, 0, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Minute), 1, time.Date(1970, 1, 1, 0, 1, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 1, 0, 0, time.UTC)},
		{Raw(1, Minute), 1, time.Date(1970, 1, 1, 0, 1, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 1, 0, 0, time.UTC)},
		{Raw(1, Minute), 2, time.Date(1970, 1, 1, 0, 2, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 2, 0, 0, time.UTC)},
		{Raw(1, Minute), -1, time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 59, 0, 0, time.UTC)},
		{Raw(1, Minute), -1, time.Date(1969, 12, 31, 23, 59, 0, 0, time.UTC), time.Date(1969, 12, 31, 23, 59, 0, 0, time.UTC)},
		{Raw(1, Minute), -2, time.Date(1969, 12, 31, 23, 58, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 58, 0, 0, time.UTC)},
		{Raw(1, Minute), -2, time.Date(1969, 12, 31, 23, 58, 0, 0, time.UTC), time.Date(1969, 12, 31, 23, 58, 0, 0, time.UTC)},
		{Raw(1, Minute), -3, time.Date(1969, 12, 31, 23, 57, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 57, 0, 0, time.UTC)},

		// 4 minute
		{Raw(4, Minute), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Minute), 0, time.Date(1970, 1, 1, 0, 3, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Minute), 1, time.Date(1970, 1, 1, 0, 4, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 4, 0, 0, time.UTC)},
		{Raw(4, Minute), 1, time.Date(1970, 1, 1, 0, 7, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 4, 0, 0, time.UTC)},
		{Raw(4, Minute), 2, time.Date(1970, 1, 1, 0, 8, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 8, 0, 0, time.UTC)},
		{Raw(4, Minute), -1, time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 56, 0, 0, time.UTC)},
		{Raw(4, Minute), -1, time.Date(1969, 12, 31, 23, 57, 0, 0, time.UTC), time.Date(1969, 12, 31, 23, 56, 0, 0, time.UTC)},
		{Raw(4, Minute), -1, time.Date(1969, 12, 31, 23, 56, 0, 0, time.UTC), time.Date(1969, 12, 31, 23, 56, 0, 0, time.UTC)},
		{Raw(4, Minute), -2, time.Date(1969, 12, 31, 23, 55, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 52, 0, 0, time.UTC)},
		{Raw(4, Minute), -2, time.Date(1969, 12, 31, 23, 53, 0, 0, time.UTC), time.Date(1969, 12, 31, 23, 52, 0, 0, time.UTC)},
		{Raw(4, Minute), -2, time.Date(1969, 12, 31, 23, 52, 0, 0, time.UTC), time.Date(1969, 12, 31, 23, 52, 0, 0, time.UTC)},
		{Raw(4, Minute), -3, time.Date(1969, 12, 31, 23, 51, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 48, 0, 0, time.UTC)},

		// 1 hour
		{Raw(1, Hour), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Hour), 0, time.Date(1970, 1, 1, 0, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Hour), 1, time.Date(1970, 1, 1, 1, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 1, 0, 0, 0, time.UTC)},
		{Raw(1, Hour), 1, time.Date(1970, 1, 1, 1, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 1, 0, 0, 0, time.UTC)},
		{Raw(1, Hour), 2, time.Date(1970, 1, 1, 2, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 2, 0, 0, 0, time.UTC)},
		{Raw(1, Hour), -1, time.Date(1969, 12, 31, 23, 0, 0, 0, time.UTC), time.Date(1969, 12, 31, 23, 0, 0, 0, time.UTC)},
		{Raw(1, Hour), -1, time.Date(1969, 12, 31, 23, 59, 0, 0, time.UTC), time.Date(1969, 12, 31, 23, 0, 0, 0, time.UTC)},
		{Raw(1, Hour), -2, time.Date(1969, 12, 31, 22, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 22, 0, 0, 0, time.UTC)},
		{Raw(1, Hour), -2, time.Date(1969, 12, 31, 22, 0, 0, 0, time.UTC), time.Date(1969, 12, 31, 22, 0, 0, 0, time.UTC)},
		{Raw(1, Hour), -3, time.Date(1969, 12, 31, 21, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 21, 0, 0, 0, time.UTC)},

		// 4 hour
		{Raw(4, Hour), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Hour), 0, time.Date(1970, 1, 1, 3, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Hour), 1, time.Date(1970, 1, 1, 4, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 4, 0, 0, 0, time.UTC)},
		{Raw(4, Hour), 1, time.Date(1970, 1, 1, 7, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 4, 0, 0, 0, time.UTC)},
		{Raw(4, Hour), 2, time.Date(1970, 1, 1, 8, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 8, 0, 0, 0, time.UTC)},
		{Raw(4, Hour), -1, time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 20, 0, 0, 0, time.UTC)},
		{Raw(4, Hour), -1, time.Date(1969, 12, 31, 21, 0, 0, 0, time.UTC), time.Date(1969, 12, 31, 20, 0, 0, 0, time.UTC)},
		{Raw(4, Hour), -1, time.Date(1969, 12, 31, 20, 0, 0, 0, time.UTC), time.Date(1969, 12, 31, 20, 0, 0, 0, time.UTC)},
		{Raw(4, Hour), -2, time.Date(1969, 12, 31, 19, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 16, 0, 0, 0, time.UTC)},
		{Raw(4, Hour), -2, time.Date(1969, 12, 31, 19, 0, 0, 0, time.UTC), time.Date(1969, 12, 31, 16, 0, 0, 0, time.UTC)},
		{Raw(4, Hour), -2, time.Date(1969, 12, 31, 17, 0, 0, 0, time.UTC), time.Date(1969, 12, 31, 16, 0, 0, 0, time.UTC)},
		{Raw(4, Hour), -2, time.Date(1969, 12, 31, 16, 0, 0, 0, time.UTC), time.Date(1969, 12, 31, 16, 0, 0, 0, time.UTC)},
		{Raw(4, Hour), -3, time.Date(1969, 12, 31, 15, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 12, 0, 0, 0, time.UTC)},

		// 1 day
		{Raw(1, Day), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Day), 0, time.Date(1970, 1, 1, 10, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Day), 0, time.Date(1970, 1, 1, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Day), 1, time.Date(1970, 1, 2, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 2, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Day), 1, time.Date(1970, 1, 2, 10, 0, 0, 0, time.UTC), time.Date(1970, 1, 2, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Day), 1, time.Date(1970, 1, 2, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 2, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Day), 2, time.Date(1970, 1, 3, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 3, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Day), -1, time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Day), -1, time.Date(1969, 12, 31, 0, 0, 0, 0, time.UTC), time.Date(1969, 12, 31, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Day), -2, time.Date(1969, 12, 30, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 30, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Day), -2, time.Date(1969, 12, 30, 0, 0, 0, 0, time.UTC), time.Date(1969, 12, 30, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Day), -3, time.Date(1969, 12, 29, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC)},

		// 4 day
		{Raw(4, Day), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Day), 0, time.Date(1970, 1, 3, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Day), 1, time.Date(1970, 1, 5, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 5, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Day), 1, time.Date(1970, 1, 5, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 5, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Day), 2, time.Date(1970, 1, 9, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 9, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Day), -1, time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 28, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Day), -1, time.Date(1969, 12, 30, 0, 0, 0, 0, time.UTC), time.Date(1969, 12, 28, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Day), -1, time.Date(1969, 12, 28, 0, 0, 0, 0, time.UTC), time.Date(1969, 12, 28, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Day), -2, time.Date(1969, 12, 27, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 24, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Day), -2, time.Date(1969, 12, 24, 0, 0, 0, 0, time.UTC), time.Date(1969, 12, 24, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Day), -3, time.Date(1969, 12, 23, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 20, 0, 0, 0, 0, time.UTC)},

		// 1 week - epoch week does not begin on 1970-01-01, it begins on 1969-12-29
		{Raw(1, Week), 0, time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC), time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Week), 0, time.Date(1970, 1, 4, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Week), 1, time.Date(1970, 1, 5, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 5, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Week), 1, time.Date(1970, 1, 11, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 5, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Week), 2, time.Date(1970, 1, 12, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 12, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Week), -1, time.Date(1969, 12, 28, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 22, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Week), -1, time.Date(1969, 12, 22, 0, 0, 0, 0, time.UTC), time.Date(1969, 12, 22, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Week), -2, time.Date(1969, 12, 21, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 15, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Week), -2, time.Date(1969, 12, 15, 0, 0, 0, 0, time.UTC), time.Date(1969, 12, 15, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Week), -3, time.Date(1969, 12, 14, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 8, 0, 0, 0, 0, time.UTC)},

		// 4 weeks - epoch week does not begin on 1970-01-01, it begins on 1969-12-29
		{Raw(4, Week), 0, time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC), time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Week), 0, time.Date(1970, 1, 25, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Week), 1, time.Date(1970, 1, 26, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 26, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Week), 1, time.Date(1970, 2, 22, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 26, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Week), 2, time.Date(1970, 2, 23, 0, 0, 0, 0, time.UTC), time.Date(1970, 2, 23, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Week), -1, time.Date(1969, 12, 28, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Week), -1, time.Date(1969, 12, 1, 0, 0, 0, 0, time.UTC), time.Date(1969, 12, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Week), -2, time.Date(1969, 11, 30, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 11, 3, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Week), -2, time.Date(1969, 11, 26, 0, 0, 0, 0, time.UTC), time.Date(1969, 11, 3, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Week), -2, time.Date(1969, 11, 3, 0, 0, 0, 0, time.UTC), time.Date(1969, 11, 3, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Week), -3, time.Date(1969, 11, 2, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 10, 6, 0, 0, 0, 0, time.UTC)},

		// 1 month
		{Raw(1, Month), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Month), 0, time.Date(1970, 1, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Month), 1, time.Date(1970, 2, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 2, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Month), 1, time.Date(1970, 2, 28, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 2, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Month), 2, time.Date(1970, 3, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 3, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Month), 2, time.Date(1970, 3, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 3, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Month), -1, time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Month), -1, time.Date(1969, 12, 1, 0, 0, 59, 999999999, time.UTC), time.Date(1969, 12, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Month), -2, time.Date(1969, 11, 30, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 11, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Month), -2, time.Date(1969, 11, 1, 0, 0, 0, 0, time.UTC), time.Date(1969, 11, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Month), -3, time.Date(1969, 10, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 10, 1, 0, 0, 0, 0, time.UTC)},

		// 4 months
		{Raw(4, Month), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Month), 0, time.Date(1970, 4, 30, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Month), 1, time.Date(1970, 5, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 5, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Month), 1, time.Date(1970, 8, 28, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 5, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Month), 2, time.Date(1970, 9, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 9, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Month), 2, time.Date(1970, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 9, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Month), 3, time.Date(1971, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1971, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Month), -1, time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 9, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Month), -1, time.Date(1969, 9, 1, 0, 0, 0, 0, time.UTC), time.Date(1969, 9, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Month), -2, time.Date(1969, 8, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 5, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Month), -2, time.Date(1969, 5, 1, 0, 0, 0, 0, time.UTC), time.Date(1969, 5, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Month), -3, time.Date(1969, 4, 30, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Month), -3, time.Date(1969, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1969, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Month), -4, time.Date(1968, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1968, 9, 1, 0, 0, 0, 0, time.UTC)},

		// 1 year
		{Raw(1, Year), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Year), 0, time.Date(1970, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Year), 1, time.Date(1971, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1971, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Year), 1, time.Date(1971, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1971, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Year), 2, time.Date(1972, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1972, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Year), 2, time.Date(1972, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1972, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Year), -1, time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Year), -1, time.Date(1969, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1969, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Year), -2, time.Date(1968, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1968, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Year), -2, time.Date(1968, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1968, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(1, Year), -3, time.Date(1967, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1967, 1, 1, 0, 0, 0, 0, time.UTC)},

		// 4 years
		{Raw(4, Year), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Year), 0, time.Date(1973, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Year), 1, time.Date(1974, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1974, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Year), 1, time.Date(1977, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1974, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Year), 2, time.Date(1978, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1978, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Year), 2, time.Date(1981, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1978, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Year), -1, time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1966, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Year), -1, time.Date(1966, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1966, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Year), -2, time.Date(1965, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1962, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Year), -2, time.Date(1962, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1962, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Raw(4, Year), -3, time.Date(1961, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1958, 1, 1, 0, 0, 0, 0, time.UTC)},
	} {
		t.Run(fmt.Sprintf("%d:%s/%d/%s", i, c.Interval, c.Period, c.TestTime), func(t *testing.T) {
			p := c.Interval.Period(c.TestTime)
			if c.Period != p {
				t.Fatal("period:", p)
			}
			b := c.Interval.Time(p, c.PeriodTime.Location())
			if c.PeriodTime != b {
				t.Fatal("periodtime:", b)
			}
		})
	}
}

func TestCanCombine(t *testing.T) {
	for _, tc := range []struct {
		from, to Interval
		expected bool
	}{
		{Of1Minute, Of1Minute, false},
		{Of1Minute, Of2Minutes, true},
		{Of1Minute, Raw(6, Minutes), true},
		{Of1Minute, Of60Minutes, true},
		{Of1Minute, Of1Hour, true},
		{Of1Minute, Of2Hours, true},
		{Of1Minute, Of1Day, false},
		{Of1Minute, Of1Week, false},
		{Of1Minute, Of1Month, false},
		{Of1Minute, Raw(1, Year), false},
		{Of1Minute, Raw(120, Seconds), true},

		{Of1Hour, Of2Hours, true},
		{Of1Hour, Of24Hours, true},
		{Of1Hour, Of48Hours, true},
		{Of2Hours, Of4Hours, true},
		{Of1Hour, Raw(120, Minute), true}, // Can combine downwards into smaller units if it's clean

		{Of1Hour, Of1Day, false},  // Hours don't generally combine into days due to DST
		{Of1Hour, Of2Days, false}, // Hours don't generally combine into days due to DST
		{Of4Hours, Of1Day, false},
		{Of1Hour, Of1Week, false},
		{Of12Hours, Of1Week, false},
		{Of12Hours, Raw(3, Week), false},
		{Of1Hour, Of1Month, false},
		{Of1Hour, Of1Year, false},

		{Of1Hour, Of1Minute, false},
		{Of1Hour, Of60Minutes, false},
		{Of1Hour, Raw(119, Minute), false},
		{Of1Hour, Raw(121, Minute), false},

		{Of1Day, Of1Hour, false},
		{Of1Day, Of1Day, false},
		{Of1Day, Of1Week, true},
		{Of2Days, Of1Week, false},
		{Of7Days, Raw(2, Week), false}, // No way to specify how these line up, so it makes sense that you can't convert.

		{Of1Week, Raw(1, Minute), false},
		{Of1Week, Of1Hour, false},
		{Of1Week, Of1Day, false},
		{Of1Week, Raw(14, Day), false},
		{Of1Week, Of1Week, false},
		{Of1Week, Raw(2, Week), true},
		{Of1Week, Of1Month, false},
		{Of1Week, Raw(1, Year), false},

		{Of1Month, Of1Minute, false},
		{Of1Month, Of1Day, false},
		{Of1Month, Of1Week, false},
		{Of1Month, Of1Month, false},
		{Of1Month, Raw(2, Month), true},
		{Of1Month, Raw(1, Year), true},
		{Raw(2, Month), Raw(3, Month), false},
		{Raw(2, Month), Raw(4, Month), true},
		{Raw(2, Month), Raw(1, Year), true},
	} {
		t.Run(fmt.Sprintf("%s-%s-%v", tc.from, tc.to, tc.expected), func(t *testing.T) {
			result := tc.from.CanCombineTo(tc.to)
			if result != tc.expected {
				t.Fatal(result, "!=", tc.expected)
			}
		})
	}
}

func TestStart(t *testing.T) {
	for idx, tc := range []struct {
		i Interval
		t time.Time
		x time.Time
	}{
		{OfValid(1, Minute), tm("2020-01-01T00:00:00Z"), tm("2020-01-01T00:00:00Z")},
		{OfValid(1, Minute), tm("2020-01-01T00:01:00Z"), tm("2020-01-01T00:01:00Z")},
		{OfValid(1, Minute), tm("2020-01-01T00:01:30Z"), tm("2020-01-01T00:01:00Z")},

		{OfValid(1, Hour), tm("2020-01-01T00:00:00Z"), tm("2020-01-01T00:00:00Z")},
		{OfValid(1, Hour), tm("2020-01-01T01:00:00Z"), tm("2020-01-01T01:00:00Z")},
		{OfValid(1, Hour), tm("2020-01-01T01:30:00Z"), tm("2020-01-01T01:00:00Z")},

		{OfValid(1, Day), tm("2020-01-01T00:00:00Z"), tm("2020-01-01T00:00:00Z")},
		{OfValid(1, Day), tm("2020-01-01T12:00:00Z"), tm("2020-01-01T00:00:00Z")},

		{OfValid(2, Days), tm("2020-01-01T00:00:00Z"), tm("2020-01-01T00:00:00Z")},
		{OfValid(2, Days), tm("2020-01-01T12:00:00Z"), tm("2020-01-01T00:00:00Z")},
		{OfValid(2, Days), tm("2020-01-02T00:00:00Z"), tm("2020-01-01T00:00:00Z")},

		{OfValid(1, Week), tm("2019-12-30T00:00:00Z"), tm("2019-12-30T00:00:00Z")},
		{OfValid(1, Week), tm("2020-01-01T12:00:00Z"), tm("2019-12-30T00:00:00Z")},

		{OfValid(2, Weeks), tm("2019-12-23T00:00:00Z"), tm("2019-12-23T00:00:00Z")},
		{OfValid(2, Weeks), tm("2020-01-01T12:00:00Z"), tm("2019-12-23T00:00:00Z")},

		{OfValid(1, Month), tm("2020-01-01T00:00:00Z"), tm("2020-01-01T00:00:00Z")},
		{OfValid(1, Month), tm("2020-01-03T12:00:00Z"), tm("2020-01-01T00:00:00Z")},

		{OfValid(2, Months), tm("2020-01-01T00:00:00Z"), tm("2020-01-01T00:00:00Z")},
		{OfValid(2, Months), tm("2020-01-03T12:00:00Z"), tm("2020-01-01T00:00:00Z")},
		{OfValid(2, Months), tm("2020-02-01T00:00:00Z"), tm("2020-01-01T00:00:00Z")},

		{OfValid(1, Year), tm("2020-01-01T00:00:00Z"), tm("2020-01-01T00:00:00Z")},
		{OfValid(1, Year), tm("2020-02-03T12:00:00Z"), tm("2020-01-01T00:00:00Z")},

		{OfValid(2, Year), tm("2020-01-01T00:00:00Z"), tm("2020-01-01T00:00:00Z")},
		{OfValid(2, Year), tm("2020-02-03T12:00:00Z"), tm("2020-01-01T00:00:00Z")},
		{OfValid(2, Year), tm("2021-02-03T12:00:00Z"), tm("2020-01-01T00:00:00Z")},
	} {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			start := tc.i.Start(tc.t)
			if !start.Equal(tc.x) {
				t.Fatal(start, "!=", tc.x)
			}
		})
	}
}

func TestNextEnd(t *testing.T) {
	for idx, tc := range []struct {
		i Interval
		t time.Time
		x time.Time
	}{
		{OfValid(1, Minute), tm("2020-01-01T00:00:00Z"), tm("2020-01-01T00:01:00Z")},
		{OfValid(1, Minute), tm("2020-01-01T00:00:59Z"), tm("2020-01-01T00:01:00Z")},
		{OfValid(1, Minute), tm("2020-01-01T00:01:00Z"), tm("2020-01-01T00:02:00Z")},

		{OfValid(1, Hour), tm("2020-01-01T00:00:00Z"), tm("2020-01-01T01:00:00Z")},
		{OfValid(1, Hour), tm("2020-01-01T00:59:59Z"), tm("2020-01-01T01:00:00Z")},
		{OfValid(1, Hour), tm("2020-01-01T01:00:00Z"), tm("2020-01-01T02:00:00Z")},

		{OfValid(1, Day), tm("2020-01-01T00:00:00Z"), tm("2020-01-02T00:00:00Z")},
		{OfValid(1, Day), tm("2020-01-01T23:59:59Z"), tm("2020-01-02T00:00:00Z")},
		{OfValid(1, Day), tm("2020-01-02T00:00:00Z"), tm("2020-01-03T00:00:00Z")},

		{OfValid(2, Days), tm("2020-01-01T00:00:00Z"), tm("2020-01-03T00:00:00Z")},
		{OfValid(2, Days), tm("2020-01-02T23:59:59Z"), tm("2020-01-03T00:00:00Z")},
		{OfValid(2, Days), tm("2020-01-03T00:00:00Z"), tm("2020-01-05T00:00:00Z")},

		{OfValid(1, Week), tm("2019-12-30T00:00:00Z"), tm("2020-01-06T00:00:00Z")},
		{OfValid(1, Week), tm("2020-01-05T23:59:59Z"), tm("2020-01-06T00:00:00Z")},
		{OfValid(1, Week), tm("2020-01-06T00:00:00Z"), tm("2020-01-13T00:00:00Z")},

		{OfValid(4, Weeks), tm("2019-12-23T00:00:00Z"), tm("2020-01-20T00:00:00Z")},
		{OfValid(4, Weeks), tm("2020-01-19T23:59:59Z"), tm("2020-01-20T00:00:00Z")},
		{OfValid(4, Weeks), tm("2020-01-20T00:00:00Z"), tm("2020-02-17T00:00:00Z")},

		{OfValid(1, Month), tm("2020-01-01T00:00:00Z"), tm("2020-02-01T00:00:00Z")},
		{OfValid(1, Month), tm("2020-01-31T23:59:59Z"), tm("2020-02-01T00:00:00Z")},
		{OfValid(1, Month), tm("2020-02-01T00:00:00Z"), tm("2020-03-01T00:00:00Z")},

		{OfValid(2, Months), tm("2020-01-01T00:00:00Z"), tm("2020-03-01T00:00:00Z")},
		{OfValid(2, Months), tm("2020-02-29T23:59:59Z"), tm("2020-03-01T00:00:00Z")},
		{OfValid(2, Months), tm("2020-03-01T00:00:00Z"), tm("2020-05-01T00:00:00Z")},

		{OfValid(1, Year), tm("2020-01-01T00:00:00Z"), tm("2021-01-01T00:00:00Z")},
		{OfValid(1, Year), tm("2020-12-31T23:59:59Z"), tm("2021-01-01T00:00:00Z")},
		{OfValid(1, Year), tm("2021-01-01T00:00:00Z"), tm("2022-01-01T00:00:00Z")},

		{OfValid(2, Year), tm("2020-01-01T00:00:00Z"), tm("2022-01-01T00:00:00Z")},
		{OfValid(2, Year), tm("2021-12-31T23:59:59Z"), tm("2022-01-01T00:00:00Z")},
		{OfValid(2, Year), tm("2022-01-01T00:00:00Z"), tm("2024-01-01T00:00:00Z")},
	} {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			end := tc.i.End(tc.t)
			if !end.Equal(tc.x) {
				t.Fatal(end, "!=", tc.x)
			}
			if tc.x != tc.i.Next(tc.t) {
				t.Fatal(tc.i.Next(tc.t), "!=", tc.x)
			}
		})
	}
}

var benchStart, benchEnd time.Time

func BenchmarkRangeMonth(b *testing.B) {
	iv := Raw(2, Month)
	tm := time.Date(2017, 4, 3, 2, 1, 0, 5, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchStart, benchEnd = iv.Range(tm)
	}
}

func BenchmarkRangeWeek(b *testing.B) {
	iv := Raw(2, Week)
	tm := time.Date(2017, 4, 3, 2, 1, 0, 5, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchStart, benchEnd = iv.Range(tm)
	}
}

func BenchmarkRangeDay(b *testing.B) {
	iv := Raw(2, Day)
	tm := time.Date(2017, 4, 3, 2, 1, 0, 5, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchStart, benchEnd = iv.Range(tm)
	}
}

func BenchmarkRangeHour(b *testing.B) {
	iv := Raw(2, Hour)
	tm := time.Date(2017, 4, 3, 2, 1, 0, 5, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchStart, benchEnd = iv.Range(tm)
	}
}

func BenchmarkRangeMinute(b *testing.B) {
	iv := Raw(2, Minute)
	tm := time.Date(2017, 4, 3, 2, 1, 0, 5, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchStart, benchEnd = iv.Range(tm)
	}
}

func BenchmarkRangeSecond(b *testing.B) {
	iv := Raw(2, Second)
	tm := time.Date(2017, 4, 3, 2, 1, 0, 5, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchStart, benchEnd = iv.Range(tm)
	}
}

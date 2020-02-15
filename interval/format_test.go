package interval

import (
	"fmt"
	"testing"
	"time"
)

func TestFormat(t *testing.T) {
	for _, tc := range []struct {
		intvl  Interval
		period Period
		out    string
	}{
		{Of1Second, 10, "1970-01-01T00:00:10Z"},
		{Of1Minute, 10, "1970-01-01T00:10Z"},
		{Of1Hour, 10, "1970-01-01T10:00Z"},
		{Of1Day, 10, "1970-01-11Z"},
		{Of1Week, 10, "1970-03-09Z"},
		{Of1Month, 10, "1970-11-01Z"},
	} {
		t.Run("", func(t *testing.T) {
			result := tc.intvl.Format(tc.period)
			if tc.out != result {
				t.Fatal(result)
			}
		})
	}
}

func TestFormatIn(t *testing.T) {
	loc := time.FixedZone("AEST", 36000)
	for _, tc := range []struct {
		intvl  Interval
		loc    *time.Location
		period Period
		out    string
	}{
		{Of1Second, loc, 10, "1970-01-01T10:00:10+10:00"},
		{Of1Minute, loc, 10, "1970-01-01T10:10+10:00"},
		{Of1Hour, loc, 10, "1970-01-01T20:00+10:00"},
		{Of1Hour, loc, 20, "1970-01-02T06:00+10:00"},
		{Of1Day, loc, 10, "1970-01-11+10:00"},
		{Of1Week, loc, 10, "1970-03-09+10:00"},
		{Of1Month, loc, 10, "1970-11-01+10:00"},
	} {
		t.Run("", func(t *testing.T) {
			result := tc.intvl.FormatIn(tc.period, tc.loc)
			if result != tc.out {
				t.Fatal(result)
			}
		})
	}
}

func TestFormatShort(t *testing.T) {
	for _, tc := range []struct {
		intvl  Interval
		period Period
		out    string
	}{
		{Of1Second, 0, "1970"},
		{Of1Second, 10, "00:00:10"},
		{Of1Second, 60, "00:01:00"},
		{Of1Second, 86400, "1970-01-02"},
		{Of1Second, 2678400, "1970-02"},

		{Of1Minute, 0, "1970"},
		{Of1Minute, 10, "00:10"},
		{Of1Minute, 60, "01:00"},
		{Of1Minute, 1000, "16:40"},
		{Of1Minute, 44640, "1970-02"},

		{Of1Hour, 0, "1970"},
		{Of1Hour, 24, "1970-01-02"},
		{Of1Hour, 744, "1970-02"},
		{Of1Hour, 10, "10:00"},

		{Of1Day, 0, "1970"},
		{Of1Day, 10, "1970-01-11"},
		{Of1Day, 31, "1970-02"},

		{Of1Week, 0, "1969-12-29"},
		{Of1Week, 10, "1970-03-09"},

		{Of1Month, 0, "1970"},
		{Of1Month, 10, "1970-11"},
		{Of1Month, 12, "1971"},
	} {
		t.Run(fmt.Sprintf("%s-%s", tc.intvl, tc.out), func(t *testing.T) {
			result := tc.intvl.FormatShort(tc.period)
			if result != tc.out {
				t.Fatal(result)
			}
		})
	}
}

func TestFormatAfter(t *testing.T) {
	// This minute officially has a 60th second:
	leapSecondMinute := time.Date(2005, 12, 31, 23, 59, 0, 0, time.UTC)

	// fmt.Println(Of1Second.Time(Of1Second.Period(leapSecondMinute)+60, time.UTC))

	for _, tc := range []struct {
		intvl Interval
		prev  Period
		cur   Period
		out   string
	}{
		{Of1Second, 0, 1, "00:00:01"},
		{Of1Second, 0, 86400, "02-Jan"},
		{Of1Second, 0, 86401, "02-Jan 00:00:01"},
		{Of1Second, 0, 86460, "02-Jan 00:01"},
		{Of1Second, 0, 90000, "02-Jan 01:00"},

		// Go's time package does not support leap seconds, this will roll over to 2006
		// just like any other year:
		{Of1Second, Of1Second.Period(leapSecondMinute), Of1Second.Period(leapSecondMinute) + 60, "2006-01-01T00:00"},

		{Of1Minute, 0, 1, "00:01"},
		{Of1Minute, 0, 60, "01:00"},
		{Of1Minute, 0, 1440, "02-Jan"},

		{Of1Hour, 0, 1, "01:00"},
		{Of1Hour, 0, 24, "02-Jan"},
		{Of1Hour, 0, 25, "02-Jan 01:00"},
		{Of1Hour, 24, 25, "01:00"},
		{Of1Hour, 0, 48, "03-Jan"},
		{Of1Hour, 0, 49, "03-Jan 01:00"},

		{Of1Day, 0, 10, "11-Jan"},
		{Of1Day, 0, 31, "01-Feb"},
		{Of1Day, 0, 365, "1971"},
		{Of1Day, 0, 366, "1971-01-02"},
		{Of1Day, 0, 396, "1971-02"},
	} {
		t.Run(fmt.Sprintf("%s-%s", tc.intvl, tc.out), func(t *testing.T) {
			result := tc.intvl.FormatAfter(tc.cur, tc.prev)
			if result != tc.out {
				t.Fatal(result)
			}
		})
	}
}

package interval

import (
	"fmt"
	"testing"
	"time"

	"github.com/shabbyrobe/golib/assert"
)

func TestFormat(t *testing.T) {
	for _, tc := range []struct {
		intvl  Interval
		period Period
		out    string
	}{
		{Seconds1, 10, "1970-01-01T00:00:10Z"},
		{Mins1, 10, "1970-01-01T00:10Z"},
		{Hours1, 10, "1970-01-01T10:00Z"},
		{Days1, 10, "1970-01-11Z"},
		{Weeks1, 10, "1970-03-09Z"},
		{Months1, 10, "1970-11-01Z"},
	} {
		t.Run("", func(t *testing.T) {
			tt := assert.WrapTB(t)
			tt.MustEqual(tc.out, tc.intvl.Format(tc.period))
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
		{Seconds1, loc, 10, "1970-01-01T10:00:10+10:00"},
		{Mins1, loc, 10, "1970-01-01T10:10+10:00"},
		{Hours1, loc, 10, "1970-01-01T20:00+10:00"},
		{Hours1, loc, 20, "1970-01-02T06:00+10:00"},
		{Days1, loc, 10, "1970-01-11+10:00"},
		{Weeks1, loc, 10, "1970-03-09+10:00"},
		{Months1, loc, 10, "1970-11-01+10:00"},
	} {
		t.Run("", func(t *testing.T) {
			tt := assert.WrapTB(t)
			tt.MustEqual(tc.out, tc.intvl.FormatIn(tc.period, tc.loc))
		})
	}
}

func TestFormatShort(t *testing.T) {
	for _, tc := range []struct {
		intvl  Interval
		period Period
		out    string
	}{
		{Seconds1, 0, "1970"},
		{Seconds1, 10, "00:00:10"},
		{Seconds1, 60, "00:01:00"},
		{Seconds1, 86400, "1970-01-02"},
		{Seconds1, 2678400, "1970-02"},

		{Mins1, 0, "1970"},
		{Mins1, 10, "00:10"},
		{Mins1, 60, "01:00"},
		{Mins1, 1000, "16:40"},
		{Mins1, 44640, "1970-02"},

		{Hours1, 0, "1970"},
		{Hours1, 24, "1970-01-02"},
		{Hours1, 744, "1970-02"},
		{Hours1, 10, "10:00"},

		{Days1, 0, "1970"},
		{Days1, 10, "1970-01-11"},
		{Days1, 31, "1970-02"},

		{Weeks1, 0, "1969-12-29"},
		{Weeks1, 10, "1970-03-09"},

		{Months1, 0, "1970"},
		{Months1, 10, "1970-11"},
		{Months1, 12, "1971"},
	} {
		t.Run(fmt.Sprintf("%s-%s", tc.intvl, tc.out), func(t *testing.T) {
			tt := assert.WrapTB(t)
			tt.MustEqual(tc.out, tc.intvl.FormatShort(tc.period))
		})
	}
}

func TestFormatAfter(t *testing.T) {
	// This minute officially has a 60th second:
	leapSecondMinute := time.Date(2005, 12, 31, 23, 59, 0, 0, time.UTC)

	// fmt.Println(Seconds1.Time(Seconds1.Period(leapSecondMinute)+60, time.UTC))

	for _, tc := range []struct {
		intvl Interval
		prev  Period
		cur   Period
		out   string
	}{
		{Seconds1, 0, 1, "00:00:01"},
		{Seconds1, 0, 86400, "02-Jan"},
		{Seconds1, 0, 86401, "02-Jan 00:00:01"},
		{Seconds1, 0, 86460, "02-Jan 00:01"},
		{Seconds1, 0, 90000, "02-Jan 01:00"},

		// Go's time package does not support leap seconds, this will roll over to 2006
		// just like any other year:
		{Seconds1, Seconds1.Period(leapSecondMinute), Seconds1.Period(leapSecondMinute) + 60, "2006-01-01T00:00"},

		{Mins1, 0, 1, "00:01"},
		{Mins1, 0, 60, "01:00"},
		{Mins1, 0, 1440, "02-Jan"},

		{Hours1, 0, 1, "01:00"},
		{Hours1, 0, 24, "02-Jan"},
		{Hours1, 0, 25, "02-Jan 01:00"},
		{Hours1, 24, 25, "01:00"},
		{Hours1, 0, 48, "03-Jan"},
		{Hours1, 0, 49, "03-Jan 01:00"},

		{Days1, 0, 10, "11-Jan"},
		{Days1, 0, 31, "01-Feb"},
		{Days1, 0, 365, "1971"},
		{Days1, 0, 366, "1971-01-02"},
		{Days1, 0, 396, "1971-02"},
	} {
		t.Run(fmt.Sprintf("%s-%s", tc.intvl, tc.out), func(t *testing.T) {
			tt := assert.WrapTB(t)
			tt.MustEqual(tc.out, tc.intvl.FormatAfter(tc.cur, tc.prev))
		})
	}
}

package interval

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/shabbyrobe/golib/assert"
)

func TestInterval(t *testing.T) {
	tt := assert.WrapTB(t)

	tt.MustEqual(time.Date(2017, 5, 1, 0, 0, 0, 0, time.UTC),
		New(4, Month).Start(time.Date(2017, 7, 1, 0, 0, 0, 0, time.UTC)))

	tt.MustEqual(time.Date(2017, 6, 26, 0, 0, 0, 0, time.UTC),
		New(3, Week).Start(time.Date(2017, 7, 1, 0, 0, 0, 0, time.UTC)))

	tt.MustEqual(time.Date(2017, 6, 26, 4, 0, 0, 0, time.UTC),
		New(4, Hour).Start(time.Date(2017, 6, 26, 5, 30, 0, 0, time.UTC)))

	aest, err := time.LoadLocation("Australia/Melbourne")
	tt.MustOK(err)

	tt.MustEqual(time.Date(2017, 5, 1, 0, 0, 0, 0, aest),
		New(4, Month).Start(time.Date(2017, 7, 1, 0, 0, 0, 0, aest)))

	tt.MustEqual(time.Date(2017, 6, 26, 0, 0, 0, 0, aest),
		New(3, Week).Start(time.Date(2017, 7, 1, 0, 0, 0, 0, aest)))

	// This is 02:00 instead of 04:00 because the timezone offset is 10 hours
	// and the buckets start from UTC.
	tt.MustEqual(time.Date(2017, 6, 26, 2, 0, 0, 0, aest),
		New(4, Hour).Start(time.Date(2017, 6, 26, 5, 30, 0, 0, aest)))
}

func TestParse(t *testing.T) {
	tt := assert.WrapTB(t)
	tt.MustEqual(New(1, Second), MustParse("1s"))
	tt.MustEqual(New(10, Second), MustParse("10s"))
	tt.MustEqual(New(1, Second), MustParse("1sec"))
	tt.MustEqual(New(10, Second), MustParse("10 s"))
	tt.MustEqual(New(10, Second), MustParse(" 10 s "))
	tt.MustEqual(New(10, Second), MustParse(" 10 secs "))
	tt.MustEqual(New(10, Second), MustParse(" 10  secs "))
	tt.MustEqual(New(10, Second), MustParse("10 seconds"))
	tt.MustEqual(New(10, Second), MustParse("10second"))

	tt.MustEqual(New(10, Minute), MustParse("10 min"))
	tt.MustEqual(New(10, Minute), MustParse("10 mins"))
	tt.MustEqual(New(10, Minute), MustParse("10 minute"))
	tt.MustEqual(New(10, Minute), MustParse("10 minutes"))
}

func TestString(t *testing.T) {
	tt := assert.WrapTB(t)
	tt.MustEqual("1sec", New(1, Second).String())
	tt.MustEqual("2sec", New(2, Second).String())
	tt.MustEqual("1min", New(1, Minute).String())
	tt.MustEqual("2min", New(2, Minute).String())
	tt.MustEqual("1hr", New(1, Hour).String())
	tt.MustEqual("2hr", New(2, Hour).String())
	tt.MustEqual("1wk", New(1, Week).String())
	tt.MustEqual("2wk", New(2, Week).String())
	tt.MustEqual("1d", New(1, Day).String())
	tt.MustEqual("2d", New(2, Day).String())
	tt.MustEqual("1mo", New(1, Month).String())
	tt.MustEqual("2mo", New(2, Month).String())
	tt.MustEqual("1yr", New(1, Year).String())
	tt.MustEqual("2yr", New(2, Year).String())
}

func TestSort(t *testing.T) {
	tt := assert.WrapTB(t)

	in := []Interval{New(61, Minute), New(1, Hour), New(59, Minute)}
	ex := []Interval{New(59, Minute), New(1, Hour), New(61, Minute)}
	sort.Slice(in, func(i, j int) bool { return in[i].Less(in[j]) })
	tt.MustEqual(ex, in)

	in = []Interval{New(61, Second), New(1, Minute), New(59, Second)}
	ex = []Interval{New(59, Second), New(1, Minute), New(61, Second)}
	sort.Slice(in, func(i, j int) bool { return in[i].Less(in[j]) })
	tt.MustEqual(ex, in)
}

func TestPeriodTime(t *testing.T) {
	for i, c := range []struct {
		Interval   Interval
		Period     Period
		TestTime   time.Time
		PeriodTime time.Time
	}{
		// 1 second
		{New(1, Second), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Second), 0, time.Date(1970, 1, 1, 0, 0, 0, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Second), 1, time.Date(1970, 1, 1, 0, 0, 1, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 1, 0, time.UTC)},
		{New(1, Second), 1, time.Date(1970, 1, 1, 0, 0, 1, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 1, 0, time.UTC)},
		{New(1, Second), 2, time.Date(1970, 1, 1, 0, 0, 2, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 2, 0, time.UTC)},
		{New(1, Second), -1, time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 59, 59, 0, time.UTC)},
		{New(1, Second), -1, time.Date(1969, 12, 31, 23, 59, 59, 0, time.UTC), time.Date(1969, 12, 31, 23, 59, 59, 0, time.UTC)},
		{New(1, Second), -2, time.Date(1969, 12, 31, 23, 59, 58, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 59, 58, 0, time.UTC)},
		{New(1, Second), -2, time.Date(1969, 12, 31, 23, 59, 58, 0, time.UTC), time.Date(1969, 12, 31, 23, 59, 58, 0, time.UTC)},
		{New(1, Second), -3, time.Date(1969, 12, 31, 23, 59, 57, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 59, 57, 0, time.UTC)},

		// 4 second
		{New(4, Second), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Second), 0, time.Date(1970, 1, 1, 0, 0, 3, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Second), 1, time.Date(1970, 1, 1, 0, 0, 4, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 4, 0, time.UTC)},
		{New(4, Second), 1, time.Date(1970, 1, 1, 0, 0, 7, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 4, 0, time.UTC)},
		{New(4, Second), 2, time.Date(1970, 1, 1, 0, 0, 8, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 8, 0, time.UTC)},
		{New(4, Second), -1, time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 59, 56, 0, time.UTC)},
		{New(4, Second), -1, time.Date(1969, 12, 31, 23, 59, 57, 0, time.UTC), time.Date(1969, 12, 31, 23, 59, 56, 0, time.UTC)},
		{New(4, Second), -1, time.Date(1969, 12, 31, 23, 59, 56, 0, time.UTC), time.Date(1969, 12, 31, 23, 59, 56, 0, time.UTC)},
		{New(4, Second), -2, time.Date(1969, 12, 31, 23, 59, 55, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 59, 52, 0, time.UTC)},
		{New(4, Second), -2, time.Date(1969, 12, 31, 23, 59, 55, 0, time.UTC), time.Date(1969, 12, 31, 23, 59, 52, 0, time.UTC)},
		{New(4, Second), -2, time.Date(1969, 12, 31, 23, 59, 52, 0, time.UTC), time.Date(1969, 12, 31, 23, 59, 52, 0, time.UTC)},
		{New(4, Second), -3, time.Date(1969, 12, 31, 23, 59, 51, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 59, 48, 0, time.UTC)},

		// 1 minute
		{New(1, Minute), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Minute), 0, time.Date(1970, 1, 1, 0, 0, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Minute), 1, time.Date(1970, 1, 1, 0, 1, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 1, 0, 0, time.UTC)},
		{New(1, Minute), 1, time.Date(1970, 1, 1, 0, 1, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 1, 0, 0, time.UTC)},
		{New(1, Minute), 2, time.Date(1970, 1, 1, 0, 2, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 2, 0, 0, time.UTC)},
		{New(1, Minute), -1, time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 59, 0, 0, time.UTC)},
		{New(1, Minute), -1, time.Date(1969, 12, 31, 23, 59, 0, 0, time.UTC), time.Date(1969, 12, 31, 23, 59, 0, 0, time.UTC)},
		{New(1, Minute), -2, time.Date(1969, 12, 31, 23, 58, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 58, 0, 0, time.UTC)},
		{New(1, Minute), -2, time.Date(1969, 12, 31, 23, 58, 0, 0, time.UTC), time.Date(1969, 12, 31, 23, 58, 0, 0, time.UTC)},
		{New(1, Minute), -3, time.Date(1969, 12, 31, 23, 57, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 57, 0, 0, time.UTC)},

		// 4 minute
		{New(4, Minute), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Minute), 0, time.Date(1970, 1, 1, 0, 3, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Minute), 1, time.Date(1970, 1, 1, 0, 4, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 4, 0, 0, time.UTC)},
		{New(4, Minute), 1, time.Date(1970, 1, 1, 0, 7, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 4, 0, 0, time.UTC)},
		{New(4, Minute), 2, time.Date(1970, 1, 1, 0, 8, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 8, 0, 0, time.UTC)},
		{New(4, Minute), -1, time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 56, 0, 0, time.UTC)},
		{New(4, Minute), -1, time.Date(1969, 12, 31, 23, 57, 0, 0, time.UTC), time.Date(1969, 12, 31, 23, 56, 0, 0, time.UTC)},
		{New(4, Minute), -1, time.Date(1969, 12, 31, 23, 56, 0, 0, time.UTC), time.Date(1969, 12, 31, 23, 56, 0, 0, time.UTC)},
		{New(4, Minute), -2, time.Date(1969, 12, 31, 23, 55, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 52, 0, 0, time.UTC)},
		{New(4, Minute), -2, time.Date(1969, 12, 31, 23, 53, 0, 0, time.UTC), time.Date(1969, 12, 31, 23, 52, 0, 0, time.UTC)},
		{New(4, Minute), -2, time.Date(1969, 12, 31, 23, 52, 0, 0, time.UTC), time.Date(1969, 12, 31, 23, 52, 0, 0, time.UTC)},
		{New(4, Minute), -3, time.Date(1969, 12, 31, 23, 51, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 23, 48, 0, 0, time.UTC)},

		// 1 hour
		{New(1, Hour), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Hour), 0, time.Date(1970, 1, 1, 0, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Hour), 1, time.Date(1970, 1, 1, 1, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 1, 0, 0, 0, time.UTC)},
		{New(1, Hour), 1, time.Date(1970, 1, 1, 1, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 1, 0, 0, 0, time.UTC)},
		{New(1, Hour), 2, time.Date(1970, 1, 1, 2, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 2, 0, 0, 0, time.UTC)},
		{New(1, Hour), -1, time.Date(1969, 12, 31, 23, 0, 0, 0, time.UTC), time.Date(1969, 12, 31, 23, 0, 0, 0, time.UTC)},
		{New(1, Hour), -1, time.Date(1969, 12, 31, 23, 59, 0, 0, time.UTC), time.Date(1969, 12, 31, 23, 0, 0, 0, time.UTC)},
		{New(1, Hour), -2, time.Date(1969, 12, 31, 22, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 22, 0, 0, 0, time.UTC)},
		{New(1, Hour), -2, time.Date(1969, 12, 31, 22, 0, 0, 0, time.UTC), time.Date(1969, 12, 31, 22, 0, 0, 0, time.UTC)},
		{New(1, Hour), -3, time.Date(1969, 12, 31, 21, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 21, 0, 0, 0, time.UTC)},

		// 4 hour
		{New(4, Hour), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Hour), 0, time.Date(1970, 1, 1, 3, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Hour), 1, time.Date(1970, 1, 1, 4, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 4, 0, 0, 0, time.UTC)},
		{New(4, Hour), 1, time.Date(1970, 1, 1, 7, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 4, 0, 0, 0, time.UTC)},
		{New(4, Hour), 2, time.Date(1970, 1, 1, 8, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 8, 0, 0, 0, time.UTC)},
		{New(4, Hour), -1, time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 20, 0, 0, 0, time.UTC)},
		{New(4, Hour), -1, time.Date(1969, 12, 31, 21, 0, 0, 0, time.UTC), time.Date(1969, 12, 31, 20, 0, 0, 0, time.UTC)},
		{New(4, Hour), -1, time.Date(1969, 12, 31, 20, 0, 0, 0, time.UTC), time.Date(1969, 12, 31, 20, 0, 0, 0, time.UTC)},
		{New(4, Hour), -2, time.Date(1969, 12, 31, 19, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 16, 0, 0, 0, time.UTC)},
		{New(4, Hour), -2, time.Date(1969, 12, 31, 19, 0, 0, 0, time.UTC), time.Date(1969, 12, 31, 16, 0, 0, 0, time.UTC)},
		{New(4, Hour), -2, time.Date(1969, 12, 31, 17, 0, 0, 0, time.UTC), time.Date(1969, 12, 31, 16, 0, 0, 0, time.UTC)},
		{New(4, Hour), -2, time.Date(1969, 12, 31, 16, 0, 0, 0, time.UTC), time.Date(1969, 12, 31, 16, 0, 0, 0, time.UTC)},
		{New(4, Hour), -3, time.Date(1969, 12, 31, 15, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 12, 0, 0, 0, time.UTC)},

		// 1 day
		{New(1, Day), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Day), 0, time.Date(1970, 1, 1, 10, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Day), 0, time.Date(1970, 1, 1, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Day), 1, time.Date(1970, 1, 2, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 2, 0, 0, 0, 0, time.UTC)},
		{New(1, Day), 1, time.Date(1970, 1, 2, 10, 0, 0, 0, time.UTC), time.Date(1970, 1, 2, 0, 0, 0, 0, time.UTC)},
		{New(1, Day), 1, time.Date(1970, 1, 2, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 2, 0, 0, 0, 0, time.UTC)},
		{New(1, Day), 2, time.Date(1970, 1, 3, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 3, 0, 0, 0, 0, time.UTC)},
		{New(1, Day), -1, time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 31, 0, 0, 0, 0, time.UTC)},
		{New(1, Day), -1, time.Date(1969, 12, 31, 0, 0, 0, 0, time.UTC), time.Date(1969, 12, 31, 0, 0, 0, 0, time.UTC)},
		{New(1, Day), -2, time.Date(1969, 12, 30, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 30, 0, 0, 0, 0, time.UTC)},
		{New(1, Day), -2, time.Date(1969, 12, 30, 0, 0, 0, 0, time.UTC), time.Date(1969, 12, 30, 0, 0, 0, 0, time.UTC)},
		{New(1, Day), -3, time.Date(1969, 12, 29, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC)},

		// 4 day
		{New(4, Day), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Day), 0, time.Date(1970, 1, 3, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Day), 1, time.Date(1970, 1, 5, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 5, 0, 0, 0, 0, time.UTC)},
		{New(4, Day), 1, time.Date(1970, 1, 5, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 5, 0, 0, 0, 0, time.UTC)},
		{New(4, Day), 2, time.Date(1970, 1, 9, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 9, 0, 0, 0, 0, time.UTC)},
		{New(4, Day), -1, time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 28, 0, 0, 0, 0, time.UTC)},
		{New(4, Day), -1, time.Date(1969, 12, 30, 0, 0, 0, 0, time.UTC), time.Date(1969, 12, 28, 0, 0, 0, 0, time.UTC)},
		{New(4, Day), -1, time.Date(1969, 12, 28, 0, 0, 0, 0, time.UTC), time.Date(1969, 12, 28, 0, 0, 0, 0, time.UTC)},
		{New(4, Day), -2, time.Date(1969, 12, 27, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 24, 0, 0, 0, 0, time.UTC)},
		{New(4, Day), -2, time.Date(1969, 12, 24, 0, 0, 0, 0, time.UTC), time.Date(1969, 12, 24, 0, 0, 0, 0, time.UTC)},
		{New(4, Day), -3, time.Date(1969, 12, 23, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 20, 0, 0, 0, 0, time.UTC)},

		// 1 week - epoch week does not begin on 1970-01-01, it begins on 1969-12-29
		{New(1, Week), 0, time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC), time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC)},
		{New(1, Week), 0, time.Date(1970, 1, 4, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC)},
		{New(1, Week), 1, time.Date(1970, 1, 5, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 5, 0, 0, 0, 0, time.UTC)},
		{New(1, Week), 1, time.Date(1970, 1, 11, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 5, 0, 0, 0, 0, time.UTC)},
		{New(1, Week), 2, time.Date(1970, 1, 12, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 12, 0, 0, 0, 0, time.UTC)},
		{New(1, Week), -1, time.Date(1969, 12, 28, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 22, 0, 0, 0, 0, time.UTC)},
		{New(1, Week), -1, time.Date(1969, 12, 22, 0, 0, 0, 0, time.UTC), time.Date(1969, 12, 22, 0, 0, 0, 0, time.UTC)},
		{New(1, Week), -2, time.Date(1969, 12, 21, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 15, 0, 0, 0, 0, time.UTC)},
		{New(1, Week), -2, time.Date(1969, 12, 15, 0, 0, 0, 0, time.UTC), time.Date(1969, 12, 15, 0, 0, 0, 0, time.UTC)},
		{New(1, Week), -3, time.Date(1969, 12, 14, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 8, 0, 0, 0, 0, time.UTC)},

		// 4 weeks - epoch week does not begin on 1970-01-01, it begins on 1969-12-29
		{New(4, Week), 0, time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC), time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC)},
		{New(4, Week), 0, time.Date(1970, 1, 25, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 29, 0, 0, 0, 0, time.UTC)},
		{New(4, Week), 1, time.Date(1970, 1, 26, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 26, 0, 0, 0, 0, time.UTC)},
		{New(4, Week), 1, time.Date(1970, 2, 22, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 26, 0, 0, 0, 0, time.UTC)},
		{New(4, Week), 2, time.Date(1970, 2, 23, 0, 0, 0, 0, time.UTC), time.Date(1970, 2, 23, 0, 0, 0, 0, time.UTC)},
		{New(4, Week), -1, time.Date(1969, 12, 28, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Week), -1, time.Date(1969, 12, 1, 0, 0, 0, 0, time.UTC), time.Date(1969, 12, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Week), -2, time.Date(1969, 11, 30, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 11, 3, 0, 0, 0, 0, time.UTC)},
		{New(4, Week), -2, time.Date(1969, 11, 26, 0, 0, 0, 0, time.UTC), time.Date(1969, 11, 3, 0, 0, 0, 0, time.UTC)},
		{New(4, Week), -2, time.Date(1969, 11, 3, 0, 0, 0, 0, time.UTC), time.Date(1969, 11, 3, 0, 0, 0, 0, time.UTC)},
		{New(4, Week), -3, time.Date(1969, 11, 2, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 10, 6, 0, 0, 0, 0, time.UTC)},

		// 1 month
		{New(1, Month), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Month), 0, time.Date(1970, 1, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Month), 1, time.Date(1970, 2, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 2, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Month), 1, time.Date(1970, 2, 28, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 2, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Month), 2, time.Date(1970, 3, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 3, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Month), 2, time.Date(1970, 3, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 3, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Month), -1, time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 12, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Month), -1, time.Date(1969, 12, 1, 0, 0, 59, 999999999, time.UTC), time.Date(1969, 12, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Month), -2, time.Date(1969, 11, 30, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 11, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Month), -2, time.Date(1969, 11, 1, 0, 0, 0, 0, time.UTC), time.Date(1969, 11, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Month), -3, time.Date(1969, 10, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 10, 1, 0, 0, 0, 0, time.UTC)},

		// 4 months
		{New(4, Month), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Month), 0, time.Date(1970, 4, 30, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Month), 1, time.Date(1970, 5, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 5, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Month), 1, time.Date(1970, 8, 28, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 5, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Month), 2, time.Date(1970, 9, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 9, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Month), 2, time.Date(1970, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 9, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Month), 3, time.Date(1971, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1971, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Month), -1, time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 9, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Month), -1, time.Date(1969, 9, 1, 0, 0, 0, 0, time.UTC), time.Date(1969, 9, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Month), -2, time.Date(1969, 8, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 5, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Month), -2, time.Date(1969, 5, 1, 0, 0, 0, 0, time.UTC), time.Date(1969, 5, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Month), -3, time.Date(1969, 4, 30, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Month), -3, time.Date(1969, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1969, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Month), -4, time.Date(1968, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1968, 9, 1, 0, 0, 0, 0, time.UTC)},

		// 1 year
		{New(1, Year), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Year), 0, time.Date(1970, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Year), 1, time.Date(1971, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1971, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Year), 1, time.Date(1971, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1971, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Year), 2, time.Date(1972, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1972, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Year), 2, time.Date(1972, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1972, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Year), -1, time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1969, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Year), -1, time.Date(1969, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1969, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Year), -2, time.Date(1968, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1968, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Year), -2, time.Date(1968, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1968, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(1, Year), -3, time.Date(1967, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1967, 1, 1, 0, 0, 0, 0, time.UTC)},

		// 4 years
		{New(4, Year), 0, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Year), 0, time.Date(1973, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Year), 1, time.Date(1974, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1974, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Year), 1, time.Date(1977, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1974, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Year), 2, time.Date(1978, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1978, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Year), 2, time.Date(1981, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1978, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Year), -1, time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1966, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Year), -1, time.Date(1966, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1966, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Year), -2, time.Date(1965, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1962, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Year), -2, time.Date(1962, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1962, 1, 1, 0, 0, 0, 0, time.UTC)},
		{New(4, Year), -3, time.Date(1961, 12, 31, 23, 59, 59, 999999999, time.UTC), time.Date(1958, 1, 1, 0, 0, 0, 0, time.UTC)},
	} {
		t.Run(fmt.Sprintf("%d:%s/%d/%s", i, c.Interval, c.Period, c.TestTime), func(t *testing.T) {
			tt := assert.WrapTB(t)
			p := c.Interval.Period(c.TestTime)
			tt.MustEqual(c.Period, p)

			b := c.Interval.Time(p, c.PeriodTime.Location())
			tt.MustEqual(c.PeriodTime, b)
		})
	}
}

func TestCanCombine(t *testing.T) {
	for _, tc := range []struct {
		from, to Interval
		result   bool
	}{
		{Mins1, Mins1, false},
		{Mins1, Mins2, true},
		{Mins1, Mins60, true},
		{Mins1, Hours1, true},
		{Mins1, Hours2, true},
		{Mins1, Days1, true},
		{Mins1, Weeks1, true},
		{Mins1, Months1, true},
		{Mins1, New(1, Year), true},

		{Hours1, Hours2, true},
		{Hours1, Hours24, true},
		{Hours1, Hours48, true},
		{Hours1, Days1, true},
		{Hours1, Days2, true},
		{Hours1, Weeks1, true},
		{Hours1, New(2, Week), true},
		{Hours1, Months1, true},
		{Hours1, New(1, Year), true},
		{Hours2, Hours4, true},
		{Hours4, Days1, true},
		{Hours12, Weeks1, true},
		{Hours12, New(3, Week), true},
		{Hours1, Mins1, false},
		{Hours1, Mins60, false},
		{Hours1, New(120, Minute), true},
		{Hours1, New(119, Minute), false},
		{Hours1, New(121, Minute), false},

		{Days1, Hours1, false},
		{Days1, Days1, false},
		{Days1, Weeks1, true},
		{Days2, Weeks1, false},
		{Days7, New(2, Week), false}, // No way to specify how these line up, so it makes sense that you can't convert.

		{Weeks1, New(1, Minute), false},
		{Weeks1, Hours1, false},
		{Weeks1, Days1, false},
		{Weeks1, New(14, Day), false},
		{Weeks1, Weeks1, false},
		{Weeks1, New(2, Week), true},
		{Weeks1, Months1, false},
		{Weeks1, New(1, Year), false},

		{Months1, Mins1, false},
		{Months1, Days1, false},
		{Months1, Weeks1, false},
		{Months1, Months1, false},
		{Months1, New(2, Month), true},
		{Months1, New(1, Year), true},
		{New(2, Month), New(3, Month), false},
		{New(2, Month), New(4, Month), true},
		{New(2, Month), New(1, Year), true},
	} {
		t.Run(fmt.Sprintf("%s-%s-%v", tc.from, tc.to, tc.result), func(t *testing.T) {
			tt := assert.WrapTB(t)
			tt.MustAssert(tc.from.CanCombine(tc.to) == tc.result)
		})
	}
}

var benchStart, benchEnd time.Time

func BenchmarkRangeMonth(b *testing.B) {
	iv := New(2, Month)
	tm := time.Date(2017, 4, 3, 2, 1, 0, 5, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchStart, benchEnd = iv.Range(tm)
	}
}

func BenchmarkRangeWeek(b *testing.B) {
	iv := New(2, Week)
	tm := time.Date(2017, 4, 3, 2, 1, 0, 5, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchStart, benchEnd = iv.Range(tm)
	}
}

func BenchmarkRangeDay(b *testing.B) {
	iv := New(2, Day)
	tm := time.Date(2017, 4, 3, 2, 1, 0, 5, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchStart, benchEnd = iv.Range(tm)
	}
}

func BenchmarkRangeHour(b *testing.B) {
	iv := New(2, Hour)
	tm := time.Date(2017, 4, 3, 2, 1, 0, 5, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchStart, benchEnd = iv.Range(tm)
	}
}

func BenchmarkRangeMinute(b *testing.B) {
	iv := New(2, Minute)
	tm := time.Date(2017, 4, 3, 2, 1, 0, 5, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchStart, benchEnd = iv.Range(tm)
	}
}

func BenchmarkRangeSecond(b *testing.B) {
	iv := New(2, Second)
	tm := time.Date(2017, 4, 3, 2, 1, 0, 5, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchStart, benchEnd = iv.Range(tm)
	}
}

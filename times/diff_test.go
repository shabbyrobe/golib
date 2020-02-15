package times

import (
	"testing"
	"time"
)

func TestDiff(t *testing.T) {
	mdt := func(yr int, mon time.Month, d, h, min, s, ns int) time.Time {
		return time.Date(yr, mon, d, h, min, s, ns, time.UTC)
	}
	md := func(yr int, mon time.Month, d int) time.Time {
		return time.Date(yr, mon, d, 0, 0, 0, 0, time.UTC)
	}
	_ = md

	for _, tc := range []struct {
		a, b     time.Time
		expected TimeDiff
	}{
		{mdt(2018, 1, 1, 12, 0, 0, 0), mdt(2018, 1, 1, 12, 0, 0, 1), TimeDiff{Nanoseconds: 1}},
		{mdt(2018, 1, 1, 12, 0, 0, 0), mdt(2018, 1, 1, 12, 0, 1, 1), TimeDiff{Seconds: 1, Nanoseconds: 1}},
		{mdt(2018, 2, 2, 2, 2, 2, 2), mdt(2018, 2, 2, 2, 2, 2, 1), TimeDiff{Nanoseconds: 1, Inverted: true}},

		{mdt(2019, 1, 1, 1, 1, 1, 1), mdt(2018, 2, 2, 2, 2, 2, 2), TimeDiff{Months: 10, Days: 29, Hours: 22, Minutes: 58, Seconds: 58, Nanoseconds: 999999999, Inverted: true}},
		{mdt(2018, 2, 2, 2, 2, 2, 2), mdt(2019, 1, 1, 1, 1, 1, 1), TimeDiff{Months: 10, Days: 29, Hours: 22, Minutes: 58, Seconds: 58, Nanoseconds: 999999999}},

		{mdt(2018, 2, 2, 1, 1, 1, 1), mdt(2019, 1, 1, 0, 0, 0, 0), TimeDiff{Months: 10, Days: 29, Hours: 22, Minutes: 58, Seconds: 58, Nanoseconds: 999999999}},

		{mdt(2015, 5, 1, 0, 0, 0, 0), mdt(2016, 6, 2, 1, 1, 1, 1), TimeDiff{Years: 1, Months: 1, Days: 1, Hours: 1, Minutes: 1, Seconds: 1, Nanoseconds: 1}},

		{md(2018, 1, 2), md(2018, 2, 1), TimeDiff{Days: 30}},
		{md(2018, 2, 2), md(2018, 3, 1), TimeDiff{Days: 27}},
		{md(2017, 2, 11), md(2018, 1, 12), TimeDiff{Months: 11, Days: 1}},

		{mdt(2005, 12, 31, 23, 59, 0, 0), mdt(2006, 1, 1, 0, 0, 0, 0), TimeDiff{Minutes: 1}},
	} {
		t.Run("", func(t *testing.T) {
			result := Diff(tc.a, tc.b)
			if tc.expected != result {
				t.Fatal("diff", tc.a, tc.b, "!=", result)
			}
		})
	}
}

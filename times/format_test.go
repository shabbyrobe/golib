package times

import (
	"fmt"
	"testing"
	"time"
)

func TestTimeToComparableRFC3339(t *testing.T) {
	for idx, tc := range []struct {
		in  time.Time
		out string
	}{
		{time.Date(2020, 1, 1, 12, 0, 0, 100_000_000, time.UTC),
			"2020-01-01T12:00:00.100000000Z"},
		{time.Date(2020, 1, 1, 12, 0, 0, 10_000_000, time.UTC),
			"2020-01-01T12:00:00.010000000Z"},
		{time.Date(2020, 1, 1, 12, 0, 0, 1_000_000, time.UTC),
			"2020-01-01T12:00:00.001000000Z"},
		{time.Date(2020, 1, 1, 12, 0, 0, 100_000, time.UTC),
			"2020-01-01T12:00:00.000100000Z"},
		{time.Date(2020, 1, 1, 12, 0, 0, 10_000, time.UTC),
			"2020-01-01T12:00:00.000010000Z"},
		{time.Date(2020, 1, 1, 12, 0, 0, 1_000, time.UTC),
			"2020-01-01T12:00:00.000001000Z"},
		{time.Date(2020, 1, 1, 12, 0, 0, 100, time.UTC),
			"2020-01-01T12:00:00.000000100Z"},
		{time.Date(2020, 1, 1, 12, 0, 0, 10, time.UTC),
			"2020-01-01T12:00:00.000000010Z"},
		{time.Date(2020, 1, 1, 12, 0, 0, 1, time.UTC),
			"2020-01-01T12:00:00.000000001Z"},

		{time.Date(2020, 1, 1, 12, 0, 0, 1_000_000_000, time.UTC),
			"2020-01-01T12:00:01.000000000Z"},

		{time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
			"2020-01-01T12:00:00.000000000Z"},
		{time.Date(2020, 1, 1, 12, 0, 0, 0, time.FixedZone("yep", 3600)),
			"2020-01-01T11:00:00.000000000Z"},
	} {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			result := TimeToComparableRFC3339(tc.in)
			if result != tc.out {
				t.Fatal(result, "!=", tc.out)
			}
		})
	}
}

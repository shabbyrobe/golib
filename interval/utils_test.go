package interval

import (
	"fmt"
	"testing"
	"time"
)

func TestDivideNicely(t *testing.T) {
	for _, tc := range []struct {
		in       Interval
		by       int
		limit    Interval
		expected Interval
	}{
		{Of5Minutes, 5, 0, Of1Minute},
		{Of5Minutes, 10, 0, Of30Seconds},
		{Of5Minutes, 10, Of1Minute, Of1Minute},
		{Of10Minutes, 5, 0, Of2Minutes},
		{Of2Days, 10, 0, Of4Hours},
		{Of2Days, 10, Of8Hours, Of8Hours},
		{Of2Days, 10, Of2Hours, Of4Hours},
		{Of1Month, 4, 0, Of1Week},
		{Of1Month, 5, 0, Raw(6, Days)},
		{Of1Month, 7, 0, Raw(4, Days)},
		{Raw(10, Years), 3, 0, Raw(3, Years)},
	} {
		t.Run(fmt.Sprintf("%s/%d==%s", tc.in, tc.by, tc.expected), func(t *testing.T) {
			result := DivideNicely(tc.in, tc.by, tc.limit)
			if result != tc.expected {
				t.Fatal(result)
			}
		})
	}
}

func TestDivideNicelyFor(t *testing.T) {
	for _, tc := range []struct {
		in       Interval
		by       int
		forv     Interval
		ok       bool
		expected Interval
	}{
		{Of5Minutes, 5, Of1Minute, true, Of1Minute},
		{Of3Hours, 5, Of5Minutes, true, Of30Minutes},
		{Raw(26, Minutes), 5, Of1Minute, true, Of5Minutes},
		{Raw(26, Minutes), 40, Of1Minute, false, Of1Minute},
	} {
		t.Run(fmt.Sprintf("%s/%d==%s", tc.in, tc.by, tc.expected), func(t *testing.T) {
			result, ok := DivideNicelyFor(tc.in, tc.by, tc.forv)
			if ok != tc.ok || result != tc.expected {
				t.Fatal(result)
			}
		})
	}
}

func TestFind(t *testing.T) {
	for _, tc := range []struct {
		dur      time.Duration
		expected Interval
	}{
		{1 * time.Minute, Of1Minute},
		{-1 * time.Minute, Of1Minute},

		{2 * time.Minute, Of2Minutes},
		{121 * time.Second, Of3Minutes},
		{86400 * time.Second, Of1Day},
		{86401 * time.Second, Of2Days},

		{24 * time.Hour * 3, Of3Days},
		{(24 * time.Hour * 3) + (1 * time.Minute), Raw(4, Days)},

		{0 * time.Second, Of1Second},
		{1 * time.Nanosecond, Of1Second},
		{999 * time.Millisecond, Of1Second},
		{-1 * time.Millisecond, Of1Second},
	} {
		t.Run(fmt.Sprintf("%s==%s", tc.dur, tc.expected), func(t *testing.T) {
			result := Find(tc.dur)
			if result != tc.expected {
				t.Fatal(result)
			}
		})
	}
}

package interval

import (
	"fmt"
	"testing"
	"time"

	"github.com/shabbyrobe/golib/assert"
)

func TestDivideNicely(t *testing.T) {
	for _, tc := range []struct {
		in     Interval
		by     int
		result Interval
	}{
		{Mins5, 5, Mins1},
		{Mins5, 10, Seconds30},
		{Mins10, 5, Mins2},
		{Months1, 4, Weeks1},
		{Months1, 5, Raw(6, Days)},
		{Months1, 7, Raw(4, Days)},
		{Raw(10, Years), 3, Raw(3, Years)},
	} {
		t.Run(fmt.Sprintf("%s/%d==%s", tc.in, tc.by, tc.result), func(t *testing.T) {
			tt := assert.WrapTB(t)
			tt.MustEqual(tc.result, DivideNicely(tc.in, tc.by))
		})
	}
}

func TestFind(t *testing.T) {
	for _, tc := range []struct {
		dur    time.Duration
		result Interval
	}{
		{1 * time.Minute, Mins1},
		{-1 * time.Minute, Mins1},

		{2 * time.Minute, Mins2},
		{121 * time.Second, Mins3},
		{86400 * time.Second, Days1},
		{86401 * time.Second, Days2},

		{24 * time.Hour * 3, Days3},
		{(24 * time.Hour * 3) + (1 * time.Minute), Raw(4, Days)},

		{0 * time.Second, Seconds1},
		{1 * time.Nanosecond, Seconds1},
		{999 * time.Millisecond, Seconds1},
		{-1 * time.Millisecond, Seconds1},
	} {
		t.Run(fmt.Sprintf("%s==%s", tc.dur, tc.result), func(t *testing.T) {
			tt := assert.WrapTB(t)
			tt.MustEqual(tc.result, Find(tc.dur))
		})
	}
}

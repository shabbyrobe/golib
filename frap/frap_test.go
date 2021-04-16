package frap

import (
	"fmt"
	"testing"
)

func TestRationalApprox(t *testing.T) {
	for idx, tc := range []struct {
		rat      float64
		maxDenom int
		num      int
		denom    int
	}{
		{0, 1, 0, 1},
		{0, 0, 1, 0}, // ??
		{1, 1, 1, 1},

		{0.25, 4, 1, 4},
		{0.25, 1000, 1, 4},
		{0.25, 3, 0, 1}, // This seems a bit off, we should return an 'ok' bool for this.

		{0.5, 100, 1, 2},
		{0.95, 100, 19, 20},
		{0.999, 100, 1, 1},
		{0.999, 1000, 999, 1000},
	} {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			rnum, rdenom := RationalApprox(tc.rat, tc.maxDenom)
			if rnum != tc.num || rdenom != tc.denom {
				t.Fatal(rnum, "/", rdenom, "!=", tc.num, "/", tc.denom)
			}
		})
	}
}

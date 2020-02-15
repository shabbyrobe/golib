package stringtools

import (
	"fmt"
	"math"
	"testing"
)

func TestSimilar(t *testing.T) {
	for i, c := range []struct {
		a, b string
		n    int
		pct  float64
	}{
		{"", "", 0, 1.0},
		{"foobar", "foobaz", 5, .833333},
		{"aaa", "bbb", 0, 0},
		{"awekfjawe", "awekjawe", 8, .94117647},
		{"Einsteinium", "Einsteinium", 11, 1.0},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			n, pct := Similar(c.a, c.b)
			if c.n != n {
				t.Fatal(c.n, "!=", n)
			}

			diff := math.Abs(c.pct - pct)
			if diff >= 0.001 {
				t.Fatal(c.pct, "!=", pct)
			}
		})
	}
}

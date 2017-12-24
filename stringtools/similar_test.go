package stringtools

import (
	"fmt"
	"math"
	"testing"

	"github.com/shabbyrobe/golib/assert"
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
			tt := assert.WrapTB(t)
			n, pct := Similar(c.a, c.b)
			tt.MustEqual(c.n, n)
			tt.MustAssert(math.Abs(c.pct-pct) < 0.001, fmt.Sprintf("%f != %f", c.pct, pct))
		})
	}
}

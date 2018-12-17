package num

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/shabbyrobe/golib/assert"
)

var i64 = I128From64

func i128s(s string) I128 {
	b, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic(s)
	}
	return I128FromBigInt(b)
}

func TestI128FromSize(t *testing.T) {
	tt := assert.WrapTB(t)
	tt.MustEqual(I128From8(127), i128s("127"))
	tt.MustEqual(I128From8(-128), i128s("-128"))
	tt.MustEqual(I128From16(32767), i128s("32767"))
	tt.MustEqual(I128From16(-32768), i128s("-32768"))
	tt.MustEqual(I128From32(2147483647), i128s("2147483647"))
	tt.MustEqual(I128From32(-2147483648), i128s("-2147483648"))
}

func TestI128Add(t *testing.T) {
	for idx, tc := range []struct {
		a, b, c I128
	}{
		{i64(-2), i64(-1), i64(-3)},
		{i64(-2), i64(1), i64(-1)},
		{i64(-1), i64(1), i64(0)},
		{i64(1), i64(2), i64(3)},
		{i64(10), i64(3), i64(13)},
		{MaxI128, i64(1), MinI128}, // Overflow wraps
		// {i64(maxInt64), i64(1), i128s("18446744073709551616")}, // lo carries to hi
		// {i128s("18446744073709551615"), i128s("18446744073709551615"), i128s("36893488147419103230")},
	} {
		t.Run(fmt.Sprintf("%d_%s+%s=%s", idx, tc.a, tc.b, tc.c), func(t *testing.T) {
			tt := assert.WrapTB(t)
			tt.MustAssert(tc.c.Equal(tc.a.Add(tc.b)))
		})
	}
}

func TestI128Inc(t *testing.T) {
	for _, tc := range []struct {
		a, b I128
	}{
		{i64(-1), i64(0)},
		{i64(1), i64(2)},
		{i64(10), i64(11)},
		{i64(maxInt64), i128s("9223372036854775808")},
		{i128s("18446744073709551616"), i128s("18446744073709551617")},
		{i128s("-18446744073709551617"), i128s("-18446744073709551616")},
	} {
		t.Run(fmt.Sprintf("%s+1=%s", tc.a, tc.b), func(t *testing.T) {
			tt := assert.WrapTB(t)
			inc := tc.a.Inc()
			tt.MustAssert(tc.b.Equal(inc), "%s + 1 != %s, found %s", tc.a, tc.b, inc)
		})
	}
}

func TestI128Dec(t *testing.T) {
	for _, tc := range []struct {
		a, b I128
	}{
		{i64(1), i64(0)},
		{i64(10), i64(9)},
		// {i64(maxUint64), i128s("18446744073709551614")},
		// {i64(0), MaxI128},
		// {i64(maxUint64).Add(i64(1)), i64(maxUint64)},
	} {
		t.Run(fmt.Sprintf("%s-1=%s", tc.a, tc.b), func(t *testing.T) {
			tt := assert.WrapTB(t)
			dec := tc.a.Dec()
			tt.MustAssert(tc.b.Equal(dec), "%s - 1 != %s, found %s", tc.a, tc.b, dec)
		})
	}
}

func TestI128Mul(t *testing.T) {
	for _, tc := range []struct {
		a, b, out I128
	}{
		{i64(1), i64(0), i64(0)},
		{i64(-2), i64(2), i64(-4)},
		{i64(-2), i64(-2), i64(4)},
		{i64(10), i64(9), i64(90)},
		{i64(maxInt64), i64(maxInt64), i128s("85070591730234615847396907784232501249")},
	} {
		t.Run(fmt.Sprintf("%s*%s=%s", tc.a, tc.b, tc.out), func(t *testing.T) {
			tt := assert.WrapTB(t)

			v := tc.a.Mul(tc.b)
			tt.MustAssert(tc.out.Equal(v), "%s * %s != %s, found %s", tc.a, tc.b, tc.out, v)
		})
	}
}

func TestI128Div(t *testing.T) {
	for _, tc := range []struct {
		i, by, q, r I128
	}{
		{i: i64(1), by: i64(2), q: i64(0), r: i64(1)},
		{i: i64(10), by: i64(3), q: i64(3), r: i64(1)},
		{i: i64(10), by: i64(-3), q: i64(-3), r: i64(1)},
	} {
		t.Run(fmt.Sprintf("%s÷%s=%s,%s", tc.i, tc.by, tc.q, tc.r), func(t *testing.T) {
			tt := assert.WrapTB(t)
			q, r := tc.i.QuoRem(tc.by)
			tt.MustEqual(tc.q.String(), q.String())
			tt.MustEqual(tc.r.String(), r.String())

			iBig := tc.i.AsBigInt()
			byBig := tc.by.AsBigInt()

			qBig, rBig := new(big.Int).Set(&iBig), new(big.Int).Set(&iBig)
			qBig = qBig.Div(qBig, &byBig)
			rBig = rBig.Mod(rBig, &byBig)

			tt.MustEqual(tc.q.String(), qBig.String())
			tt.MustEqual(tc.r.String(), rBig.String())
		})
	}
}

func TestI128Float64Random(t *testing.T) {
	tt := assert.WrapTB(t)

	bts := make([]byte, 16)

	// The ratio of the difference between the input number and the output
	// number relative to the input number after performing the transform
	// I128(float64(I128)) must not be more than this very reasonable limit:
	limit := new(big.Float).SetFloat64(0.00000000000001)

	for i := 0; i < 100000; i++ {
		rand.Read(bts)

		num := I128{}
		num.lo = binary.LittleEndian.Uint64(bts)

		if bts[0]%2 == 1 {
			// if we always generate hi bits, the universe will die before we
			// test a number < maxInt64
			num.hi = int64(binary.LittleEndian.Uint64(bts[8:]))
		}

		f := num.AsFloat64()
		r := I128FromFloat64(f)
		diff := DifferenceI128(num, r)

		ubig, diffBig := num.AsBigFloat(), diff.AsBigFloat()
		pct := new(big.Float).Quo(&diffBig, &ubig)
		// spew.Dump(num, f, r, pct, "---")

		tt.MustAssert(pct.Cmp(limit) < 0, "%f", pct)
	}
}

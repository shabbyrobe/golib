package num

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"testing"

	"github.com/shabbyrobe/golib/assert"
)

var i64 = I128From64

func bigI64(i int64) *big.Int { return new(big.Int).SetInt64(i) }
func bigs(s string) *big.Int {
	v, _ := new(big.Int).SetString(strings.Replace(s, " ", "", -1), 0)
	return v
}

func i128s(s string) I128 {
	b, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic(s)
	}
	i, acc := I128FromBigInt(b)
	if !acc {
		panic(fmt.Errorf("num: inaccurate i128 %s", s))
	}
	return i
}

func i128sx(s string) I128 {
	s = strings.TrimPrefix(s, "0x")
	b, ok := new(big.Int).SetString(s, 16)
	if !ok {
		panic(s)
	}
	i, acc := I128FromBigInt(b)
	if !acc {
		panic(fmt.Errorf("num: inaccurate i128 %s", s))
	}
	return i
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

func TestI128AsBigInt(t *testing.T) {
	for idx, tc := range []struct {
		a I128
		b *big.Int
	}{
		{I128{0, 2}, bigI64(2)},
		{I128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFE}, bigI64(-2)},
		{I128{0x1, 0x0}, bigs("18446744073709551616")},
		{I128{0x1, 0xFFFFFFFFFFFFFFFF}, bigs("36893488147419103231")}, // (1<<65) - 1
		{I128{0x1, 0x8AC7230489E7FFFF}, bigs("28446744073709551615")},
		{I128{0x7FFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}, bigs("170141183460469231731687303715884105727")},
		{I128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}, bigs("-1")},
		{I128{0x8000000000000000, 0}, bigs("-170141183460469231731687303715884105728")},
	} {
		t.Run(fmt.Sprintf("%d/%d,%d=%s", idx, tc.a.hi, tc.a.lo, tc.b), func(t *testing.T) {
			tt := assert.WrapTB(t)
			v := tc.a.AsBigInt()
			tt.MustAssert(tc.b.Cmp(v) == 0, "found: %s", v)
		})
	}
}

func TestI128FromBigInt(t *testing.T) {
	for idx, tc := range []struct {
		a *big.Int
		b I128
	}{
		{bigI64(0), i64(0)},
		{bigI64(2), i64(2)},
		{bigI64(-2), i64(-2)},
		{bigs("18446744073709551616"), I128{0x1, 0x0}}, // 1 << 64
		{bigs("36893488147419103231"), I128{0x1, 0xFFFFFFFFFFFFFFFF}}, // (1<<65) - 1
		{bigs("28446744073709551615"), i128s("28446744073709551615")},
		{bigs("170141183460469231731687303715884105727"), i128s("170141183460469231731687303715884105727")},
		{bigs("-1"), I128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}},
	} {
		t.Run(fmt.Sprintf("%d/%s=%d,%d", idx, tc.a, tc.b.lo, tc.b.hi), func(t *testing.T) {
			tt := assert.WrapTB(t)
			v := accI128FromBigInt(tc.a)
			tt.MustAssert(tc.b.Cmp(v) == 0, "found: (%d, %d), expected (%d, %d)", v.hi, v.lo, tc.b.hi, tc.b.lo)
		})
	}
}

func TestI128Neg(t *testing.T) {
	for idx, tc := range []struct {
		a, b I128
	}{
		{i64(0), i64(0)},
		{i64(-2), i64(2)},
		{i64(2), i64(-2)},
		{i128s("28446744073709551615"), i128s("-28446744073709551615")},
		{i128s("-28446744073709551615"), i128s("28446744073709551615")},
		// FIXME: test overflow cases
	} {
		t.Run(fmt.Sprintf("%d/-%s=%s", idx, tc.a, tc.b), func(t *testing.T) {
			tt := assert.WrapTB(t)
			tt.MustAssert(tc.b.Equal(tc.a.Neg()))
		})
	}
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
		t.Run(fmt.Sprintf("%d/%s+%s=%s", idx, tc.a, tc.b, tc.c), func(t *testing.T) {
			tt := assert.WrapTB(t)
			tt.MustAssert(tc.c.Equal(tc.a.Add(tc.b)))
		})
	}
}

func TestI128Sub(t *testing.T) {
	for idx, tc := range []struct {
		a, b, c I128
	}{
		{i64(-2), i64(-1), i64(-1)},
		{i64(-2), i64(1), i64(-3)},
		{i64(2), i64(1), i64(1)},
		{i64(2), i64(-1), i64(3)},
		{i64(1), i64(2), i64(-1)},  // crossing zero
		{i64(-1), i64(-2), i64(1)}, // crossing zero

		{MinI128, i64(1), MaxI128},  // Overflow wraps
		{MaxI128, i64(-1), MinI128}, // Overflow wraps

		{i128sx("0x10000000000000000"), i64(1), i128sx("0xFFFFFFFFFFFFFFFF")},  // carry down
		{i128sx("0xFFFFFFFFFFFFFFFF"), i64(-1), i128sx("0x10000000000000000")}, // carry up

		// {i64(maxInt64), i64(1), i128s("18446744073709551616")}, // lo carries to hi
		// {i128s("18446744073709551615"), i128s("18446744073709551615"), i128s("36893488147419103230")},
	} {
		t.Run(fmt.Sprintf("%d/%s-%s=%s", idx, tc.a, tc.b, tc.c), func(t *testing.T) {
			tt := assert.WrapTB(t)
			tt.MustAssert(tc.c.Equal(tc.a.Sub(tc.b)))
		})
	}
}

func TestI128Inc(t *testing.T) {
	for _, tc := range []struct {
		a, b I128
	}{
		{i64(-1), i64(0)},
		{i64(-2), i64(-1)},
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

			qBig, rBig := new(big.Int).Set(iBig), new(big.Int).Set(iBig)
			qBig = qBig.Div(qBig, byBig)
			rBig = rBig.Mod(rBig, byBig)

			tt.MustEqual(tc.q.String(), qBig.String())
			tt.MustEqual(tc.r.String(), rBig.String())
		})
	}
}

func TestI128Float64Random(t *testing.T) {
	tt := assert.WrapTB(t)

	bts := make([]byte, 16)

	for i := 0; i < 100000; i++ {
		rand.Read(bts)

		num := I128{}
		num.lo = binary.LittleEndian.Uint64(bts)
		num.hi = binary.LittleEndian.Uint64(bts[8:])

		f := num.AsFloat64()
		r := I128FromFloat64(f)
		diff := DifferenceI128(num, r)

		ibig, diffBig := num.AsBigFloat(), diff.AsBigFloat()
		pct := new(big.Float).Quo(diffBig, ibig)
		// spew.Dump(num, f, r, pct, "---")

		tt.MustAssert(pct.Cmp(floatDiffLimit) < 0, "%f", pct)
	}
}

func TestI128AsFloat(t *testing.T) {
	for _, tc := range []struct {
		a   I128
		out string
	}{
		{i128s("-120"), "-120"},
	} {
		t.Run(fmt.Sprintf("float64(%s)=%s", tc.a, tc.out), func(t *testing.T) {
			tt := assert.WrapTB(t)
			tt.MustEqual(tc.out, cleanFloatStr(fmt.Sprintf("%f", tc.a.AsFloat64())))
		})
	}
}

var (
	BenchIResult I128
)

func BenchmarkI128Sub(b *testing.B) {
	sub := i64(1)
	for _, iv := range []I128{i64(1), i128sx("0x10000000000000000"), MaxI128} {
		b.Run(fmt.Sprintf("%s", iv), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				BenchIResult = iv.Sub(sub)
			}
		})
	}
}

func BenchmarkI128LessThan(b *testing.B) {
	for _, iv := range []struct {
		a, b I128
	}{
		{i64(1), i64(1)},
		{i64(2), i64(1)},
		{i64(1), i64(2)},
		{i64(-1), i64(-1)},
		{i64(-1), i64(-2)},
		{i64(-2), i64(-1)},
	} {
		b.Run(fmt.Sprintf("%s<%s", iv.a, iv.b), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				BenchBoolResult = iv.a.LessThan(iv.b)
			}
		})
	}
}

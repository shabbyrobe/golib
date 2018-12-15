package num

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/shabbyrobe/golib/assert"
)

var u64 = U128From64

func u128s(s string) U128 {
	b, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic(s)
	}
	return U128FromBigInt(b)
}

func TestU128Add(t *testing.T) {
	for _, tc := range []struct {
		a, b, c U128
	}{
		{u64(1), u64(2), u64(3)},
		{u64(10), u64(3), u64(13)},
		{u64(maxUint64), u64(1), u128s("18446744073709551616")},
		{u128s("18446744073709551615"), u128s("18446744073709551615"), u128s("36893488147419103230")},
	} {
		t.Run(fmt.Sprintf("%s+%s=%s", tc.a, tc.b, tc.c), func(t *testing.T) {
			tt := assert.WrapTB(t)
			tt.MustAssert(tc.c.Equal(tc.a.Add(tc.b)))
		})
	}
}

func TestU128Mul(t *testing.T) {
	tt := assert.WrapTB(t)

	u := U128From64(maxUint64)
	v := u.Mul(U128From64(maxUint64))

	var v1, v2 big.Int
	v1.SetUint64(maxUint64)
	v2.SetUint64(maxUint64)
	tt.MustEqual(v.String(), v1.Mul(&v1, &v2).String())
}

func TestU128Div(t *testing.T) {
	for _, tc := range []struct {
		u, by, q, r U128
	}{
		{u: u64(1), by: u64(2), q: u64(0), r: u64(1)},
		{u: u64(10), by: u64(3), q: u64(3), r: u64(1)},
	} {
		t.Run(fmt.Sprintf("%s÷%s=%s,%s", tc.u, tc.by, tc.q, tc.r), func(t *testing.T) {
			tt := assert.WrapTB(t)
			q, r := tc.u.DivMod(tc.by)
			tt.MustEqual(tc.q.String(), q.String())
			tt.MustEqual(tc.r.String(), r.String())

			uBig := tc.u.AsBigInt()
			byBig := tc.by.AsBigInt()

			qBig, rBig := new(big.Int).Set(&uBig), new(big.Int).Set(&uBig)
			qBig = qBig.Div(qBig, &byBig)
			rBig = rBig.Mod(rBig, &byBig)

			tt.MustEqual(tc.q.String(), qBig.String())
			tt.MustEqual(tc.r.String(), rBig.String())
		})
	}
}

var BenchUResult U128

var BenchIntResult int

func BenchmarkU128Mul(b *testing.B) {
	u := U128From64(maxUint64)
	for i := 0; i < b.N; i++ {
		BenchUResult = u.Mul(u)
	}
}

func BenchmarkU128Add(b *testing.B) {
	u := U128From64(maxUint64)
	for i := 0; i < b.N; i++ {
		BenchUResult = u.Add(u)
	}
}

func BenchmarkU128Div(b *testing.B) {
	u := U128From64(maxUint64)
	by := U128From64(121525124)
	for i := 0; i < b.N; i++ {
		BenchUResult, _ = u.DivMod(by)
	}
}

func BenchmarkU128CmpEqual(b *testing.B) {
	u := U128From64(maxUint64)
	n := U128From64(maxUint64)
	for i := 0; i < b.N; i++ {
		BenchIntResult = u.Cmp(n)
	}
}

func BenchmarkU128Lsh(b *testing.B) {
	for _, tc := range []struct {
		in U128
		sh uint
	}{
		{u64(maxUint64), 1},
		{u64(maxUint64), 2},
		{u64(maxUint64), 8},
		{u64(maxUint64), 64},
		{u64(maxUint64), 126},
		{u64(maxUint64), 127},
		{u64(maxUint64), 128},
		{MaxU128, 1},
		{MaxU128, 2},
		{MaxU128, 8},
		{MaxU128, 64},
		{MaxU128, 126},
		{MaxU128, 127},
		{MaxU128, 128},
	} {
		b.Run(fmt.Sprintf("%s>>%d", tc.in, tc.sh), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				BenchUResult = tc.in.Lsh(tc.sh)
			}
		})
	}
}

func BenchmarkBigIntMul(b *testing.B) {
	var max big.Int
	max.SetUint64(maxUint64)

	for i := 0; i < b.N; i++ {
		var dest big.Int
		dest.Mul(&dest, &max)
	}
}

func BenchmarkBigIntAdd(b *testing.B) {
	var max big.Int
	max.SetUint64(maxUint64)

	for i := 0; i < b.N; i++ {
		var dest big.Int
		dest.Add(&dest, &max)
	}
}

func BenchmarkBigIntDiv(b *testing.B) {
	u := new(big.Int).SetUint64(maxUint64)
	by := new(big.Int).SetUint64(121525124)
	for i := 0; i < b.N; i++ {
		var z big.Int
		z.Div(u, by)
	}
}

func BenchmarkBigIntCmpEqual(b *testing.B) {
	var v1, v2 big.Int
	v1.SetUint64(maxUint64)
	v2.SetUint64(maxUint64)

	for i := 0; i < b.N; i++ {
		BenchIntResult = v1.Cmp(&v2)
	}
}

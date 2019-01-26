package fixvarnum

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strings"
	"testing"

	num "github.com/shabbyrobe/go-num"
	"github.com/shabbyrobe/golib/assert"
)

var u64 = num.U128From64

func u128s(s string) num.U128 {
	s = strings.Replace(s, " ", "", -1)
	b, ok := new(big.Int).SetString(s, 0)
	if !ok {
		panic(fmt.Errorf("num: u128 string %q invalid", s))
	}
	out, acc := num.U128FromBigInt(b)
	if !acc {
		panic(fmt.Errorf("num: inaccurate u128 %s", s))
	}
	return out
}

const FuzzIterations = 1e6

func assertU128(tt assert.T, v num.U128, scratch []byte, args ...interface{}) {
	tt.Helper()
	n := PutU128(scratch, v)

	vd, _ := U128(scratch[:n])
	tt.MustEqual(v, vd, args...)

	// vd, _ = UvarintTurbo(scratch[:n])
	// tt.MustEqual(v, vd)
}

func assertU128Sz(tt assert.T, v num.U128, sz int, scratch []byte) {
	tt.Helper()
	n := PutU128(scratch, v)

	vd, osz := U128(scratch[:n])
	tt.MustEqual(v, vd)
	tt.MustEqual(sz, n)
	tt.MustEqual(sz, osz)

	// vd, osz = UvarintTurbo(scratch[:n])
	// tt.MustEqual(v, vd)
	// tt.MustEqual(sz, n)
	// tt.MustEqual(sz, osz)
}

// func assertInt(tt assert.T, v int64, scratch []byte) {
//     tt.Helper()
//     n := PutVarint(scratch, v)
//
//     vd, _ := Varint(scratch[:n])
//     tt.MustEqual(v, vd)
//
//     vd, _ = VarintTurbo(scratch[:n])
//     tt.MustEqual(v, vd)
// }
//
// func assertIntSz(tt assert.T, v int64, sz int, scratch []byte) {
//     tt.Helper()
//     n := PutVarint(scratch, v)
//
//     vd, osz := Varint(scratch[:n])
//     tt.MustEqual(v, vd)
//     tt.MustEqual(sz, n)
//     tt.MustEqual(sz, osz)
//
//     vd, osz = VarintTurbo(scratch[:n])
//     tt.MustEqual(v, vd)
//     tt.MustEqual(sz, n)
//     tt.MustEqual(sz, osz)
// }
//

/*
func TestVarUintOverflow(t *testing.T) {
	tt := assert.WrapTB(t)

	// The number represented here is 18446744073709551615 + 1, which is
	// one past the largest representable 64-bit integer:
	in := []byte{0x80, 0xfe, 0xfe, 0xfe, 0xfe, 0xfe, 0xfe, 0xfe, 0xff, 0x00}
	_, n, err := DecodeUint(in)
	tt.MustAssert(IsOverflow(err))

	// We successfully decoded 9 bytes, but failed at the 10th:
	tt.MustEqual(9, n)
}
*/

func TestVarUintZero(t *testing.T) {
	tt := assert.WrapTB(t)
	b := make([]byte, 16)
	assertU128Sz(tt, num.U128{}, 1, b)
}

func TestVarUintSz(t *testing.T) {
	b := make([]byte, 16)

	for _, tc := range []struct {
		sz int
		in num.U128
	}{
		{1, u64(1)},
		{1, u64(7)},
		{2, u64(8)},
		{2, u64(701)},

		{1, u64(1e1)},
		{1, u64(1e2)},
		{1, u64(1e3)},
		{1, u64(1e4)},
		{1, u64(1e5)},
		{1, u64(1e6)},
		{1, u64(1e7)},
		{1, u64(1e8)},
		{1, u64(1e9)},
		{1, u64(1e10)},
		{1, u64(1e11)},
		{1, u64(1e12)},
		{1, u64(1e13)},
		{1, u64(1e14)},
		{1, u64(1e15)},
		{2, u64(1e16)}, // exceeded 4 "zero bits"
		{9, u64(11111111111111111)},
		{8, u64(11111111111111110)},
		{8, u64(11111111111111100)},
		{7, u64(11111111111111000)},
		{7, u64(11111111111110000)},
		{6, u64(11111111111100000)},
		{6, u64(11111111111000000)},
		{5, u64(11111111110000000)},
		{5, u64(11111111100000000)},
		{4, u64(11111111000000000)},
		{4, u64(11111110000000000)},
		{3, u64(11111100000000000)},
		{3, u64(11111000000000000)},
		{3, u64(11110000000000000)},
		{2, u64(11100000000000000)},
		{2, u64(11000000000000000)},
		{6, u64(1<<36 - 1)},
		{7, u64(1<<41 - 1)},
		{8, u64(1<<48 - 1)},
		{9, u64(1<<55 - 1)},
		{10, u64(math.MaxUint64)},
	} {
		t.Run(fmt.Sprintf("%d", tc.in), func(t *testing.T) {
			tt := assert.WrapTB(t)
			assertU128Sz(tt, tc.in, tc.sz, b)
		})
	}
}

func TestVarUintFuzz(t *testing.T) {
	tt := assert.WrapTB(t)
	scratch := make([]byte, MaxLen128)

	var seed int64
	// seed = time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))
	next := func() (num.U128, uint, bool) {
		var mask num.U128
		bits := uint(rng.Intn(128) + 1)
		if bits == 128 {
			mask = num.MaxU128
		} else if bits > 0 {
			mask = num.U128From64(1).Lsh(bits).Sub(num.U128From64(1))
		}
		uv := num.RandU128(rng).And(mask)
		return uv, bits, false
	}

	// var x bool
	// next = func() (num.U128, uint, bool) {
	//     if x {
	//         return num.U128{}, 0, true
	//     }
	//     x = true
	//     return num.MustU128FromString("447573512691987709388639"), 128, false
	// }

	for i := 0; i < FuzzIterations; i++ {
		uv, bits, stop := next()
		if stop {
			break
		}
		assertU128(tt, uv, scratch, "failed at index %d with bits %d", i, bits)
	}
}

/*
func TestVarIntFuzz(t *testing.T) {
	tt := assert.WrapTB(t)
	b := make([]byte, 16)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < FuzzIterations; i++ {
		bits := rand.Intn(63) + 1
		mask := uint64((1 << uint(bits)) - 1)
		uv := rng.Uint64() & mask
		iv := int64(uv)
		if rand.Intn(2) == 1 {
			iv = -iv
		}
		assertInt(tt, iv, b)
	}
}

func TestVarInt(t *testing.T) {
	tt := assert.WrapTB(t)
	b := make([]byte, 16)

	assertInt(tt, -3, b)
	assertInt(tt, int64(-1)<<32, b)
	assertInt(tt, int64(-1<<63), b)
}

func TestVarIntZero(t *testing.T) {
	tt := assert.WrapTB(t)
	b := make([]byte, 16)
	assertIntSz(tt, 0, 1, b)
}
*/

func BenchmarkU128Zero(b *testing.B) {
	for _, v := range []num.U128{
		u64(1),
		u64(1e1),
		u64(1e2),
		u64(1e3),
		u64(1e4),
		u64(1e5),
		u64(1e6),
		u64(1e7),
		u64(1e8),
		u64(1e9),
		u64(1e10),
		u64(1e11),
		u64(1e12),
		u64(1e13),
		u64(1e14),
		u64(1e15),
		u64(1e16),
		u64(1111111111111111),
		u64(1111111111111110),
		u64(1111111111111100),
		u64(1111111111111000),
		u64(1111111111110000),
		u64(1111111111100000),
		u64(1111111111000000),
		u64(1111111110000000),
		u64(1111111100000000),
		u64(1111111000000000),
		u64(1111110000000000),
		u64(1111100000000000),
		u64(1111000000000000),
		u64(1110000000000000),
		u64(1100000000000000),
	} {
		b.Run(fmt.Sprintf("%d", v), func(b *testing.B) {
			buf := make([]byte, 16)
			for i := 0; i < b.N; i++ {
				x, _ := U128(buf[:PutU128(buf, v)])
				BenchmarkDecodeUintResult = x
			}
		})
	}
}

func BenchmarkAppendUint1(b *testing.B)  { benchmarkAppendUint(b, u64(1)) }
func BenchmarkAppendUint2(b *testing.B)  { benchmarkAppendUint(b, u64(128)) }
func BenchmarkAppendUint3(b *testing.B)  { benchmarkAppendUint(b, u64(16512)) }
func BenchmarkAppendUint4(b *testing.B)  { benchmarkAppendUint(b, u64(2113664)) }
func BenchmarkAppendUint5(b *testing.B)  { benchmarkAppendUint(b, u64(270549120)) }
func BenchmarkAppendUint6(b *testing.B)  { benchmarkAppendUint(b, u64(34630287488)) }
func BenchmarkAppendUint7(b *testing.B)  { benchmarkAppendUint(b, u64(4432676798592)) }
func BenchmarkAppendUint8(b *testing.B)  { benchmarkAppendUint(b, u64(567382630219904)) }
func BenchmarkAppendUint9(b *testing.B)  { benchmarkAppendUint(b, u64(72624976668147840)) }
func BenchmarkAppendUint10(b *testing.B) { benchmarkAppendUint(b, u64(9295997013522923648)) }
func BenchmarkAppendUint11(b *testing.B) { benchmarkAppendUint(b, u128s("295147905179352825856")) }

func benchmarkAppendUint(b *testing.B, v num.U128) {
	buf := make([]byte, 16)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PutU128(buf, v)
	}
}

var BenchmarkDecodeUintResult num.U128

func benchmarkDecodeUint(b *testing.B, buf []byte) {
	for i := 0; i < b.N; i++ {
		v, _ := U128(buf)
		BenchmarkDecodeUintResult = v
	}
}

func BenchmarkDecodeUintNoZeros1(b *testing.B) { benchmarkDecodeUint(b, []byte{0x3}) }
func BenchmarkDecodeUintNoZeros2(b *testing.B) { benchmarkDecodeUint(b, []byte{0x3, 0x7f}) }

func BenchmarkDecodeUintWithZeros1(b *testing.B) { benchmarkDecodeUint(b, []byte{0x7f}) }
func BenchmarkDecodeUint2(b *testing.B)          { benchmarkDecodeUint(b, []byte{0xff, 0x7f}) }
func BenchmarkDecodeUint3(b *testing.B)          { benchmarkDecodeUint(b, []byte{0xff, 0xff, 0x7f}) }
func BenchmarkDecodeUint4(b *testing.B)          { benchmarkDecodeUint(b, []byte{0xff, 0xff, 0xff, 0x7f}) }
func BenchmarkDecodeUint5(b *testing.B)          { benchmarkDecodeUint(b, []byte{0xff, 0xff, 0xff, 0xff, 0x7f}) }
func BenchmarkDecodeUint6(b *testing.B) {
	benchmarkDecodeUint(b, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0x7f})
}
func BenchmarkDecodeUint7(b *testing.B) {
	benchmarkDecodeUint(b, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f})
}
func BenchmarkDecodeUint8(b *testing.B) {
	benchmarkDecodeUint(b, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f})
}
func BenchmarkDecodeUint9(b *testing.B) {
	benchmarkDecodeUint(b, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f})
}
func BenchmarkDecodeUint10(b *testing.B) {
	benchmarkDecodeUint(b, []byte{0x80, 0xfe, 0xfe, 0xfe, 0xfe, 0xfe, 0xfe, 0xfe, 0xfe, 0x7f})
}

/*
func benchmarkDecodeUintTurbo(b *testing.B, buf []byte) {
	for i := 0; i < b.N; i++ {
		v, _ := UvarintTurbo(buf)
		BenchmarkDecodeUintResult += v
	}
}

func BenchmarkDecodeUintTurbo1(b *testing.B) { benchmarkDecodeUintTurbo(b, []byte{0x7f}) }
func BenchmarkDecodeUintTurbo(b *testing.B)  { benchmarkDecodeUintTurbo(b, []byte{0xff, 0x7f}) }
func BenchmarkDecodeUintTurbo3(b *testing.B) { benchmarkDecodeUintTurbo(b, []byte{0xff, 0xff, 0x7f}) }
func BenchmarkDecodeUintTurbo4(b *testing.B) {
	benchmarkDecodeUintTurbo(b, []byte{0xff, 0xff, 0xff, 0x7f})
}
func BenchmarkDecodeUintTurbo5(b *testing.B) {
	benchmarkDecodeUintTurbo(b, []byte{0xff, 0xff, 0xff, 0xff, 0x7f})
}
func BenchmarkDecodeUintTurbo6(b *testing.B) {
	benchmarkDecodeUintTurbo(b, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0x7f})
}
func BenchmarkDecodeUintTurbo7(b *testing.B) {
	benchmarkDecodeUintTurbo(b, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f})
}
func BenchmarkDecodeUintTurbo8(b *testing.B) {
	benchmarkDecodeUintTurbo(b, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f})
}
func BenchmarkDecodeUintTurbo9(b *testing.B) {
	benchmarkDecodeUintTurbo(b, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f})
}
func BenchmarkDecodeUintTurbo10(b *testing.B) {
	benchmarkDecodeUintTurbo(b, []byte{0x80, 0xfe, 0xfe, 0xfe, 0xfe, 0xfe, 0xfe, 0xfe, 0xfe, 0x7f})
}

var BenchmarkDecodeIntResult int64

func benchmarkDecodeInt(b *testing.B, buf []byte) {
	for i := 0; i < b.N; i++ {
		v, _ := Varint(buf)
		BenchmarkDecodeIntResult += v
	}
}

func BenchmarkDecodeInt1(b *testing.B) { benchmarkDecodeInt(b, []byte{0x7f}) }
func BenchmarkDecodeInt2(b *testing.B) { benchmarkDecodeInt(b, []byte{0xff, 0x7f}) }
func BenchmarkDecodeInt3(b *testing.B) { benchmarkDecodeInt(b, []byte{0xff, 0xff, 0x7f}) }
func BenchmarkDecodeInt4(b *testing.B) { benchmarkDecodeInt(b, []byte{0xff, 0xff, 0xff, 0x7f}) }
func BenchmarkDecodeInt5(b *testing.B) { benchmarkDecodeInt(b, []byte{0xff, 0xff, 0xff, 0xff, 0x7f}) }
func BenchmarkDecodeInt6(b *testing.B) {
	benchmarkDecodeInt(b, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0x7f})
}
func BenchmarkDecodeInt7(b *testing.B) {
	benchmarkDecodeInt(b, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f})
}
func BenchmarkDecodeInt8(b *testing.B) {
	benchmarkDecodeInt(b, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f})
}
func BenchmarkDecodeInt9(b *testing.B) {
	benchmarkDecodeInt(b, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f})
}

var BenchmarkDecodeIntTurboResult int64

func benchmarkDecodeIntTurbo(b *testing.B, buf []byte) {
	for i := 0; i < b.N; i++ {
		v, _ := VarintTurbo(buf)
		BenchmarkDecodeIntTurboResult += v
	}
}

func BenchmarkDecodeIntTurbo1(b *testing.B) {
	benchmarkDecodeIntTurbo(b, []byte{0x7f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}
func BenchmarkDecodeIntTurbo(b *testing.B) {
	benchmarkDecodeIntTurbo(b, []byte{0xff, 0x7f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}
func BenchmarkDecodeIntTurbo3(b *testing.B) {
	benchmarkDecodeIntTurbo(b, []byte{0xff, 0xff, 0x7f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}
func BenchmarkDecodeIntTurbo4(b *testing.B) {
	benchmarkDecodeIntTurbo(b, []byte{0xff, 0xff, 0xff, 0x7f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}
func BenchmarkDecodeIntTurbo5(b *testing.B) {
	benchmarkDecodeIntTurbo(b, []byte{0xff, 0xff, 0xff, 0xff, 0x7f, 0x00, 0x00, 0x00, 0x00, 0x00})
}
func BenchmarkDecodeIntTurbo6(b *testing.B) {
	benchmarkDecodeIntTurbo(b, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0x7f, 0x00, 0x00, 0x00, 0x00})
}
func BenchmarkDecodeIntTurbo7(b *testing.B) {
	benchmarkDecodeIntTurbo(b, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f, 0x00, 0x00, 0x00})
}
func BenchmarkDecodeIntTurbo8(b *testing.B) {
	benchmarkDecodeIntTurbo(b, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f, 0x00, 0x00})
}
func BenchmarkDecodeIntTurbo9(b *testing.B) {
	benchmarkDecodeIntTurbo(b, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f, 0x00})
}
*/
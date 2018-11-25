package fixvarint

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/shabbyrobe/golib/assert"
)

func assertUint(tt assert.T, v uint64, scratch []byte) {
	tt.Helper()
	n := PutUvarint(scratch, v)
	vd, _ := Uvarint(scratch[:n])
	tt.MustEqual(v, vd)
}

func assertUintSz(tt assert.T, v uint64, sz int, scratch []byte) {
	tt.Helper()
	n := PutUvarint(scratch, v)
	vd, osz := Uvarint(scratch[:n])
	tt.MustEqual(v, vd)
	tt.MustEqual(sz, n)
	tt.MustEqual(sz, osz)
}

func assertInt(tt assert.T, v int64, scratch []byte) {
	tt.Helper()
	n := PutVarint(scratch, v)
	vd, _ := Varint(scratch[:n])
	tt.MustEqual(v, vd)
}

func assertIntSz(tt assert.T, v int64, sz int, scratch []byte) {
	tt.Helper()
	n := PutVarint(scratch, v)
	vd, osz := Varint(scratch[:n])
	tt.MustEqual(v, vd)
	tt.MustEqual(sz, n)
	tt.MustEqual(sz, osz)
}

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
	assertUintSz(tt, 0, 1, b)
}

func TestVarUintSz(t *testing.T) {
	b := make([]byte, 16)

	for _, tc := range []struct {
		sz int
		in uint64
	}{
		{1, 1},
		{1, 7},
		{2, 8},
		{2, 701},

		{1, 1e1},
		{1, 1e2},
		{1, 1e3},
		{1, 1e4},
		{1, 1e5},
		{1, 1e6},
		{1, 1e7},
		{1, 1e8},
		{1, 1e9},
		{1, 1e10},
		{1, 1e11},
		{1, 1e12},
		{1, 1e13},
		{1, 1e14},
		{1, 1e15},
		{2, 1e16}, // exceeded 4 "zero bits"
		{9, 11111111111111111},
		{8, 11111111111111110},
		{8, 11111111111111100},
		{7, 11111111111111000},
		{7, 11111111111110000},
		{6, 11111111111100000},
		{6, 11111111111000000},
		{5, 11111111110000000},
		{5, 11111111100000000},
		{4, 11111111000000000},
		{4, 11111110000000000},
		{3, 11111100000000000},
		{3, 11111000000000000},
		{3, 11110000000000000},
		{2, 11100000000000000},
		{2, 11000000000000000},
	} {
		t.Run(fmt.Sprintf("%d", tc.in), func(t *testing.T) {
			tt := assert.WrapTB(t)
			assertUintSz(tt, tc.in, tc.sz, b)
		})
	}
}

func TestVarUintFuzz(t *testing.T) {
	tt := assert.WrapTB(t)
	b := make([]byte, 16)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 100000; i++ {
		var mask uint64
		bits := rand.Intn(64) + 1
		if bits == 64 {
			mask = ^uint64(0)
		} else {
			mask = (1 << uint(bits)) - 1
		}
		uv := rng.Uint64() & mask
		assertUint(tt, uv, b)
	}
}

func TestVarIntFuzz(t *testing.T) {
	tt := assert.WrapTB(t)
	b := make([]byte, 16)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 100000; i++ {
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

func TestVarUint(t *testing.T) {
	tt := assert.WrapTB(t)
	b := make([]byte, 16)

	for _, v := range []uint64{
		3,
		7,
		8,
		63,
		64,
		65535,
		65536,
		500000,

		1,
		1e1,
		1e2,
		1e3,
		1e4,
		1e5,
		1e6,
		1e7,
		1e8,
		1e9,
		1e10,
		1e11,
		1e12,
		1e13,
		1e14,
		1e15,
		1e16,
		1111111111111111,
		1111111111111110,
		1111111111111100,
		1111111111111000,
		1111111111110000,
		1111111111100000,
		1111111111000000,
		1111111110000000,
		1111111100000000,
		1111111000000000,
		1111110000000000,
		1111100000000000,
		1111000000000000,
		1110000000000000,
		1100000000000000,
	} {
		t.Run("", func(t *testing.T) {
			assertUint(tt, v, b)
		})
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

func BenchmarkZero(b *testing.B) {
	for _, v := range []uint64{
		1,
		1e1,
		1e2,
		1e3,
		1e4,
		1e5,
		1e6,
		1e7,
		1e8,
		1e9,
		1e10,
		1e11,
		1e12,
		1e13,
		1e14,
		1e15,
		1e16,
		1111111111111111,
		1111111111111110,
		1111111111111100,
		1111111111111000,
		1111111111110000,
		1111111111100000,
		1111111111000000,
		1111111110000000,
		1111111100000000,
		1111111000000000,
		1111110000000000,
		1111100000000000,
		1111000000000000,
		1110000000000000,
		1100000000000000,
	} {
		b.Run(fmt.Sprintf("%d", v), func(b *testing.B) {
			buf := make([]byte, 16)
			for i := 0; i < b.N; i++ {
				x, _ := Uvarint(buf[:PutUvarint(buf, v)])
				BenchmarkDecodeUintResult += x
			}
		})
	}
}

func BenchmarkAppendUint1(b *testing.B)  { benchmarkAppendUint(b, 1) }
func BenchmarkAppendUint2(b *testing.B)  { benchmarkAppendUint(b, 128) }
func BenchmarkAppendUint3(b *testing.B)  { benchmarkAppendUint(b, 16512) }
func BenchmarkAppendUint4(b *testing.B)  { benchmarkAppendUint(b, 2113664) }
func BenchmarkAppendUint5(b *testing.B)  { benchmarkAppendUint(b, 270549120) }
func BenchmarkAppendUint6(b *testing.B)  { benchmarkAppendUint(b, 34630287488) }
func BenchmarkAppendUint7(b *testing.B)  { benchmarkAppendUint(b, 4432676798592) }
func BenchmarkAppendUint8(b *testing.B)  { benchmarkAppendUint(b, 567382630219904) }
func BenchmarkAppendUint9(b *testing.B)  { benchmarkAppendUint(b, 72624976668147840) }
func BenchmarkAppendUint10(b *testing.B) { benchmarkAppendUint(b, 9295997013522923648) }

func benchmarkAppendUint(b *testing.B, v uint64) {
	buf := make([]byte, 16)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PutUvarint(buf, v)
	}
}

var BenchmarkDecodeUintResult uint64

func benchmarkDecodeUint(b *testing.B, buf []byte) {
	for i := 0; i < b.N; i++ {
		v, _ := Uvarint(buf)
		BenchmarkDecodeUintResult += v
	}
}

func BenchmarkDecodeUint1(b *testing.B) { benchmarkDecodeUint(b, []byte{0x7f}) }
func BenchmarkDecodeUint2(b *testing.B) { benchmarkDecodeUint(b, []byte{0xff, 0x7f}) }
func BenchmarkDecodeUint3(b *testing.B) { benchmarkDecodeUint(b, []byte{0xff, 0xff, 0x7f}) }
func BenchmarkDecodeUint4(b *testing.B) { benchmarkDecodeUint(b, []byte{0xff, 0xff, 0xff, 0x7f}) }
func BenchmarkDecodeUint5(b *testing.B) { benchmarkDecodeUint(b, []byte{0xff, 0xff, 0xff, 0xff, 0x7f}) }
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

package fixvarint

import (
	"fmt"
	"math"
	"testing"

	"github.com/shabbyrobe/golib/assert"
)

func assertUintSz(tt assert.T, v uint64, sz int, scratch []byte, args ...interface{}) {
	tt.Helper()
	assertUintSzInternal(tt, v, sz, scratch, args)
}

func assertUintSzInternal(tt assert.T, v uint64, sz int, scratch []byte, args ...interface{}) {
	n := PutUvarint(scratch, v)

	// NOTE: there's a lot of duplicated stuff here and in assertDecodeIntSz, but the fuzzer
	// is significantly slower if we try to reuse.

	vd, osz := Uvarint(scratch[:n])
	if v != vd {
		fatalfArgs(tt, fmt.Sprintf("decoded value %d did not match input %d", vd, v), args...)
	}

	if sz != osz {
		fatalfArgs(tt, fmt.Sprintf("decoded size %d did not match expected size %d", osz, sz), args...)
	}
	if sz != n {
		fatalfArgs(tt, fmt.Sprintf("encoded size %d did not match expected size %d", n, sz), args...)
	}

	vd, osz = UvarintTurbo(scratch[:n])
	if v != vd {
		fatalfArgs(tt, fmt.Sprintf("turbo decoded value %d did not match input %d", vd, v), args...)
	}
	if sz != osz {
		fatalfArgs(tt, fmt.Sprintf("turbo decoded size %d did not match expected size %d", osz, sz), args...)
	}
}

func assertIntSz(tt assert.T, v int64, sz int, scratch []byte, args ...interface{}) {
	tt.Helper()
	assertIntSzInternal(tt, v, sz, scratch, args...)
}

func assertIntSzInternal(tt assert.T, v int64, sz int, scratch []byte, args ...interface{}) {
	n := PutVarint(scratch, v)

	// NOTE: there's a lot of duplicated stuff here and in assertDecodeIntSz, but the fuzzer
	// is significantly slower if we try to reuse.

	vd, osz := Varint(scratch[:n])
	if v != vd {
		fatalfArgs(tt, fmt.Sprintf("decoded value %d did not match input %d", vd, v), args...)
	}

	if sz != osz {
		fatalfArgs(tt, fmt.Sprintf("decoded size %d did not match expected size %d", osz, sz), args...)
	}
	if sz != n {
		fatalfArgs(tt, fmt.Sprintf("encoded size %d did not match expected size %d", n, sz), args...)
	}

	vd, osz = VarintTurbo(scratch[:n])
	if v != vd {
		fatalfArgs(tt, fmt.Sprintf("turbo decoded value %d did not match input %d", vd, v), args...)
	}
	if sz != osz {
		fatalfArgs(tt, fmt.Sprintf("turbo decoded size %d did not match expected size %d", osz, sz), args...)
	}
}

func assertDecodeIntSz(tt assert.T, data []byte, v int64, sz int, args ...interface{}) {
	tt.Helper()

	vd, osz := Varint(data)
	if v != vd {
		fatalfArgs(tt, fmt.Sprintf("decoded value %d did not match input %d", vd, v), args...)
	}
	if sz != osz {
		fatalfArgs(tt, fmt.Sprintf("decoded size %d did not match expected size %d", osz, sz), args...)
	}

	vd, osz = VarintTurbo(data)
	if v != vd {
		fatalfArgs(tt, fmt.Sprintf("turbo decoded value %d did not match input %d", vd, v), args...)
	}
	if sz != osz {
		fatalfArgs(tt, fmt.Sprintf("turbo decoded size %d did not match expected size %d", osz, sz), args...)
	}
}

func assertDecodeUintSz(tt assert.T, data []byte, v uint64, sz int, args ...interface{}) {
	tt.Helper()

	vd, osz := Uvarint(data)
	if v != vd {
		fatalfArgs(tt, fmt.Sprintf("decoded value %d did not match input %d", vd, v), args...)
	}
	if sz != osz {
		fatalfArgs(tt, fmt.Sprintf("decoded size %d did not match expected size %d", osz, sz), args...)
	}

	vd, osz = UvarintTurbo(data)
	if v != vd {
		fatalfArgs(tt, fmt.Sprintf("turbo decoded value %d did not match input %d", vd, v), args...)
	}
	if sz != osz {
		fatalfArgs(tt, fmt.Sprintf("turbo decoded size %d did not match expected size %d", osz, sz), args...)
	}
}

func TestVarUintOverflow(t *testing.T) {
	tt := assert.WrapTB(t)

	scr := make([]byte, 11)

	{
		// Sanity check MaxUint64 equals what we expect:
		n := PutUvarint(scr, math.MaxUint64)
		tt.MustEqual(10, n)
		tt.MustEqual([]byte{0x87, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x1f}, scr[:n])
	}

	{
		// MaxUint64 + 1:
		in := []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x20}
		assertDecodeUintSz(tt, in, 0, -10)
	}

	{
		// MaxUint128:
		in := []byte{0x87, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x3f}
		u, n := Uvarint(in)
		tt.MustEqual(uint64(0), u)

		// Should read until the end of the number, but then report overflow
		tt.MustEqual(-19, n)
	}

	{
		// Overflow with no terminating byte:
		in := []byte{0x87, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
		u, n := Uvarint(in)
		tt.MustEqual(uint64(0), u)

		// Should read until the end of the number, but then report overflow
		// (slightly different for turbo version)
		tt.MustEqual(-18, n)
	}
}

func TestVarUintZero(t *testing.T) {
	tt := assert.WrapTB(t)
	b := make([]byte, MaxLen64)
	assertUintSz(tt, 0, 1, b)
}

func TestVarUintSz(t *testing.T) {
	b := make([]byte, MaxLen64)

	for idx, tc := range []struct {
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
		{6, 1<<36 - 1},
		{7, 1<<41 - 1},
		{8, 1<<48 - 1},
		{9, 1<<55 - 1},
		{10, math.MaxUint64},
	} {
		t.Run(fmt.Sprintf("%d/%d", idx, tc.in), func(t *testing.T) {
			tt := assert.WrapTB(t)
			assertUintSz(tt, tc.in, tc.sz, b)
		})
	}
}

func TestVarintOverflow(t *testing.T) {
	scratch := make([]byte, 11)

	// Sanity check MaxInt64 equals what we expect:
	t.Run("", func(t *testing.T) {
		tt := assert.WrapTB(t)
		n := PutVarint(scratch, math.MaxInt64)
		tt.MustEqual(10, n)
		tt.MustEqual([]byte{0x86, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x1f}, scratch[:n])
	})

	// MaxInt64 + 1:
	t.Run("", func(t *testing.T) {
		tt := assert.WrapTB(t)
		in := []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x20}
		assertDecodeIntSz(tt, in, 0, -10)
	})

	// MaxInt128:
	t.Run("", func(t *testing.T) {
		tt := assert.WrapTB(t)
		in := []byte{0x87, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x3f}
		u, n := Varint(in)
		tt.MustEqual(int64(0), u)

		// Should read until the end of the number, but then report overflow
		// (slightly different for turbo version)
		tt.MustEqual(-19, n)
	})

	// Overflow with no terminating byte:
	t.Run("", func(t *testing.T) {
		tt := assert.WrapTB(t)
		in := []byte{0x87, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
		u, n := Varint(in)
		tt.MustEqual(int64(0), u)

		// Should read until the end of the number, but then report overflow
		// (slightly different for turbo version)
		tt.MustEqual(-18, n)
	})
}

func TestUvarintOverflowZero(t *testing.T) {
	scratch := make([]byte, 11)

	// If we add a zero to the zero bits, do we overflow?
	for idx, tc := range []struct {
		v    uint64
		add  byte
		over bool
	}{
		{1, 15, false},

		{1<<50 - 1, 4, false}, // Fast path, number fits inside 64 bits after multiplication.
		{1<<50 - 1, 5, true},  // Slow path, number does not fit inside 64 bits after multiplication.
		{1 << 50, 4, false},   // Slow path
		{1 << 50, 5, true},    // Slow path
	} {
		t.Run(fmt.Sprintf("extrazero/%d", idx), func(t *testing.T) {
			if tc.add <= 0 {
				panic("not enough zeros!")
			}

			tt := assert.WrapTB(t)

			sz := PutUvarint(scratch, tc.v)
			inZeros := (scratch[0] & 0x78) >> 3
			if int(inZeros)+int(tc.add) > 15 {
				panic("too many zeros!")
			}

			scratch[0] = (scratch[0] & 0x80) | // Continuation bit
				((inZeros + tc.add) << 3) | // Zeros
				(scratch[0] & 0x7) // Remaining data

			{
				v, n := Uvarint(scratch[:sz])
				if tc.over {
					tt.MustAssert(n < 0)
				} else {
					tt.MustAssert(n > 0)
					tt.MustEqual(tc.v*zumul[tc.add], v)
				}
			}

			{
				v, n := UvarintTurbo(scratch[:sz])
				if tc.over {
					tt.MustAssert(n < 0)
				} else {
					tt.MustAssert(n > 0)
					tt.MustEqual(tc.v*zumul[tc.add], v)
				}
			}
		})
	}
}

func TestVarintSpillToNextByte(t *testing.T) {
	// Encoded values that should spill to the next size, indexed by length. Note that
	// length MaxLen64 is skipped here as the largest number that 10 bytes can hold
	// overflows a 64-bit integer:
	templates := make([][]byte, MaxLen64)

	// Build the templates
	for i := range templates {
		if i == 0 {
			continue
		} else if i == 1 {
			templates[i] = []byte{0x7}
		} else {
			template := make([]byte, i)
			template[0] = 0x87
			for j := 1; j < i-1; j++ {
				template[j] = 0xff
			}
			template[i-1] = 0x7f
			templates[i] = template
		}
	}

	scratch := make([]byte, MaxLen64)

	for sz, data := range templates {
		if sz == 0 {
			continue
		}

		// FIXME: we are not testing all zeros because we aren't properly handling the
		// overflow that can happen when the encoded trailing decimal zeros cause the
		// final multiplication stage to overflow
		for zeros := byte(0); zeros <= 1; zeros++ {

			// Take the first byte of the template and set the number of trailing decimal
			// zeros, located in bits 2 to 5. For example 0b0ZZZZ000, where Z is a
			// 'trailing decimal zero bit'. Only the Z bits are touched:
			data[0] = data[0]&0x87 | (zeros << 3)

			t.Run(fmt.Sprintf("uv/sz=%d/z=%d", sz, zeros), func(t *testing.T) {
				tt := assert.WrapTB(t)

				v, osz := Uvarint(data)
				if sz != osz {
					fatalfArgs(tt, fmt.Sprintf("decoded size %d did not match expected size %d", osz, sz))
				}

				if v >= 0 {
					v++
				}
				nsz := PutUvarint(scratch, v)
				if osz >= nsz {
					fatalfArgs(tt, fmt.Sprintf("next number after %d expected size to be greater than %d", v, osz))
				}
			})

			t.Run(fmt.Sprintf("iv/sz=%d/z=%d", sz, zeros), func(t *testing.T) {
				tt := assert.WrapTB(t)

				v, osz := Varint(data)
				if sz != osz {
					fatalfArgs(tt, fmt.Sprintf("decoded size %d did not match expected size %d", osz, sz))
				}

				if v >= 0 {
					v++
				} else {
					v--
				}
				nsz := PutVarint(scratch, v)
				if osz >= nsz {
					fatalfArgs(tt, fmt.Sprintf("next number after %d expected size to be greater than %d", v, osz))
				}
			})
		}
	}
}

func TestVarUintFuzz(t *testing.T) {
	tt := assert.WrapTB(t)
	b := make([]byte, MaxLen64)

	rng := globalRNG

	for i := 0; i < fuzzIterations; i++ {
		var mask uint64
		bits := uint(rng.Intn(64) + 1)
		if bits == 64 {
			mask = ^uint64(0)
		} else {
			mask = (1 << bits) - 1
		}
		uv := rng.Uint64() & mask
		uv |= 1 << (bits - 1) // Ensure that the number is definitely the expected number of bits

		sz := expectedBytesFromUint64(uv)
		assertUintSzInternal(tt, uv, sz, b, "failed at index %d with bits %d, number %d", i, bits, uv)
	}
}

func TestVarIntFuzz(t *testing.T) {
	tt := assert.WrapTB(t)
	b := make([]byte, MaxLen64)

	rng := globalRNG

	for i := 0; i < fuzzIterations; i++ {
		// bit 0 == signed or unsigned. bits 1-6 == bit size of number used in this fuzz interation
		randStuff := rng.Intn(1 << 7)

		signed := randStuff>>6 == 1
		bits := (randStuff & 0x3f) + 1

		mask := uint64((1 << uint(bits)) - 1)
		uv := rng.Uint64() & mask
		iv := int64(uv)
		if signed {
			iv = -iv
		}
		sz := expectedBytesFromInt64(iv)
		assertIntSzInternal(tt, iv, sz, b, "failed at index %d with bits %d, number %d", i, bits, iv)
	}
}

func TestVarUint(t *testing.T) {
	b := make([]byte, MaxLen64)

	for _, tc := range []struct {
		sz int
		in uint64
	}{
		{1, 3},
		{1, 7},
		{2, 8},
		{2, 63},
		{2, 64},
		{3, 65535},
		{3, 65536},
		{1, 500000},
		{4, 500001},

		{1, 1},
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
		{2, 1e16},
		{8, 1111111111111111},
		{8, 1111111111111110},
		{7, 1111111111111100},
		{7, 1111111111111000},
		{6, 1111111111110000},
		{6, 1111111111100000},
		{5, 1111111111000000},
		{5, 1111111110000000},
		{4, 1111111100000000},
		{4, 1111111000000000},
		{3, 1111110000000000},
		{3, 1111100000000000},
		{3, 1111000000000000},
		{2, 1110000000000000},
		{2, 1100000000000000},
	} {
		t.Run(fmt.Sprintf("%d", tc.in), func(t *testing.T) {
			tt := assert.WrapTB(t)
			assertUintSz(tt, tc.in, tc.sz, b)
		})
	}
}

func TestVarIntSz(t *testing.T) {
	b := make([]byte, MaxLen64)

	for idx, tc := range []struct {
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
		{6, 1<<36 - 1},
		{7, 1<<41 - 1},
		{8, 1<<48 - 1},
		{9, 1<<55 - 1},
		{10, math.MaxUint64},
	} {
		t.Run(fmt.Sprintf("%d/%d", idx, tc.in), func(t *testing.T) {
			tt := assert.WrapTB(t)
			assertUintSz(tt, tc.in, tc.sz, b)
		})
	}
}

func TestVarInt(t *testing.T) {
	tt := assert.WrapTB(t)
	b := make([]byte, MaxLen64)

	assertIntSz(tt, -3, 1, b)
	assertIntSz(tt, int64(-1)<<32, 6, b)
	assertIntSz(tt, int64(-1<<63), 10, b)
}

func TestVarIntZero(t *testing.T) {
	tt := assert.WrapTB(t)
	b := make([]byte, MaxLen64)
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
	benchmarkDecodeUintTurbo(b, []byte{0x80, 0xfe, 0xfe, 0xfe, 0xfe, 0xfe, 0xfe, 0xfe, 0xfe, 0x1f})
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

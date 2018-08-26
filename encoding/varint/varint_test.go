package varint

import (
	"math/rand"
	"testing"
	"time"

	"github.com/shabbyrobe/golib/assert"
)

func assertUint(tt assert.T, v uint64, scratch []byte) {
	tt.Helper()
	_, vb := AppendUint(v, scratch)
	vd, _, err := DecodeUint(vb)
	tt.MustOK(err)
	tt.MustEqual(v, vd)
}

func assertInt(tt assert.T, v int64, scratch []byte) {
	tt.Helper()
	_, vb := AppendInt(v, scratch)
	vd, _, err := DecodeInt(vb)
	tt.MustOK(err)
	tt.MustEqual(v, vd)
}

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

func TestVarUintFuzz(t *testing.T) {
	tt := assert.WrapTB(t)
	b := make([]byte, 0, 16)

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
	b := make([]byte, 0, 16)

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

func TestVarUintBoundaries(t *testing.T) {
	for _, tc := range []struct {
		v uint64
		b []byte
	}{
		{v: 127, b: []byte{0x7f}},
		{v: 128, b: []byte{0x80, 0x00}},
		{v: 16511, b: []byte{0xff, 0x7f}},
		{v: 16512, b: []byte{0x80, 0x80, 0x00}},
		{v: 2113663, b: []byte{0xff, 0xff, 0x7f}},
		{v: 2113664, b: []byte{0x80, 0x80, 0x80, 0x00}},
		{v: 270549119, b: []byte{0xff, 0xff, 0xff, 0x7f}},
		{v: 270549120, b: []byte{0x80, 0x80, 0x80, 0x80, 0x00}},
		{v: 34630287487, b: []byte{0xff, 0xff, 0xff, 0xff, 0x7f}},
		{v: 34630287488, b: []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x00}},
		{v: 4432676798591, b: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
		{v: 4432676798592, b: []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x00}},
		{v: 567382630219903, b: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
		{v: 567382630219904, b: []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x00}},
		{v: 72624976668147839, b: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
		{v: 72624976668147840, b: []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x00}},
		{v: 9295997013522923647, b: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
		{v: 9295997013522923648, b: []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x00}},
		{v: 18446744073709551615, b: []byte{0x80, 0xfe, 0xfe, 0xfe, 0xfe, 0xfe, 0xfe, 0xfe, 0xfe, 0x7f}},
	} {
		t.Run("", func(t *testing.T) {
			tt := assert.WrapTB(t)
			rv, n, err := DecodeUint(tc.b)
			tt.MustEqual(tc.v, rv)
			tt.MustEqual(len(tc.b), n)
			tt.MustOK(err)
		})
	}
}

func TestVarIntBoundaries(t *testing.T) {
	for _, tc := range []struct {
		v int64
		b []byte
	}{
		{v: -64, b: []byte{0x7f}},
		{v: 64, b: []byte{0x80, 0x00}},
		{v: -8256, b: []byte{0xff, 0x7f}},
		{v: 8256, b: []byte{0x80, 0x80, 0x00}},
		{v: -1056832, b: []byte{0xff, 0xff, 0x7f}},
		{v: 1056832, b: []byte{0x80, 0x80, 0x80, 0x00}},
		{v: -135274560, b: []byte{0xff, 0xff, 0xff, 0x7f}},
		{v: 135274560, b: []byte{0x80, 0x80, 0x80, 0x80, 0x00}},
		{v: -17315143744, b: []byte{0xff, 0xff, 0xff, 0xff, 0x7f}},
		{v: 17315143744, b: []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x00}},
		{v: -2216338399296, b: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
		{v: 2216338399296, b: []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x00}},
		{v: -283691315109952, b: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
		{v: 283691315109952, b: []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x00}},
		{v: -36312488334073920, b: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
		{v: 36312488334073920, b: []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x00}},
	} {
		t.Run("", func(t *testing.T) {
			tt := assert.WrapTB(t)
			rv, n, err := DecodeInt(tc.b)
			tt.MustEqual(tc.v, rv)
			tt.MustEqual(len(tc.b), n)
			tt.MustOK(err)
		})
	}
}

func TestVarUint(t *testing.T) {
	tt := assert.WrapTB(t)
	b := make([]byte, 0, 16)

	assertUint(tt, 3, b)
	assertUint(tt, 63, b)
	assertUint(tt, 64, b)
	assertUint(tt, 65535, b)
	assertUint(tt, 65536, b)
}

func TestVarInt(t *testing.T) {
	tt := assert.WrapTB(t)
	b := make([]byte, 0, 16)

	assertInt(tt, -3, b)
	assertInt(tt, int64(-1)<<32, b)
	assertInt(tt, int64(-1<<63), b)
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
	buf := make([]byte, 0, 16)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AppendUint(v, buf)
	}
}

var BenchmarkDecodeUintResult uint64

func benchmarkDecodeUint(b *testing.B, buf []byte) {
	for i := 0; i < b.N; i++ {
		v, _, _ := DecodeUint(buf)
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
		v, _, _ := DecodeInt(buf)
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

/*
def zigzag_encode (i):
	return (i >> 31) ^ (i << 1)

def zigzag_decode (i):
	return (i >> 1) ^ -(i & 1)
*/

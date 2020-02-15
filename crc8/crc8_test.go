package crc8

import (
	"crypto/rand"
	"reflect"
	"testing"
)

func TestCCITT(t *testing.T) {
	for _, tc := range []struct {
		in  []byte
		out uint8
	}{
		{[]byte{}, 0},
		{[]byte{0}, 0},
		{[]byte{1}, 7},
		{[]byte{0, 1}, 7},
		{[]byte{0, 1, 2}, 27},
		{[]byte{0, 1, 2, 3}, 72},
		{[]byte{0, 1, 2, 3, 4}, 227},
		{[]byte{0, 1, 2, 3, 4, 5}, 188},
		{[]byte{0, 1, 2, 3, 4, 5, 6}, 47},
		{[]byte{0, 1, 2, 3, 4, 5, 6, 7}, 216},
	} {
		t.Run("", func(t *testing.T) {
			if tc.out != CCITT(tc.in) {
				t.Fail()
			}
			if len(tc.in) == 8 {
				if tc.out != CCITTFirst8(tc.in) {
					t.Fail()
				}
			}
		})
	}
}

func TestCCITTFirst8Rand(t *testing.T) {
	buf := make([]byte, 32)
	rand.Read(buf)

	if !reflect.DeepEqual(CCITT(buf[:8]), CCITTFirst8(buf)) {
		t.Fatal(CCITT(buf[:8]), "!=", CCITTFirst8(buf))
	}
}

var BenchResult uint8

func BenchmarkCCITTFull(b *testing.B) {
	for _, sz := range []int{8, 50, 100, 1000} {
		buf := make([]byte, sz)
		rand.Read(buf)

		b.Run("", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				BenchResult = CCITT(buf)
			}
		})
	}
}

func BenchmarkCCITTFirst8(b *testing.B) {
	buf := make([]byte, 8)
	rand.Read(buf)

	b.Run("", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			BenchResult = CCITTFirst8(buf)
		}
	})
}

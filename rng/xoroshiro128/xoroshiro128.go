package xoroshiro128

import (
	"math/bits"

	"github.com/shabbyrobe/golib/rng/splitmix64"
)

const (
	p1, p2, p3 = 55, 14, 36 // 2018 version
	// p1, p2, p3 = 24, 16, 37 // 2020 version
)

type Source struct {
	a, b uint64
}

func NewSource(seed int64) *Source {
	r := &Source{}
	r.Seed(seed)
	return r
}

func (xo *Source) Seed(seed int64) {
	// The state must be seeded so that it is not all zero. If you have a 64-bit seed, we
	// suggest to seed a splitmix64 generator and use its output to fill s.
	sm := splitmix64.NewSource(seed)
	for xo.a == 0 && xo.b == 0 {
		xo.a, xo.b = sm.Uint64(), sm.Uint64()
	}
}

func (xo *Source) Int63() int64 {
	return int64(xo.Uint64() >> 1)
}

func (xo *Source) Uint64() uint64 {
	s0, s1 := xo.a, xo.b
	result := s0 + s1

	s1 ^= s0
	xo.a = bits.RotateLeft64(s0, p1) ^ s1 ^ (s1 << p2)
	xo.b = bits.RotateLeft64(s1, p3)

	return result
}

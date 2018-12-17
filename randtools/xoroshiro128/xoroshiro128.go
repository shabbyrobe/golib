package xoroshiro128

import "github.com/shabbyrobe/golib/randtools/splitmix64"

type Source struct {
	a, b uint64
}

func NewSource(seed int64) *Source {
	r := &Source{}
	r.Seed(seed)
	return r
}

func (xo *Source) Seed(seed int64) {
	// The state must be seeded so that it is not everywhere zero. If you have
	// a 64-bit seed, we suggest to seed a splitmix64 generator and use its
	// output to fill s.
	sm := splitmix64.NewSource(seed)
	xo.a, xo.b = sm.Uint64(), sm.Uint64()
}

func (xo *Source) Int63() int64 {
	return int64(xo.Uint64() >> 1)
}

func (xo *Source) Uint64() uint64 {
	s0, s1 := xo.a, xo.b
	result := s0 + s1

	s1 ^= s0
	xo.a = ((s0 << 55) | (s0 >> (64 - 55))) ^ s1 ^ (s1 << 14)
	xo.b = ((s1 << 36) | (s1 >> (64 - 36)))

	return result
}

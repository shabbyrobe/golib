package xoshiro256

import "github.com/shabbyrobe/golib/rng/splitmix64"

type Source struct {
	a, b, c, d uint64
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
	xo.a, xo.b, xo.c, xo.d = sm.Uint64(), sm.Uint64(), sm.Uint64(), sm.Uint64()
}

func (xo *Source) Int63() int64 {
	return int64(xo.Uint64() >> 1)
}

func (xo *Source) Uint64() uint64 {
	var smul = xo.b * 5
	var result uint64 = ((smul << 7) | (smul >> (64 - 7))) * 9
	var t = xo.b << 17

	xo.c ^= xo.a
	xo.d ^= xo.b
	xo.b ^= xo.c
	xo.a ^= xo.d

	xo.c ^= t

	xo.d = (xo.d << 45) | (xo.d >> (64 - 45))

	return result
}

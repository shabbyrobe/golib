package splitmix64

/*
This is a fixed-increment version of Java 8's SplittableRandom generator
See http://dx.doi.org/10.1145/2714064.2660195 and
http://docs.oracle.com/javase/8/docs/api/java/util/SplittableRandom.html

It is a very fast generator passing BigCrush, and it can be useful if
for some reason you absolutely want 64 bits of state; otherwise, we
rather suggest to use a xoroshiro128+ (for moderately parallel
computations) or xorshift1024* (for massively parallel computations)
generator.

http://xoshiro.di.unimi.it/splitmix64.c
*/
type Source struct {
	state uint64
}

func NewSource(seed int64) *Source {
	return &Source{state: uint64(seed)}
}

func (sm64 *Source) Seed(seed int64) { sm64.state = uint64(seed) }
func (sm64 *Source) Int63() int64    { return int64(sm64.Uint64() >> 1) }

func (sm64 *Source) Uint64() uint64 {
	sm64.state += 0x9E3779B97F4A7C15
	var z uint64 = sm64.state
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	return z ^ (z >> 31)
}

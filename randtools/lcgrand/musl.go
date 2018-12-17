package lcgrand

// Musl implements the LCG random number generation strategy used in musl libc.
type Musl struct {
	state uint64
}

const (
	muslA     = uint64(6364136223846793005)
	muslC     = uint64(1)
	muslShift = uint64(33)
)

func NewMusl(seed int64) *Musl {
	m := &Musl{}
	m.Seed(seed)
	return m
}

func (rng *Musl) Seed(seed int64) { rng.state = uint64(seed) - 1 }

func (rng *Musl) Int63() int64 {
	rng.state = muslA*rng.state + muslC
	return int64(rng.state >> muslShift)
}

func (rng *Musl) Uint64() uint64 {
	rng.state = muslA*rng.state + muslC
	return rng.state >> muslShift
}

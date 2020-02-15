package xoshiro256

import (
	"fmt"
	"math/rand"
)

func Example() {
	seed := int64(1)
	rng := rand.New(NewSource(seed))

	fmt.Println(rng.Int63())
	fmt.Println(rng.Int63())
	fmt.Println(rng.Float64())

	// Output:
	// 6483309580052039778
	// 4800180567299270261
	// 0.5741057000197226
}

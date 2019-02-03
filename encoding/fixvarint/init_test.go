package fixvarint

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"
)

const (
	fuzzDefaultIterations = int(1e6)
)

var (
	fuzzIterations = fuzzDefaultIterations
	fuzzSeed       int64

	globalRNG *rand.Rand
)

func TestMain(m *testing.M) {
	flag.IntVar(&fuzzIterations, "fixvarint.iter", fuzzIterations, "Number of iterations to fuzz each op")
	flag.Int64Var(&fuzzSeed, "fixvarint.seed", fuzzSeed, "Seed the RNG (0 == current nanotime)")
	flag.Parse()

	if fuzzSeed == 0 {
		fuzzSeed = time.Now().UnixNano()
	}

	globalRNG = rand.New(rand.NewSource(fuzzSeed))

	log.Println("rando seed:", fuzzSeed) // classic rando!
	log.Println("iterations:", fuzzIterations)

	code := m.Run()
	os.Exit(code)
}

package fixvarint

import (
	"flag"
	"log"
	"math"
	"math/bits"
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

	log.Println("-fixvarint.seed", fuzzSeed)
	log.Println("-fixvarint.iter", fuzzIterations)

	code := m.Run()
	os.Exit(code)
}

func expectedBytesFromBits(bits int) (bytes int) {
	if bits == 0 {
		return 1
	}
	bytes++
	bits -= 3 // first byte only contains 3 bits of the number
	if bits <= 0 {
		return bytes
	}

	bytes += bits / 7
	if bits%7 > 0 {
		bytes++
	}
	return bytes
}

func expectedBytesFromUint64(u uint64) (bytes int) {
	zeros := 0
	for i := 0; i < 15; i++ {
		if u != 0 && u%10 == 0 {
			u = u / 10
			zeros++
		} else {
			break
		}
	}

	return expectedBytesFromBits(bits.Len64(u))
}

func expectedBytesFromInt64(i int64) (bytes int) {
	zeros := 0
	for n := 0; n < 15; n++ {
		if i != 0 && i%10 == 0 {
			i = i / 10
			zeros++
		} else {
			break
		}
	}

	// Room is always made for the sign bit, so we always +1.
	if i == math.MinInt64 {
		return 10
	} else if i >= 0 {
		return expectedBytesFromBits(bits.Len64(uint64(i)) + 1)
	} else {
		// "-i - 1" is needed because PutVarint inverts the bits to get rid of
		// the leading run of 1-bits, which converts '-4' into '3'.
		return expectedBytesFromBits(bits.Len64(uint64(-i-1)) + 1)
	}
}

// fatalfArgs allows you to call tb.Fatalf(msg, args...) using a single, fully
// optional "args ...interface{}" param. If v[0] is not a string, FatalfArgs
// will panic.
func fatalfArgs(t testing.TB, msg string, v ...interface{}) {
	t.Helper()
	if len(v) == 0 {
		t.Fatal(msg)
	}
	t.Fatalf(v[0].(string)+": "+msg, v[1:]...)
}

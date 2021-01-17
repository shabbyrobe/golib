package fixvarnum

import (
	"flag"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	num "github.com/shabbyrobe/go-num"
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
	flag.IntVar(&fuzzIterations, "fixvarnum.iter", fuzzIterations, "Number of iterations to fuzz each op")
	flag.Int64Var(&fuzzSeed, "fixvarnum.seed", fuzzSeed, "Seed the RNG (0 == current nanotime)")
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

var u64 = num.U128From64

func u128s(s string) num.U128 {
	s = strings.Replace(s, " ", "", -1)
	b, ok := new(big.Int).SetString(s, 0)
	if !ok {
		panic(fmt.Errorf("num: u128 string %q invalid", s))
	}
	out, acc := num.U128FromBigInt(b)
	if !acc {
		panic(fmt.Errorf("num: inaccurate u128 %s", s))
	}
	return out
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

func expectedBytesFromU128(u num.U128) (bytes int) {
	zeros := 0
	ten := num.U128From64(10)

	for i := 0; i < 15; i++ {
		if !u.IsZero() {
			q, r := u.QuoRem(ten)
			if r.Equal(zeroU128) {
				u = q
				zeros++
				continue
			}
		}

		break
	}

	return expectedBytesFromBits(u.BitLen())
}

func expectedBytesFromI128(i num.I128) (bytes int) {
	zeros := 0
	ten := num.I128From64(10)

	for n := 0; n < 16; n++ {
		if !i.IsZero() {
			q, r := i.QuoRem(ten)
			if r.Equal(zeroI128) {
				i = q
				zeros++
				continue
			}
		}
		break
	}

	// Room is always made for the sign bit, so we always +1.
	if i.Equal(num.MaxI128) {
		return MaxLen128
	} else if i.GreaterOrEqualTo(zeroI128) {
		return expectedBytesFromBits(i.AsU128().BitLen() + 1)
	} else {
		// "-i - 1" is needed because PutVarint inverts the bits to get rid of
		// the leading run of 1-bits, which converts '-4' into '3'.
		return expectedBytesFromBits(i.Neg().Sub(num.I128From64(1)).AsU128().BitLen() + 1)
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

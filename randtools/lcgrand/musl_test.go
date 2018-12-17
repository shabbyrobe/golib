package lcgrand_test

import (
	"math/rand"
	"testing"

	"github.com/shabbyrobe/golib/randtools/lcgrand"
)

var U64Result uint64

func BenchmarkMusl(b *testing.B) {
	src := lcgrand.NewMusl(1)
	for i := 0; i < b.N; i++ {
		U64Result = src.Uint64()
	}
}

func BenchmarkStdlib(b *testing.B) {
	src := rand.NewSource(1).(rand.Source64)
	for i := 0; i < b.N; i++ {
		U64Result = src.Uint64()
	}
}

package splitmix64

import (
	"math/rand"
	"testing"
)

var U64Result uint64

func BenchmarkSplitMix64(b *testing.B) {
	src := NewSource(1)
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

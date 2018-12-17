package splitmix64

import "testing"

var U64Result uint64

func BenchmarkSplitMix64(b *testing.B) {
	src := NewSource(1)
	for i := 0; i < b.N; i++ {
		U64Result = src.Uint64()
	}
}

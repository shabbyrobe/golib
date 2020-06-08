package sorttools

import (
	"math/rand"
	"sort"
	"testing"
)

type uint64Slice []uint64

func (u uint64Slice) Len() int               { return len(u) }
func (u uint64Slice) Less(i int, j int) bool { return u[i] < u[j] }
func (u uint64Slice) Swap(i int, j int)      { u[i], u[j] = u[j], u[i] }

var sortSizes = []int{0, 1, 4, 16, 128, 1024}

const seed = 0

func BenchmarkSortUint64sThis(b *testing.B) {
	for _, sz := range sortSizes {
		b.Run("", func(b *testing.B) {
			rng := rand.NewSource(seed).(rand.Source64)
			var buf = make([]uint64, sz)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				for j := 0; j < len(buf); j++ {
					buf[j] = rng.Uint64()
				}
				SortUint64s(buf)
			}
		})
	}
}

func BenchmarkSortUint64sInterface(b *testing.B) {
	for _, sz := range sortSizes {
		b.Run("", func(b *testing.B) {
			rng := rand.NewSource(seed).(rand.Source64)
			var buf = make([]uint64, sz)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				for j := 0; j < len(buf); j++ {
					buf[j] = rng.Uint64()
				}
				sort.Sort(uint64Slice(buf))
			}
		})
	}
}

func BenchmarkSortUint64sCallback(b *testing.B) {
	for _, sz := range sortSizes {
		rng := rand.NewSource(seed).(rand.Source64)
		b.Run("", func(b *testing.B) {
			var buf = make([]uint64, sz)
			b.ResetTimer()
			b.ReportAllocs()
			cb := func(i, j int) bool {
				return buf[i] < buf[j]
			}
			for i := 0; i < b.N; i++ {
				for j := 0; j < len(buf); j++ {
					buf[j] = rng.Uint64()
				}
				sort.Slice(buf, cb)
			}
		})
	}
}

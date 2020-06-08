package sorttools

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func TestSortUint64s(t *testing.T) {
	var max = 2000
	var step = 2

	var buf1 = make([]uint64, max)
	var buf2 = make([]uint64, max)

	for i := 0; i < max; i += step {
		rng := rand.NewSource(int64(i)).(rand.Source64)
		buf1 = buf1[:i]
		buf2 = buf2[:i]

		for j := 0; j < i; j++ {
			v := rng.Uint64()
			buf1[j] = v
			buf2[j] = v
		}

		SortUint64s(buf1)
		sort.Slice(buf2, func(i, j int) bool {
			return buf2[i] < buf2[j]
		})

		if !reflect.DeepEqual(buf1, buf2) {
			t.Fatal(i, buf1, buf2)
		}
	}
}

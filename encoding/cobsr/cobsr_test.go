package cobsr

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestCOBSR(t *testing.T) {
	for i := 0; i < 10000; i++ {
		t.Run("", func(t *testing.T) {
			in := make([]byte, i)
			result := make([]byte, i+1)
			out := make([]byte, (i+1)*2)
			rand.Read(in)
			n, err := Encode(in, out)
			if err != nil {
				t.Fatal(err)
			}
			if n > MaxEncodedSize(len(in)) {
				t.Fatal()
			}

			n, err = Decode(out[:n], result)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(in, result[:n]) {
				t.Fatal()
			}
			if i != n {
				t.Fatal()
			}
		})
	}
}

var BenchCOBSRResult int

var sizes = []struct {
	name string
	sz   int
}{
	{"sz=1", 1},
	{"sz=10", 10},
	{"sz=30", 30},
	{"sz=50", 50},
	{"sz=100", 100},
	{"sz=10000", 10000},
	{"sz=1000000", 1000000},
}

func BenchmarkCOBSREncode(b *testing.B) {
	buf := make([]byte, 1000000)
	out := make([]byte, 1100000)
	rand.Read(buf)

	for _, tc := range sizes {
		b.Run(tc.name, func(b *testing.B) {
			var err error
			cur := buf[:tc.sz]
			b.SetBytes(int64(tc.sz))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				BenchCOBSRResult, err = Encode(cur, out)
				if err != nil {
					panic(err)
				}
			}
		})
	}
}

func BenchmarkCOBSRDecode(b *testing.B) {
	for _, tc := range sizes {
		b.Run(tc.name, func(b *testing.B) {
			buf := make([]byte, tc.sz)
			out := make([]byte, tc.sz*2)
			result := make([]byte, tc.sz)

			rand.Read(buf)
			n, err := Encode(buf, out)
			if err != nil {
				panic(err)
			}

			cur := out[:n]
			b.SetBytes(int64(tc.sz))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				BenchCOBSRResult, err = Decode(cur, result)
				if err != nil {
					panic(err)
				}
			}
		})
	}
}

func BenchmarkBaseline(b *testing.B) {
	sz := 1000000
	b.SetBytes(int64(sz))
	buf := make([]byte, sz)
	rand.Read(buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, b := range buf {
			BenchCOBSRResult += int(b)
		}
	}
}

package incrementer

import (
	"strconv"
	"strings"
	"testing"

	"github.com/shabbyrobe/golib/assert"
)

func TestIncZeroValue(t *testing.T) {
	tt := assert.WrapTB(t)

	var inc Inc
	tt.MustEqual("0", inc.Current())
	tt.MustEqual("1", inc.Next())
	tt.MustEqual("2", inc.Next())
}

func TestIncTrim(t *testing.T) {
	tt := assert.WrapTB(t)

	var inc Inc
	for i := 0; i < 100; i++ {
		v := strings.Repeat("0", i) + "1"
		inc.Set(v)
		tt.MustEqual("1", inc.Current())
		tt.MustEqual("2", inc.Next())
	}
}

func TestIncRange(t *testing.T) {
	tt := assert.WrapTB(t)

	var inc Inc
	for i := int64(0); i < 10000; i++ {
		tt.MustEqual(strconv.FormatInt(i, 10), inc.Current())
		nxt := inc.Next()
		// fmt.Println(i, inc.Current(), nxt)
		tt.MustEqual(strconv.FormatInt(i+1, 10), nxt)
	}
}

func TestIncIntBoundaries(t *testing.T) {
	tt := assert.WrapTB(t)

	var inc Inc
	tt.MustOK(inc.Set("2147483647"))
	tt.MustEqual("2147483647", inc.Current())
	tt.MustEqual("2147483648", inc.Next())

	tt.MustOK(inc.Set("4294967295"))
	tt.MustEqual("4294967295", inc.Current())
	tt.MustEqual("4294967296", inc.Next())

	tt.MustOK(inc.Set("9223372036854775807"))
	tt.MustEqual("9223372036854775807", inc.Current())
	tt.MustEqual("9223372036854775808", inc.Next())

	tt.MustOK(inc.Set("18446744073709551615"))
	tt.MustEqual("18446744073709551615", inc.Current())
	tt.MustEqual("18446744073709551616", inc.Next())
}

func TestIncRidiculous(t *testing.T) {
	tt := assert.WrapTB(t)

	for i := 0; i < 10000; i++ {
		in := strings.Repeat("9", i)
		out := "1" + strings.Repeat("0", i)

		var inc Inc
		tt.MustOK(inc.Set(in))
		tt.MustEqual(out, inc.Next())
	}
}

var Result string

func BenchmarkIncGrow1(b *testing.B)   { benchmarkIncGrowN(b, 1) }
func BenchmarkIncGrow10(b *testing.B)  { benchmarkIncGrowN(b, 10) }
func BenchmarkIncGrow100(b *testing.B) { benchmarkIncGrowN(b, 100) }

func benchmarkIncGrowN(b *testing.B, n int) {
	in := strings.Repeat("9", n)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var inc Inc
		inc.buf = []byte(in)
		inc.len, inc.cap = n, n
		Result = inc.Next()
	}
}

func BenchmarkIncCountFrom0To100(b *testing.B)     { benchmarkIncCountFromTo(b, 0, 100) }
func BenchmarkIncCountFrom0To10000(b *testing.B)   { benchmarkIncCountFromTo(b, 0, 10000) }
func BenchmarkIncCountFrom0To1000000(b *testing.B) { benchmarkIncCountFromTo(b, 0, 1000000) }

func benchmarkIncCountFromTo(b *testing.B, from, to int64) {
	start := strconv.FormatInt(from, 10)
	n := len(start)

	b.ResetTimer()
	for i := 0; i < b.N; {
		var inc Inc
		inc.buf = []byte(start)
		inc.len, inc.cap = n, n
		for j := from; j < to; j++ {
			Result = inc.Next()
			i++
		}
	}
}

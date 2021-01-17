package incrementer

import (
	"strconv"
	"strings"
	"testing"
)

func TestIncZeroValue(t *testing.T) {
	var inc Inc
	if inc.Current() != "0" {
		t.Fatal()
	}
	if inc.Next() != "1" {
		t.Fatal()
	}
	if inc.Next() != "2" {
		t.Fatal()
	}
}

func TestIncTrim(t *testing.T) {
	var inc Inc
	for i := 0; i < 100; i++ {
		v := strings.Repeat("0", i) + "1"
		inc.Set(v)
		if inc.Current() != "1" {
			t.Fatal()
		}
		if inc.Next() != "2" {
			t.Fatal()
		}
	}
}

func TestIncRange(t *testing.T) {
	var inc Inc
	for i := int64(0); i < 10000; i++ {
		if inc.Current() != strconv.FormatInt(i, 10) {
			t.Fatal(i)
		}
		nxt := inc.Next()
		if nxt != strconv.FormatInt(i+1, 10) {
			t.Fatal(nxt)
		}
	}
}

func TestIncIntBoundaries(t *testing.T) {
	var inc Inc
	if err := inc.Set("2147483647"); err != nil {
		t.Fatal(err)
	}
	if inc.Current() != "2147483647" {
		t.Fatal()
	}
	if inc.Next() != "2147483648" {
		t.Fatal()
	}

	if err := inc.Set("4294967295"); err != nil {
		t.Fatal(err)
	}
	if inc.Current() != "4294967295" {
		t.Fatal()
	}
	if inc.Next() != "4294967296" {
		t.Fatal()
	}

	if err := inc.Set("9223372036854775807"); err != nil {
		t.Fatal(err)
	}
	if inc.Current() != "9223372036854775807" {
		t.Fatal()
	}
	if inc.Next() != "9223372036854775808" {
		t.Fatal()
	}

	if err := inc.Set("18446744073709551615"); err != nil {
		t.Fatal(err)
	}
	if inc.Current() != "18446744073709551615" {
		t.Fatal()
	}
	if inc.Next() != "18446744073709551616" {
		t.Fatal()
	}
}

func TestIncRidiculous(t *testing.T) {
	for i := 0; i < 10000; i++ {
		in := strings.Repeat("9", i)
		out := "1" + strings.Repeat("0", i)

		var inc Inc
		if err := inc.Set(in); err != nil {
			t.Fatal(err)
		}
		if inc.Next() != out {
			t.Fatal()
		}
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

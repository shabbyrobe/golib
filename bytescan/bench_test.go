package bytescan

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"math/rand"
	"testing"
)

func BenchmarkByteScanSplitChunk(b *testing.B) {
	data := randBytes(0, 10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scn := NewScanner(data)
		scn.Split(splitChunk(100))
		j := 0
		for scn.Scan() {
			j++
		}
		if j != 100 {
			b.Fatal(j)
		}
	}
}

func BenchmarkBufioScanSplitChunk(b *testing.B) {
	data := randBytes(0, 10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scn := bufio.NewScanner(bytes.NewReader(data))
		scn.Split(splitChunk(100))
		j := 0
		for scn.Scan() {
			j++
		}
		if j != 101 { // FIXME: work out why this is 101
			b.Fatal(j)
		}
	}
}

func BenchmarkByteScanLines(b *testing.B) {
	lines := randLines(0, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scn := NewScanner(lines)
		j := 0
		for scn.Scan() {
			j++
		}
		if j != 100 {
			b.Fatal(j)
		}
	}
}

func BenchmarkBufioScanLines(b *testing.B) {
	lines := randLines(0, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scn := bufio.NewScanner(bytes.NewReader(lines))
		j := 0
		for scn.Scan() {
			j++
		}
		if j != 100 {
			b.Fatal(j)
		}
	}
}

func splitChunk(sz int) func(data []byte, atEOF bool) (advance int, token []byte, err error) {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if len(data) > 0 {
			end := sz
			if end > len(data) {
				end = len(data)
			}
			return end, data[:end], nil
		}
		return 0, nil, nil
	}
}

func splitByte(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) > 0 {
		return 1, data[:1], nil
	}
	return 0, nil, nil
}

func randBytes(seed int64, n int) []byte {
	rng := rand.New(rand.NewSource(seed))
	var rbuf = make([]byte, n)
	rng.Read(rbuf)
	return rbuf
}

func randLines(seed int64, n int) []byte {
	rng := rand.New(rand.NewSource(seed))
	var buf bytes.Buffer

	var rbuf [32]byte
	for i := 0; i < n; i++ {
		rng.Read(rbuf[:])
		buf.WriteString(hex.EncodeToString(rbuf[:]))
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}

// +build ignore

package main

import (
	"bufio"
	"flag"
	"os"
	"time"

	"github.com/shabbyrobe/golib/rng/xoroshiro128"
)

// go run out.go | dieharder -a -g 200
func main() {
	seed := flag.Int64("seed", time.Now().UnixNano(), "RNG seed")
	src := xoroshiro128.NewSource(*seed)

	var d [8]byte
	w := bufio.NewWriter(os.Stdout)
	b := d[:]
	_ = b[7]

	for {
		i := src.Uint64()
		b[0] = byte(i >> 56)
		b[1] = byte(i >> 48)
		b[2] = byte(i >> 40)
		b[3] = byte(i >> 32)
		b[4] = byte(i >> 24)
		b[5] = byte(i >> 16)
		b[6] = byte(i >> 8)
		b[7] = byte(i)
		w.Write(b)
	}
}

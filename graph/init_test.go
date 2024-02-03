package graph

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempGraph[T ID](tb testing.TB, g *Digraph[T], fname string) string {
	tb.Helper()

	d, err := Dot(g)
	if err != nil {
		tb.Fatal(err)
	}

	full := filepath.Join(os.TempDir(), fname)
	if err := os.WriteFile(full, []byte(d), 0600); err != nil {
		tb.Fatal(err)
	}
	return full
}

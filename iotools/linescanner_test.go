package iotools

import (
	"reflect"
	"strings"
	"testing"
)

func trackDiscard() (fn func(limit, start int) error, discardStarts *[]int) {
	var starts []int
	return func(limit, start int) error {
		starts = append(starts, start)
		return nil
	}, &starts
}

func assertLines(t *testing.T, scn *LineScanner, expectedLines ...string) {
	t.Helper()

	var lines []string
	for scn.Scan() {
		lines = append(lines, string(scn.Bytes()))
	}
	if err := scn.Err(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expectedLines, lines) {
		t.Fatalf("(-want +got): %v != %v", expectedLines, lines)
	}
}

func assertDiscards(t *testing.T, starts *[]int, expected ...int) {
	t.Helper()
	if !reflect.DeepEqual(expected, *starts) {
		t.Fatalf("(-want +got): %v != %v", expected, *starts)
	}
}

func TestLineScanner(t *testing.T) {
	t.Run("nothing", func(t *testing.T) {
		scn := NewScanner(strings.NewReader(""), 2)
		assertLines(t, scn)
	})

	t.Run("one empty line", func(t *testing.T) {
		scn := NewScanner(strings.NewReader("\n"), 2)
		assertLines(t, scn, "")
	})

	t.Run("two empty lines", func(t *testing.T) {
		scn := NewScanner(strings.NewReader("\n\n"), 2)
		assertLines(t, scn, "", "")
	})

	t.Run("lines-without-trailing-nl", func(t *testing.T) {
		scn := NewScanner(strings.NewReader("a\nb\nc"), 2)
		assertLines(t, scn, "a", "b", "c")
	})

	t.Run("lines-with-trailing-nl", func(t *testing.T) {
		scn := NewScanner(strings.NewReader("a\nb\nc\n"), 2)
		assertLines(t, scn, "a", "b", "c")
	})

	t.Run("sheared-read", func(t *testing.T) {
		for i := 4; i < 8; i++ {
			scn := NewScanner(strings.NewReader("abcd\nefgh\nijkl"), i)
			assertLines(t, scn, "abcd", "efgh", "ijkl")
		}
	})

	t.Run("discard", func(t *testing.T) {
		discard, starts := trackDiscard()
		scn := NewScanner(strings.NewReader("a\nbcde\nf\n"), 2).OnDiscard(discard)
		assertLines(t, scn, "a", "f")
		assertDiscards(t, starts, 2)
	})

	t.Run("multi-read-discard", func(t *testing.T) {
		discard, starts := trackDiscard()
		scn := NewScanner(strings.NewReader("a\nbcdefghi\nj\n"), 2).OnDiscard(discard)
		assertLines(t, scn, "a", "j")
		assertDiscards(t, starts, 2)
	})
}

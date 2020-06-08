package pathtools

import (
	"fmt"
	"testing"
)

func TestAppendBeforeExt(t *testing.T) {
	for idx, tc := range []struct {
		in  string
		bit string
		n   int
		out string
	}{
		{"foo", "yep", 0, "fooyep"},
		{"foo", "yep", 1, "fooyep"},
		{"foo", "", 1, "foo"},
		{"foo.bar", "yep", 0, "foo.baryep"},
		{"foo.bar", "yep", 1, "fooyep.bar"},
		{"foo.bar", "yep", 2, "fooyep.bar"},
		{"foo.bar.baz", "yep", 0, "foo.bar.bazyep"},
		{"foo.bar.baz", "yep", 1, "foo.baryep.baz"},
		{"foo.bar.baz", "yep", 2, "fooyep.bar.baz"},
		{"foo.bar.baz", "yep", -1, "fooyep.bar.baz"},
	} {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			out, err := AppendBeforeExt(tc.in, tc.bit, tc.n)
			if err != nil {
				t.Fatal(err)
			}
			if out != tc.out {
				t.Fatal(out, "!=", tc.out)
			}
		})
	}
}

func TestAppendBeforeExtFails(t *testing.T) {
	for idx, tc := range []struct {
		in  string
		bit string
		n   int
	}{
		{"", "wat", 0},
	} {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			_, err := AppendBeforeExt(tc.in, tc.bit, tc.n)
			if err == nil {
				t.Fatal()
			}
		})
	}
}

package termfmt

import (
	"fmt"
	"strconv"
	"testing"
)

func TestFmt(t *testing.T) {
	for idx, tc := range []struct {
		format string
		in     any
		style  Style
		out    string
	}{
		{"%s", "hi", Style{}, "hi"},
		{"%d", 9, Style{}, "9"},
		{"%s", "hi", Bold(), "\x1b[1mhi\x1b[0m"},
		{"%s", "hi", Fg(1, 2, 3, 212, Red), "\x1b[38;2;1;2;3mhi\x1b[0m"},
		{"%s", "hi", FgRGB(1, 2, 3).BgRGB(9, 9, 9), "\x1b[38;2;1;2;3m\x1b[48;2;9;9;9mhi\x1b[0m\x1b[0m"},
		{"%s", "hi", Linked("http://google.com"), "\x1b]8;;http://google.com\x1b\\hi\x1b]8;;\x1b\\"},
		{"%s", "hi", FgRGB(1, 2, 3).Linked("http://google.com"), "\x1b[38;2;1;2;3m\x1b]8;;http://google.com\x1b\\hi\x1b]8;;\x1b\\\x1b[0m"},
		{"%d", 9, FgRGB(1, 2, 3).Linked("http://google.com"), "\x1b[38;2;1;2;3m\x1b]8;;http://google.com\x1b\\9\x1b]8;;\x1b\\\x1b[0m"},

		// Allow unprintable?
		{"%s", "\x1bwhatever you say mate\x1b", (Style{}).AllowUnprintable(true),
			"\x1bwhatever you say mate\x1b"},
		{"%s", "\x1bwhatever you say mate\x1b", FgRGB(1, 2, 3).AllowUnprintable(true),
			"\x1b[38;2;1;2;3m\x1bwhatever you say mate\x1b\x1b[0m"},
		{"%s", "whatever i say mate", FgRGB(1, 2, 3),
			"\x1b[38;2;1;2;3mwhatever i say mate\x1b[0m"},
	} {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			result := fmt.Sprintf(tc.format, tc.style.V(tc.in))
			if result != tc.out {
				t.Fatal("result", strconv.Quote(result), "!=", "expected", strconv.Quote(tc.out))
			}
		})
	}
}

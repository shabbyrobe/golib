package stringtools

import (
	"bufio"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestLineSplitGrab(t *testing.T) {
	cases := []struct {
		in     string
		strs   []string
		splits []string
	}{
		{"", nil, nil},
		{"\n", []string{""}, []string{"\n"}},
		{"a", []string{"a"}, []string{""}},
		{"a\n", []string{"a"}, []string{"\n"}},
		{"a\nb", []string{"a", "b"}, []string{"\n", ""}},
		{"a\nb\n", []string{"a", "b"}, []string{"\n", "\n"}},
		{"a\nb\nc", []string{"a", "b", "c"}, []string{"\n", "\n", ""}},
		{"a\nb\nc\n", []string{"a", "b", "c"}, []string{"\n", "\n", "\n"}},

		{"\r\n", []string{""}, []string{"\r\n"}},
		{"a", []string{"a"}, []string{""}},
		{"a\r\n", []string{"a"}, []string{"\r\n"}},
		{"a\r\nb", []string{"a", "b"}, []string{"\r\n", ""}},
		{"a\r\nb\r\n", []string{"a", "b"}, []string{"\r\n", "\r\n"}},
		{"a\r\nb\r\nc", []string{"a", "b", "c"}, []string{"\r\n", "\r\n", ""}},
		{"a\r\nb\r\nc\r\n", []string{"a", "b", "c"}, []string{"\r\n", "\r\n", "\r\n"}},
	}

	for idx, tc := range cases {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			in := tc.in

			var split string
			var strs, splits []string
			grb := bufio.NewScanner(strings.NewReader(in))
			grb.Split(LineSplitGrab(&split))
			for grb.Scan() {
				strs = append(strs, grb.Text())
				splits = append(splits, split)
			}

			if !reflect.DeepEqual(tc.strs, strs) {
				t.Fatal(tc.strs, "!=", strs)
			}
			if !reflect.DeepEqual(tc.splits, splits) {
				t.Fatal(fmt.Sprintf("%q", fmt.Sprint(tc.splits, "!=", splits)))
			}
		})
	}
}

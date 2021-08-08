package times

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	for idx, tc := range []struct {
		ok  bool
		in  string
		out Date
	}{
		{true, "2020-01-01", Date{2020, 1, 1}},
		{true, "9999-01-01", Date{9999, 1, 1}},
		{true, "0000-00-00", Date{0, 0, 0}},

		{false, "200-01-01", Date{}},
		{false, "20-01-01", Date{}},
		{false, "2-01-01", Date{}},
		{false, "abcd", Date{}},
		{false, "2000-1-01", Date{}},
		{false, "2000-01-1", Date{}},
		{false, "2000-01-100", Date{}},
		{false, "2020-00-01", Date{}},
		{false, "2020-13-01", Date{}},
		{false, "2020-01-00", Date{}},
		{false, "2020-01-32", Date{}},
		{false, "2020-02-30", Date{}},
		{false, "20201-01-01", Date{}},
	} {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			result, err := ParseDate(tc.in)
			if tc.ok && err != nil {
				t.Fatal("unexpected error", err)
			} else if !tc.ok && err == nil {
				t.Fatal("expected error")
			}
			if result != tc.out {
				t.Fatal(result, "!=", tc.out)
			}
		})
	}
}

func TestMarshalZeroValue(t *testing.T) {
	d := Date{}
	b, err := d.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "" {
		t.Fatal()
	}
}

func TestUnmarshalZeroValue(t *testing.T) {
	for idx, tc := range []struct {
		v string
	}{
		{""},
		{"0000-00-00"},
	} {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			d := Date{}
			if err := d.UnmarshalText([]byte(tc.v)); err != nil {
				t.Fatal(err)
			}
			if !d.IsZero() {
				t.Fatal()
			}
		})
	}
}

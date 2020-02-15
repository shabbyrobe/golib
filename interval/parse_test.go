package interval

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	for idx, tc := range []struct {
		in       string
		expected Interval
	}{
		{"1s", Raw(1, Second)},
		{"10s", Raw(10, Second)},
		{"1sec", Raw(1, Second)},
		{"10 s", Raw(10, Second)},
		{" 10 s ", Raw(10, Second)},
		{" 10 secs ", Raw(10, Second)},
		{" 10  secs ", Raw(10, Second)},
		{"10 seconds", Raw(10, Second)},
		{"10second", Raw(10, Second)},

		{"10 min", Raw(10, Minute)},
		{"10 mins", Raw(10, Minute)},
		{"10 minute", Raw(10, Minute)},
		{"10 minutes", Raw(10, Minute)},
	} {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			result := MustParse(tc.in)
			if result != tc.expected {
				t.Fatal(result != tc.expected)
			}
		})
	}
}

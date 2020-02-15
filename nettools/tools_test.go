package nettools

import (
	"fmt"
	"testing"
)

func TestHostDefaultPort(t *testing.T) {
	for idx, tc := range []struct {
		host     string
		dflt     string
		expected string
	}{
		{"localhost", "123", "localhost:123"},
		{"localhost", ":123", "localhost:123"},
		{"localhost:123", "", "localhost:123"},
		{"localhost:123", "456", "localhost:123"},
		{"localhost:", "", "localhost"},
		{"localhost:", ":", "localhost:"},
		{"localhost", ":http", "localhost:http"},
		{"localhost:", ":http", "localhost:http"},
		{"localhost:", "http", "localhost:http"},
	} {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			result := HostDefaultPort(tc.host, tc.dflt)
			if result != tc.expected {
				t.Fatal(result, "!=", tc.expected)
			}
		})
	}
}

package retools

import (
	"fmt"
	"regexp"
)

// MustSubexpIndex closes a gap in the stdlib by providing a panic on startup when a
// SubexpIndex is not found (instead of at runtime when negative array accesses happen).
//
// The proposal for this function was rejected: https://github.com/golang/go/issues/47593
func MustSubexpIndex(p *regexp.Regexp, name string) int {
	idx := p.SubexpIndex(name)
	if idx < 0 {
		panic(fmt.Errorf("regexp %q does not have named capture %q", p, name))
	}
	return idx
}

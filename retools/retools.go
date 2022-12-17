package retools

import (
	"fmt"
	"regexp"
)

func MustSubexpIndex(p *regexp.Regexp, name string) int {
	// Infinite sadness:
	// https://github.com/golang/go/issues/47593
	idx := p.SubexpIndex(name)
	if idx < 0 {
		panic(fmt.Errorf("regexp %q does not have named capture %q", p, name))
	}
	return idx
}

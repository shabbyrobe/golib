package cli

import (
	"flag"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var (
	_ flag.Value = &StringList{}
	_ flag.Value = &IntList{}
	_ flag.Value = &StringMap{}
)

var splitPattern = regexp.MustCompile(`,\s*`)

type StringList []string

func (s *StringList) String() string {
	return strings.Join(*s, ",")
}

func (s StringList) Strings() []string {
	out := make([]string, len(s))
	copy(out, s)
	return out
}

func (s *StringList) Set(v string) error {
	*s = append(*s, v)
	return nil
}

type IntList []int

func (s *IntList) String() string {
	if s == nil {
		return ""
	}
	var out []string
	for _, i := range *s {
		out = append(out, fmt.Sprintf("%d", i))
	}
	return strings.Join(out, ",")
}

func (s IntList) Ints() []int {
	out := make([]int, len(s))
	copy(out, s)
	return out
}

func (s *IntList) Set(v string) error {
	for _, part := range splitPattern.Split(v, -1) {
		if len(part) == 0 {
			continue
		}
		i, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return err
		}
		*s = append(*s, int(i))
	}
	return nil
}

type StringMap map[string]string

func (s *StringMap) String() string {
	out := ""
	first := true
	for k, v := range *s {
		if !first {
			out += ","
		}
		out += fmt.Sprintf("%s:%s", k, v)
	}
	return out
}

func (s StringMap) Map() map[string]string {
	out := make(map[string]string, len(s))
	for k, v := range s {
		out[k] = v
	}
	return out
}

func (s *StringMap) Set(v string) error {
	if *s == nil {
		*s = make(StringMap)
	}
	parts := strings.SplitN(v, ":", 2)
	if len(parts) != 2 {
		return errors.Errorf("invalid map arg %s, expected key:value", v)
	}
	(*s)[parts[0]] = parts[1]
	return nil
}

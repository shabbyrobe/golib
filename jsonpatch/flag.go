package jsonpatch

import (
	"encoding/json"
	"flag"
	"fmt"
	"regexp"
)

type Arg struct {
	Patch
	Raw string
}

var _ flag.Value = (*Arg)(nil)

func (p *Arg) String() string {
	return p.Raw
}

func (p *Arg) Type() string {
	// XXX: pflag compat
	return ""
}

func (p *Arg) Set(s string) error {
	if len(p.Raw) > 0 {
		p.Raw += " "
	}
	p.Raw += s

	// FIXME: implement a real parser so we can escape things properly
	match := patchArgPattern.FindStringSubmatch(s)
	if match == nil {
		return fmt.Errorf("invalid patch arg. expected format: --patch=op:path=value, " +
			"e.g. --patch=add:/port/magickraum=10202/tcp")
	}

	path, err := ParsePath(match[patchArgPathIndex])
	if err != nil {
		return err
	}

	value := json.RawMessage(match[patchArgValueIndex])

	var op Operation

	switch match[patchArgOpIndex] {
	case "add":
		op = Add{Path: path, Value: value}
	case "replace":
		op = Replace{Path: path, Value: value}
	case "remove":
		if value != nil {
			return fmt.Errorf("remove op must not have a value")
		}
		op = Remove{Path: path}
	case "move":
		to, err := ParsePath(string(value))
		if err != nil {
			return err
		}
		op = Move{From: path, Path: to}
	case "copy":
		to, err := ParsePath(string(value))
		if err != nil {
			return err
		}
		op = Copy{From: path, Path: to}
	}

	p.Patch.Operations = append(p.Patch.Operations, op)

	return nil
}

var (
	patchArgPattern = regexp.MustCompile(`^` +
		`(?P<op>add|remove|replace|move|copy)` +
		`:` +
		`(?P<path>[^=]+)` +
		`(` +
		`` + `=` +
		`` + `(?P<value>.*)` +
		`)?` +
		`$`)

	patchArgOpIndex    = mustIndex(patchArgPattern.SubexpIndex("op"))
	patchArgPathIndex  = mustIndex(patchArgPattern.SubexpIndex("path"))
	patchArgValueIndex = mustIndex(patchArgPattern.SubexpIndex("value"))
)

func mustIndex(idx int) int {
	if idx < 0 {
		panic(nil)
	}
	return idx
}

// Copy-pastable library of structures that can be deserialised from a JSON
// patch file.

package jsonpatch

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type OpKind string

const (
	OpAdd     OpKind = "add"
	OpReplace OpKind = "replace"
	OpRemove  OpKind = "remove"
	OpCopy    OpKind = "copy"
	OpMove    OpKind = "move"
	OpTest    OpKind = "test"
)

type Patch struct {
	Operations []Operation
}

type Operation interface {
	OpKind() OpKind
}

type Add struct {
	Path  Path
	Value json.RawMessage
}

func (Add) OpKind() OpKind { return OpAdd }

type Remove struct {
	Path Path
}

func (Remove) OpKind() OpKind { return OpRemove }

type Replace struct {
	Path  Path
	Value json.RawMessage
}

func (Replace) OpKind() OpKind { return OpReplace }

type Copy struct {
	From Path
	Path Path
}

func (Copy) OpKind() OpKind { return OpCopy }

type Move struct {
	From Path
	Path Path
}

func (Move) OpKind() OpKind { return OpMove }

type Test struct {
	Path  Path
	Value json.RawMessage
}

func (Test) OpKind() OpKind { return OpTest }

// JSON Pointer
// https://datatracker.ietf.org/doc/html/rfc6901
type Path []PathSegment

func (p Path) String() string {
	var sb strings.Builder
	sb.WriteByte('/')
	for idx, s := range p {
		if idx > 0 {
			sb.WriteByte('/')
		}
		if s.String != "" {
			sb.WriteString(s.String)
		} else {
			sb.WriteString(strconv.FormatInt(int64(s.Int), 10))
		}
	}
	return sb.String()
}

// Segments can be string for fields, or integer for indexes.
// Can't do `interface{ int | ~string }` like one would expect/hope
// so we have to go back to 'any' and type assertions, or old-style
// "let's pretend it's a union".
//
// Invalid if String != "" && Int > 0
type PathSegment struct {
	String string
	Int    int // If string is empty, assume this is an array index.
}

func (path *Path) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("path must be a string: %w", err)
	}
	parsed, err := ParsePath(raw)
	if err != nil {
		return err
	}
	*path = parsed
	return nil
}

func ParsePath(raw string) (Path, error) {
	if len(raw) == 0 || raw[0] != '/' {
		return nil, fmt.Errorf("path must start with '/'; found %q", raw)
	}

	parts := strings.Split(raw[1:], "/")
	path := make([]PathSegment, len(parts))
	for idx, part := range parts {
		part = strings.Replace(part, "~1", "/", -1)
		part = strings.Replace(part, "~0", "~", -1)

		if part == "" {
			return nil, fmt.Errorf("path %q contained empty part at index %d", path, idx)
		}

		intval, err := strconv.ParseInt(part, 10, 0)
		if err == nil {
			path[idx].Int = int(intval)
		} else {
			path[idx].String = part
		}
	}
	return path, nil
}

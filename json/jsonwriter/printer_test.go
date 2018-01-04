package jsonwriter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func writeUnmarshal(into interface{}, do func(p *printer) error) bool {
	var buf bytes.Buffer
	p := &printer{Writer: bufio.NewWriter(&buf)}
	if err := do(p); err != nil {
		return false
	}
	if err := p.Flush(); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(buf.Bytes(), &into); err != nil {
		return false
	}
	return true
}

func TestPrintStringDecodableByStdlib(t *testing.T) {
	properties := gopter.NewProperties(nil)

	sgen := gen.OneGenOf(
		MessyChar(),
		gen.AnyString(),
		gen.UnicodeString(unicode.Han),
	)

	properties.Property("printer.string works with json.Unmarshal", prop.ForAll(
		func(v string) bool {
			var check string
			return writeUnmarshal(&check, func(p *printer) error { p.string(v); return nil })
		},
		sgen,
	))

	properties.Property("printer.stringBytes works with json.Unmarshal", prop.ForAll(
		func(v string) bool {
			var check string
			return writeUnmarshal(&check, func(p *printer) error { p.stringBytes([]byte(v)); return nil })
		},
		sgen,
	))

	properties.Property("floatEncoder works with json.Unmarshal", prop.ForAll(
		func(v float64) bool {
			var check float64
			return writeUnmarshal(&check, func(p *printer) error { return float64Encoder(p, v, false) })
		},
		gen.Float64(),
	))

	properties.TestingRun(t)
}

func MessyChar() gopter.Gen {
	return genString(gen.Frequency(map[int]gopter.Gen{
		0:  gen.NumChar(),
		14: gen.AlphaChar(),
		2:  gen.RuneRange(0, 32),
		3:  gen.RuneRange(0, 255),
		1:  gen.RuneRange(2028, 2029),
		4:  gen.Rune(),
	}), utf8.ValidRune)
}

func genString(runeGen gopter.Gen, runeSieve func(ch rune) bool) gopter.Gen {
	return gen.SliceOf(runeGen).Map(runesToString).SuchThat(func(v string) bool {
		for _, ch := range v {
			if !runeSieve(ch) {
				return false
			}
		}
		return true
	}).WithShrinker(gen.StringShrinker)
}

func runesToString(v []rune) string {
	return string(v)
}

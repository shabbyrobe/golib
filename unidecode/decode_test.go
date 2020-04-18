package unidecode

import (
	"fmt"
	"strings"
	"testing"
)

var cases = []struct {
	in      string
	out     string
	inPlace bool
}{
	{"", "", true},
	{"f", "f", true},
	{"foo", "foo", true},
	{"épée", "epee", true},
	{"Épée", "Epee", true},

	{"北京", "Bei Jing", false},
	{"abc北京", "abcBei Jing", false},
	{"ネオアームストロングサイクロンジェットアームストロング砲", "neoamusutorongusaikuronzietsutoamusutoronguPao", false},
	{"30 𝗄𝗆/𝗁", "30 km/h", true},
	{"kožušček", "kozuscek", true},
	{"ⓐⒶ⑳⒇⒛⓴⓾⓿", "aA20(20)20.20100", false},
	{"Hello, World!", "Hello, World!", true},
	{`\n`, `\n`, true},
	{`北京abc\n`, `Bei Jing abc\n`, false},
	{`'"\r\n`, `'"\r\n`, true},
	{"ČŽŠčžš", "CZSczs", true},
	{"ア", "a", true},
	{"α", "a", true},
	{"a", "a", true},
	{"ch\u00e2teau", "chateau", true},
	{"vi\u00f1edos", "vinedos", true},
	{"Efﬁcient", "Efficient", false},
	{"příliš žluťoučký kůň pěl ďábelské ódy", "prilis zlutoucky kun pel dabelske ody", true},
	{"PŘÍLIŠ ŽLUŤOUČKÝ KŮŇ PĚL ĎÁBELSKÉ ÓDY", "PRILIS ZLUTOUCKY KUN PEL DABELSKE ODY", true},
	{strings.Repeat("éfficient", 1000), strings.Repeat("efficient", 1000), true},
	{strings.Repeat("efficient", 1000), strings.Repeat("efficient", 1000), true},
}

func TestDecode(t *testing.T) {
	for idx, tc := range cases {
		buf := make([]byte, 65536)
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			dec := DecodeString(tc.in)
			if dec != tc.out {
				t.Fatal(dec, "!=", tc.out)
			}

			decb := Decode([]byte(tc.in), buf)
			if string(decb) != tc.out {
				t.Fatal(string(decb), "!=", tc.out)
			}

			if tc.inPlace {
				decin := DecodeInPlace([]byte(tc.in))
				if string(decin) != tc.out {
					t.Fatal(string(decin), "!=", tc.out)
				}
			}
		})
	}
}

var BenchStringResult string
var BenchBytesResult []byte

func BenchmarkDecodeString(b *testing.B) {
	for idx, tc := range cases {
		b.Run(fmt.Sprintf("%d", idx), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				BenchStringResult = DecodeString(tc.in)
			}
		})
	}
}

func BenchmarkDecode(b *testing.B) {
	for idx, tc := range cases {
		sz := 1000
		if len(tc.in) > 1000 {
			sz = 65536
		}
		buf := make([]byte, sz)
		in := []byte(tc.in)
		b.Run(fmt.Sprintf("%d", idx), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				BenchBytesResult = Decode(in, buf)
			}
		})
	}
}

func BenchmarkDecodeInPlace(b *testing.B) {
	for idx, tc := range cases {
		if !tc.inPlace {
			continue
		}
		in := []byte(tc.in)
		cur := []byte(tc.in)

		b.Run(fmt.Sprintf("%d", idx), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				copy(cur, in) // XXX: confounds benchmark slightly
				BenchBytesResult = DecodeInPlace(cur)
			}
		})
	}
}

var BenchIntResult = 0

func BenchmarkDecodeInPlaceConfound(b *testing.B) {
	for idx, tc := range cases {
		if !tc.inPlace {
			continue
		}
		in := []byte(tc.in)
		cur := []byte(tc.in)

		b.Run(fmt.Sprintf("%d", idx), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				BenchIntResult = copy(cur, in)
			}
		})
	}
}

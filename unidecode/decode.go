package unidecode

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func MaxDecodedLen(sz int) int {
	return sz * maxLen
}

// DecodeInPlace applies all of the unidecode transliterations which are only 1 byte in
// length to the input slice b, then returns b, which will be truncated.
func DecodeInPlace(b []byte) []byte {
	var r, w int
	for r < len(b) {
		if b[r] <= unicode.MaxASCII {
			if r != w {
				b[w] = b[r]
			}
			r, w = r+1, w+1
			continue
		}

		rn, sz := utf8.DecodeRune(b[r:])
		if rn == utf8.RuneError || rn > maxRune || single[rn] == 0 {
			r += sz
			continue
		}

		b[w] = single[rn]
		r, w = r+sz, w+1
	}
	return b[:w]
}

// Decode copies all of the unidecode transliterations found in 'v' into 'buf', returning
// 'buf'. If 'buf' is not large enough to hold the result, this will panic. Call
// MaxDecodedLen to work out the worst case allocation prior to calling.
func Decode(v []byte, buf []byte) []byte {
	var r, w int
	for r < len(v) {
		if v[r] <= unicode.MaxASCII {
			buf[w] = v[r]
			r++
			w++
			continue
		}

		rn, sz := utf8.DecodeRune(v[r:])
		if rn == utf8.RuneError || rn > maxRune {
			r += sz
			continue
		}
		if multi[rn] != nil {
			w += copy(buf[w:], multi[rn])
		}
		r += sz
	}

	if w > 0 && buf[w-1] == ' ' {
		w--
	}

	return buf[:w]
}

func DecodeString(v string) string {
	var buf strings.Builder
	for _, rn := range v {
		if rn == utf8.RuneError || rn > maxRune {
			continue
		}
		if rn <= unicode.MaxASCII {
			buf.WriteByte(byte(rn))
		} else if multi[rn] != nil {
			buf.Write(multi[rn])
		}
	}

	out := buf.String()
	last := len(out) - 1
	if len(out) > 0 && out[last] == ' ' {
		out = out[:last]
	}

	return out
}

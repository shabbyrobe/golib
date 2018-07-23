package stringtools

import (
	"strings"
	"unicode/utf8"
)

// RightPad is a startling and regrettable omission from the stdlib.
// There really does not appear to be any way to accomplish the same thing with
// fmt.Sprintf(), or any other damn package.  Every developer who has ever
// contributed to golang has hereby forefeited the right to criticise
// javascript for the left-pad.io fiasco.
func RightPad(s string, c byte, total int) string {
	pad := total - len(s)
	if pad <= 0 {
		return s
	}
	return s + strings.Repeat(string(c), pad)
}

func RightPadUTF8(s string, r rune, total int) string {
	rc := utf8.RuneCountInString(s)
	pad := total - rc
	if pad <= 0 {
		return s
	}
	return s + strings.Repeat(string(r), pad)
}

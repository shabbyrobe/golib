package termfmt

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type Escape interface {
	Wrap(out string) string
}

func With(escs ...Escape) Style                       { return (Style{}).With(escs...) }
func Bold() Style                                     { return (Style{}).Bold() }
func Italic() Style                                   { return (Style{}).Italic() }
func Linked(link string) Style                        { return (Style{}).Linked(link) }
func FgRGB(r, g, b uint8) Style                       { return (Style{}).FgRGB(r, g, b) }
func BgRGB(r, g, b uint8) Style                       { return (Style{}).BgRGB(r, g, b) }
func Fg(r, g, b uint8, c256 uint8, c16 C16Name) Style { return (Style{}).Fg(r, g, b, c256, c16) }
func Bg(r, g, b uint8, c256 uint8, c16 C16Name) Style { return (Style{}).Bg(r, g, b, c256, c16) }

type Style struct {
	escapes          []Escape
	allowUnprintable bool
	v                any
}

var _ fmt.Formatter = Style{}

func (c Style) With(escs ...Escape) Style {
	c.escapes = append(c.escapes, escs...)
	return c
}

func (c Style) Bold() Style               { return c.With(BoldEscape{}) }
func (c Style) Italic() Style             { return c.With(ItalicEscape{}) }
func (c Style) Linked(link string) Style  { return c.With(Link{link}) }
func (c Style) FgRGB(r, g, b uint8) Style { return c.With(RGBColor{r, g, b, false}) }
func (c Style) BgRGB(r, g, b uint8) Style { return c.With(RGBColor{r, g, b, true}) }

func (c Style) AllowUnprintable(yep bool) Style {
	c.allowUnprintable = true
	return c
}

func (c Style) Fg(r, g, b uint8, c256 uint8, c16 C16Name) Style {
	return c.With((ColorCascade{}).
		RGB(RGBColor{r, g, b, false}).
		C256(C256Color{c256, false}).
		C16(C16Color{c16, false}))
}

func (c Style) Bg(r, g, b uint8, c256 uint8, c16 C16Name) Style {
	return c.With((ColorCascade{}).
		Background().
		RGB(RGBColor{r, g, b, true}).
		C256(C256Color{c256, true}).
		C16(C16Color{c16, true}))
}

func (c Style) V(v any) Style {
	c.v = v
	return c
}

func (c Style) Format(f fmt.State, verb rune) {
	v := fmt.Sprintf(buildValueFormat(f, verb), c.v)
	if !c.allowUnprintable {
		v = printable(v)
	}
	for i := len(c.escapes) - 1; i >= 0; i-- {
		v = c.escapes[i].Wrap(v)
	}
	f.Write([]byte(v))
}

func buildValueFormat(f fmt.State, verb rune) string {
	s := "%"
	if f.Flag(' ') {
		s += " "
	}
	if f.Flag('+') {
		s += "+"
	}
	if f.Flag('-') {
		s += "-"
	}
	if f.Flag('0') {
		s += "0"
	}
	if f.Flag('#') {
		s += "#"
	}
	width, ok := f.Width()
	if ok {
		s += strconv.Itoa(width)
	}
	prec, ok := f.Precision()
	if ok {
		s += "." + strconv.Itoa(prec)
	}
	s += string(verb)
	return s
}

type Link struct {
	URL string
}

func (l Link) Wrap(out string) string {
	return fmt.Sprintf(""+
		"\x1b]8;;"+
		"%s"+
		"\x1b\\"+
		"%s"+
		"\x1b]8;;\x1b\\",
		printable(l.URL),
		out)
}

type BoldEscape struct{}

func (b BoldEscape) Wrap(v string) string { return fmt.Sprintf("\x1b[1m%s\x1b[0m", v) }

type ItalicEscape struct{}

func (b ItalicEscape) Wrap(v string) string { return fmt.Sprintf("\x1b[3m%s\x1b[0m", v) }

// https://github.com/termstandard/colors
type RGBColor struct {
	R, G, B uint8
	Bg      bool
}

func (rgb RGBColor) Background() RGBColor {
	rgb.Bg = true
	return rgb
}

func (rgb RGBColor) Wrap(out string) string {
	esc := 38
	if rgb.Bg {
		esc = 48
	}
	return fmt.Sprintf("\x1b[%d;2;%d;%d;%dm"+"%s"+"\x1b[0m", esc, rgb.R, rgb.G, rgb.B, out)
}

// https://gist.github.com/fnky/458719343aabd01cfb17a3a4f7296797?permalink_comment_id=4619910#gistcomment-4619910
type C256Color struct {
	C  uint8
	Bg bool
}

func (c C256Color) Background() C256Color {
	c.Bg = true
	return c
}

func (c C256Color) Wrap(out string) string {
	esc := 38
	if c.Bg {
		esc = 48
	}
	return fmt.Sprintf("\x1b[%d;5;%dm"+"%s"+"\x1b[0m", esc, c.C, out)
}

type C16Name uint8

const (
	DefaultColor C16Name = iota

	Black
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	LightGrey

	DarkGrey
	LightRed
	LightGreen
	LightYellow
	LightBlue
	LightMagenta
	LightCyan
	White
)

type C16Color struct {
	Name C16Name
	Bg   bool
}

func (c C16Color) Background() C16Color {
	c.Bg = true
	return c
}

func (c C16Color) Wrap(out string) string {
	var cv uint8
	if c.Name == DefaultColor {
		cv = 39
	} else {
		// Our enum starts at one, adjust so it starts at 0:
		cv = uint8(c.Name) - 1

		// If fg, the lower 8 colours run from 30 to 37, the upper 8 from 90 to 97.
		// We take care of bg later.
		if c.Name < DarkGrey {
			cv += 30
		} else {
			cv += 90
		}
	}

	if c.Bg {
		cv += 10
	}

	return fmt.Sprintf("\x1b[%dm"+"%s"+"\x1b[0m", cv, out)
}

func Reset() {
	rgbSupported = true
	c256Supported = true
}

var (
	rgbSupported  bool
	c256Supported bool
)

func init() {
	Reset()
}

func RGBSupported(supported bool)  { rgbSupported = supported }
func C256Supported(supported bool) { c256Supported = supported }

type ColorCascade struct {
	rgb    RGBColor
	rgbSet bool

	c256    C256Color
	c256Set bool

	c16    C16Color
	c16Set bool

	bg bool
}

func (cc ColorCascade) Wrap(out string) string {
	if rgbSupported && cc.rgbSet {
		return cc.rgb.Wrap(out)
	} else if c256Supported && cc.c256Set {
		return cc.c256.Wrap(out)
	} else if cc.c16Set {
		return cc.c16.Wrap(out)
	}
	return ""
}

func (cc ColorCascade) Background() ColorCascade {
	cc.bg = true
	cc.rgb.Bg = true
	cc.c256.Bg = true
	cc.c16.Bg = true
	return cc
}

func (cc ColorCascade) RGB(c RGBColor) ColorCascade {
	if cc.bg {
		c = c.Background()
	}
	cc.rgb = c
	cc.rgbSet = true
	return cc
}

func (cc ColorCascade) C256(c C256Color) ColorCascade {
	if cc.bg {
		c = c.Background()
	}
	cc.c256 = c
	cc.c256Set = true
	return cc
}

func (cc ColorCascade) C16(c C16Color) ColorCascade {
	if cc.bg {
		c = c.Background()
	}
	cc.c16 = c
	cc.c16Set = true
	return cc
}

func (cc ColorCascade) C16Name(c C16Color) ColorCascade {
	if cc.bg {
		c = c.Background()
	}
	cc.c16 = c
	cc.c16Set = true
	return cc
}

func mapPrintable(r rune) rune {
	if unicode.IsGraphic(r) {
		return r
	}
	return -1
}

func printable(v string) string {
	return strings.Map(mapPrintable, v)
}

package fmttools

import "fmt"

// FormatInt makes it a bit easier to implement a '%d' format branch for a type
// that implements fmt.Formatter. It assumes a rune of 'd' is passed.
//
// For example:
//
//   type Test int
//
//   func (t Test) String() { return fmt.Sprintf("YEP %d YEP", t) }
//
//   func (t Test) Format(f fmt.State, c rune) {
//   	if c == 's' || c == 'v' {
//   		io.WriteString(f, m.String())
//   	} else {
//   		fmttools.FormatInt(int64(m), f)
//   	}
//   }
//
func FormatInt(v int64, f fmt.State) {
	zero := f.Flag('0')
	w, wok := f.Width()
	if zero && wok {
		fmt.Fprintf(f, "%0*d", w, v)
	} else if wok {
		fmt.Fprintf(f, "%*d", w, v)
	} else {
		fmt.Fprintf(f, "%d", v)
	}
}

// FormatInt makes it a bit easier to implement a '%d' format branch for a type
// that implements fmt.Formatter. It assumes a rune of 'd' is passed.
func FormatUint(v uint64, f fmt.State) {
	zero := f.Flag('0')
	w, wok := f.Width()
	if zero && wok {
		fmt.Fprintf(f, "%0*d", w, v)
	} else if wok {
		fmt.Fprintf(f, "%*d", w, v)
	} else {
		fmt.Fprintf(f, "%d", v)
	}
}

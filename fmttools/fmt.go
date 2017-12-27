package fmttools

import "fmt"

// FormatInt makes it a bit easier to implement custom formatting
// that can fall back to formatting as decimal. It assumes a rune of
// 'd' is passed.
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

// FormatUint makes it a bit easier to implement custom formatting
// that can fall back to formatting as decimal. It assumes a rune of
// 'd' is passed.
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

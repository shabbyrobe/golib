package stringtools

import "bufio"

// LineSplitGrab returns a bufio.SplitFunc that will place the line separator
// used for each call to Scan() into the string pointed to by 'split'.
func LineSplitGrab(split *string) bufio.SplitFunc {
	// XXX: This does not work with any splitfunc; SplitWords can trim whitespace
	// off the _front_ of the scanned token as well, which is harder to detect.
	fn := bufio.ScanLines

	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		advance, token, err = fn(data, atEOF)
		*split = string(data[len(token):advance])
		return advance, token, err
	}
}

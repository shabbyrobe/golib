package mathtools

// AbsInt64 returns the absolute value of n.
// http://cavaliercoder.com/blog/optimized-abs-for-int64-in-go.html
func AbsInt64(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}

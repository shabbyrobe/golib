package num

const (
	maxUint64Float    = float64(maxUint64) + 1
	maxUint64NegFloat = -(float64(maxUint64) + 1)
	maxUint64         = 1<<64 - 1

	maxInt64Float = float64(maxInt64) + 1
	maxInt64      = 1<<63 - 1
	minInt64      = -1 << 63
)

var (
	zeroI128 I128
	zeroU128 U128
)

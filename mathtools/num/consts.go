package num

import (
	"math/big"
)

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
	zeroBig  = new(big.Int).SetInt64(0)

	maxBigUint64  = new(big.Int).SetUint64(maxUint64)
	maxBigU128, _ = new(big.Int).SetString("340282366920938463463374607431768211455", 10)
	maxBigInt64   = new(big.Int).SetUint64(maxInt64)

	minBigI128, _ = new(big.Int).SetString("-170141183460469231731687303715884105728", 10)
	maxBigI128, _ = new(big.Int).SetString("170141183460469231731687303715884105727", 10)

	// wrapBigU128 is 1 << 128, used to simulate over/underflow:
	wrapBigU128, _ = new(big.Int).SetString("340282366920938463463374607431768211456", 10)
)

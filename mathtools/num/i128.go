package num

import (
	"math/big"
)

type I128 struct {
	hi uint64
	lo uint64
}

const (
	signBit  = 0x8000000000000000
	signMask = 0x7FFFFFFFFFFFFFFF
)

func I128FromRaw(hi, lo uint64) I128 {
	return I128{hi: hi, lo: lo}
}

func I128From64(v int64) I128 {
	var hi uint64
	if v < 0 {
		hi = maxUint64
	}
	return I128{hi: hi, lo: uint64(v)}
}

func I128From32(v int32) I128 { return I128From64(int64(v)) }
func I128From16(v int16) I128 { return I128From64(int64(v)) }
func I128From8(v int8) I128   { return I128From64(int64(v)) }

func I128FromBigInt(v *big.Int) (out I128, accurate bool) {
	neg := v.Cmp(big0) < 0
	var a, b big.Int
	if neg {
		a.Neg(v).Sub(&a, big1).Xor(&a, maxBigU128)
	} else {
		a.Set(v)
	}
	out.lo = b.And(&a, maxBigUint64).Uint64()
	out.hi = a.Rsh(&a, 64).Uint64()
	return out, v.Cmp(minBigI128) >= 0 && v.Cmp(maxBigI128) <= 0
}

func I128FromFloat32(f float32) I128 { return I128FromFloat64(float64(f)) }

func I128FromFloat64(f float64) (out I128) {
	const spillPos = float64(maxUint64) // (1<<64) - 1
	const spillNeg = -float64(maxUint64) - 1

	if f == 0 {
		return out

	} else if f < 0 {
		if f >= spillNeg {
			return I128{hi: maxUint64, lo: uint64(f)}
		} else if f >= minI128Float {
			f = -f
			lo := mod(f, wrapUint64Float)
			return I128{hi: ^uint64(f / wrapUint64Float), lo: ^uint64(lo)}
		} else {
			return MinI128
		}

	} else {
		if f <= spillPos {
			return I128{lo: uint64(f)}
		} else if f <= maxI128Float {
			lo := mod(f, wrapUint64Float)
			return I128{hi: uint64(f / wrapUint64Float), lo: uint64(lo)}
		} else {
			return MaxI128
		}
	}
}

// RandI128 generates a positive signed 128-bit random integer from an external
// source.
func RandI128(source RandSource) (out I128) {
	return I128{hi: source.Uint64() & maxInt64, lo: source.Uint64()}
}

func (i I128) Raw() (hi uint64, lo uint64) { return uint64(i.hi), i.lo }

func (i I128) String() string {
	v := i.AsBigInt() // This is good enough for now
	return v.String()
}

func (i I128) IntoBigInt(b *big.Int) {
	neg := i.hi&signBit != 0
	if i.hi > 0 {
		b.SetUint64(i.hi)
		b.Lsh(b, 64)
	}
	var lo big.Int
	lo.SetUint64(i.lo)
	b.Add(b, &lo)

	if neg {
		b.Xor(b, maxBigU128).Add(b, big1).Neg(b)
	}
}

func (i I128) AsBigInt() (b *big.Int) {
	b = new(big.Int)
	neg := i.hi&signBit != 0
	if i.hi > 0 {
		b.SetUint64(i.hi)
		b.Lsh(b, 64)
	}
	var lo big.Int
	lo.SetUint64(i.lo)
	b.Add(b, &lo)

	if neg {
		b.Xor(b, maxBigU128).Add(b, big1).Neg(b)
	}

	return b
}

func (i I128) AsU128() U128 {
	return U128{lo: i.lo, hi: uint64(i.hi)}
}

func (i I128) Sign() int {
	if i == zeroI128 {
		return 0
	} else if i.hi&signBit == 0 {
		return 1
	}
	return -1
}

func (i I128) AsFloat64() float64 {
	if i.hi == 0 && i.lo == 0 {
		return 0
	} else if i.hi&signBit != 0 {
		if i.hi == maxUint64 {
			return -float64((^i.lo) + 1)
		} else {
			return (-float64(^i.hi) * maxUint64Float) + -float64(^i.lo)
		}
	} else {
		if i.hi == 0 {
			return float64(i.lo)
		} else {
			return (float64(i.hi) * maxUint64Float) + float64(i.lo)
		}
	}
}

func (i I128) AsBigFloat() (b *big.Float) {
	return new(big.Float).SetInt(i.AsBigInt())
}

func (i I128) Inc() (v I128) {
	v.lo = i.lo + 1
	v.hi = i.hi
	if i.lo > v.lo {
		v.hi++
	}
	return v
}

func (i I128) Dec() (v I128) {
	v.lo = i.lo - 1
	v.hi = i.hi
	if i.lo < v.lo {
		v.hi--
	}
	return v
}

func (i I128) Add(n I128) (v I128) {
	v.lo = i.lo + n.lo
	v.hi = i.hi + n.hi
	if i.lo > v.lo {
		v.hi++
	}
	return v
}

func (i I128) Sub(n I128) (v I128) {
	v.lo = i.lo - n.lo
	v.hi = i.hi - n.hi
	if i.lo < v.lo {
		v.hi--
	}
	return v
}

func (i I128) Neg() (v I128) {
	if i.hi == 0 && i.lo == 0 {
		return v
	}
	if i.hi&signBit != 0 {
		v.hi = ^i.hi
		v.lo = ^(i.lo - 1)
	} else {
		v.hi = ^i.hi
		v.lo = (^i.lo) + 1
	}
	return v
}

func (i I128) Abs() I128 {
	if i.hi&signBit != 0 {
		i.hi = ^i.hi
		i.lo = ^(i.lo - 1)
	}
	return i
}

func (i I128) Cmp(n I128) int {
	if i.hi == n.hi && i.lo == n.lo {
		return 0
	} else if i.hi&signBit == n.hi&signBit {
		if i.hi > n.hi || (i.hi == n.hi && i.lo > n.lo) {
			return 1
		}
	} else if i.hi&signBit == 0 {
		return 1
	}
	return -1
}

func (i I128) Equal(n I128) bool {
	return i.hi == n.hi && i.lo == n.lo
}

func (i I128) GreaterThan(n I128) bool {
	if i.hi&signBit == n.hi&signBit {
		return i.hi > n.hi || (i.hi == n.hi && i.lo > n.lo)
	} else if i.hi&signBit == 0 {
		return true
	}
	return false
}

func (i I128) GreaterOrEqualTo(n I128) bool {
	if i.hi == n.hi && i.lo == n.lo {
		return true
	}
	if i.hi&signBit == n.hi&signBit {
		return i.hi > n.hi || (i.hi == n.hi && i.lo > n.lo)
	} else if i.hi&signBit == 0 {
		return true
	}
	return false
}

func (i I128) LessThan(n I128) bool {
	if i.hi&signBit == n.hi&signBit {
		return i.hi < n.hi || (i.hi == n.hi && i.lo < n.lo)
	} else if i.hi&signBit != 0 {
		return true
	}
	return false
}

func (i I128) LessOrEqualTo(n I128) bool {
	if i.hi == n.hi && i.lo == n.lo {
		return true
	}
	if i.hi&signBit == n.hi&signBit {
		return i.hi < n.hi || (i.hi == n.hi && i.lo < n.lo)
	} else if i.hi&signBit != 0 {
		return true
	}
	return false
}

func (i I128) Mul(n I128) (dest I128) {
	// Adapted from Warren, Hacker's Delight, p. 132.
	hl := i.hi*n.lo + i.lo*n.hi

	dest.lo = i.lo * n.lo // lower 64 bits are easy

	// break the multiplication into (x1 << 32 + x0)(y1 << 32 + y0)
	// which is x1*y1 << 64 + (x0*y1 + x1*y0) << 32 + x0*y0
	// so now we can do 64 bit multiplication and addition and
	// shift the results into the right place
	x0, x1 := i.lo&0x00000000ffffffff, i.lo>>32
	y0, y1 := n.lo&0x00000000ffffffff, n.lo>>32
	t := x1*y0 + (x0*y0)>>32
	w1 := (t & 0x00000000ffffffff) + (x0 * y1)
	dest.hi = (x1 * y1) + (t >> 32) + (w1 >> 32) + hl

	return dest
}

// QuoRem returns the quotient q and remainder r for y != 0. If y == 0, a
// division-by-zero run-time panic occurs.
//
// QuoRem implements T-division and modulus (like Go):
//
//	q = x/y      with the result truncated to zero
//	r = x - y*q
//
// U128 does not support big.Int.DivMod()-style Euclidean division.
//
func (i I128) QuoRem(by I128) (q, r I128) {
	qSign, rSign := 1, 1
	if i.LessThan(zeroI128) {
		qSign, rSign = -1, -1
		i = i.Neg()
	}
	if by.LessThan(zeroI128) {
		qSign = -qSign
		by = by.Neg()
	}

	qu, ru := i.AsU128().QuoRem(by.AsU128())
	q, r = qu.AsI128(), ru.AsI128()
	if qSign < 0 {
		q = q.Neg()
	}
	if rSign < 0 {
		r = r.Neg()
	}
	return q, r
}

// Quo returns the quotient x/y for y != 0. If y == 0, a division-by-zero
// run-time panic occurs. Quo implements truncated division (like Go); see
// QuoRem for more details.
func (i I128) Quo(by I128) (q I128) {
	// FIXME: can do much better than this.
	q, _ = i.QuoRem(by)
	return q
}

// Rem returns the remainder of x%y for y != 0. If y == 0, a division-by-zero
// run-time panic occurs. Rem implements truncated modulus (like Go); see
// QuoRem for more details.
func (i I128) Rem(by I128) (r I128) {
	// FIXME: can do much better than this.
	_, r = i.QuoRem(by)
	return r
}

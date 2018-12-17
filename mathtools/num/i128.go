package num

import (
	"math/big"
)

type I128 struct {
	hi int64
	lo uint64
}

var MaxI128 = I128{hi: maxInt64, lo: maxUint64}
var MinI128 = I128{hi: minInt64, lo: 0}

func I128FromRaw(hi, lo uint64) I128 {
	return I128{hi: int64(hi), lo: lo}
}

func I128From64(v int64) I128 {
	var hi int64
	if v < 0 {
		hi = -1
	}
	return I128{hi: hi, lo: uint64(v)}
}

func I128From32(v int32) I128 {
	var hi int64
	if v < 0 {
		hi = -1
	}
	return I128{hi: hi, lo: uint64(int64(v))}
}

func I128From16(v int16) I128 {
	var hi int64
	if v < 0 {
		hi = -1
	}
	return I128{hi: hi, lo: uint64(int64(v))}
}

func I128From8(v int8) I128 {
	var hi int64
	if v < 0 {
		hi = -1
	}
	return I128{hi: hi, lo: uint64(int64(v))}
}

func I128FromBigInt(v *big.Int) (out I128) {
	var a, b big.Int
	out.lo = b.And(v, maxBigUint64).Uint64()
	out.hi = a.Rsh(v, 64).Int64()
	return out
}

func I128FromFloat32(f float32) I128 { return I128FromFloat64(float64(f)) }

func I128FromFloat64(f float64) I128 {
	if f == 0 {
		return I128{}

	} else if f < 0 {
		if f >= maxUint64NegFloat {
			return I128{hi: -1, lo: uint64(-f)}
		} else {
			return I128{hi: int64(f / maxUint64Float), lo: uint64(f)}
		}

	} else {
		if f <= maxUint64Float {
			return I128{lo: uint64(f)}
		} else {
			return I128{hi: int64(f / maxUint64Float), lo: uint64(f)}
		}
	}
}

func (i I128) Raw() (hi uint64, lo uint64) { return uint64(i.hi), i.lo }

func (i I128) String() string {
	v := i.AsBigInt() // This is good enough for now
	return v.String()
}

func (i I128) IntoBigInt(b *big.Int) {
	b.SetInt64(i.hi)
	b.Lsh(b, 64)

	var lo big.Int
	lo.SetUint64(i.lo)
	b.Add(b, &lo)
}

func (i I128) AsBigInt() (b big.Int) {
	i.IntoBigInt(&b)
	return b
}

func (i I128) AsU128() U128 {
	return U128{lo: i.lo, hi: uint64(i.hi)}
}

func (i I128) AsFloat64() float64 {
	if i.hi == 0 && i.lo == 0 {
		return 0
	} else if i.hi == 0 {
		return float64(i.lo)
	} else {
		return (float64(i.hi) * maxUint64Float) + float64(i.lo)
	}
}

func (i I128) AsBigFloat() (b big.Float) {
	v := i.AsBigInt()
	b.SetInt(&v)
	return b
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
	v.hi = -i.hi
	v.lo = -i.lo
	if v.lo > 0 {
		v.hi--
	}
	return v
}

func (i I128) Abs() I128 {
	if i.hi < 0 {
		i.hi = -i.hi
		i.lo = -i.lo
		if i.lo > 0 {
			i.hi--
		}
	}
	return i
}

func (i I128) Cmp(n I128) int {
	if i.hi > n.hi {
		return 1
	} else if i.hi < n.hi {
		return -1
	} else if i.lo > n.lo {
		return 1
	} else if i.lo < n.lo {
		return -1
	}
	return 0
}

func (i I128) Equal(n I128) bool {
	return i.hi == n.hi && i.lo == n.lo
}

func (i I128) GreaterThan(n I128) bool {
	if i.hi > n.hi {
		return true
	} else if i.hi < n.hi {
		return false
	} else if i.lo > n.lo {
		return true
	} else if i.lo < n.lo {
		return false
	}
	return false
}

func (i I128) GreaterOrEqualTo(n I128) bool {
	if i.hi > n.hi {
		return true
	} else if i.hi < n.hi {
		return false
	} else if i.lo > n.lo {
		return true
	} else if i.lo < n.lo {
		return false
	}
	return true
}

func (i I128) LessThan(n I128) bool {
	if i.hi > n.hi {
		return false
	} else if i.hi < n.hi {
		return true
	} else if i.lo > n.lo {
		return false
	} else if i.lo < n.lo {
		return true
	}
	return false
}

func (i I128) LessOrEqualTo(n I128) bool {
	if i.hi > n.hi {
		return false
	} else if i.hi < n.hi {
		return true
	} else if i.lo > n.lo {
		return false
	} else if i.lo < n.lo {
		return true
	}
	return true
}

func (i I128) Mul(n I128) (dest I128) {
	// Adapted from Warren, Hacker's Delight, p. 132.
	ih, nh := uint64(i.hi), uint64(n.hi)
	hl := ih*n.lo + i.lo*nh

	dest.lo = i.lo * n.lo // lower 64 bits are easy

	// break the multiplication into (x1 << 32 + x0)(y1 << 32 + y0)
	// which is x1*y1 << 64 + (x0*y1 + x1*y0) << 32 + x0*y0
	// so now we can do 64 bit multiplication and addition and
	// shift the results into the right place
	x0, x1 := i.lo&0x00000000ffffffff, i.lo>>32
	y0, y1 := n.lo&0x00000000ffffffff, n.lo>>32
	t := x1*y0 + (x0*y0)>>32
	w1 := (t & 0x00000000ffffffff) + (x0 * y1)
	dest.hi = int64((x1 * y1) + (t >> 32) + (w1 >> 32) + hl)

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

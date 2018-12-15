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

var bigLoMask = new(big.Int).SetUint64(maxUint64)

func I128FromBigInt(v *big.Int) (out I128) {
	var a, b big.Int
	out.lo = b.And(v, bigLoMask).Uint64()
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

// func RandI128From(rand *rand.Rand) (out I128) {
//     return I128{hi: rand.Int63(), lo: rand.Uint64()}
// }

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
	// FIXME: build big.Float directly
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

func (i I128) DivMod(by I128) (q, r I128) {
	qSign, rSign := 1, 1
	if i.LessThan(zeroI128) {
		qSign, rSign = -1, -1
		i = i.Neg()
	}
	if by.LessThan(zeroI128) {
		qSign = -qSign
		by = by.Neg()
	}

	qu, ru := i.AsU128().DivMod(by.AsU128())
	q, r = qu.AsI128(), ru.AsI128()
	if qSign < 0 {
		q = q.Neg()
	}
	if rSign < 0 {
		r = r.Neg()
	}
	return q, r
}

func (i I128) Mod(by I128) (r I128) {
	// FIXME: can do much better than this.
	_, r = i.DivMod(by)
	return r
}

func (i I128) Div(by I128) (q I128) {
	// FIXME: can do much better than this.
	q, _ = i.DivMod(by)
	return q
}

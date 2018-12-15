package num

import (
	"math"
	"math/big"
)

const maxUint64Float = float64(math.MaxUint64) + 1

type U128 struct {
	lo, hi uint64
}

func U128From64(v uint64) U128 { return U128{hi: 0, lo: v} }
func U128From32(v uint32) U128 { return U128{hi: 0, lo: uint64(v)} }
func U128From16(v uint16) U128 { return U128{hi: 0, lo: uint64(v)} }
func U128From8(v uint8) U128   { return U128{hi: 0, lo: uint64(v)} }

func U128FromBigInt(v *big.Int) (out U128) {
	if v.Sign() < 0 {
		return out
	}

	tmp := new(big.Int).Set(v)
	tmp.Rsh(tmp, 64)
	out.lo = v.Uint64()
	out.hi = tmp.Uint64()
	return out
}

func U128FromFloat32(f float32) U128 { return U128FromFloat64(float64(f)) }

func U128FromFloat64(f float64) U128 {
	if f <= 0 {
		return U128{}
	} else if f <= math.MaxUint64 {
		return U128{lo: uint64(f)}
	} else {
		// FIXME: dunno about this.
		return U128{hi: uint64(f / maxUint64Float), lo: uint64(f)}
	}
}

func (u U128) String() string {
	v := u.AsBigInt() // This is good enough for now
	return v.String()
}

func (u U128) IntoBigInt(b *big.Int) {
	b.SetUint64(u.hi)
	b.Lsh(b, 64)

	var lo big.Int
	lo.SetUint64(u.lo)
	b.Add(b, &lo)
}

func (u U128) AsBigInt() (b big.Int) {
	u.IntoBigInt(&b)
	return b
}

func (u U128) AsFloat64() float64 {
	if u.hi == 0 && u.lo == 0 {
		return 0
	} else if u.hi == 0 {
		return float64(u.lo)
	} else {
		return (float64(u.hi) * maxUint64Float) + float64(u.lo)
	}
}

func (u U128) Inc() (v U128) {
	v.lo = u.lo + 1
	v.hi = u.hi + ((v.lo^u.lo)&v.lo)>>63
	return v
}

func (u U128) Dec() (v U128) {
	v.lo = u.lo - 1
	v.hi = u.hi - ((v.lo^u.lo)&v.lo)>>63
	return v
}

func (u U128) Add(n U128) U128 {
	lo := u.lo + n.lo
	hi := u.hi + n.hi
	if u.lo > lo {
		hi++
	}
	return U128{hi: hi, lo: lo}
}

func (u U128) Sub(n U128) (v U128) {
	v.lo = u.lo - n.lo
	v.hi = u.hi - n.hi
	if u.lo < v.lo {
		v.hi--
	}
	return v
}

func (u U128) Cmp(n U128) int {
	if u.hi > n.hi {
		return 1
	} else if u.hi < n.hi {
		return -1
	} else if u.lo > n.lo {
		return 1
	} else if u.lo < n.lo {
		return -1
	}
	return 0
}

func (u U128) Equal(n U128) bool {
	return u.hi == n.hi && u.lo == n.lo
}

func (u U128) Gt(n U128) bool {
	if u.hi > n.hi {
		return true
	} else if u.hi < n.hi {
		return false
	} else if u.lo > n.lo {
		return true
	} else if u.lo < n.lo {
		return false
	}
	return false
}

func (u U128) Gte(n U128) bool {
	if u.hi > n.hi {
		return true
	} else if u.hi < n.hi {
		return false
	} else if u.lo > n.lo {
		return true
	} else if u.lo < n.lo {
		return false
	}
	return true
}

func (u U128) Lt(n U128) bool {
	if u.hi > n.hi {
		return false
	} else if u.hi < n.hi {
		return true
	} else if u.lo > n.lo {
		return false
	} else if u.lo < n.lo {
		return true
	}
	return false
}

func (u U128) Lte(n U128) bool {
	if u.hi > n.hi {
		return false
	} else if u.hi < n.hi {
		return true
	} else if u.lo > n.lo {
		return false
	} else if u.lo < n.lo {
		return true
	}
	return true
}

func (u U128) And(v U128) (out U128) {
	out.hi = u.hi & v.hi
	out.lo = u.lo & v.lo
	return out
}

func (u U128) Or(v U128) (out U128) {
	out.hi = u.hi | v.hi
	out.lo = u.lo | v.lo
	return out
}

func (u U128) Xor(v U128) (out U128) {
	out.hi = u.hi ^ v.hi
	out.lo = u.lo ^ v.lo
	return out
}

func (u U128) Lsh(n uint) (v U128) {
	if n >= 128 {
		return v
	} else if n >= 64 {
		v.hi = u.lo << (n - 64)
		v.lo = 0
		return v
	} else {
		v.hi = u.hi<<n | u.lo>>(64-n)
		v.lo = u.lo << n
		return v
	}
}

func (u U128) Rsh(n uint) (v U128) {
	if n >= 128 {
		return v
	} else if n >= 64 {
		v.hi = 0
		v.lo = u.hi >> (n - 64)
		return v
	} else {
		v.hi = u.hi >> n
		v.lo = u.lo>>n | u.hi<<(64-n)
		return v
	}
}

func (u U128) Mul(n U128) (dest U128) {
	// Adapted from Warren, Hacker's Delight, p. 132.
	hl := u.hi*n.lo + u.lo*n.hi

	dest.lo = u.lo * n.lo // lower 64 bits are easy

	// break the multiplication into (x1 << 32 + x0)(y1 << 32 + y0)
	// which is x1*y1 << 64 + (x0*y1 + x1*y0) << 32 + x0*y0
	// so now we can do 64 bit multiplication and addition and
	// shift the results into the right place
	x0, x1 := u.lo&0x00000000ffffffff, u.lo>>32
	y0, y1 := n.lo&0x00000000ffffffff, n.lo>>32
	t := x1*y0 + (x0*y0)>>32
	w1 := (t & 0x00000000ffffffff) + (x0 * y1)
	dest.hi = (x1 * y1) + (t >> 32) + (w1 >> 32) + hl

	return dest
}

func (u U128) Div(by U128) (q U128) {
	// FIXME: can do much better than this.
	q, _ = u.DivMod(by)
	return q
}

func (u U128) DivMod(by U128) (q, r U128) {
	if by.lo == 0 && by.hi == 0 {
		panic("u128: division by zero")
	}

	var (
		uLeading0   = leadingZeros128(u)
		byLeading0  = leadingZeros128(by)
		byTrailing0 = trailingZeros128(by)
	)

	if u.hi|by.hi == 0 {
		q.lo = u.lo / by.lo // FIXME: div/0 risk?
		r.lo = u.lo % by.lo
		return q, r

	} else if byLeading0 == 127 {
		return u, r

	} else if (byLeading0 + byTrailing0) == 127 {
		q = u.Rsh(byTrailing0)
		by = by.Dec()
		r = by.And(u)
		return
	}

	if cmp := u.Cmp(by); cmp < 0 {
		return q, u // it's 100% remainder

	} else if cmp == 0 {
		q.lo = 1
		return q, r
	}

	// The original author of this method claims choosing to spill at 5 was
	// the result of a benchmark, but that's in a C context. This should be
	// benchmarked as Go and tuned:
	if byLeading0-uLeading0 > 5 {
		return divmod128by128(u, by)
	} else {
		return divmod128bin(u, by)
	}
}

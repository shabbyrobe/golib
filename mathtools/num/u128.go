package num

import (
	"fmt"
	"math/big"
	"math/bits"
)

type U128 struct {
	lo, hi uint64
}

var MaxU128 = U128{hi: maxUint64, lo: maxUint64}

func U128FromRaw(hi, lo uint64) U128 { return U128{hi: hi, lo: lo} }
func U128From64(v uint64) U128       { return U128{hi: 0, lo: v} }
func U128From32(v uint32) U128       { return U128{hi: 0, lo: uint64(v)} }
func U128From16(v uint16) U128       { return U128{hi: 0, lo: uint64(v)} }
func U128From8(v uint8) U128         { return U128{hi: 0, lo: uint64(v)} }

func U128FromString(s string) (out U128, err error) {
	b, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return out, fmt.Errorf("num: u128 string %q invalid", s)
	}
	return U128FromBigInt(b), nil
}

func U128FromBigInt(v *big.Int) (out U128) {
	if v.Sign() < 0 {
		return out
	}

	var a, b big.Int
	out.lo = b.And(v, maxBigUint64).Uint64()
	out.hi = a.Rsh(v, 64).Uint64()
	return out
}

func U128FromFloat32(f float32) U128 { return U128FromFloat64(float64(f)) }

func U128FromFloat64(f float64) U128 {
	if f <= 0 {
		return U128{}
	} else if f <= maxUint64Float {
		return U128{lo: uint64(f)}
	} else {
		return U128{hi: uint64(f / maxUint64Float), lo: uint64(f)}
	}
}

// func RandU128From(rand *rand.Rand) (out U128) {
//     return U128{hi: rand.Uint64(), lo: rand.Uint64()}
// }

func (u U128) Raw() (hi, lo uint64) { return u.hi, u.lo }

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

func (u U128) AsBigInt() (b *big.Int) {
	b = new(big.Int).SetUint64(u.hi)
	b.Lsh(b, 64)

	var lo big.Int
	lo.SetUint64(u.lo)
	b.Add(b, &lo)
	return b
}

func (u U128) AsBigFloat() (b big.Float) {
	i := u.AsBigInt()
	b.SetInt(i)
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

func (u U128) AsI128() I128 {
	return I128{lo: u.lo, hi: int64(u.hi)}
}

// AsUint64 truncates the U128 to fit in a uint64.
func (u U128) AsUint64() uint64 {
	return u.lo
}

func (u U128) Inc() (v U128) {
	v.lo = u.lo + 1
	v.hi = u.hi
	if u.lo > v.lo {
		v.hi++
	}
	return v
}

func (u U128) Dec() (v U128) {
	v.lo = u.lo - 1
	v.hi = u.hi
	if u.lo < v.lo {
		v.hi--
	}
	return v
}

func (u U128) Add(n U128) (v U128) {
	v.lo = u.lo + n.lo
	v.hi = u.hi + n.hi
	if u.lo > v.lo {
		v.hi++
	}
	return v
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

func (u U128) GreaterThan(n U128) bool {
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

func (u U128) GreaterOrEqualTo(n U128) bool {
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

func (u U128) LessThan(n U128) bool {
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

func (u U128) LessOrEqualTo(n U128) bool {
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
	if n == 0 {
		return u
	} else if n > 64 {
		v.hi = u.lo << (n - 64)
		v.lo = 0
	} else if n < 64 {
		v.hi = (u.hi << n) | (u.lo >> (64 - n))
		v.lo = u.lo << n
	} else if n == 64 {
		v.hi = u.lo
		v.lo = 0
	}
	return v
}

func (u U128) Rsh(n uint) (v U128) {
	if n == 0 {
		return u
	} else if n > 64 {
		v.lo = u.hi >> (n - 64)
		v.hi = 0
	} else if n < 64 {
		v.lo = (u.lo >> n) | (u.hi << (64 - n))
		v.hi = u.hi >> n
	} else if n == 64 {
		v.lo = u.hi
		v.hi = 0
	}

	return v
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

// Quo returns the quotient x/y for y != 0. If y == 0, a division-by-zero
// run-time panic occurs. Quo implements truncated division (like Go); see
// QuoRem for more details.
func (u U128) Quo(by U128) (q U128) {
	if by.lo == 0 && by.hi == 0 {
		panic("u128: division by zero")
	}

	var (
		uLeading0   = u.leadingZeros()
		byLeading0  = by.leadingZeros()
		byTrailing0 = by.trailingZeros()
	)

	if u.hi|by.hi == 0 {
		q.lo = u.lo / by.lo // FIXME: div/0 risk?
		return q

	} else if byLeading0 == 127 {
		return u

	} else if (byLeading0 + byTrailing0) == 127 {
		return u.Rsh(byTrailing0)
	}

	if cmp := u.Cmp(by); cmp < 0 {
		return q

	} else if cmp == 0 {
		q.lo = 1
		return q
	}

	if byLeading0-uLeading0 > 5 {
		q, _ = quorem128by128(u, by)
		return q
	} else {
		return div128bin(u, by)
	}
}

// Rem returns the remainder of x%y for y != 0. If y == 0, a division-by-zero
// run-time panic occurs. Rem implements truncated modulus (like Go); see
// QuoRem for more details.
func (u U128) Rem(by U128) (r U128) {
	// FIXME: can do much better than this.
	_, r = u.QuoRem(by)
	return r
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
func (u U128) QuoRem(by U128) (q, r U128) {
	if by.lo == 0 && by.hi == 0 {
		panic("u128: division by zero")
	}

	var (
		uLeading0   = u.leadingZeros()
		byLeading0  = by.leadingZeros()
		byTrailing0 = by.trailingZeros()
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
		return quorem128by128(u, by)
	} else {
		return quorem128bin(u, by)
	}
}

func (u U128) leadingZeros() uint {
	if u.hi == 0 {
		return uint(bits.LeadingZeros64(u.lo)) + 64
	} else {
		return uint(bits.LeadingZeros64(u.hi))
	}
}

func (u U128) trailingZeros() uint {
	if u.lo == 0 {
		return uint(bits.TrailingZeros64(u.hi)) + 64
	} else {
		return uint(bits.TrailingZeros64(u.lo))
	}
}

// Hacker's delight 9-4, divlu:
func div128by64(u1, u0, v uint64) (q uint64) {
	var b uint64 = 1 << 32
	var un1, un0, vn1, vn0, q1, q0, un32, un21, un10, rhat, vs, left, right uint64

	s := uint(bits.LeadingZeros64(v))
	vs = v << s

	vn1 = vs >> 32
	vn0 = vs & 0xffffffff

	if s > 0 {
		un32 = (u1 << s) | (u0 >> (64 - s))
		un10 = u0 << s
	} else {
		un32 = u1
		un10 = u0
	}

	un1 = un10 >> 32
	un0 = un10 & 0xffffffff

	q1 = un32 / vn1
	rhat = un32 % vn1

	left = q1 * vn0
	right = (rhat << 32) | un1

again1:
	if (q1 >= b) || (left > right) {
		q1--
		rhat += vn1
		if rhat < b {
			left -= vn0
			right = (rhat << 32) | un1
			goto again1
		}
	}

	un21 = (un32 << 32) + (un1 - (q1 * vs))

	q0 = un21 / vn1
	rhat = un21 % vn1

	left = q0 * vn0
	right = (rhat << 32) | un0

again2:
	if (q0 >= b) || (left > right) {
		q0--
		rhat += vn1
		if rhat < b {
			left -= vn0
			right = (rhat << 32) | un0
			goto again2
		}
	}

	return (q1 << 32) | q0
}

// Hacker's delight 9-4, divlu:
func quorem128by64(u1, u0, v uint64) (q, r uint64) {
	var b uint64 = 1 << 32
	var un1, un0, vn1, vn0, q1, q0, un32, un21, un10, rhat, left, right uint64

	s := uint(bits.LeadingZeros64(v))
	v <<= s

	vn1 = v >> 32
	vn0 = v & 0xffffffff

	if s > 0 {
		un32 = (u1 << s) | (u0 >> (64 - s))
		un10 = u0 << s
	} else {
		un32 = u1
		un10 = u0
	}

	un1 = un10 >> 32
	un0 = un10 & 0xffffffff

	q1 = un32 / vn1
	rhat = un32 % vn1

	left = q1 * vn0
	right = (rhat << 32) + un1

again1:
	if (q1 >= b) || (left > right) {
		q1--
		rhat += vn1
		if rhat < b {
			left -= vn0
			right = (rhat << 32) | un1
			goto again1
		}
	}

	un21 = (un32 << 32) + (un1 - (q1 * v))

	q0 = un21 / vn1
	rhat = un21 % vn1

	left = q0 * vn0
	right = (rhat << 32) | un0

again2:
	if (q0 >= b) || (left > right) {
		q0--
		rhat += vn1
		if rhat < b {
			left -= vn0
			right = (rhat << 32) | un0
			goto again2
		}
	}

	return (q1 << 32) | q0, ((un21 << 32) + (un0 - (q0 * v))) >> s
}

func quorem128by128(m, v U128) (q, r U128) {
	if v.hi == 0 {
		if m.hi < v.lo {
			q.lo, r.lo = quorem128by64(m.hi, m.lo, v.lo)
			return q, r

		} else {
			q.hi = m.hi / v.lo
			r.hi = m.hi % v.lo
			q.lo, r.lo = quorem128by64(r.hi, m.lo, v.lo)
			r.hi = 0
			return q, r
		}

	} else {
		sh := uint(bits.LeadingZeros64(v.hi))

		v1 := v.Lsh(sh)
		u1 := m.Rsh(1)

		var q1 U128
		q1.lo = div128by64(u1.hi, u1.lo, v1.hi)
		q1 = q1.Rsh(63 - sh)

		if q1.hi|q1.lo != 0 {
			q1 = q1.Dec()
		}
		q = q1
		q1 = q1.Mul(v)
		r = m.Sub(q1)

		if r.Cmp(v) >= 0 {
			q = q.Inc()
			r = r.Sub(v)
		}

		return q, r
	}
}

func quorem128bin(u, by U128) (q, r U128) {
	shift := int(by.leadingZeros() - u.leadingZeros())
	by = by.Lsh(uint(shift))

	for {
		q = q.Lsh(1)
		if u.Cmp(by) >= 0 {
			u = u.Sub(by)
			q.lo |= 1
		}

		by = by.Rsh(1)

		if shift <= 0 {
			break
		}
		shift--
	}

	r = u
	return q, r
}

func div128bin(u, by U128) (q U128) {
	shift := int(by.leadingZeros() - u.leadingZeros())
	by = by.Lsh(uint(shift))

	for {
		q = q.Lsh(1)
		if u.Cmp(by) >= 0 {
			u = u.Sub(by)
			q.lo |= 1
		}

		by = by.Rsh(1)

		if shift <= 0 {
			break
		}
		shift--
	}

	return q
}

func (u U128) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

func (u *U128) UnmarshalText(bts []byte) (err error) {
	v, err := U128FromString(string(bts))
	if err != nil {
		return err
	}
	*u = v
	return nil
}

func (u U128) MarshalJSON() ([]byte, error) {
	return []byte(`"` + u.String() + `"`), nil
}

func (u *U128) UnmarshalJSON(bts []byte) (err error) {
	if bts[0] == '"' {
		ln := len(bts)
		if bts[ln-1] != '"' {
			return fmt.Errorf("num: u128 invalid JSON %q", string(bts))
		}
		bts = bts[1 : ln-1]
	}

	v, err := U128FromString(string(bts))
	if err != nil {
		return err
	}
	*u = v
	return nil
}

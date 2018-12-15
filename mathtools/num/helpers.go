package num

import "math/bits"

func leadingZeros128(u U128) uint {
	if u.hi == 0 {
		return uint(bits.LeadingZeros64(u.lo)) + 64
	} else {
		return uint(bits.LeadingZeros64(u.hi))
	}
}

func trailingZeros128(u U128) uint {
	if u.lo == 0 {
		return uint(bits.TrailingZeros64(u.hi)) + 64
	} else {
		return uint(bits.TrailingZeros64(u.lo))
	}
}

func mul(x, y uint64) (z1, z0 uint64) {
	z0 = x * y // lower 64 bits are easy
	// break the multiplication into (x1 << 32 + x0)(y1 << 32 + y0)
	// which is x1*y1 << 64 + (x0*y1 + x1*y0) << 32 + x0*y0
	// so now we can do 64 bit multiplication and addition and
	// shift the results into the right place
	x0, x1 := x&0x00000000ffffffff, x>>32
	y0, y1 := y&0x00000000ffffffff, y>>32
	w0 := x0 * y0
	t := x1*y0 + w0>>32
	w1 := t & 0x00000000ffffffff
	w2 := t >> 32
	w1 += x0 * y1
	z1 = x1*y1 + w2 + w1>>32
	return
}

// Hacker's delight 9-4, divlu:
func divmod128by64(u1, u0, v uint64) (q, r uint64) {
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

func divmod128by128(m, v U128) (q, r U128) {
	if v.hi == 0 {
		if m.hi < v.lo {
			q.lo, r.lo = divmod128by64(m.hi, m.lo, v.lo)
			return q, r

		} else {
			q.hi = m.hi / v.lo
			r.hi = m.hi % v.lo
			q.lo, r.lo = divmod128by64(r.hi, m.lo, v.lo)
			r.hi = 0
			return q, r
		}

	} else {
		sh := uint(bits.LeadingZeros64(v.hi))

		v1 := v.Lsh(sh)
		u1 := m.Rsh(1)

		var q1 U128
		_, q1.lo = divmod128by64(u1.hi, u1.lo, v1.hi)
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

		return
	}
}

func divmod128bin(u, by U128) (q, r U128) {
	sz := leadingZeros128(by) - leadingZeros128(u)
	by = by.Lsh(sz)

	for {
		q = q.Lsh(1)
		if u.Cmp(by) >= 0 {
			u = u.Sub(by)
			q.lo |= 1
		}

		by = by.Rsh(1)

		sz--
		if sz == 0 { // Careful: sz is unsigned.
			break
		}
	}

	r = u
	return q, r
}

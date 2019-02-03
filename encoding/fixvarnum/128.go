package fixvarnum

import (
	"errors"

	num "github.com/shabbyrobe/go-num"
)

const (
	// MaxLen128 reports the largest buffer that PutU128 or PutI128 will ever
	// create. This consists of 4 bits encoding the number of trailing decimal
	// zeros, then 128 bits of data. 132 bits broken into 7 bit chunks == 19
	// bytes.
	MaxLen128 = 19
)

var zeroU128 num.U128

// PutU128 encodes a variable-length num.U128 into buf and returns the number
// of bytes written. If the buffer is too small, PutUvarint will panic.
func PutU128(buf []byte, x num.U128) int {
	var xhi, xlo = x.Raw()
	var zeros byte

	hasTen := x.Rem(zumul[1]).Equal(zeroU128)
	if x != zeroU128 && hasTen {
		if !x.Rem(zumul[8]).Equal(zeroU128) { // <8
			if !x.Rem(zumul[4]).Equal(zeroU128) { // <4
				if !x.Rem(zumul[2]).Equal(zeroU128) { // <2
					if !hasTen {
						// all good
					} else { // == 1
						zeros, x = 1, x.Quo(zumul[1])
					}
				} else { // >=2, <4
					if !x.Rem(zumul[3]).Equal(zeroU128) { // == 2
						zeros, x = 2, x.Quo(zumul[2])
					} else { // == 3
						zeros, x = 3, x.Quo(zumul[3])
					}
				}

			} else { // >=4, <8
				if !x.Rem(zumul[6]).Equal(zeroU128) { // >=4, <6
					if !x.Rem(zumul[5]).Equal(zeroU128) { // == 4
						zeros, x = 4, x.Quo(zumul[4])
					} else { // == 5
						zeros, x = 5, x.Quo(zumul[5])
					}
				} else { // >= 6, <8
					if !x.Rem(zumul[7]).Equal(zeroU128) { // == 6
						zeros, x = 6, x.Quo(zumul[6])
					} else { // == 7
						zeros, x = 7, x.Quo(zumul[7])
					}
				}
			}

		} else { // >= 8, <16
			if !x.Rem(zumul[12]).Equal(zeroU128) { // >= 8, <12
				if !x.Rem(zumul[10]).Equal(zeroU128) { // >= 8, <10
					if !x.Rem(zumul[9]).Equal(zeroU128) { // == 8
						zeros, x = 8, x.Quo(zumul[8])
					} else { // == 9
						zeros, x = 9, x.Quo(zumul[9])
					}
				} else { // >=10, <12
					if !x.Rem(zumul[11]).Equal(zeroU128) { // == 10
						zeros, x = 10, x.Quo(zumul[10])
					} else { // == 11
						zeros, x = 11, x.Quo(zumul[11])
					}
				}

			} else { // >=12, <16
				if !x.Rem(zumul[14]).Equal(zeroU128) { // >=12, <14
					if !x.Rem(zumul[13]).Equal(zeroU128) { // == 12
						zeros, x = 12, x.Quo(zumul[12])
					} else { // == 13
						zeros, x = 13, x.Quo(zumul[13])
					}
				} else { // >= 14, <16
					if !x.Rem(zumul[15]).Equal(zeroU128) { // == 14
						zeros, x = 14, x.Quo(zumul[14])
					} else { // == 15
						zeros, x = 15, x.Quo(zumul[15])
					}
				}
			}
		}

		xhi, xlo = x.Raw()
	}

	i := 0

	xv := xlo

	var cont byte
	if xv >= 0x8 || xhi > 0 {
		cont = 0x80
	}

	buf[0] = cont | (zeros << 3) | byte(xv&0x7)
	xv >>= 3
	i++

	for bits := x.BitLen() - 3; bits > 0; bits -= 7 {
		if i == 9 {
			// If we have written 9 bytes, we have shifted 59 bits off xlo (3
			// initial bits + 8x7 bits).
			//
			// This means the composition of the join byte (the byte that crosses the gap
			// between hi and lo) is like so (where x is the continuation bit):
			//   x H H L L L L L
			xv = (xhi << 5) | (xlo >> 59)

			if bits <= 7 {
				buf[i], i = byte(xv), i+1
			} else {
				buf[i], i = byte(xv)|0x80, i+1
			}

			// The previous byte uses 2 bits of xhi, so replace our bit queue
			// with the remaining 62:
			xv = xhi >> 2

		} else {
			if bits <= 7 {
				buf[i], i = byte(xv), i+1
			} else {
				buf[i], i = byte(xv)|0x80, i+1
			}
			xv >>= 7
		}
	}

	return i
}

// U128 decodes a num.U128 from buf and returns that value and the
// number of bytes read (> 0). If an error occurred, the value is 0
// and the number of bytes n is <= 0 meaning:
//
// 	n == 0: buf too small
// 	n  < 0: value larger than 128 bits (overflow)
// 	        and -n is the number of bytes read
//
func U128(buf []byte) (out num.U128, n int) {
	var shift uint
	var lo, hi uint64
	var b byte
	var i int

	lim := len(buf)
	iter := lim

	zeros := (buf[0] >> 3) & 0xF
	shift = 3
	lo = uint64(buf[0] & 0x7)

	n = 1
	if buf[0] < 0x80 {
		goto done
	}

	if iter > 9 {
		iter = 9
	}

	for i = 1; i < iter; i++ {
		b = buf[i]

		if b < 0x80 {
			lo, n = lo|uint64(b)<<shift, i+1 // +1 to convert from 0-index
			goto done
		}
		lo |= uint64(b&0x7f) << shift
		shift += 7
	}

	{ // i == 9
		b = buf[9]
		// if we have read 9 bytes, we have accumulated 59 bits of the lo number.
		// after the continuation bit, the high 2 bits of the current byte belong
		// to the hi number of the U128, and the low 5 belong to the lo:
		//   x H H L L L L L
		if b < 0x80 {
			lo, hi, n = lo|(uint64(b)<<shift), uint64(b>>5), i+1 // +1 to convert from 0-index
			goto done
		}
		lo, hi = lo|(uint64(b&0x7f)<<shift), (uint64(b&0x7f) >> 5)
		shift = 2
	}

	for i = 10; i < lim; i++ {
		b = buf[i]
		if i > MaxLen128 || (i == MaxLen128 && b > 1) {
			return out, -(i + 1) // overflow
		}

		if b < 0x80 {
			hi, n = hi|uint64(b)<<shift, i+1 // +1 to convert from 0-index
			goto done
		}
		hi |= uint64(b&0x7f) << shift
		shift += 7
	}

done:
	if zeros == 0 {
		return num.U128FromRaw(hi, lo), n
	}

	zmlo := zumul64[zeros]

	hl := hi * zmlo
	olo := lo * zmlo // Subsequent calculations must use the original value

	// break the multiplication into (x1 << 32 + x0)(y1 << 32 + y0)
	// which is x1*y1 << 64 + (x0*y1 + x1*y0) << 32 + x0*y0
	// so now we can do 64 bit multiplication and addition and
	// shift the results into the right place
	x0, x1 := lo&0x00000000ffffffff, lo>>32
	y0, y1 := zmlo&0x00000000ffffffff, zmlo>>32
	t := x1*y0 + (x0*y0)>>32
	w1 := (t & 0x00000000ffffffff) + (x0 * y1)
	hi = (x1 * y1) + (t >> 32) + (w1 >> 32) + hl

	return num.U128FromRaw(hi, olo), n
}

/*
// UvarintTurbo is an experimental replacement for Uvarint. It's substantially
// faster for larger input at the cost of being truly disgusting to read.
func UvarintTurbo(buf []byte) (num.U128, int) {
	var x uint64

	zeros := (buf[0] >> 3) & 0xF
	x = uint64(buf[0] & 0x7)

	n := 1
	if buf[0] < 0x80 {
		goto done
	}

	if buf[1] < 0x80 {
		x, n = x|uint64(buf[1])<<3, 2
		goto done
	}

	if buf[2] < 0x80 {
		x = x |
			(uint64(buf[1]&0x7f) << 3) |
			(uint64(buf[2]) << 10)
		n = 3
		goto done
	}

	if buf[3] < 0x80 {
		x = x |
			(uint64(buf[1]&0x7f) << 3) |
			(uint64(buf[2]&0x7f) << 10) |
			(uint64(buf[3]) << 17)
		n = 4
		goto done
	}

	if buf[4] < 0x80 {
		x = x |
			(uint64(buf[1]&0x7f) << 3) |
			(uint64(buf[2]&0x7f) << 10) |
			(uint64(buf[3]&0x7f) << 17) |
			(uint64(buf[4]) << 24)
		n = 5
		goto done
	}

	if buf[5] < 0x80 {
		x = x |
			(uint64(buf[1]&0x7f) << 3) |
			(uint64(buf[2]&0x7f) << 10) |
			(uint64(buf[3]&0x7f) << 17) |
			(uint64(buf[4]&0x7f) << 24) |
			(uint64(buf[5]) << 31)
		n = 6
		goto done
	}

	if buf[6] < 0x80 {
		x = x | (uint64(buf[1]&0x7f) << 3) |
			(uint64(buf[2]&0x7f) << 10) |
			(uint64(buf[3]&0x7f) << 17) |
			(uint64(buf[4]&0x7f) << 24) |
			(uint64(buf[5]&0x7f) << 31) |
			(uint64(buf[6]) << 38)
		n = 7
		goto done
	}

	if buf[7] < 0x80 {
		x = x | (uint64(buf[1]&0x7f) << 3) |
			(uint64(buf[2]&0x7f) << 10) |
			(uint64(buf[3]&0x7f) << 17) |
			(uint64(buf[4]&0x7f) << 24) |
			(uint64(buf[5]&0x7f) << 31) |
			(uint64(buf[6]&0x7f) << 38) |
			(uint64(buf[7]) << 45)
		n = 8
		goto done
	}

	if buf[8] < 0x80 {
		x = x |
			(uint64(buf[1]&0x7f) << 3) |
			(uint64(buf[2]&0x7f) << 10) |
			(uint64(buf[3]&0x7f) << 17) |
			(uint64(buf[4]&0x7f) << 24) |
			(uint64(buf[5]&0x7f) << 31) |
			(uint64(buf[6]&0x7f) << 38) |
			(uint64(buf[7]&0x7f) << 45) |
			(uint64(buf[8]) << 52)
		n = 9
		goto done
	}

	if buf[9] < 0x80 {
		x = x |
			(uint64(buf[1]&0x7f) << 3) |
			(uint64(buf[2]&0x7f) << 10) |
			(uint64(buf[3]&0x7f) << 17) |
			(uint64(buf[4]&0x7f) << 24) |
			(uint64(buf[5]&0x7f) << 31) |
			(uint64(buf[6]&0x7f) << 38) |
			(uint64(buf[7]&0x7f) << 45) |
			(uint64(buf[8]&0x7f) << 52) |
			(uint64(buf[9]) << 59)
		n = 10
		goto done
	}

	return 0, -11

done:
	if zeros > 0 {
		x *= zumul[zeros]
	}

	return x, n
}

// PutVarint encodes an int64 into buf and returns the number of bytes written.
// If the buffer is too small, PutVarint will panic.
func PutVarint(buf []byte, x num.I128) int {
	var zeros byte

	if x != 0 && x%10 == 0 {
		if x%1e8 != 0 { // <8
			if x%1e4 != 0 { // <4
				if x%1e2 != 0 { // <2
					if x%1e1 != 0 { // == 0
						// all good
					} else { // == 1
						zeros, x = 1, x/1e1
					}
				} else { // >=2, <4
					if x%1e3 != 0 { // == 2
						zeros, x = 2, x/100
					} else { // == 3
						zeros, x = 3, x/1000
					}
				}

			} else { // >=4, <8
				if x%1e6 != 0 { // >=4, <6
					if x%1e5 != 0 { // == 4
						zeros, x = 4, x/1e4
					} else { // == 5
						zeros, x = 5, x/1e5
					}
				} else { // >= 6, <8
					if x%1e7 != 0 { // == 6
						zeros, x = 6, x/1e6
					} else { // == 7
						zeros, x = 7, x/1e7
					}
				}
			}

		} else { // >= 8, <16
			if x%1e12 != 0 { // >= 8, <12
				if x%1e10 != 0 { // >= 8, <10
					if x%1e9 != 0 { // == 8
						zeros, x = 8, x/1e8
					} else { // == 9
						zeros, x = 9, x/1e9
					}
				} else { // >=10, <12
					if x%1e11 != 0 { // == 10
						zeros, x = 10, x/1e10
					} else { // == 11
						zeros, x = 11, x/1e11
					}
				}

			} else { // >=12, <16
				if x%1e14 != 0 { // >=12, <14
					if x%1e13 != 0 { // == 12
						zeros, x = 12, x/1e12
					} else { // == 13
						zeros, x = 13, x/1e13
					}
				} else { // >= 14, <16
					if x%1e15 != 0 { // == 14
						zeros, x = 14, x/1e14
					} else { // == 15
						zeros, x = 15, x/1e15
					}
				}
			}
		}
	}

	ux := uint64(x) << 1
	if x < 0 {
		ux = ^ux
	}

	i := 0

	var cont byte
	if ux >= 0x8 {
		cont = 0x80
	}

	buf[i] = cont | (zeros << 3) | byte(ux&0x7)
	ux >>= 3
	if ux == 0 {
		return i + 1
	}
	i++

	for ux >= 0x80 {
		buf[i] = byte(ux) | 0x80
		ux >>= 7
		i++
	}
	buf[i] = byte(ux)
	return i + 1
}

// Varint decodes an int64 from buf and returns that value and the
// number of bytes read (> 0). If an error occurred, the value is 0
// and the number of bytes n is <= 0 with the following meaning:
//
// 	n == 0: buf too small
// 	n  < 0: value larger than 64 bits (overflow)
// 	        and -n is the number of bytes read
//
func Varint(buf []byte) (v num.I128, n int) {
	var ix int64
	var ux uint64
	var s uint = 3

	zeros := (buf[0] >> 3) & 0xF
	ux = uint64(buf[0] & 0x7)

	n = 1
	if buf[0] < 0x80 {
		goto done
	}

	for i, b := range buf[1:] {
		if b < 0x80 {
			if i > 9 || i == 9 && b > 1 {
				return 0, -(i + 1) // overflow
			}
			ux, n = ux|uint64(b)<<s, i+2 // +1 for the slice offset, +1 to convert from 0-index

			goto done
		}
		ux |= uint64(b&0x7f) << s
		s += 7
	}

done:
	ix = int64(ux >> 1)
	if ux&1 != 0 {
		ix = ^ix
	}

	if zeros > 0 {
		ix *= zmul[zeros]
	}

	return ix, n
}

// VarintTurbo is an experimental replacement for Varint. It's substantially
// faster for larger input at the cost of being truly disgusting to read.
func VarintTurbo(buf []byte) (num.I128, int) {
	var ix int64
	var ux uint64

	zeros := (buf[0] >> 3) & 0xF
	ux = uint64(buf[0] & 0x7)

	n := 1
	if buf[0] < 0x80 {
		goto done
	}

	if buf[1] < 0x80 {
		ux, n = ux|uint64(buf[1])<<3, 2
		goto done
	}

	if buf[2] < 0x80 {
		ux = ux |
			(uint64(buf[1]&0x7f) << 3) |
			(uint64(buf[2]) << 10)
		n = 3
		goto done
	}

	if buf[3] < 0x80 {
		ux = ux |
			(uint64(buf[1]&0x7f) << 3) |
			(uint64(buf[2]&0x7f) << 10) |
			(uint64(buf[3]) << 17)
		n = 4
		goto done
	}

	if buf[4] < 0x80 {
		ux = ux |
			(uint64(buf[1]&0x7f) << 3) |
			(uint64(buf[2]&0x7f) << 10) |
			(uint64(buf[3]&0x7f) << 17) |
			(uint64(buf[4]) << 24)
		n = 5
		goto done
	}

	if buf[5] < 0x80 {
		ux = ux |
			(uint64(buf[1]&0x7f) << 3) |
			(uint64(buf[2]&0x7f) << 10) |
			(uint64(buf[3]&0x7f) << 17) |
			(uint64(buf[4]&0x7f) << 24) |
			(uint64(buf[5]) << 31)
		n = 6
		goto done
	}

	if buf[6] < 0x80 {
		ux = ux | (uint64(buf[1]&0x7f) << 3) |
			(uint64(buf[2]&0x7f) << 10) |
			(uint64(buf[3]&0x7f) << 17) |
			(uint64(buf[4]&0x7f) << 24) |
			(uint64(buf[5]&0x7f) << 31) |
			(uint64(buf[6]) << 38)
		n = 7
		goto done
	}

	if buf[7] < 0x80 {
		ux = ux | (uint64(buf[1]&0x7f) << 3) |
			(uint64(buf[2]&0x7f) << 10) |
			(uint64(buf[3]&0x7f) << 17) |
			(uint64(buf[4]&0x7f) << 24) |
			(uint64(buf[5]&0x7f) << 31) |
			(uint64(buf[6]&0x7f) << 38) |
			(uint64(buf[7]) << 45)
		n = 8
		goto done
	}

	if buf[8] < 0x80 {
		ux = ux |
			(uint64(buf[1]&0x7f) << 3) |
			(uint64(buf[2]&0x7f) << 10) |
			(uint64(buf[3]&0x7f) << 17) |
			(uint64(buf[4]&0x7f) << 24) |
			(uint64(buf[5]&0x7f) << 31) |
			(uint64(buf[6]&0x7f) << 38) |
			(uint64(buf[7]&0x7f) << 45) |
			(uint64(buf[8]) << 52)
		n = 9
		goto done
	}

	if buf[9] < 0x80 {
		ux = ux |
			(uint64(buf[1]&0x7f) << 3) |
			(uint64(buf[2]&0x7f) << 10) |
			(uint64(buf[3]&0x7f) << 17) |
			(uint64(buf[4]&0x7f) << 24) |
			(uint64(buf[5]&0x7f) << 31) |
			(uint64(buf[6]&0x7f) << 38) |
			(uint64(buf[7]&0x7f) << 45) |
			(uint64(buf[8]&0x7f) << 52) |
			(uint64(buf[9]) << 59)
		n = 10
		goto done
	}

	return 0, -11

done:
	ix = int64(ux >> 1)
	if ux&1 != 0 {
		ix = ^ix
	}

	if zeros > 0 {
		ix *= zmul[zeros]
	}

	return ix, n
}

*/

var (
	zmul  = [...]int64{0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9, 1e10, 1e11, 1e12, 1e13, 1e14, 1e15}
	zumul = [...]num.U128{
		num.U128From64(0),
		num.U128From64(1e1),
		num.U128From64(1e2),
		num.U128From64(1e3),
		num.U128From64(1e4),
		num.U128From64(1e5),
		num.U128From64(1e6),
		num.U128From64(1e7),
		num.U128From64(1e8),
		num.U128From64(1e9),
		num.U128From64(1e10),
		num.U128From64(1e11),
		num.U128From64(1e12),
		num.U128From64(1e13),
		num.U128From64(1e14),
		num.U128From64(1e15),
	}

	zumul64 = [...]uint64{
		0,
		1e1,
		1e2,
		1e3,
		1e4,
		1e5,
		1e6,
		1e7,
		1e8,
		1e9,
		1e10,
		1e11,
		1e12,
		1e13,
		1e14,
		1e15,
	}

	overflow = errors.New("fixvarint: varint overflows a 64-bit integer")
)

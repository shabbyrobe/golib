package fixvarint

import (
	"errors"
)

const (
	// 4 bits encoding the number of trailing zeros, then 64 bits of data.
	// 68 bits broken into 7 bit chunks == 10 bytes.
	MaxLen64 = 10
)

// PutUvarint encodes a uint64 into buf and returns the number of bytes written.
// If the buffer is too small, PutUvarint will panic.
func PutUvarint(buf []byte, x uint64) int {
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
						zeros, x = 2, x/1e2
					} else { // == 3
						zeros, x = 3, x/1e3
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

	i := 0

	var cont byte
	if x >= 0x8 {
		cont = 0x80
	}

	buf[i] = cont | (zeros << 3) | byte(x&0x7)
	x >>= 3
	if x == 0 {
		return i + 1
	}
	i++

	for x >= 0x80 {
		buf[i] = byte(x) | 0x80
		x >>= 7
		i++
	}
	buf[i] = byte(x)
	return i + 1
}

// Uvarint decodes a uint64 from buf and returns that value and the
// number of bytes read (> 0). If an error occurred, the value is 0
// and the number of bytes n is <= 0 meaning:
//
// 	n == 0: buf too small
// 	n  < 0: (value larger than 64 bits (overflow) or no terminating byte)
// 	        and -n is the number of bytes read
//
func Uvarint(buf []byte) (uint64, int) {
	var x uint64
	var s uint = 3

	zeros := (buf[0] >> 3) & 0xF
	x = uint64(buf[0] & 0x7)

	n := 1
	if buf[0] < 0x80 {
		goto done
	}

	for i, b := range buf[1:] {
		if b < 0x80 {
			// this is a bit cryptic; the 8th index here is actually the 10th byte due
			// to the buf[1:]. if the last byte is greater than 0x1f, we have run out of
			// space in a 64-bit number to accomodate what's left.
			if i > 8 || i == 8 && b > 0x1f {
				return 0, -(i + 2) // overflow
			}
			x, n = x|uint64(b)<<s, i+2 // +1 for the slice offset, +1 to convert from 0-index

			goto done
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}

	// If we do not exit the loop early, we must fail as we never found a
	// terminating byte:
	n = len(buf)
	return 0, -n

done:
	if zeros > 0 {
		if x&zumask[zeros] == 0 {
			// Definitely doesn't overflow; fast path!
			x *= zumul[zeros]

		} else {
			a := x
			x *= zumul[zeros]

			// Use division by constants to encourage compiler to use a significantly
			// faster MUL instruction:
			switch zeros {
			case 1:
				if x/1e1 != a {
					return 0, -n
				}
			case 2:
				if x/1e2 != a {
					return 0, -n
				}
			case 3:
				if x/1e3 != a {
					return 0, -n
				}
			case 4:
				if x/1e4 != a {
					return 0, -n
				}
			case 5:
				if x/1e5 != a {
					return 0, -n
				}
			case 6:
				if x/1e6 != a {
					return 0, -n
				}
			case 7:
				if x/1e7 != a {
					return 0, -n
				}
			case 8:
				if x/1e8 != a {
					return 0, -n
				}
			case 9:
				if x/1e9 != a {
					return 0, -n
				}
			case 10:
				if x/1e10 != a {
					return 0, -n
				}
			case 11:
				if x/1e11 != a {
					return 0, -n
				}
			case 12:
				if x/1e12 != a {
					return 0, -n
				}
			case 13:
				if x/1e13 != a {
					return 0, -n
				}
			case 14:
				if x/1e14 != a {
					return 0, -n
				}
			case 15:
				if x/1e15 != a {
					return 0, -n
				}
			}
		}
	}
	return x, n
}

// UvarintTurbo is an experimental replacement for Uvarint. It's substantially
// faster for larger input at the cost of being truly disgusting to read.
func UvarintTurbo(buf []byte) (uint64, int) {
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
		if buf[9] > 0x1f {
			return 0, -10 // overflow
		}
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

	return 0, -10

done:
	if zeros > 0 {
		if x&zumask[zeros] == 0 {
			// Definitely doesn't overflow; fast path!
			x *= zumul[zeros]

		} else {
			a := x
			x *= zumul[zeros]

			// Use division by constants to encourage compiler to use a significantly
			// faster MUL instruction:
			switch zeros {
			case 1:
				if x/1e1 != a {
					return 0, -n
				}
			case 2:
				if x/1e2 != a {
					return 0, -n
				}
			case 3:
				if x/1e3 != a {
					return 0, -n
				}
			case 4:
				if x/1e4 != a {
					return 0, -n
				}
			case 5:
				if x/1e5 != a {
					return 0, -n
				}
			case 6:
				if x/1e6 != a {
					return 0, -n
				}
			case 7:
				if x/1e7 != a {
					return 0, -n
				}
			case 8:
				if x/1e8 != a {
					return 0, -n
				}
			case 9:
				if x/1e9 != a {
					return 0, -n
				}
			case 10:
				if x/1e10 != a {
					return 0, -n
				}
			case 11:
				if x/1e11 != a {
					return 0, -n
				}
			case 12:
				if x/1e12 != a {
					return 0, -n
				}
			case 13:
				if x/1e13 != a {
					return 0, -n
				}
			case 14:
				if x/1e14 != a {
					return 0, -n
				}
			case 15:
				if x/1e15 != a {
					return 0, -n
				}
			}
		}
	}

	return x, n
}

// PutVarint encodes an int64 into buf and returns the number of bytes written.
// If the buffer is too small, PutVarint will panic.
func PutVarint(buf []byte, x int64) int {
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

	// Shifting the entire number one to the left creates space for a sign bit as
	// the LSB, rather than the MSB. If the number is negative, we invert it so
	// the leading run of 1-bits becomes a leading run of 0-bits. The ^ operation
	// will flip our new LSB sign-bit to 'on' as well.
	ux := uint64(x) << 1
	if x < 0 {
		ux = ^ux
	}

	i := 0

	var cont byte
	if ux > 0x7 {
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
func Varint(buf []byte) (v int64, n int) {
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
			// this is a bit cryptic; the 8th index here is actually the 10th byte due
			// to the buf[1:]. if the last byte is greater than 0x1f, we have run out of
			// space in a 64-bit number to accomodate what's left.
			if i > 8 || i == 8 && b > 0x1f {
				return 0, -(i + 2) // overflow
			}

			ux, n = ux|uint64(b)<<s, i+2 // +1 for the slice offset, +1 to convert from 0-index

			goto done
		}
		ux |= uint64(b&0x7f) << s
		s += 7
	}

	// If we do not exit the loop early, we must fail as we never found a
	// terminating byte:
	n = len(buf)
	return 0, -n

done:
	ix = int64(ux >> 1)
	if ux&1 != 0 {
		ix = ^ix
	}

	if zeros > 0 {
		if ix&zimask[zeros] == 0 {
			// Definitely doesn't overflow; fast path!
			ix *= zimul[zeros]

		} else {
			a := ix
			ix *= zimul[zeros]

			// Use division by constants to encourage compiler to use a significantly
			// faster MUL instruction:
			switch zeros {
			case 1:
				if ix/1e1 != a {
					return 0, -n
				}
			case 2:
				if ix/1e2 != a {
					return 0, -n
				}
			case 3:
				if ix/1e3 != a {
					return 0, -n
				}
			case 4:
				if ix/1e4 != a {
					return 0, -n
				}
			case 5:
				if ix/1e5 != a {
					return 0, -n
				}
			case 6:
				if ix/1e6 != a {
					return 0, -n
				}
			case 7:
				if ix/1e7 != a {
					return 0, -n
				}
			case 8:
				if ix/1e8 != a {
					return 0, -n
				}
			case 9:
				if ix/1e9 != a {
					return 0, -n
				}
			case 10:
				if ix/1e10 != a {
					return 0, -n
				}
			case 11:
				if ix/1e11 != a {
					return 0, -n
				}
			case 12:
				if ix/1e12 != a {
					return 0, -n
				}
			case 13:
				if ix/1e13 != a {
					return 0, -n
				}
			case 14:
				if ix/1e14 != a {
					return 0, -n
				}
			case 15:
				if ix/1e15 != a {
					return 0, -n
				}
			}
		}
	}

	return ix, n
}

// VarintTurbo is an experimental replacement for Varint. It's substantially
// faster for larger input at the cost of being truly disgusting to read.
func VarintTurbo(buf []byte) (int64, int) {
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
		if buf[9] > 0x1f {
			return 0, -10 // overflow
		}
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

	return 0, -10

done:
	ix = int64(ux >> 1)
	if ux&1 != 0 {
		ix = ^ix
	}

	if zeros > 0 {
		if ix&zimask[zeros] == 0 {
			// Definitely doesn't overflow; fast path!
			ix *= zimul[zeros]

		} else {
			a := ix
			ix *= zimul[zeros]

			// Use division by constants to encourage compiler to use a significantly
			// faster MUL instruction:
			switch zeros {
			case 1:
				if ix/1e1 != a {
					return 0, -n
				}
			case 2:
				if ix/1e2 != a {
					return 0, -n
				}
			case 3:
				if ix/1e3 != a {
					return 0, -n
				}
			case 4:
				if ix/1e4 != a {
					return 0, -n
				}
			case 5:
				if ix/1e5 != a {
					return 0, -n
				}
			case 6:
				if ix/1e6 != a {
					return 0, -n
				}
			case 7:
				if ix/1e7 != a {
					return 0, -n
				}
			case 8:
				if ix/1e8 != a {
					return 0, -n
				}
			case 9:
				if ix/1e9 != a {
					return 0, -n
				}
			case 10:
				if ix/1e10 != a {
					return 0, -n
				}
			case 11:
				if ix/1e11 != a {
					return 0, -n
				}
			case 12:
				if ix/1e12 != a {
					return 0, -n
				}
			case 13:
				if ix/1e13 != a {
					return 0, -n
				}
			case 14:
				if ix/1e14 != a {
					return 0, -n
				}
			case 15:
				if ix/1e15 != a {
					return 0, -n
				}
			}
		}
	}

	return ix, n
}

var (
	zimul = [...]int64{0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9, 1e10, 1e11, 1e12, 1e13, 1e14, 1e15}
	zumul = [...]uint64{0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9, 1e10, 1e11, 1e12, 1e13, 1e14, 1e15}

	// These masks represent the maximum number you can multiply by each
	// corresponding index in zumul before overflowing before you have to
	// calculate zero multiplication the slow way:
	zumask = [...]uint64{
		0,
		^(uint64(1<<(64-4)) - 1),
		^(uint64(1<<(64-7)) - 1),
		^(uint64(1<<(64-10)) - 1),
		^(uint64(1<<(64-14)) - 1),
		^(uint64(1<<(64-17)) - 1),
		^(uint64(1<<(64-20)) - 1),
		^(uint64(1<<(64-24)) - 1),
		^(uint64(1<<(64-27)) - 1),
		^(uint64(1<<(64-30)) - 1),
		^(uint64(1<<(64-34)) - 1),
		^(uint64(1<<(64-37)) - 1),
		^(uint64(1<<(64-40)) - 1),
		^(uint64(1<<(64-44)) - 1),
		^(uint64(1<<(64-47)) - 1),
		^(uint64(1<<(64-50)) - 1),
	}

	// These masks represent the maximum number you can multiply by each
	// corresponding index in zimul before overflowing before you have to
	// calculate zero multiplication the slow way:
	zimask = [...]int64{
		0,
		int64(zumask[1] & signMask),
		int64(zumask[2] & signMask),
		int64(zumask[3] & signMask),
		int64(zumask[4] & signMask),
		int64(zumask[5] & signMask),
		int64(zumask[6] & signMask),
		int64(zumask[7] & signMask),
		int64(zumask[8] & signMask),
		int64(zumask[9] & signMask),
		int64(zumask[10] & signMask),
		int64(zumask[11] & signMask),
		int64(zumask[12] & signMask),
		int64(zumask[13] & signMask),
		int64(zumask[14] & signMask),
		int64(zumask[15] & signMask),
	}

	overflow = errors.New("fixvarint: varint overflows a 64-bit integer")
)

const signMask = ^uint64(1 << 63)

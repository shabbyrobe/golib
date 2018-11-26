package fixvarint

import (
	"errors"
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
// 	n  < 0: value larger than 64 bits (overflow)
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
			if i > 9 || i == 9 && b > 1 {
				return 0, -(i + 1) // overflow
			}
			x, n = x|uint64(b)<<s, i+2 // +1 for the slice offset, +1 to convert from 0-index

			goto done
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}

done:
	switch zeros {
	case 1:
		x *= 1e1
	case 2:
		x *= 1e2
	case 3:
		x *= 1e3
	case 4:
		x *= 1e4
	case 5:
		x *= 1e5
	case 6:
		x *= 1e6
	case 7:
		x *= 1e7
	case 8:
		x *= 1e8
	case 9:
		x *= 1e9
	case 10:
		x *= 1e10
	case 11:
		x *= 1e11
	case 12:
		x *= 1e12
	case 13:
		x *= 1e13
	case 14:
		x *= 1e14
	case 15:
		x *= 1e15
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
	} else {
		x |= uint64(buf[1]&0x7f) << 3
	}

	if buf[2] < 0x80 {
		x, n = x|uint64(buf[2])<<10, 3
		goto done
	} else {
		x |= uint64(buf[2]&0x7f) << 10
	}

	if buf[3] < 0x80 {
		x, n = x|uint64(buf[3])<<17, 4
		goto done
	} else {
		x |= uint64(buf[3]&0x7f) << 17
	}

	if buf[4] < 0x80 {
		x, n = x|uint64(buf[4])<<24, 5
		goto done
	} else {
		x |= uint64(buf[4]&0x7f) << 24
	}

	if buf[5] < 0x80 {
		x, n = x|uint64(buf[5])<<31, 6
		goto done
	} else {
		x |= uint64(buf[5]&0x7f) << 31
	}

	if buf[6] < 0x80 {
		x, n = x|uint64(buf[6])<<38, 7
		goto done
	} else {
		x |= uint64(buf[6]&0x7f) << 38
	}

	if buf[7] < 0x80 {
		x, n = x|uint64(buf[7])<<45, 8
		goto done
	} else {
		x |= uint64(buf[7]&0x7f) << 45
	}

	if buf[8] < 0x80 {
		x, n = x|uint64(buf[8])<<52, 9
		goto done
	} else {
		x |= uint64(buf[8]&0x7f) << 52
	}

	if buf[9] < 0x80 {
		x, n = x|uint64(buf[9])<<59, 10
		goto done
	} else {
		x |= uint64(buf[9]&0x7f) << 59
	}

	return 0, -11

done:
	switch zeros {
	case 1:
		x *= 1e1
	case 2:
		x *= 1e2
	case 3:
		x *= 1e3
	case 4:
		x *= 1e4
	case 5:
		x *= 1e5
	case 6:
		x *= 1e6
	case 7:
		x *= 1e7
	case 8:
		x *= 1e8
	case 9:
		x *= 1e9
	case 10:
		x *= 1e10
	case 11:
		x *= 1e11
	case 12:
		x *= 1e12
	case 13:
		x *= 1e13
	case 14:
		x *= 1e14
	case 15:
		x *= 1e15
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

	switch zeros {
	case 1:
		ix *= 1e1
	case 2:
		ix *= 1e2
	case 3:
		ix *= 1e3
	case 4:
		ix *= 1e4
	case 5:
		ix *= 1e5
	case 6:
		ix *= 1e6
	case 7:
		ix *= 1e7
	case 8:
		ix *= 1e8
	case 9:
		ix *= 1e9
	case 10:
		ix *= 1e10
	case 11:
		ix *= 1e11
	case 12:
		ix *= 1e12
	case 13:
		ix *= 1e13
	case 14:
		ix *= 1e14
	case 15:
		ix *= 1e15
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
	} else {
		ux |= uint64(buf[1]&0x7f) << 3
	}

	if buf[2] < 0x80 {
		ux, n = ux|uint64(buf[2])<<10, 3
		goto done
	} else {
		ux |= uint64(buf[2]&0x7f) << 10
	}

	if buf[3] < 0x80 {
		ux, n = ux|uint64(buf[3])<<17, 4
		goto done
	} else {
		ux |= uint64(buf[3]&0x7f) << 17
	}

	if buf[4] < 0x80 {
		ux, n = ux|uint64(buf[4])<<24, 5
		goto done
	} else {
		ux |= uint64(buf[4]&0x7f) << 24
	}

	if buf[5] < 0x80 {
		ux, n = ux|uint64(buf[5])<<31, 6
		goto done
	} else {
		ux |= uint64(buf[5]&0x7f) << 31
	}

	if buf[6] < 0x80 {
		ux, n = ux|uint64(buf[6])<<38, 7
		goto done
	} else {
		ux |= uint64(buf[6]&0x7f) << 38
	}

	if buf[7] < 0x80 {
		ux, n = ux|uint64(buf[7])<<45, 8
		goto done
	} else {
		ux |= uint64(buf[7]&0x7f) << 45
	}

	if buf[8] < 0x80 {
		ux, n = ux|uint64(buf[8])<<52, 9
		goto done
	} else {
		ux |= uint64(buf[8]&0x7f) << 52
	}

	if buf[9] < 0x80 {
		ux, n = ux|uint64(buf[9])<<59, 10
		goto done
	} else {
		ux |= uint64(buf[9]&0x7f) << 59
	}

	return 0, -11

done:
	ix = int64(ux >> 1)
	if ux&1 != 0 {
		ix = ^ix
	}

	switch zeros {
	case 1:
		ix *= 1e1
	case 2:
		ix *= 1e2
	case 3:
		ix *= 1e3
	case 4:
		ix *= 1e4
	case 5:
		ix *= 1e5
	case 6:
		ix *= 1e6
	case 7:
		ix *= 1e7
	case 8:
		ix *= 1e8
	case 9:
		ix *= 1e9
	case 10:
		ix *= 1e10
	case 11:
		ix *= 1e11
	case 12:
		ix *= 1e12
	case 13:
		ix *= 1e13
	case 14:
		ix *= 1e14
	case 15:
		ix *= 1e15
	}

	return ix, n
}

var overflow = errors.New("fixvarint: varint overflows a 64-bit integer")

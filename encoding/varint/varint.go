/*
Package varint implements a variable length integer encode/decode pattern.

It uses a similar algorithm to Git for the layoutof unsigned integers, and uses
zig-zag encoding for signed ones.
*/
package varint

import (
	"errors"
)

const varintOverflow = ^(uint64(1<<57) - 1)

var (
	errShort      = errors.New("varint: short buf")
	errIncomplete = errors.New("varint: incomplete")
	errOverflow   = errors.New("varint: overflow")
)

func IsOverflow(err error) bool    { return err == errOverflow }
func IsShortBuffer(err error) bool { return err == errShort }
func IsIncomplete(err error) bool  { return err == errIncomplete }

func DecodeUint(buf []byte) (v uint64, n int, rerr error) {
	blen := len(buf)
	if blen < 1 {
		return 0, 0, errShort
	}
	end := blen - 1

	c := buf[0]
	v = uint64(c & 127)

	idx := 0
	for c&128 != 0 {
		if idx == end {
			return 0, idx + 1, errIncomplete
		}

		v += 1
		if v == 0 || (v&varintOverflow) != 0 {
			return 0, idx + 1, errOverflow
		}
		idx++
		c = buf[idx]
		v = (v << 7) + uint64(c&127)
	}

	return v, idx + 1, nil
}

func DecodeInt(buf []byte) (vi int64, n int, rerr error) {
	const mask = ^uint64(0) >> 1

	blen := len(buf)
	if blen < 1 {
		return 0, 0, errShort
	}
	end := blen - 1

	c := buf[0]
	v := uint64(c & 127)

	idx := 0
	for c&128 != 0 {
		if idx == end {
			return 0, idx + 1, errIncomplete
		}

		v += 1
		if v == 0 || (v&varintOverflow) != 0 {
			return 0, idx + 1, errOverflow
		}
		idx++
		c = buf[idx]
		v = (v << 7) + uint64(c&127)
	}

	vi = int64((v>>1)&mask) ^ -(int64(v) & 1)
	return vi, idx + 1, nil
}

func AppendUint(v uint64, into []byte) (n int, buf []byte) {
	var scratch [16]byte
	pos := 15
	scratch[pos] = byte(v & 127)
	for {
		v >>= 7
		if v == 0 {
			break
		}
		pos--
		v--
		scratch[pos] = 128 | (byte(v) & 127)
	}
	buf = append(into, scratch[pos:]...)
	return 15 - pos, buf
}

func AppendInt(vi int64, into []byte) (n int, buf []byte) {
	v := uint64((vi >> 63) ^ (vi << 1))

	var scratch [16]byte
	pos := 15
	scratch[pos] = byte(v & 127)
	for {
		v >>= 7
		if v == 0 {
			break
		}
		pos--
		v--
		scratch[pos] = 128 | (byte(v) & 127)
	}
	buf = append(into, scratch[pos:]...)
	return 15 - pos, buf
}

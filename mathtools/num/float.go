// This file contains a heavily modified version of math.Mod
// that only supports our specific range of values.
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package num

import "math"

const (
	mask  = 0x7FF
	shift = 64 - 11 - 1
	bias  = 1023
)

func mod(x, y float64) float64 {
	yfr, yexp := frexp(y)
	neg := false
	r := x
	if x < 0 {
		r = -x
		neg = true
	}

	for r >= y {
		rfr, rexp := frexp(r)
		if rfr < yfr {
			rexp = rexp - 1
		}
		r = r - ldexp(y, rexp-yexp)
	}
	if neg {
		r = -r
	}
	return r
}

func frexp(f float64) (frac float64, exp int) {
	bits := math.Float64bits(f)
	exp += int((bits>>shift)&mask) - bias + 1
	bits &^= mask << shift
	bits |= (-1 + bias) << shift
	frac = math.Float64frombits(bits)
	return
}

func ldexp(frac float64, exp int) float64 {
	x := math.Float64bits(frac)
	exp += int(x>>shift)&mask - bias
	x &^= mask << shift
	x |= uint64(exp+bias) << shift
	return math.Float64frombits(x)
}

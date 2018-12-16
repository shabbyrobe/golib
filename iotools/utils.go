package iotools

import "io"

func ComputeSeekStart(current, size int64, offset int64, whence int) (computedOffset int64) {
	switch whence {
	case io.SeekStart:
		return offset
	case io.SeekCurrent:
		return current + offset
	case io.SeekEnd:
		// The SeekEnd offset is supposed to be negative if you want it to do
		// the thing you expect it to do, but this is not documented anywhere.
		return current + offset
	default:
		panic("iotools: unsupported whence when computing seek")
	}
}

package iotools

import "bytes"

type splitReader struct {
	input  *bytes.Reader
	lens   []int
	curLen int
	done   bool
}

func newSplitReader(bts []byte, lens ...int) *splitReader {
	return &splitReader{input: bytes.NewReader(bts), lens: lens}
}

func (rdr *splitReader) Read(buf []byte) (n int, err error) {
	if rdr.curLen < len(rdr.lens) {
		curLen := rdr.lens[rdr.curLen]
		rdr.curLen++
		if curLen < len(buf) {
			return rdr.input.Read(buf[:curLen])
		} else {
			return rdr.input.Read(buf)
		}
	}

	return rdr.input.Read(buf)
}

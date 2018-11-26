package iotools

import (
	"io"

	"github.com/pkg/errors"
)

type BytePrefixMessageReader struct {
	rdr    io.Reader
	buf    []byte
	bufPos int
	bufLen int
}

func NewBytePrefixMessageReader(rdr io.Reader, scratch []byte) *BytePrefixMessageReader {
	if len(scratch) == 0 {
		scratch = make([]byte, 65536)
	}

	if len(scratch) < 512 {
		// Double the maximum message size. It probably only needs to be 256
		// bytes plus one for the length plus one more for the overhang.
		panic("scratch must be 512 bytes or more")
	}

	return &BytePrefixMessageReader{
		rdr: rdr,
		buf: scratch,
	}
}

// ReadNext returns a slice containing the next message and the length of the
// message. The memory returned is valid only until the next call to ReadNext.
func (pr *BytePrefixMessageReader) ReadNext() (out []byte, n int, err error) {
again:
	if pr.bufPos >= pr.bufLen {
		n, err := pr.rdr.Read(pr.buf)
		pr.bufLen = n
		pr.bufPos = 0

		if err == io.EOF {
			return nil, 0, io.EOF // EOF is used to allow users to terminate the loop
		} else if err != nil {
			return nil, 0, errors.Wrap(err, "iotools: messagereader read failed")
		} else if n == 0 {
			return nil, 0, nil
		}
	}

	msgLen := int(pr.buf[pr.bufPos])
	pr.bufPos++
	if msgLen == 0 {
		goto again
	}

	if pr.bufPos+msgLen >= pr.bufLen {
		left := pr.bufLen - pr.bufPos
		copy(pr.buf, pr.buf[pr.bufPos:pr.bufPos+left])

		n, err := pr.rdr.Read(pr.buf[left:])
		pr.bufLen = n + left

		if err != nil {
			if err == io.EOF {
				if pr.bufLen == 0 {
					return nil, 0, io.EOF // EOF is used to allow users to terminate the loop
				}
			} else {
				return nil, 0, errors.Wrap(err, "iotools: messagereader read failed")
			}

		} else if n == 0 {
			return nil, 0, nil
		}

		pr.bufPos = 0
	}

	if pr.bufLen < msgLen {
		return nil, 0, errors.Errorf("iotools: short message read; expected %d bytes, found %d", msgLen, pr.bufLen)
	}

	out = pr.buf[pr.bufPos : pr.bufPos+msgLen]
	pr.bufPos += msgLen
	return out, msgLen, nil
}

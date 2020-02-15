package iotools

import (
	"encoding/binary"
	"fmt"
	"io"
)

type MessageReaderBytePrefix struct {
	rdr    io.Reader
	buf    []byte
	bufPos int
	bufLen int
}

func NewMessageReaderBytePrefix(rdr io.Reader, scratch []byte) *MessageReaderBytePrefix {
	if len(scratch) == 0 {
		scratch = make([]byte, 65536)
	}

	if len(scratch) < 512 {
		// Double the maximum message size. It probably only needs to be 256
		// bytes plus one for the length plus one more for the overhang.
		panic("scratch must be 512 bytes or more")
	}

	return &MessageReaderBytePrefix{
		rdr: rdr,
		buf: scratch,
	}
}

// ReadNext returns a slice containing the next message and the length of the
// message. The memory returned is valid only until the next call to ReadNext.
func (pr *MessageReaderBytePrefix) ReadNext() (out []byte, n int, err error) {
again:
	if pr.bufPos >= pr.bufLen {
		n, err := pr.rdr.Read(pr.buf)
		pr.bufLen = n
		pr.bufPos = 0

		if err == io.EOF {
			return nil, 0, io.EOF // EOF is used to allow users to terminate the loop
		} else if err != nil {
			return nil, 0, fmt.Errorf("iotools: messagereader read failed: %w", err)
		} else if n == 0 {
			return nil, 0, nil
		}
	}

	msgLen := int(pr.buf[pr.bufPos])
	pr.bufPos++
	if msgLen == 0 {
		goto again
	}

	for pr.bufPos+msgLen > pr.bufLen {
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
				return nil, 0, fmt.Errorf("iotools: messagereader read failed: %w", err)
			}

		} else if n == 0 {
			return nil, 0, nil
		}

		pr.bufPos = 0
	}

	if pr.bufLen < msgLen {
		return nil, 0, fmt.Errorf("iotools: short message read; expected %d bytes, found %d", msgLen, pr.bufLen)
	}

	out = pr.buf[pr.bufPos : pr.bufPos+msgLen]
	pr.bufPos += msgLen
	return out, msgLen, nil
}

type MessageReaderShortPrefix struct {
	rdr    io.Reader
	buf    []byte
	bufPos int
	bufLen int
}

func NewMessageReaderShortPrefix(rdr io.Reader, scratch []byte) *MessageReaderShortPrefix {
	if len(scratch) == 0 {
		scratch = make([]byte, 65536)
	} else if len(scratch) < 65536 {
		panic("scratch must be >= 65536 or nil")
	}
	return &MessageReaderShortPrefix{
		rdr: rdr,
		buf: scratch,
	}
}

// ReadNext returns a slice containing the next message and the length of the
// message. The memory returned is valid only until the next call to ReadNext.
func (pr *MessageReaderShortPrefix) ReadNext() (out []byte, n int, err error) {
again:
	if pr.bufPos >= pr.bufLen {
		n, err := io.ReadAtLeast(pr.rdr, pr.buf, 2)
		pr.bufLen = n
		pr.bufPos = 0

		// io.UnexpectedEOF is an error here - it means a short length read.

		if err == io.EOF {
			return nil, 0, io.EOF // EOF is used to allow users to terminate the loop
		} else if err != nil {
			return nil, 0, fmt.Errorf("iotools: messagereader read failed: %w", err)
		} else if n == 0 {
			return nil, 0, nil
		}
	}

	msgLen := int(binary.LittleEndian.Uint16(pr.buf[pr.bufPos:]))
	pr.bufPos += 2
	if msgLen == 0 {
		goto again
	}

	for pr.bufPos+msgLen > pr.bufLen {
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
				return nil, 0, fmt.Errorf("iotools: messagereader read failed: %w", err)
			}

		} else if n == 0 {
			return nil, 0, nil
		}

		pr.bufPos = 0
	}

	if pr.bufLen < msgLen {
		return nil, 0, fmt.Errorf("iotools: short message read; expected %d bytes, found %d", msgLen, pr.bufLen)
	}

	out = pr.buf[pr.bufPos : pr.bufPos+msgLen]
	pr.bufPos += msgLen
	return out, msgLen, nil
}

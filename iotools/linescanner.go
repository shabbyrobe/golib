package iotools

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

func failOnDiscard(limit int, start int) error {
	return fmt.Errorf("line starting at position %d exceeded limit %d", start, limit)
}

type LineScanner struct {
	rdr *bufio.Reader

	streamPos int
	buf       []byte
	bufSize   int
	bufPos    int
	readLimit int
	discard   func(limit int, start int) error

	line []byte
	err  error
}

func NewScanner(
	rdr io.Reader,
	readLimit int,
) *LineScanner {
	if readLimit <= 0 {
		readLimit = 8192
	}

	scn := &LineScanner{
		buf:       make([]byte, readLimit*2),
		readLimit: readLimit,
		rdr:       bufio.NewReader(rdr),
		discard:   failOnDiscard,
	}
	return scn
}

func (lscn *LineScanner) OnDiscard(discard func(limit int, start int) error) *LineScanner {
	lscn.discard = discard
	return lscn
}

func (lscn *LineScanner) Reset(rdr io.Reader) {
	lscn.rdr.Reset(rdr)
	lscn.bufSize = 0
	lscn.bufPos = 0
	lscn.streamPos = 0
	lscn.line = nil
	lscn.err = nil
}

func (lscn *LineScanner) nextLine() (line []byte, ok bool) {
	if lscn.err != nil && lscn.err != io.EOF {
		return nil, false
	}

	var discard bool
	var idx int
	var start = lscn.streamPos

search:
	for {
		// Is there a newline in the buffer?
		idx = bytes.IndexByte(lscn.buf[lscn.bufPos:lscn.bufSize], '\n')

		if idx >= 0 && discard {
			// If so, and we are in "discard" mode, chuck everything away up to the
			// newline and exit discard mode:
			lscn.bufPos += idx + 1
			lscn.streamPos += idx + 1
			discard = false
			if err := lscn.discard(len(lscn.buf), start); err != nil {
				lscn.err = err
				return nil, false
			}
			start = lscn.streamPos

		} else if idx >= 0 && !discard {
			// If so, and we are _not_ in "discard" mode, we have a line and we're done:
			lscn.streamPos += idx + 1
			break search

		} else if discard || lscn.bufSize-lscn.bufPos > lscn.readLimit {
			lscn.streamPos += lscn.bufSize

			// If there is no newline in the buffer AND:
			// - We are in discard mode OR:
			// - The buffer does not have enough room in it for a full read if we
			//   slide what's left in it to the start:
			//
			// Replace the entire buffer with a fresh read and enter discard mode
			// if we're not in it already.
			n, err := lscn.rdr.Read(lscn.buf[:lscn.readLimit])
			if err != nil {
				lscn.err = err
				break search
			}
			lscn.bufPos = 0
			lscn.bufSize = n
			discard = true

		} else {
			// If there is no newline in the buffer and there is enough room for a read
			// (after we take bufPos into account), move the existing data to the start,
			// read some, and try again:
			lscn.bufSize = copy(lscn.buf, lscn.buf[lscn.bufPos:lscn.bufSize])
			lscn.bufPos = 0

			n, err := lscn.rdr.Read(lscn.buf[lscn.bufSize : lscn.bufSize+lscn.readLimit])
			if err != nil {
				lscn.err = err
				break search
			}
			lscn.bufSize += n
		}
	}

	// If we never found a newline, but we are at EOF, this is the last line:
	if idx < 0 && lscn.err == io.EOF {
		line := lscn.buf[lscn.bufPos:lscn.bufSize]
		if len(line) == 0 {
			return nil, false
		}
		lscn.bufPos = lscn.bufSize
		return line, true

	} else if lscn.err != nil {
		return nil, false
	}

	// If we never found a newline but we are not at EOF, something went very wrong:
	if idx < 0 {
		panic(fmt.Errorf("expected newline"))
	}

	line = lscn.buf[lscn.bufPos : lscn.bufPos+idx]
	lscn.bufPos += idx + 1
	return line, true
}

func (lscn *LineScanner) Scan() bool {
	line, ok := lscn.nextLine()
	if ok {
		lscn.line = line
	} else {
		lscn.line = nil
	}
	return ok
}

func (lscn *LineScanner) Bytes() []byte {
	return lscn.line
}

func (lscn *LineScanner) Err() error {
	if lscn.err == io.EOF {
		return nil
	}
	return lscn.err
}

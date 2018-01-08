package iotools

import (
	"bytes"
	"fmt"
	"io"
)

// CommitReader allows you to Commit a series of reads you have just made,
// or Rewind to the position the reader was at before the last Commit or
// Advance.
type CommitReader struct {
	r     io.Reader
	buf   bytes.Buffer
	pos   int
	rbuf  []byte
	eof   bool
	valid bool
}

func NewCommitReader(r io.Reader) *CommitReader {
	return NewCommitReaderSize(r, 8192)
}

func NewCommitReaderSize(r io.Reader, sz int) *CommitReader {
	return &CommitReader{
		r:     r,
		rbuf:  make([]byte, sz),
		valid: true,
	}
}

func (c *CommitReader) Pos() int {
	return c.pos
}

// Rest returns the remaining data as a reader. The commit reader becomes invalid.
func (c *CommitReader) Rest() io.Reader {
	if !c.valid {
		panic("commit reader is not valid")
	}
	c.Commit()
	c.valid = false
	return NewLeadingReader(c.buf.Bytes(), c.r)
}

func (c *CommitReader) Read(p []byte) (n int, err error) {
	if !c.valid {
		panic("commit reader is not valid")
	}

	max := len(p)
	ln := c.buf.Len()
	left := ln - c.pos

	if left == 0 {
		rn, rerr := c.r.Read(c.rbuf)
		if rn > 0 {
			left += rn
			ln += rn
			_, berr := c.buf.Write(c.rbuf[0:rn])
			if berr != nil {
				return 0, berr
			}
		}

		if rerr == io.EOF {
			c.eof = true
		} else if rerr != nil {
			err = rerr
			return
		}
	}

	if c.eof && left == 0 {
		return 0, io.EOF
	}

	n = max
	if left < max {
		n = left
	}

	b := c.buf.Bytes()
	if n > 0 {
		copy(p, b[c.pos:c.pos+n])
	}
	c.pos += n
	return n, nil
}

func (c *CommitReader) Commit() {
	if !c.valid {
		panic("commit reader is not valid")
	}
	x := c.buf.Next(c.pos)
	if len(x) != c.pos {
		panic(fmt.Errorf("unexpected buffer size"))
	}
	c.pos = 0
}

func (c *CommitReader) Rewind() {
	if !c.valid {
		panic("commit reader is not valid")
	}
	c.pos = 0
}

func (c *CommitReader) Advance(n int) int {
	if !c.valid {
		panic("commit reader is not valid")
	}
	x := c.buf.Next(n)
	out := len(x)
	c.pos = 0
	return out
}

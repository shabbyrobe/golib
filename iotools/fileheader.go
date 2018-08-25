package iotools

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

const fileHeaderLengthBytes = 8

// FileHeaderBytes extracts the length-delimited header portion of a file from
// the supplied byte array, if and only if the magic string is present at the
// start.
//
// This API is not stable.
func FileHeaderBytes(magic string, bts []byte) (hdr []byte, rest []byte, err error) {
	magicLen := len(magic)

	expected := magicLen + fileHeaderLengthBytes
	if len(bts) < expected {
		err = fmt.Errorf("iotools: file header expected at least %d bytes, found %d", expected, len(bts))
		return
	}

	for i := 0; i < magicLen; i++ {
		if bts[i] != magic[i] {
			err = fmt.Errorf("iotools: file header expected magic in first %d bytes of file", fileHeaderLengthBytes)
			return
		}
	}

	hlen := binary.LittleEndian.Uint64(bts[expected:])
	eu64 := uint64(expected)
	return bts[eu64 : eu64+hlen], bts[eu64+hlen:], nil
}

// FileHeaderRead reads the length-delimited header portion of a file from the
// supplied io.Reader, if and only if the magic string is present at the start.
//
// This API is not stable.
func FileHeaderRead(magic string, rdr io.Reader) (hdr []byte, err error) {
	magicLen := len(magic)

	expected := magicLen + fileHeaderLengthBytes
	bts := make([]byte, expected)
	n, err := io.ReadFull(rdr, bts)
	if n != expected {
		return nil, errors.Errorf("iotools: file header short read preamble")
	} else if err != nil {
		return nil, err
	}

	for i := 0; i < magicLen; i++ {
		if bts[i] != magic[i] {
			err = fmt.Errorf("iotools: file header expected magic in first %d bytes of file", fileHeaderLengthBytes)
		}
	}

	hlen := binary.LittleEndian.Uint32(bts[magicLen:])
	if hlen > 0 {
		hdr = make([]byte, hlen)
		n, err = io.ReadFull(rdr, hdr)
		if uint32(n) != hlen {
			return nil, errors.Errorf("iotools: file header short read")
		} else if err != nil {
			return nil, err
		}
	}
	return hdr, nil
}

// FileHeaderWrite writes a length-delimited header section to an io.Writer,
// preceded by a magic string to help identify the file.
func FileHeaderWrite(magic string, w io.Writer, hdr []byte) (n int, err error) {
	magicLen := len(magic)
	preamble := make([]byte, magicLen+fileHeaderLengthBytes)
	copy(preamble, magic)

	binary.LittleEndian.PutUint64(preamble[magicLen:], uint64(len(hdr)))

	cn, err := w.Write(preamble)
	n += cn
	if err != nil {
		return n, err
	} else if cn != len(preamble) {
		return n, errors.Errorf("iotools: file header short length write")
	}

	cn, err = w.Write(hdr)
	n += cn
	if err != nil {
		return n, err
	} else if cn != len(hdr) {
		return n, errors.Errorf("iotools: file header short write")
	}

	return n, nil
}

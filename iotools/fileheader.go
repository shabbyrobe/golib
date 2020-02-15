package iotools

import (
	"encoding/binary"
	"fmt"
	"io"
)

const fileHeaderLengthBytes = 4

// FileHeaderBytes extracts the length-delimited header portion of a file from
// the supplied byte array, if and only if the magic string is present at the
// start.
//
// This API is not stable.
func FileHeaderBytes(magic string, bts []byte) (hdr []byte, n int, err error) {
	magicLen := len(magic)

	expected := magicLen + fileHeaderLengthBytes
	if len(bts) < expected {
		err = fmt.Errorf("iotools: file header expected at least %d bytes, found %d", expected, len(bts))
		return
	}

	for i := 0; i < magicLen; i++ {
		if bts[i] != magic[i] {
			err = fmt.Errorf("iotools: file header expected magic %q in first %d bytes of file", magic, fileHeaderLengthBytes)
			return
		}
	}

	hlen := binary.LittleEndian.Uint32(bts[expected:])
	eu32 := uint32(expected)
	return bts[eu32 : eu32+hlen], int(eu32 + hlen), nil
}

// FileHeaderRead reads the length-delimited header portion of a file from the
// supplied io.Reader, if and only if the magic string is present at the start.
//
// If an error occurs, all bytes read from the reader are returned as 'hdr'.
//
// This API is not stable.
func FileHeaderRead(magic string, rdr io.Reader) (hdr []byte, n int, err error) {
	magicLen := len(magic)

	expected := magicLen + fileHeaderLengthBytes
	bts := make([]byte, expected)

	rn, err := io.ReadFull(rdr, bts)
	n += rn
	if rn != expected {
		return bts, n, fmt.Errorf("iotools: file header short read preamble")
	} else if err != nil {
		return bts, n, err
	}

	for i := 0; i < magicLen; i++ {
		if bts[i] != magic[i] {
			return bts, n, fmt.Errorf("iotools: file header expected magic %q in first %d bytes of file", magic, fileHeaderLengthBytes)
		}
	}

	hlen := binary.LittleEndian.Uint32(bts[magicLen:])
	if hlen > 0 {
		hdr = make([]byte, hlen)
		rn, err = io.ReadFull(rdr, hdr)
		n += rn
		if uint32(rn) != hlen {
			return bts, n, fmt.Errorf("iotools: file header short read")
		} else if err != nil {
			return bts, n, err
		}
	}
	return hdr, n, nil
}

// FileHeaderWrite writes a length-delimited header section to an io.Writer,
// preceded by a magic string to help identify the file.
func FileHeaderWrite(magic string, w io.Writer, hdr []byte) (n int, err error) {
	magicLen := len(magic)
	preamble := make([]byte, magicLen+fileHeaderLengthBytes)
	copy(preamble, magic)

	binary.LittleEndian.PutUint32(preamble[magicLen:], uint32(len(hdr)))

	cn, err := w.Write(preamble)
	n += cn
	if err != nil {
		return n, err
	} else if cn != len(preamble) {
		return n, fmt.Errorf("iotools: file header short length write")
	}

	cn, err = w.Write(hdr)
	n += cn
	if err != nil {
		return n, err
	} else if cn != len(hdr) {
		return n, fmt.Errorf("iotools: file header short write")
	}

	return n, nil
}

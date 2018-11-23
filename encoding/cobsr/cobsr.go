/*
Package cobsr implements Craig McQueen's COBS/R encoding method, a variation on
the Consistent Overhead Byte Stuffing algorithm.

The COBS encoding guarantees that an encoded message will never contain the
byte 0x00, which frees it up to be used as a message delimiter, in a way that
minimises the overhead.

Both COBS and COBS/R are well explained here:

	https://pythonhosted.org/cobs/intro.html
	https://pythonhosted.org/cobs/cobs.cobsr.html
*/
package cobsr

// MaxEncodedSize provides an upper bound for the output size of an encoded
// message.
//
// The encoded data length may be one byte longer than the input length.
// Additionally, it may increase by one extra byte for every 254 bytes of
// input data.
func MaxEncodedSize(inlen int) int {
	return inlen + 1 + (inlen / 254) + 1
}

// Encode a byte string according to the COBS/R encoding method.
//
// The 'into' slice must be large enough to hold the resulting encoded
// message. Using MaxEncodedSize() will guarantee the 'into' buffer is
// large enough.
//
func Encode(in []byte, into []byte) (n int, err error) {
	var searchStartIdx int
	var intoIdx int

	for idx, b := range in {
		if idx-searchStartIdx == 0xfe {
			into[intoIdx] = 0xff
			intoIdx++
			for i := searchStartIdx; i < idx; i++ {
				into[intoIdx] = in[i]
				intoIdx++
			}
			searchStartIdx = idx
		}
		if b == 0 {
			into[intoIdx] = byte(idx - searchStartIdx + 1)
			intoIdx++
			for i := searchStartIdx; i < idx; i++ {
				into[intoIdx] = in[i]
				intoIdx++
			}
			searchStartIdx = idx + 1
		}
	}

	inLen := len(in)
	var finalByte byte
	if inLen > 0 {
		finalByte = in[inLen-1]
	}

	lengthValue := byte(inLen - searchStartIdx + 1)
	if finalByte < lengthValue {
		into[intoIdx] = lengthValue
		intoIdx += copy(into[intoIdx+1:], in[searchStartIdx:]) + 1
	} else {
		into[intoIdx] = finalByte
		intoIdx += copy(into[intoIdx+1:], in[searchStartIdx:inLen-1]) + 1
	}

	return intoIdx, nil
}

// Decode a byte string according to the COBS/R method.
//
// If an unexpected zero-byte is encountered, an error is returned. This
// error can be tested using cobsr.IsErrZeroFound().
//
// The 'into' slice must be large enough to hold the resulting decoded
// message. The decoded message will never be larger than the encoded
// message.
//
func Decode(in []byte, into []byte) (n int, err error) {
	inLen := len(in)
	if inLen == 0 {
		return 0, nil
	}

	var idx int
	var intoIdx int

	for {
		length := in[idx]
		if length == 0 {
			return intoIdx, errvZeroFound
		}
		idx++

		end := idx + int(length) - 1
		endIdx := end
		if endIdx > inLen {
			endIdx = inLen
		}
		for _, b := range in[idx:endIdx] {
			if b == 0 {
				return intoIdx, errvZeroFound
			}
			into[intoIdx] = b
			intoIdx++
		}
		idx = end
		if idx > inLen {
			into[intoIdx] = length
			intoIdx++
			break

		} else if idx < inLen {
			if length < 0xFF {
				into[intoIdx] = 0
				intoIdx++
			}

		} else {
			break
		}
	}
	return intoIdx, nil
}

var errvZeroFound = &errZeroFound{}

type errZeroFound struct{}

func (e *errZeroFound) Error() string { return "cobsr: zero found in input" }

func IsErrZeroFound(err error) bool {
	return err == errvZeroFound
}

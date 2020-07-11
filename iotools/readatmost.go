package iotools

import (
	"fmt"
	"io"
	"io/ioutil"
)

var ErrReadTooLarge = fmt.Errorf("io: read too large")

func ReadAtMost(r io.Reader, limit int64) (bts []byte, err error) {
	limRdr := io.LimitedReader{R: r, N: limit}
	bts, err = ioutil.ReadAll(&limRdr)
	if err != nil {
		return bts, err
	}
	if limRdr.N <= 0 {
		return bts, ErrReadTooLarge
	}
	return bts, nil
}

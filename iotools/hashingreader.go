package iotools

import (
	"fmt"
	"hash"
	"io"
)

// HashingReader proxies an existing io.Reader, passing each read block to the
// given hash.Hash.
type HashingReader struct {
	inner io.Reader
	Hash  hash.Hash
}

func NewHashingReader(inner io.Reader, hash hash.Hash) *HashingReader {
	return &HashingReader{
		inner: inner,
		Hash:  hash,
	}
}

func (h *HashingReader) Read(p []byte) (n int, err error) {
	n, err = h.inner.Read(p)
	if err != nil {
		return n, err
	}
	if n == 0 {
		return 0, nil
	}

	wn, _ := h.Hash.Write(p[:n]) // Hash.Write never returns an error.
	if wn != n {
		return n, fmt.Errorf("iotools: short write to hasher")
	}
	return n, nil
}

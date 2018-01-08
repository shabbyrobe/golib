package iotools

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/shabbyrobe/golib/assert"
)

func TestLeadingReader(t *testing.T) {
	tt := assert.WrapTB(t)

	lr := NewLeadingReader([]byte{'a'}, bytes.NewReader([]byte{'b'}))
	all, err := ioutil.ReadAll(lr)
	tt.MustOK(err)
	tt.MustEqual([]byte{'a', 'b'}, all)
}

func TestLeadingReaderEmptyLeading(t *testing.T) {
	tt := assert.WrapTB(t)

	lr := NewLeadingReader([]byte{}, bytes.NewReader([]byte{'b'}))
	all, err := ioutil.ReadAll(lr)
	tt.MustOK(err)
	tt.MustEqual([]byte{'b'}, all)
}

func TestLeadingReaderEmptyReader(t *testing.T) {
	tt := assert.WrapTB(t)

	lr := NewLeadingReader([]byte{'a'}, bytes.NewReader([]byte{}))
	all, err := ioutil.ReadAll(lr)
	tt.MustOK(err)
	tt.MustEqual([]byte{'a'}, all)
}

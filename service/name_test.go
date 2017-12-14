package service

import (
	"math/rand"
	"testing"
)

func TestNameAppend(t *testing.T) {
	tt := WrapTB(t)
	tt.MustEqual(Name("foo/bar"), Name("foo").Append("bar"))
}

func TestNameAppendUnique(t *testing.T) {
	tt := WrapTB(t)

	rand.Seed(1)
	tt.MustEqual(Name("foo/52FDFC072182654F163F5F0F9A621D72"), Name("foo").AppendUnique())
}

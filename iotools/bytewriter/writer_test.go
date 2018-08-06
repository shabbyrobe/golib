package bytewriter

import (
	"testing"

	"github.com/shabbyrobe/golib/assert"
)

func TestByteWriterAppend(t *testing.T) {
	tt := assert.WrapTB(t)
	_ = tt

	var bw Writer
	{
		bw.Write([]byte("abc"))
		bw.Write([]byte("def"))
		bw.Write([]byte("ghi"))
		buf, n := bw.Take()
		tt.MustEqual(9, n)
		tt.MustEqual("abcdefghi", string(buf))
	}

	{ // the writer should be reset if you take the buffer
		bw.Write([]byte("abc"))
		buf, n := bw.Take()
		tt.MustEqual(3, n)
		tt.MustEqual("abc", string(buf))
	}
}

func TestByteWriterSet(t *testing.T) {
	tt := assert.WrapTB(t)
	_ = tt

	var bw Writer
	bw.Give(make([]byte, 0, 1024))
	bw.Write([]byte("abc"))
	bw.Write([]byte("def"))
	bw.Write([]byte("ghi"))
	buf, n := bw.Take()
	tt.MustEqual(9, n)
	tt.MustEqual("abcdefghi", string(buf))
	tt.MustEqual(1024, cap(buf))
	tt.MustEqual(9, len(buf))
}

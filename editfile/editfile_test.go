package editfile

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"

	"github.com/shabbyrobe/golib/assert"
	"github.com/shabbyrobe/golib/iotools"
)

type ReadWriterAt interface {
	io.ReaderAt
	io.WriterAt
}

func tempFile(tt assert.T, body []byte) string {
	tt.Helper()
	tmp, err := ioutil.TempFile("", "")
	tt.MustOK(err)
	if len(body) > 0 {
		tt.MustOKAll(tmp.Write(body))
	}
	tt.MustOK(tmp.Close())
	return tmp.Name()
}

func assertExists(tt assert.T, path string) {
	tt.Helper()
	exists, err := iotools.Exists(path)
	tt.MustOK(err)
	tt.MustAssert(exists)
}

func assertNotExists(tt assert.T, path string) {
	tt.Helper()
	exists, err := iotools.Exists(path)
	tt.MustOK(err)
	tt.MustAssert(!exists)
}

func assertFile(tt assert.T, path string, contents []byte) {
	tt.Helper()
	result, err := ioutil.ReadFile(path)
	tt.MustOK(err)
	if !bytes.Equal(contents, result) {
		tt.Fatalf("compare failed for file %s", path)
	}
}

func assertWriteAt(tt assert.T, to ReadWriterAt, data []byte, at int64) {
	tt.Helper()
	n, err := to.WriteAt(data, at)
	tt.MustOK(err)
	tt.MustEqual(len(data), n)

	check := make([]byte, len(data))
	n, err = to.ReadAt(check, at)
	tt.MustOK(err)
	tt.MustEqual(check, data)
}

func TestEditFile(t *testing.T) {
	tt := assert.WrapTB(t)

	f := tempFile(tt, nil)
	defer os.Remove(f)

	ef, err := Edit(f)
	tt.MustOK(err)
	tt.MustOK(ef.Close())

	assertExists(tt, f)
}

func TestEditFileModifyExisting(t *testing.T) {
	tt := assert.WrapTB(t)

	f := tempFile(tt, []byte{1, 2, 3})
	defer os.Remove(f)

	ef, err := Edit(f)
	tt.MustOK(err)
	assertWriteAt(tt, ef, []byte{9}, 1)
	tt.MustOK(ef.Close())

	assertFile(tt, f, []byte{1, 9, 3})
}

func TestEditFileModifyExistingExtend(t *testing.T) {
	tt := assert.WrapTB(t)

	f := tempFile(tt, []byte{1, 2, 3})
	defer os.Remove(f)

	ef, err := Edit(f)
	tt.MustOK(err)
	assertWriteAt(tt, ef, []byte{9}, 5)
	tt.MustOK(ef.Close())

	assertFile(tt, f, []byte{1, 2, 3, 0, 0, 9})
}

func TestEditFileModifyExistingTruncate(t *testing.T) {
	tt := assert.WrapTB(t)

	f := tempFile(tt, []byte{1, 2, 3})
	defer os.Remove(f)

	ef, err := Edit(f)
	tt.MustOK(err)
	tt.MustOK(ef.Truncate(1))
	tt.MustOK(ef.Close())

	assertFile(tt, f, []byte{1})
}

func TestEditFileModifyExistingTruncateExtend(t *testing.T) {
	tt := assert.WrapTB(t)

	f := tempFile(tt, []byte{1, 2, 3})
	defer os.Remove(f)

	ef, err := Edit(f)
	tt.MustOK(err)
	tt.MustOK(ef.Truncate(1))
	assertWriteAt(tt, ef, []byte{9}, 5)
	tt.MustOK(ef.Close())

	assertFile(tt, f, []byte{1, 0, 0, 0, 0, 9})
}

func TestEditFileLock(t *testing.T) {
	tt := assert.WrapTB(t)

	f := tempFile(tt, []byte{1, 2, 3})
	defer os.Remove(f)

	ef, err := Edit(f)
	tt.MustOK(err)
	defer ef.Close()

	_, err = Edit(f)
	tt.MustAssert(IsLocked(err))
	tt.MustOK(ef.Close())

	ef2, err := Edit(f)
	tt.MustOK(err)
	defer ef2.Close()
}

func TestEditFileDisabledWhenReadAtFails(t *testing.T) {
	tt := assert.WrapTB(t)

	f := tempFile(tt, []byte{1, 2, 3})
	defer os.Remove(f)

	ef, err := Edit(f)
	tt.MustOK(err)
	defer ef.Close()

	ef.buf = iotools.NewFullyBufferedWriterAt(&bungReadSeeker{}, &bungWriteDestination{})
	buf := make([]byte, 1)
	n, err := ef.ReadAt(buf, 0)
	tt.MustEqual(0, n)
	tt.MustAssert(err != nil)

	_, nextErr := ef.ReadAt(buf, 0)
	tt.MustAssert(IsDisabled(nextErr))
	tt.MustOK(ef.Close())

	assertNotExists(tt, ef.tmp.Name())
}

func TestEditFileDisabledWhenWriteAtFails(t *testing.T) {
	tt := assert.WrapTB(t)

	f := tempFile(tt, []byte{1, 2, 3})
	defer os.Remove(f)

	ef, err := Edit(f)
	tt.MustOK(err)
	defer ef.Close()

	ef.buf = iotools.NewFullyBufferedWriterAt(&bungReadSeeker{}, &bungWriteDestination{})
	buf := make([]byte, 1)
	n, err := ef.WriteAt(buf, 0)
	tt.MustEqual(0, n)
	tt.MustAssert(err != nil)

	_, nextErr := ef.ReadAt(buf, 0)
	tt.MustAssert(IsDisabled(nextErr))
}

func TestEditFileCloseLeavesOriginalWhenOpFails(t *testing.T) {
	tt := assert.WrapTB(t)

	f := tempFile(tt, []byte{1, 2, 3})
	defer os.Remove(f)

	ef, err := Edit(f)
	tt.MustOK(err)
	defer ef.Close()

	assertWriteAt(tt, ef, []byte{9, 9}, 2)

	// This is a bit nasty, we are fussing around with the internals in a potentially
	// very brittle way.
	ef.buf = iotools.NewFullyBufferedWriterAt(bytes.NewReader([]byte{0}), &bungWriteDestination{})
	tt.MustOKAll(ef.ReadAt([]byte{0}, 0))
	tt.MustAssert(ef.Flush() != nil)

	tt.MustAssert(ef.Close() != nil)
	assertFile(tt, f, []byte{1, 2, 3})
	assertNotExists(tt, ef.tmp.Name())
}

func TestEditFileFuzz(t *testing.T) {
	buf := make([]byte, 16384)
	rand.Read(buf)
	maxOffset := int64(16384)
	iters := 1
	maxWritesPerIter := 10

	for i := 0; i < iters; i++ {
		t.Run("", func(t *testing.T) {
			tt := assert.WrapTB(t)
			var check byteWriterAt

			f := tempFile(tt, nil)
			defer os.Remove(f)

			ef, err := Edit(f)
			tt.MustOK(err)

			for i := 0; i < maxWritesPerIter; i++ {
				sz := rand.Intn(len(buf)-1) + 1
				at := rand.Int63n(maxOffset)
				check.WriteAt(buf[:sz], at)
				assertWriteAt(tt, ef, buf[:sz], at)
			}

			tt.MustOK(ef.Close())
			assertFile(tt, f, check.buf)
		})
	}
}

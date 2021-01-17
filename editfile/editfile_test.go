package editfile

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"testing"

	"github.com/shabbyrobe/golib/iotools"
)

type ReadWriterAt interface {
	io.ReaderAt
	io.WriterAt
}

func tempFile(t testing.TB, body []byte) string {
	t.Helper()
	tmp, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(body) > 0 {
		if _, err := tmp.Write(body); err != nil {
			t.Fatal(err)
		}
	}
	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}
	return tmp.Name()
}

func assertExists(t testing.TB, path string) {
	t.Helper()
	exists, err := iotools.Exists(path)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal(exists)
	}
}

func assertNotExists(t testing.TB, path string) {
	t.Helper()
	exists, err := iotools.Exists(path)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal(exists)
	}
}

func assertFile(t testing.TB, path string, contents []byte) {
	t.Helper()
	result, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(contents, result) {
		t.Fatalf("compare failed for file %s", path)
	}
}

func assertWriteAt(t testing.TB, to ReadWriterAt, data []byte, at int64) {
	t.Helper()
	n, err := to.WriteAt(data, at)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(data) {
		t.Fatal()
	}

	check := make([]byte, len(data))
	n, err = to.ReadAt(check, at)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(data, check) {
		t.Fatal()
	}
}

func TestEditFile(t *testing.T) {
	f := tempFile(t, nil)
	defer os.Remove(f)

	ef, err := Edit(f)
	if err != nil {
		t.Fatal(err)
	}
	if err := ef.Close(); err != nil {
		t.Fatal(err)
	}

	assertExists(t, f)
}

func TestEditFileModifyExisting(t *testing.T) {
	f := tempFile(t, []byte{1, 2, 3})
	defer os.Remove(f)

	ef, err := Edit(f)
	if err != nil {
		t.Fatal(err)
	}

	assertWriteAt(t, ef, []byte{9}, 1)
	if err := ef.Close(); err != nil {
		t.Fatal(err)
	}

	assertFile(t, f, []byte{1, 9, 3})
}

func TestEditFileModifyExistingExtend(t *testing.T) {
	f := tempFile(t, []byte{1, 2, 3})
	defer os.Remove(f)

	ef, err := Edit(f)
	if err != nil {
		t.Fatal(err)
	}

	assertWriteAt(t, ef, []byte{9}, 5)
	if err := ef.Close(); err != nil {
		t.Fatal(err)
	}

	assertFile(t, f, []byte{1, 2, 3, 0, 0, 9})
}

func TestEditFileModifyExistingTruncate(t *testing.T) {
	f := tempFile(t, []byte{1, 2, 3})
	defer os.Remove(f)

	ef, err := Edit(f)
	if err != nil {
		t.Fatal(err)
	}

	if err := ef.Truncate(1); err != nil {
		t.Fatal(err)
	}
	if err := ef.Close(); err != nil {
		t.Fatal(err)
	}

	assertFile(t, f, []byte{1})
}

func TestEditFileModifyExistingTruncateExtend(t *testing.T) {
	f := tempFile(t, []byte{1, 2, 3})
	defer os.Remove(f)

	ef, err := Edit(f)
	if err != nil {
		t.Fatal(err)
	}
	if err := ef.Truncate(1); err != nil {
		t.Fatal(err)
	}
	assertWriteAt(t, ef, []byte{9}, 5)
	if err := ef.Close(); err != nil {
		t.Fatal(err)
	}

	assertFile(t, f, []byte{1, 0, 0, 0, 0, 9})
}

func TestEditFileLock(t *testing.T) {
	f := tempFile(t, []byte{1, 2, 3})
	defer os.Remove(f)

	ef, err := Edit(f)
	if err != nil {
		t.Fatal(err)
	}
	defer ef.Close() // Ignore errors

	_, err = Edit(f)
	if !IsLocked(err) {
		t.Fatal()
	}
	if err := ef.Close(); err != nil {
		t.Fatal(err)
	}

	ef2, err := Edit(f)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := ef2.Close(); err != nil {
			t.Fatal(err)
		}
	}()
}

func TestEditFileDisabledWhenReadAtFails(t *testing.T) {
	f := tempFile(t, []byte{1, 2, 3})
	defer os.Remove(f)

	ef, err := Edit(f)
	if err != nil {
		t.Fatal(err)
	}

	defer ef.Close()

	ef.buf = iotools.NewFullyBufferedWriterAt(&bungReadSeeker{}, &bungWriteDestination{})
	buf := make([]byte, 1)
	n, err := ef.ReadAt(buf, 0)
	if n != 0 {
		t.Fatal()
	}
	if err == nil {
		t.Fatal(err)
	}

	_, nextErr := ef.ReadAt(buf, 0)
	if !IsDisabled(nextErr) {
		t.Fatal()
	}
	if err := ef.Close(); err != nil {
		t.Fatal(err)
	}

	assertNotExists(t, ef.tmp.Name())
}

func TestEditFileDisabledWhenWriteAtFails(t *testing.T) {
	f := tempFile(t, []byte{1, 2, 3})
	defer os.Remove(f)

	ef, err := Edit(f)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := ef.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	ef.buf = iotools.NewFullyBufferedWriterAt(&bungReadSeeker{}, &bungWriteDestination{})
	buf := make([]byte, 1)
	n, err := ef.WriteAt(buf, 0)
	if err == nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatal()
	}

	_, nextErr := ef.ReadAt(buf, 0)
	if !IsDisabled(nextErr) {
		t.Fatal()
	}
}

func TestEditFileCloseLeavesOriginalWhenOpFails(t *testing.T) {
	f := tempFile(t, []byte{1, 2, 3})
	defer os.Remove(f)

	ef, err := Edit(f)
	if err != nil {
		t.Fatal(err)
	}
	defer ef.Close() // Error ignored

	assertWriteAt(t, ef, []byte{9, 9}, 2)

	// This is a bit nasty, we are fussing around with the internals in a potentially
	// very brittle way.
	ef.buf = iotools.NewFullyBufferedWriterAt(bytes.NewReader([]byte{0}), &bungWriteDestination{})
	if _, err := ef.ReadAt([]byte{0}, 0); err != nil {
		t.Fatal()
	}
	if err := ef.Flush(); err == nil { // Error is expected here
		t.Fatal(err)
	}
	if err := ef.Close(); err == nil { // Error is expected here
		t.Fatal(err)
	}

	assertFile(t, f, []byte{1, 2, 3})
	assertNotExists(t, ef.tmp.Name())
}

func TestEditFileFuzz(t *testing.T) {
	buf := make([]byte, 16384)
	rand.Read(buf)
	maxOffset := int64(16384)
	iters := 1
	maxWritesPerIter := 10

	for i := 0; i < iters; i++ {
		t.Run("", func(t *testing.T) {
			var check byteWriterAt

			f := tempFile(t, nil)
			defer os.Remove(f)

			ef, err := Edit(f)
			if err != nil {
				t.Fatal(err)
			}

			for i := 0; i < maxWritesPerIter; i++ {
				sz := rand.Intn(len(buf)-1) + 1
				at := rand.Int63n(maxOffset)
				check.WriteAt(buf[:sz], at)
				assertWriteAt(t, ef, buf[:sz], at)
			}

			if err := ef.Close(); err != nil {
				t.Fatal(err)
			}
			assertFile(t, f, check.buf)
		})
	}
}

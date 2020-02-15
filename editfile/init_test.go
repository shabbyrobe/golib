package editfile

import "fmt"

type bungReadSeeker struct {
	err error
}

func (rdr *bungReadSeeker) Read(p []byte) (n int, err error) {
	if rdr.err != nil {
		return 0, rdr.err
	}
	return 0, fmt.Errorf("bungreader: read failed")
}

func (rdr *bungReadSeeker) Seek(off int64, whence int) (n int64, err error) {
	if rdr.err != nil {
		return 0, rdr.err
	}
	return 0, fmt.Errorf("bungreader: read failed")
}

type bungReaderAt struct {
	err error
}

func (rdr *bungReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	if rdr.err != nil {
		return 0, rdr.err
	}
	return 0, fmt.Errorf("bungreader: read failed")
}

type bungWriterAt struct {
	err error
}

func (wrt *bungWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	if wrt.err != nil {
		return 0, wrt.err
	}
	return 0, fmt.Errorf("bungwriter: write failed")
}

type bungWriteDestination struct {
	err error
}

func (wrt *bungWriteDestination) Truncate(sz int64) error { return wrt.err }
func (wrt *bungWriteDestination) Close() error            { return wrt.err }
func (wrt *bungWriteDestination) Write(p []byte) (n int, err error) {
	if wrt.err != nil {
		return 0, wrt.err
	}
	return 0, fmt.Errorf("bungwriter: write failed")
}

type byteWriterAt struct {
	buf []byte
	len int64
}

func (wrt *byteWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	plen64 := int64(len(p))
	if off+plen64 > wrt.len {
		wrt.buf = append(wrt.buf, make([]byte, off+plen64-wrt.len)...)
		wrt.len = off + plen64
	}

	copy(wrt.buf[off:], p)
	return int(plen64), nil
}

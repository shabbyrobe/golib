package curl

import (
	"compress/gzip"
	"io"
)

// gzipReader wraps a response body so it can lazily
// call gzip.NewReader on the first call to Read
type gzipReadCloser struct {
	body io.Reader
	cls  io.Closer
	zr   *gzip.Reader // lazily-initialized gzip reader
	zerr error        // any error from gzip.NewReader; sticky
}

func (gz *gzipReadCloser) Read(p []byte) (n int, err error) {
	if gz.zr == nil {
		if gz.zerr == nil {
			gz.zr, gz.zerr = gzip.NewReader(gz.body)
		}
		if gz.zerr != nil {
			return 0, gz.zerr
		}
	}
	return gz.zr.Read(p)
}

func (gz *gzipReadCloser) Close() error {
	if gz.zr != nil {
		gz.zr.Close()
	}
	return gz.cls.Close()
}

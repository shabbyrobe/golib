package unsafetools

import (
	"unsafe"
)

// Reading:
// https://groups.google.com/g/golang-nuts/c/Zsfk-VMd_fU/discussion
// https://github.com/golang/go/issues/19367
// https://github.com/golang/go/issues/25484
// https://go-review.googlesource.com/c/go/+/231223/
// https://github.com/golang/go/issues/53003#issuecomment-1140276077
// https://github.com/golang/go/issues/2205

func String(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func Bytes(s string) (b []byte) {
	if s == "" {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

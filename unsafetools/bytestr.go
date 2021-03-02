package unsafetools

import (
	"reflect"
	"unsafe"
)

type StringHeader struct {
	Data unsafe.Pointer
	Len  int
}

type SliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

// Reading:
// https://groups.google.com/g/golang-nuts/c/Zsfk-VMd_fU/discussion
// https://github.com/golang/go/issues/19367
// https://github.com/golang/go/issues/25484

// THIS IS EVIL CODE.
// YOU HAVE BEEN WARNED.
func String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// Alternative:
// func String(s string) (b []byte) {
//     st := (*StringHeader)(unsafe.Pointer(&s))
//     sl := (*SliceHeader)(unsafe.Pointer(&b))
//     sl.Data = st.Data
//     sl.Len = len(s)
//     sl.Cap = len(s)
//     return b
// }

// THIS IS EVIL CODE.
// YOU HAVE BEEN WARNED.
func Bytes(s string) (b []byte) {
	st := (*StringHeader)(unsafe.Pointer(&s))
	sl := (*SliceHeader)(unsafe.Pointer(&b))
	sl.Data = st.Data
	sl.Len = len(s)
	return b
}

func init() {
	// Check to make sure string header is what reflect thinks it is.
	// They should be the same except for the type of the Data field.
	if unsafe.Sizeof(StringHeader{}) != unsafe.Sizeof(reflect.StringHeader{}) {
		panic("string layout has changed")
	}
	x := StringHeader{}
	y := reflect.StringHeader{}
	x.Data = unsafe.Pointer(y.Data)
	y.Data = uintptr(x.Data)
	x.Len = y.Len
}

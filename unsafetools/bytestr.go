package unsafetools

import (
	"reflect"
	"unsafe"
)

// https://github.com/golang/go/issues/25484
// THIS IS EVIL CODE.
// YOU HAVE BEEN WARNED.
func String(bs []byte) string {
	return *(*string)(unsafe.Pointer(&bs))
}

// Alternative:
// func String(b []byte) string {
//     sh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
//     return *(*string)(unsafe.Pointer(&reflect.StringHeader{Data: sh.Data, Len: sh.Len}))
// }

// THIS IS EVIL CODE.
// YOU HAVE BEEN WARNED.
func Bytes(str string) []byte {
	hdr := *(*reflect.StringHeader)(unsafe.Pointer(&str))
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: hdr.Data,
		Len:  hdr.Len,
		Cap:  hdr.Len,
	}))
}

package fieldc

import "unsafe"

// #cgo CFLAGS: -O3
// #include "generic.h"
import "C"

type Element struct {
	l0 uint64
	l1 uint64
	l2 uint64
	l3 uint64
	l4 uint64
}

func Mul(v, a, b *Element) {
	C.feMulGeneric(
		(*C.Element)(unsafe.Pointer(v)),
		(*C.Element)(unsafe.Pointer(a)),
		(*C.Element)(unsafe.Pointer(b)),
	)
}

func Square(v, a *Element) {
	C.feSquareGeneric(
		(*C.Element)(unsafe.Pointer(v)),
		(*C.Element)(unsafe.Pointer(a)),
	)
}

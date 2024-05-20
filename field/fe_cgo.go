//go:build fieldc
// +build fieldc

package field

import (
	"unsafe"

	"filippo.io/edwards25519/field/fieldc"
)

// feMul sets out = a * b. It works like feMulGeneric.
func feMul(out *Element, a *Element, b *Element) {
	fieldc.Mul(
		(*fieldc.Element)(unsafe.Pointer(out)),
		(*fieldc.Element)(unsafe.Pointer(a)),
		(*fieldc.Element)(unsafe.Pointer(b)),
	)
}

// feSquare sets out = a * a. It works like feSquareGeneric.
func feSquare(out *Element, a *Element) {
	fieldc.Square(
		(*fieldc.Element)(unsafe.Pointer(out)),
		(*fieldc.Element)(unsafe.Pointer(a)),
	)
}

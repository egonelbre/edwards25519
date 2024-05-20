// Copyright (c) 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (!amd64 || !gc || purego) && !fieldc
// +build !amd64 !gc purego
// +build !fieldc

package field

import (
	"unsafe"

	"filippo.io/edwards25519/field/fieldc"
)

func feMul(v, x, y *Element) {
	fieldc.Mul(
		(*fieldc.Element)(unsafe.Pointer(v)),
		(*fieldc.Element)(unsafe.Pointer(x)),
		(*fieldc.Element)(unsafe.Pointer(y)),
	)
}

func feSquare(v, x *Element) {
	fieldc.Square(
		(*fieldc.Element)(unsafe.Pointer(v)),
		(*fieldc.Element)(unsafe.Pointer(x)),
	)
}

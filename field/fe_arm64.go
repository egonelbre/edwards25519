// Copyright (c) 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build arm64 && gc && !purego
// +build arm64,gc,!purego

package field

//go:noescape
func carryPropagate(v *Element)

func (v *Element) carryPropagate() *Element {
	carryPropagate(v)
	return v
}

// feMul sets out = a * b. It works like feMulGeneric.
//
//go:noescape
//func feMul(out *Element, a *Element, b *Element)

// feSquare sets out = a * a. It works like feSquareGeneric.
//
//go:noescape
func feSquare(out *Element, a *Element)

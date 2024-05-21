// Copyright (c) 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !(arm64 && amd64) || !gc || purego
// +build !arm64 !arm64 !amd64 !gc purego

package field

func feMul(v, x, y *Element) { feMulGeneric(v, x, y) }

func feSquare(v, x *Element) { feSquareGeneric(v, x) }

func feSquare2(v, x *Element) { feSquare2Generic(v, x) }

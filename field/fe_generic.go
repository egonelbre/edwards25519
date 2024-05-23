// Copyright (c) 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package field

import "math/bits"

// uint128 holds a 128-bit number as two 64-bit limbs, for use with the
// bits.Mul64 and bits.Add64 intrinsics.
type uint128 struct {
	lo, hi uint64
}

// mul64 returns a * b.
func mul64(a, b uint64) uint128 {
	hi, lo := bits.Mul64(a, b)
	return uint128{lo, hi}
}

// addMul64 returns v + a * b.
func addMul64(v uint128, a, b uint64) uint128 {
	hi, lo := bits.Mul64(a, b)
	lo, c := bits.Add64(lo, v.lo, 0)
	hi, _ = bits.Add64(hi, v.hi, c)
	return uint128{lo, hi}
}

// double128 returns v + a * b.
func double128(v uint128) uint128 {
	c, lo := bits.Add64(v.lo, v.lo, 0)
	_, hi := bits.Add64(v.hi, v.hi, c)
	return uint128{lo, hi}
}

// shiftRightBy51 returns a >> 51. a is assumed to be at most 115 bits.
func shiftRightBy51(a uint128) uint64 {
	return (a.hi << (64 - 51)) | (a.lo >> 51)
}

func feMulGeneric(v, a, b *Element) {
	a0 := a.l0
	a1 := a.l1
	a2 := a.l2
	a3 := a.l3
	a4 := a.l4

	b0 := b.l0
	b1 := b.l1
	b2 := b.l2
	b3 := b.l3
	b4 := b.l4

	// Limb multiplication works like pen-and-paper columnar multiplication, but
	// with 51-bit limbs instead of digits.
	//
	//                          a4   a3   a2   a1   a0  x
	//                          b4   b3   b2   b1   b0  =
	//                         ------------------------
	//                        a4b0 a3b0 a2b0 a1b0 a0b0  +
	//                   a4b1 a3b1 a2b1 a1b1 a0b1       +
	//              a4b2 a3b2 a2b2 a1b2 a0b2            +
	//         a4b3 a3b3 a2b3 a1b3 a0b3                 +
	//    a4b4 a3b4 a2b4 a1b4 a0b4                      =
	//   ----------------------------------------------
	//      r8   r7   r6   r5   r4   r3   r2   r1   r0
	//
	// We can then use the reduction identity (a * 2²⁵⁵ + b = a * 19 + b) to
	// reduce the limbs that would overflow 255 bits. r5 * 2²⁵⁵ becomes 19 * r5,
	// r6 * 2³⁰⁶ becomes 19 * r6 * 2⁵¹, etc.
	//
	// Reduction can be carried out simultaneously to multiplication. For
	// example, we do not compute r5: whenever the result of a multiplication
	// belongs to r5, like a1b4, we multiply it by 19 and add the result to r0.
	//
	//            a4b0    a3b0    a2b0    a1b0    a0b0  +
	//            a3b1    a2b1    a1b1    a0b1 19×a4b1  +
	//            a2b2    a1b2    a0b2 19×a4b2 19×a3b2  +
	//            a1b3    a0b3 19×a4b3 19×a3b3 19×a2b3  +
	//            a0b4 19×a4b4 19×a3b4 19×a2b4 19×a1b4  =
	//           --------------------------------------
	//              r4      r3      r2      r1      r0
	//
	// Finally we add up the columns into wide, overlapping limbs.

	a1_19 := a1 * 19
	a2_19 := a2 * 19
	a3_19 := a3 * 19
	a4_19 := a4 * 19

	// r0 = a0×b0 + 19×(a1×b4 + a2×b3 + a3×b2 + a4×b1)
	r0 := mul64(a0, b0)
	r0 = addMul64(r0, a1_19, b4)
	r0 = addMul64(r0, a2_19, b3)
	r0 = addMul64(r0, a3_19, b2)
	r0 = addMul64(r0, a4_19, b1)

	// r1 = a0×b1 + a1×b0 + 19×(a2×b4 + a3×b3 + a4×b2)
	r1 := mul64(a0, b1)
	r1 = addMul64(r1, a1, b0)
	r1 = addMul64(r1, a2_19, b4)
	r1 = addMul64(r1, a3_19, b3)
	r1 = addMul64(r1, a4_19, b2)

	// r2 = a0×b2 + a1×b1 + a2×b0 + 19×(a3×b4 + a4×b3)
	r2 := mul64(a0, b2)
	r2 = addMul64(r2, a1, b1)
	r2 = addMul64(r2, a2, b0)
	r2 = addMul64(r2, a3_19, b4)
	r2 = addMul64(r2, a4_19, b3)

	// r3 = a0×b3 + a1×b2 + a2×b1 + a3×b0 + 19×a4×b4
	r3 := mul64(a0, b3)
	r3 = addMul64(r3, a1, b2)
	r3 = addMul64(r3, a2, b1)
	r3 = addMul64(r3, a3, b0)
	r3 = addMul64(r3, a4_19, b4)

	// r4 = a0×b4 + a1×b3 + a2×b2 + a3×b1 + a4×b0
	r4 := mul64(a0, b4)
	r4 = addMul64(r4, a1, b3)
	r4 = addMul64(r4, a2, b2)
	r4 = addMul64(r4, a3, b1)
	r4 = addMul64(r4, a4, b0)

	// After the multiplication, we need to reduce (carry) the five coefficients
	// to obtain a result with limbs that are at most slightly larger than 2⁵¹,
	// to respect the Element invariant.
	//
	// Overall, the reduction works the same as carryPropagate, except with
	// wider inputs: we take the carry for each coefficient by shifting it right
	// by 51, and add it to the limb above it. The top carry is multiplied by 19
	// according to the reduction identity and added to the lowest limb.
	//
	// The largest coefficient (r0) will be at most 111 bits, which guarantees
	// that all carries are at most 111 - 51 = 60 bits, which fits in a uint64.
	//
	//     r0 = a0×b0 + 19×(a1×b4 + a2×b3 + a3×b2 + a4×b1)
	//     r0 < 2⁵²×2⁵² + 19×(2⁵²×2⁵² + 2⁵²×2⁵² + 2⁵²×2⁵² + 2⁵²×2⁵²)
	//     r0 < (1 + 19 × 4) × 2⁵² × 2⁵²
	//     r0 < 2⁷ × 2⁵² × 2⁵²
	//     r0 < 2¹¹¹
	//
	// Moreover, the top coefficient (r4) is at most 107 bits, so c4 is at most
	// 56 bits, and c4 * 19 is at most 61 bits, which again fits in a uint64 and
	// allows us to easily apply the reduction identity.
	//
	//     r4 = a0×b4 + a1×b3 + a2×b2 + a3×b1 + a4×b0
	//     r4 < 5 × 2⁵² × 2⁵²
	//     r4 < 2¹⁰⁷
	//

	c0 := shiftRightBy51(r0)
	c1 := shiftRightBy51(r1)
	c2 := shiftRightBy51(r2)
	c3 := shiftRightBy51(r3)
	c4 := shiftRightBy51(r4)

	rr0 := r0.lo&maskLow51Bits + c4*19
	rr1 := r1.lo&maskLow51Bits + c0
	rr2 := r2.lo&maskLow51Bits + c1
	rr3 := r3.lo&maskLow51Bits + c2
	rr4 := r4.lo&maskLow51Bits + c3

	// Now all coefficients fit into 64-bit registers but are still too large to
	// be passed around as an Element. We therefore do one last carry chain,
	// where the carries will be small enough to fit in the wiggle room above 2⁵¹.
	*v = Element{rr0, rr1, rr2, rr3, rr4}
	v.carryPropagate()
}

func feSquareGeneric(v, a *Element) {
	l0 := a.l0
	l1 := a.l1
	l2 := a.l2
	l3 := a.l3
	l4 := a.l4

	// Squaring works precisely like multiplication above, but thanks to its
	// symmetry we get to group a few terms together.
	//
	//                          l4   l3   l2   l1   l0  x
	//                          l4   l3   l2   l1   l0  =
	//                         ------------------------
	//                        l4l0 l3l0 l2l0 l1l0 l0l0  +
	//                   l4l1 l3l1 l2l1 l1l1 l0l1       +
	//              l4l2 l3l2 l2l2 l1l2 l0l2            +
	//         l4l3 l3l3 l2l3 l1l3 l0l3                 +
	//    l4l4 l3l4 l2l4 l1l4 l0l4                      =
	//   ----------------------------------------------
	//      r8   r7   r6   r5   r4   r3   r2   r1   r0
	//
	//            l4l0    l3l0    l2l0    l1l0    l0l0  +
	//            l3l1    l2l1    l1l1    l0l1 19×l4l1  +
	//            l2l2    l1l2    l0l2 19×l4l2 19×l3l2  +
	//            l1l3    l0l3 19×l4l3 19×l3l3 19×l2l3  +
	//            l0l4 19×l4l4 19×l3l4 19×l2l4 19×l1l4  =
	//           --------------------------------------
	//              r4      r3      r2      r1      r0
	//
	// With precomputed 2×, 19×, and 2×19× terms, we can compute each limb with
	// only three Mul64 and four Add64, instead of five and eight.

	l0_2 := l0 * 2
	l1_2 := l1 * 2

	l1_38 := l1 * 38
	l2_38 := l2 * 38
	l3_38 := l3 * 38

	l3_19 := l3 * 19
	l4_19 := l4 * 19

	// r0 = l0×l0 + 19×(l1×l4 + l2×l3 + l3×l2 + l4×l1) = l0×l0 + 19×2×(l1×l4 + l2×l3)
	r0 := mul64(l0, l0)
	r0 = addMul64(r0, l1_38, l4)
	r0 = addMul64(r0, l2_38, l3)

	// r1 = l0×l1 + l1×l0 + 19×(l2×l4 + l3×l3 + l4×l2) = 2×l0×l1 + 19×2×l2×l4 + 19×l3×l3
	r1 := mul64(l0_2, l1)
	r1 = addMul64(r1, l2_38, l4)
	r1 = addMul64(r1, l3_19, l3)

	// r2 = l0×l2 + l1×l1 + l2×l0 + 19×(l3×l4 + l4×l3) = 2×l0×l2 + l1×l1 + 19×2×l3×l4
	r2 := mul64(l0_2, l2)
	r2 = addMul64(r2, l1, l1)
	r2 = addMul64(r2, l3_38, l4)

	// r3 = l0×l3 + l1×l2 + l2×l1 + l3×l0 + 19×l4×l4 = 2×l0×l3 + 2×l1×l2 + 19×l4×l4
	r3 := mul64(l0_2, l3)
	r3 = addMul64(r3, l1_2, l2)
	r3 = addMul64(r3, l4_19, l4)

	// r4 = l0×l4 + l1×l3 + l2×l2 + l3×l1 + l4×l0 = 2×l0×l4 + 2×l1×l3 + l2×l2
	r4 := mul64(l0_2, l4)
	r4 = addMul64(r4, l1_2, l3)
	r4 = addMul64(r4, l2, l2)

	c0 := shiftRightBy51(r0)
	c1 := shiftRightBy51(r1)
	c2 := shiftRightBy51(r2)
	c3 := shiftRightBy51(r3)
	c4 := shiftRightBy51(r4)

	rr0 := r0.lo&maskLow51Bits + c4*19
	rr1 := r1.lo&maskLow51Bits + c0
	rr2 := r2.lo&maskLow51Bits + c1
	rr3 := r3.lo&maskLow51Bits + c2
	rr4 := r4.lo&maskLow51Bits + c3

	c0 = rr0 >> 51
	c1 = rr1 >> 51
	c2 = rr2 >> 51
	c3 = rr3 >> 51
	c4 = rr4 >> 51

	// c4 is at most 64 - 51 = 13 bits, so c4*19 is at most 18 bits, and
	// the final l0 will be at most 52 bits. Similarly for the rest.
	v.l0 = rr0&maskLow51Bits + c4*19
	v.l1 = rr1&maskLow51Bits + c0
	v.l2 = rr2&maskLow51Bits + c1
	v.l3 = rr3&maskLow51Bits + c2
	v.l4 = rr4&maskLow51Bits + c3
}

func feSquare2Generic(v, a *Element) {
	l0 := a.l0
	l1 := a.l1
	l2 := a.l2
	l3 := a.l3
	l4 := a.l4

	// Squaring works precisely like multiplication above, but thanks to its
	// symmetry we get to group a few terms together.
	//
	//                          l4   l3   l2   l1   l0  x
	//                          l4   l3   l2   l1   l0  =
	//                         ------------------------
	//                        l4l0 l3l0 l2l0 l1l0 l0l0  +
	//                   l4l1 l3l1 l2l1 l1l1 l0l1       +
	//              l4l2 l3l2 l2l2 l1l2 l0l2            +
	//         l4l3 l3l3 l2l3 l1l3 l0l3                 +
	//    l4l4 l3l4 l2l4 l1l4 l0l4                      =
	//   ----------------------------------------------
	//      r8   r7   r6   r5   r4   r3   r2   r1   r0
	//
	//            l4l0    l3l0    l2l0    l1l0    l0l0  +
	//            l3l1    l2l1    l1l1    l0l1 19×l4l1  +
	//            l2l2    l1l2    l0l2 19×l4l2 19×l3l2  +
	//            l1l3    l0l3 19×l4l3 19×l3l3 19×l2l3  +
	//            l0l4 19×l4l4 19×l3l4 19×l2l4 19×l1l4  =
	//           --------------------------------------
	//              r4      r3      r2      r1      r0
	//
	// With precomputed 2×, 19×, and 2×19× terms, we can compute each limb with
	// only three Mul64 and four Add64, instead of five and eight.

	l0_2 := l0 * 2
	l1_2 := l1 * 2

	l1_38 := l1 * 38
	l2_38 := l2 * 38
	l3_38 := l3 * 38

	l3_19 := l3 * 19
	l4_19 := l4 * 19

	// r0 = l0×l0 + 19×(l1×l4 + l2×l3 + l3×l2 + l4×l1) = l0×l0 + 19×2×(l1×l4 + l2×l3)
	r0 := mul64(l0, l0)
	r0 = addMul64(r0, l1_38, l4)
	r0 = addMul64(r0, l2_38, l3)

	// r1 = l0×l1 + l1×l0 + 19×(l2×l4 + l3×l3 + l4×l2) = 2×l0×l1 + 19×2×l2×l4 + 19×l3×l3
	r1 := mul64(l0_2, l1)
	r1 = addMul64(r1, l2_38, l4)
	r1 = addMul64(r1, l3_19, l3)

	// r2 = l0×l2 + l1×l1 + l2×l0 + 19×(l3×l4 + l4×l3) = 2×l0×l2 + l1×l1 + 19×2×l3×l4
	r2 := mul64(l0_2, l2)
	r2 = addMul64(r2, l1, l1)
	r2 = addMul64(r2, l3_38, l4)

	// r3 = l0×l3 + l1×l2 + l2×l1 + l3×l0 + 19×l4×l4 = 2×l0×l3 + 2×l1×l2 + 19×l4×l4
	r3 := mul64(l0_2, l3)
	r3 = addMul64(r3, l1_2, l2)
	r3 = addMul64(r3, l4_19, l4)

	// r4 = l0×l4 + l1×l3 + l2×l2 + l3×l1 + l4×l0 = 2×l0×l4 + 2×l1×l3 + l2×l2
	r4 := mul64(l0_2, l4)
	r4 = addMul64(r4, l1_2, l3)
	r4 = addMul64(r4, l2, l2)

	r0 = double128(r0)
	r1 = double128(r1)
	r2 = double128(r2)
	r3 = double128(r3)
	r4 = double128(r4)

	c0 := shiftRightBy51(r0)
	c1 := shiftRightBy51(r1)
	c2 := shiftRightBy51(r2)
	c3 := shiftRightBy51(r3)
	c4 := shiftRightBy51(r4)

	rr0 := r0.lo&maskLow51Bits + c4*19
	rr1 := r1.lo&maskLow51Bits + c0
	rr2 := r2.lo&maskLow51Bits + c1
	rr3 := r3.lo&maskLow51Bits + c2
	rr4 := r4.lo&maskLow51Bits + c3

	c0 = rr0 >> 51
	c1 = rr1 >> 51
	c2 = rr2 >> 51
	c3 = rr3 >> 51
	c4 = rr4 >> 51

	// c4 is at most 64 - 51 = 13 bits, so c4*19 is at most 18 bits, and
	// the final l0 will be at most 52 bits. Similarly for the rest.
	v.l0 = rr0&maskLow51Bits + c4*19
	v.l1 = rr1&maskLow51Bits + c0
	v.l2 = rr2&maskLow51Bits + c1
	v.l3 = rr3&maskLow51Bits + c2
	v.l4 = rr4&maskLow51Bits + c3
}

// carryPropagateGeneric brings the limbs below 52 bits by applying the reduction
// identity (a * 2²⁵⁵ + b = a * 19 + b) to the l4 carry.
func (v *Element) carryPropagateGeneric() *Element {
	c0 := v.l0 >> 51
	c1 := v.l1 >> 51
	c2 := v.l2 >> 51
	c3 := v.l3 >> 51
	c4 := v.l4 >> 51

	// c4 is at most 64 - 51 = 13 bits, so c4*19 is at most 18 bits, and
	// the final l0 will be at most 52 bits. Similarly for the rest.
	v.l0 = v.l0&maskLow51Bits + c4*19
	v.l1 = v.l1&maskLow51Bits + c0
	v.l2 = v.l2&maskLow51Bits + c1
	v.l3 = v.l3&maskLow51Bits + c2
	v.l4 = v.l4&maskLow51Bits + c3

	return v
}

func feSquareGenericAlternate(v, a *Element) {
	var R1, R2, R3, R4, R5, R6, R7, R8, R9, R10, R11, R12, R13, R14, R15, R16, R17, c uint64
	l0, l1 := a.l0, a.l1
	l2, l3 = a.l2, a.l3
	l4 = a.l4
	R3 = l1<<2 + l1
	R13 = l0 << 1
	R17 = l0 * l0
	R9 = l1 << 1
	R2 = l2<<2 + l2
	R3 = R3 << 2
	R3 = l1 - R3
	R7 = l3<<2 + l3
	R2 = R2 << 2
	R15 = l4<<2 + l4
	R2 = l2 - R2
	R3 = R3 << 1
	R15 = R15 << 2
	R14 = l1 * l1
	R1 = R2 << 1
	R2 = R7 << 2
	R7 = l4 * R3
	R2 = l3 - R2
	R3, _ = bits.Mul64(l4, R3)
	R15 = l4 - R15
	R8 = l3 * R1
	R12 = R2 << 1
	R16, _ = bits.Mul64(l3, R1)
	c, R7 = bits.Add64(R8, R7, 0)
	R8, _ = bits.Mul64(l0, l0)
	R3 = R16 + R3 + c
	R6 = l4 * R1
	R16 = l1 * R13
	c, R7 = bits.Add64(R17, R7, 0)
	R3 = R8 + R3 + c
	R1, _ = bits.Mul64(l4, R1)
	R8, _ = bits.Mul64(l1, R13)
	c, R6 = bits.Add64(R16, R6, 0)
	R16 = l3 * R2
	R3 = (R7 << 51) | (R3 >> 51)
	R2, _ = bits.Mul64(l3, R2)
	R1 = R8 + R1 + c
	c, R16 = bits.Add64(R16, R6, 0)
	R8 = l4 * R12
	R6 = R13 * l2
	R1 = R2 + R1 + c
	R12, _ = bits.Mul64(l4, R12)
	R7 = 0x7ffffffffffff & R7
	R2, _ = bits.Mul64(R13, l2)
	c, R8 = bits.Add64(R6, R8, 0)
	R5, _ = bits.Mul64(l1, l1)
	R1 = (R16 << 51) | (R1 >> 51)
	R6 = l2 * R9
	R12 = R2 + R12 + c
	c, R8 = bits.Add64(R14, R8, 0)
	R2 = l3 * R13
	R14, _ = bits.Mul64(l2, R9)
	R12 = R5 + R12 + c
	R17 = l4 * R15
	c, R6 = bits.Add64(R2, R6, 0)
	R5, _ = bits.Mul64(l3, R13)
	R12 = (R8 << 51) | (R12 >> 51)
	R2, _ = bits.Mul64(l4, R15)
	R16 = 0x7ffffffffffff & R16
	R14 = R5 + R14 + c
	c, R15 = bits.Add64(R17, R6, 0)
	R5 = l3 * R9
	R14 = R2 + R14 + c
	R17 = l4 * R13
	R8 = 0x7ffffffffffff & R8
	R6, _ = bits.Mul64(l4, R13)
	R16 = R3 + R16
	R2, _ = bits.Mul64(l3, R9)
	c, R5 = bits.Add64(R17, R5, 0)
	R10 = l2 * l2
	R9 = (R15 << 51) | (R14 >> 51)
	R4, _ = bits.Mul64(l2, l2)
	R2 = R6 + R2 + c
	c, R5 = bits.Add64(R10, R5, 0)
	R8 = R1 + R8
	R2 = R4 + R2 + c
	R4 = 0x7ffffffffffff & R5
	R9 = R9 + R4
	R6 = 0x7ffffffffffff & R15
	R2 = (R5 << 51) | (R2 >> 51)
	R10 = R9 >> 51
	R5 = R12 + R6
	R4 = R2<<2 + R2
	R6 = 0x7ffffffffffff & R8
	R3 = R10<<2 + R10
	R9 = 0x7ffffffffffff & R9
	R1 = R4 << 2
	R4 = 0x7ffffffffffff & R16
	R1 = R2 - R1
	R2 = R3 << 2
	R1 = R7 + R1
	R2 = R10 - R2
	R7 = 0x7ffffffffffff & R5
	R3 = 0x7ffffffffffff & R1
	R5 = R5>>51 + R9
	R8 = R8>>51 + R7
	R16 = R16>>51 + R6
	R1 = R1>>51 + R4
	R2 = R3 + R2
	v.l0, v.l1 = R2, R1
	v.l2, v.l3 = R16, R8
	v.l4 = R5
}
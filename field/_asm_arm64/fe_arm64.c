#include <stdint.h>
#include <stddef.h>

#define uint128_t __uint128_t

typedef struct {
	// An element t represents the integer
	//     t.l0 + t.l1*2^51 + t.l2*2^102 + t.l3*2^153 + t.l4*2^204
	//
	// Between operations, all limbs are expected to be lower than 2^52.
	uint64_t l0;
	uint64_t l1;
	uint64_t l2;
	uint64_t l3;
	uint64_t l4;
} Element;

// uint128 holds a 128-bit number as two 64-bit limbs, for use with the
// bits.Mul64 and bits.Add64 intrinsics.

const uint64_t maskLow51Bits = ((uint64_t)(1) << 51) - 1;

// mul64 returns a * b.
uint128_t mul64(uint64_t a, uint64_t b) {
	return (uint128_t)(a) * (uint128_t)(b);
}

// addMul64 returns v + a * b.
uint128_t addMul64(uint128_t v, uint64_t a, uint64_t b) {
	return v + mul64(a, b);
}

// shiftRightBy51 returns a >> 51. a is assumed to be at most 115 bits.
uint64_t shiftRightBy51(uint128_t a) {
	return (uint64_t)(a >> 51);
}

void feSquareGeneric(Element *v, Element *a) {
	uint64_t l0 = a->l0;
	uint64_t l1 = a->l1;
	uint64_t l2 = a->l2;
	uint64_t l3 = a->l3;
	uint64_t l4 = a->l4;

	// Squaring works precisely like multiplication above, but thanks to its
	// symmetry we get to group a few terms together.
	//
	//                          l4   l3   l2   l1   l0  x
	//                          l4   l3   l2   l1   l0  =;
	//                         ------------------------
	//                        l4l0 l3l0 l2l0 l1l0 l0l0  +
	//                   l4l1 l3l1 l2l1 l1l1 l0l1       +
	//              l4l2 l3l2 l2l2 l1l2 l0l2            +
	//         l4l3 l3l3 l2l3 l1l3 l0l3                 +
	//    l4l4 l3l4 l2l4 l1l4 l0l4                      =;
	//   ----------------------------------------------
	//      r8   r7   r6   r5   r4   r3   r2   r1   r0
	//
	//            l4l0    l3l0    l2l0    l1l0    l0l0  +
	//            l3l1    l2l1    l1l1    l0l1 19×l4l1  +
	//            l2l2    l1l2    l0l2 19×l4l2 19×l3l2  +
	//            l1l3    l0l3 19×l4l3 19×l3l3 19×l2l3  +
	//            l0l4 19×l4l4 19×l3l4 19×l2l4 19×l1l4  =;
	//           --------------------------------------
	//              r4      r3      r2      r1      r0
	//
	// With precomputed 2×, 19×, and 2×19× terms, we can compute each limb with
	// only three Mul64 and four Add64, instead of five and eight.

	uint64_t l0_2 = l0 * 2;
	uint64_t l1_2 = l1 * 2;

	uint64_t l1_38 = l1 * 38;
	uint64_t l2_38 = l2 * 38;
	uint64_t l3_38 = l3 * 38;

	uint64_t l3_19 = l3 * 19;
	uint64_t l4_19 = l4 * 19;

	// r0 = l0×l0 + 19×(l1×l4 + l2×l3 + l3×l2 + l4×l1) = l0×l0 + 19×2×(l1×l4 + l2×l3);
	uint128_t r0 = mul64(l0, l0);
	r0 = addMul64(r0, l1_38, l4);
	r0 = addMul64(r0, l2_38, l3);

	// r1 = l0×l1 + l1×l0 + 19×(l2×l4 + l3×l3 + l4×l2) = 2×l0×l1 + 19×2×l2×l4 + 19×l3×l3;
	uint128_t r1 = mul64(l0_2, l1);
	r1 = addMul64(r1, l2_38, l4);
	r1 = addMul64(r1, l3_19, l3);

	// r2 = l0×l2 + l1×l1 + l2×l0 + 19×(l3×l4 + l4×l3) = 2×l0×l2 + l1×l1 + 19×2×l3×l4;
	uint128_t r2 = mul64(l0_2, l2);
	r2 = addMul64(r2, l1, l1);
	r2 = addMul64(r2, l3_38, l4);

	// r3 = l0×l3 + l1×l2 + l2×l1 + l3×l0 + 19×l4×l4 = 2×l0×l3 + 2×l1×l2 + 19×l4×l4;
	uint128_t r3 = mul64(l0_2, l3);
	r3 = addMul64(r3, l1_2, l2);
	r3 = addMul64(r3, l4_19, l4);

	// r4 = l0×l4 + l1×l3 + l2×l2 + l3×l1 + l4×l0 = 2×l0×l4 + 2×l1×l3 + l2×l2;
	uint128_t r4 = mul64(l0_2, l4);
	r4 = addMul64(r4, l1_2, l3);
	r4 = addMul64(r4, l2, l2);

	uint64_t c0 = shiftRightBy51(r0);
	uint64_t c1 = shiftRightBy51(r1);
	uint64_t c2 = shiftRightBy51(r2);
	uint64_t c3 = shiftRightBy51(r3);
	uint64_t c4 = shiftRightBy51(r4);

	uint64_t rr0 = r0&maskLow51Bits + c4*19;
	uint64_t rr1 = r1&maskLow51Bits + c0;
	uint64_t rr2 = r2&maskLow51Bits + c1;
	uint64_t rr3 = r3&maskLow51Bits + c2;
	uint64_t rr4 = r4&maskLow51Bits + c3;

	v->l0 = rr0;
	v->l1 = rr1;
	v->l2 = rr2;
	v->l3 = rr3;
	v->l4 = rr4;

	c0 = v->l0 >> 51;
	c1 = v->l1 >> 51;
	c2 = v->l2 >> 51;
	c3 = v->l3 >> 51;
	c4 = v->l4 >> 51;

	// c4 is at most 64 - 51 = 13 bits, so c4*19 is at most 18 bits, and;
	// the final l0 will be at most 52 bits. Similarly for the rest.
	v->l0 = v->l0&maskLow51Bits + c4*19;
	v->l1 = v->l1&maskLow51Bits + c0;
	v->l2 = v->l2&maskLow51Bits + c1;
	v->l3 = v->l3&maskLow51Bits + c2;
	v->l4 = v->l4&maskLow51Bits + c3;
}

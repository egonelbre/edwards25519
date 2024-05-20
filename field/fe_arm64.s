// Copyright (c) 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build arm64 && gc && !purego

#include "textflag.h"

// carryPropagate works exactly like carryPropagateGeneric and uses the
// same AND, ADD, and LSR+MADD instructions emitted by the compiler, but
// avoids loading R0-R4 twice and uses LDP and STP.
//
// See https://golang.org/issues/43145 for the main compiler issue.
//
// func carryPropagate(v *Element)
TEXT ·carryPropagate(SB),NOFRAME|NOSPLIT,$0-8
	MOVD v+0(FP), R20

	LDP 0(R20), (R0, R1)
	LDP 16(R20), (R2, R3)
	MOVD 32(R20), R4

	AND $0x7ffffffffffff, R0, R10
	AND $0x7ffffffffffff, R1, R11
	AND $0x7ffffffffffff, R2, R12
	AND $0x7ffffffffffff, R3, R13
	AND $0x7ffffffffffff, R4, R14

	ADD R0>>51, R11, R11
	ADD R1>>51, R12, R12
	ADD R2>>51, R13, R13
	ADD R3>>51, R14, R14
	// R4>>51 * 19 + R10 -> R10
	LSR $51, R4, R21
	MOVD $19, R22
	MADD R22, R10, R21, R10

	STP (R10, R11), 0(R20)
	STP (R12, R13), 16(R20)
	MOVD R14, 32(R20)

	RET

// func feSquare(out *Element, a *Element)
TEXT ·feSquare(SB),NOFRAME|NOSPLIT,$0-16
	MOVD out+0(FP), R0
	MOVD a+0(FP), R1

	LDP 8(R1), (R11, R9)
	MOVW $38, R10
	MOVW $19, R8
	LDP 24(R1), (R12, R14)
	MOVD (R1), R16
	MUL R10, R11, R13
	MUL R10, R9, R15
	MUL R8, R12, R17
	UMULH R16, R16, R1
	MUL R16, R16, R2
	LSL $1, R16, R16
	MUL R15, R12, R5
	UMULH R13, R14, R3
	MUL R13, R14, R13
	ADDS R2, R5, R2
	UMULH R15, R12, R4
	MUL R10, R12, R10
	MUL R16, R11, R7
	ADC R1, R4, R1
	UMULH R15, R14, R19
	ADDS R13, R2, R13
	MUL R15, R14, R15
	ADC R3, R1, R1
	UMULH R16, R11, R6
	UMULH R12, R17, R20
	ADDS R7, R15, R15
	MUL R12, R17, R17
	ADC R6, R19, R6
	MUL R16, R9, R4
	UMULH R11, R11, R2
	ADDS R17, R15, R15
	MUL R11, R11, R3
	LSL $1, R11, R11
	MUL R8, R14, R18
	ADC R20, R6, R6
	UMULH R16, R9, R5
	ADDS R3, R4, R3
	UMULH R10, R14, R7
	MUL R10, R14, R10
	ADC R2, R5, R2
	UMULH R16, R12, R19
	MUL R16, R12, R17
	ADDS R10, R3, R10
	UMULH R11, R9, R20
	ADC R7, R2, R2
	UMULH R11, R12, R4
	MUL R11, R12, R12
	MUL R11, R9, R11
	UMULH R14, R18, R3
	ADDS R11, R17, R11
	MUL R14, R18, R17
	ADC R20, R19, R18
	MUL R16, R14, R5
	ADDS R17, R11, R11
	MUL R9, R9, R17
	UMULH R9, R9, R9
	ADC R3, R18, R18
	ADDS R17, R12, R12
	UMULH R16, R14, R14
	ADC R9, R4, R9
	ADDS R5, R12, R12
	ADC R14, R9, R9
	EXTR $51, R11, R18, R14
	MOVD $2251799813685247, R16
	EXTR $51, R13, R1, R17
	EXTR $51, R12, R9, R9
	EXTR $51, R15, R6, R18
	EXTR $51, R10, R2, R1
	MADD R8, R16, R9, R9
	AND R13, R9, R9
	ADD R16, R14, R13
	AND R12, R13, R12
	ADD R16, R17, R14
	LSR $51, R12, R17
	ADD R16, R18, R13
	AND R10, R13, R10
	ADD R16, R1, R13
	UMADDL R8, R16, R17, R8
	AND R15, R14, R14
	AND R11, R13, R11
	ADD R9>>51, R16, R13
	AND R9, R8, R8
	AND R14, R13, R9
	ADD R14>>51, R16, R13
	ADD R10>>51, R16, R14
	AND R10, R13, R10
	AND R11, R14, R13
	ADD R11>>51, R16, R11
	STP (R8, R9), (R0)
	AND R12, R11, R8
	STP (R10, R13), 16(R0)
	MOVD R8, 32(R0)
	RET
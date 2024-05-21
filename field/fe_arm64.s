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
TEXT ·carryPropagate(SB), NOFRAME|NOSPLIT, $0-8
	MOVD v+0(FP), R20

	LDP  0(R20), (R0, R1)
	LDP  16(R20), (R2, R3)
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
	LSR  $51, R4, R21
	MOVD $19, R22
	MADD R22, R10, R21, R10

	STP  (R10, R11), 0(R20)
	STP  (R12, R13), 16(R20)
	MOVD R14, 32(R20)

	RET

/*
// func feMulX(out *Element, a *Element, b *Element)
TEXT ·feMulX(SB), NOSPLIT, $0-24
	STP.W (R29, R30), -48(RSP)
	MOVD  RSP, R29
	LDP   (R1), (R12, R10)
	STP   (R19, R20), 16(RSP)
	LDP   16(R1), (R7, R4)
	STP   (R21, R22), 32(RSP)
	MOVD  32(R1), R9
	ADD   R10<<2, R10, R3
	LDP   16(R2), (R15, R13)
	ADD   R7<<2, R7, R8
	LDP   (R2), (R11, R16)
	ADD   R4<<2, R4, R18
	ADD   R9<<2, R9, R17
	LSL   $2, R3, R3
	MOVD  32(R2), R6
	LSL   $2, R8, R8
	SUB   R7, R8, R8
	SUB   R10, R3, R3
	LSL   $2, R18, R18
	LSL   $2, R17, R17
	SUB   R9, R17, R17
	SUB   R4, R18, R18
	MUL   R13, R8, R5
	MUL   R6, R3, R1
	UMULH R13, R8, R19
	MUL   R15, R18, R14
	ADDS  R5, R1, R1
	UMULH R6, R3, R2
	MUL   R16, R17, R3
	UMULH R15, R18, R20
	ADC   R19, R2, R2
	UMULH R16, R17, R5
	ADDS  R14, R3, R3
	MUL   R11, R12, R19
	ADC   R20, R5, R5
	UMULH R11, R12, R20
	ADDS  R3, R1, R1
	MUL   R6, R8, R14
	ADC   R5, R2, R2
	ADDS  R19, R1, R1
	ADC   R20, R2, R2
	MUL   R13, R18, R3
	MUL   R12, R16, R20
	UMULH R6, R8, R8
	ADDS  R14, R3, R3
	UMULH R13, R18, R30
	EXTR  $51, R1, R2, R14
	MUL   R11, R10, R5
	UMULH R11, R10, R19
	ADC   R8, R30, R30
	UMULH R12, R16, R2
	ADDS  R20, R5, R5
	MUL   R15, R17, R20
	AND   $0x7ffffffffffff, R1, R8
	UMULH R15, R17, R21
	ADC   R2, R19, R1
	ADDS  R5, R3, R3
	MUL   R6, R18, R19
	ADC   R1, R30, R2
	MUL   R13, R17, R5
	ADDS  R20, R3, R3
	UMULH R13, R17, R1
	ADC   R21, R2, R2
	UMULH R6, R18, R18
	MUL   R12, R15, R20
	ADDS  R19, R5, R5
	MUL   R16, R10, R30
	EXTR  $51, R3, R2, R19
	ADC   R18, R1, R18
	UMULH R12, R15, R2
	UMULH R16, R10, R1
	ADDS  R20, R30, R30
	MUL   R11, R7, R20
	AND   $0x7ffffffffffff, R3, R3
	ADC   R2, R1, R1
	ADDS  R30, R5, R5
	UMULH R11, R7, R21
	ADC   R1, R18, R18
	MUL   R11, R4, R2
	ADDS  R20, R5, R5
	MUL   R16, R7, R1
	ADC   R21, R18, R18
	UMULH R16, R7, R20
	ADD   R14, R3, R3
	UMULH R11, R4, R30
	ADDS  R1, R2, R2
	MUL   R12, R13, R14
	EXTR  $51, R5, R18, R18
	MUL   R15, R10, R1
	ADC   R20, R30, R30
	UMULH R12, R13, R22
	AND   $0x7ffffffffffff, R5, R5
	UMULH R15, R10, R20
	ADDS  R14, R1, R1
	MUL   R6, R17, R21
	ADD   R19, R5, R14
	UMULH R6, R17, R17
	ADC   R22, R20, R19
	ADDS  R1, R2, R2
	MUL   R15, R7, R20
	ADC   R19, R30, R30
	MUL   R16, R4, R5
	ADDS  R21, R2, R2
	UMULH R16, R4, R1
	UMULH R15, R7, R7
	ADC   R17, R30, R15
	MUL   R13, R10, R16
	ADDS  R20, R5, R5
	MUL   R12, R6, R17
	ADC   R7, R1, R1
	UMULH R13, R10, R4
	EXTR  $51, R2, R15, R7
	UMULH R12, R6, R6
	ADDS  R17, R16, R16
	MUL   R11, R9, R10
	AND   $0x7ffffffffffff, R2, R2
	ADC   R6, R4, R4
	UMULH R11, R9, R9
	ADDS  R16, R5, R5
	ADD   R18, R2, R2
	ADC   R4, R1, R1
	ADDS  R10, R5, R5
	ADC   R9, R1, R1
	AND   $0x7ffffffffffff, R5, R4
	ADD   R7, R4, R4
	AND   $0x7ffffffffffff, R2, R11
	AND   $0x7ffffffffffff, R4, R6
	EXTR  $51, R5, R1, R1
	LSR   $51, R4, R7
	AND   $0x7ffffffffffff, R14, R10
	ADD   R1<<2, R1, R4
	AND   $0x7ffffffffffff, R3, R9
	ADD   R7<<2, R7, R5
	ADD   R2>>51, R6, R2
	LSL   $2, R4, R4
	ADD   R14>>51, R11, R14
	SUB   R1, R4, R1
	LSL   $2, R5, R4
	ADD   R8, R1, R1
	SUB   R7, R4, R4
	AND   $0x7ffffffffffff, R1, R5
	ADD   R3>>51, R10, R3
	LDP   16(RSP), (R19, R20)
	ADD   R1>>51, R9, R1
	LDP   32(RSP), (R21, R22)
	ADD   R5, R4, R4
	LDP.P 48(RSP), (R29, R30)
	STP   (R4, R1), (R0)
	STP   (R3, R14), 16(R0)
	MOVD  R2, 32(R0)
	RET
*/

// func feSquare(out *Element, a *Element)
TEXT ·feSquare(SB), NOSPLIT, $0-16
	MOVD  out+0(FP), R0
	MOVD  v+8(FP), R1

	LDP   (R1), (R6, R5)
	LDP   16(R1), (R4, R11)
	MOVD  32(R1), R10
	ADD   R5<<2, R5, R3
	LSL   $1, R6, R13
	MUL   R6, R6, R17
	LSL   $1, R5, R9
	ADD   R4<<2, R4, R2
	LSL   $2, R3, R3
	SUB   R5, R3, R3
	ADD   R11<<2, R11, R7
	LSL   $2, R2, R2
	ADD   R10<<2, R10, R15
	SUB   R4, R2, R2
	LSL   $1, R3, R3
	LSL   $2, R15, R15
	MUL   R5, R5, R14
	LSL   $1, R2, R1
	LSL   $2, R7, R2
	MUL   R10, R3, R7
	SUB   R11, R2, R2
	UMULH R10, R3, R3
	SUB   R10, R15, R15
	MUL   R11, R1, R8
	LSL   $1, R2, R12
	UMULH R11, R1, R16
	ADDS  R8, R7, R7
	UMULH R6, R6, R8
	ADC   R16, R3, R3
	MUL   R10, R1, R6
	MUL   R5, R13, R16
	ADDS  R17, R7, R7
	ADC   R8, R3, R3
	UMULH R10, R1, R1
	UMULH R5, R13, R8
	ADDS  R16, R6, R6
	MUL   R11, R2, R16
	EXTR  $51, R7, R3, R3
	UMULH R11, R2, R2
	ADC   R8, R1, R1
	ADDS  R16, R6, R16
	MUL   R10, R12, R8
	MUL   R13, R4, R6
	ADC   R2, R1, R1
	UMULH R10, R12, R12
	AND   $0x7ffffffffffff, R7, R7
	UMULH R13, R4, R2
	ADDS  R6, R8, R8
	UMULH R5, R5, R5
	EXTR  $51, R16, R1, R1
	MUL   R4, R9, R6
	ADC   R2, R12, R12
	ADDS  R14, R8, R8
	MUL   R11, R13, R2
	UMULH R4, R9, R14
	ADC   R5, R12, R12
	MUL   R10, R15, R17
	ADDS  R2, R6, R6
	UMULH R11, R13, R5
	EXTR  $51, R8, R12, R12
	UMULH R10, R15, R2
	AND   $0x7ffffffffffff, R16, R16
	ADC   R5, R14, R14
	ADDS  R17, R6, R15
	MUL   R11, R9, R5
	ADC   R2, R14, R14
	MUL   R10, R13, R17
	AND   $0x7ffffffffffff, R8, R8
	UMULH R10, R13, R6
	ADD   R3, R16, R16
	UMULH R11, R9, R2
	ADDS  R17, R5, R5
	MUL   R4, R4, R10
	EXTR  $51, R15, R14, R9
	UMULH R4, R4, R4
	ADC   R6, R2, R2
	ADDS  R10, R5, R5
	ADD   R1, R8, R8
	ADC   R4, R2, R2
	AND   $0x7ffffffffffff, R5, R4
	ADD   R9, R4, R9
	AND   $0x7ffffffffffff, R15, R6
	EXTR  $51, R5, R2, R2
	LSR   $51, R9, R10
	ADD   R12, R6, R5
	ADD   R2<<2, R2, R4
	AND   $0x7ffffffffffff, R8, R6
	ADD   R10<<2, R10, R3
	AND   $0x7ffffffffffff, R9, R9
	LSL   $2, R4, R1
	AND   $0x7ffffffffffff, R16, R4
	SUB   R2, R1, R1
	LSL   $2, R3, R2
	ADD   R7, R1, R1
	SUB   R10, R2, R2
	AND   $0x7ffffffffffff, R5, R7
	AND   $0x7ffffffffffff, R1, R3
	ADD   R5>>51, R9, R5
	ADD   R8>>51, R7, R8
	ADD   R16>>51, R6, R16
	ADD   R1>>51, R4, R1
	ADD   R3, R2, R2
	STP   (R2, R1), (R0)
	STP   (R16, R8), 16(R0)
	MOVD  R5, 32(R0)
	RET

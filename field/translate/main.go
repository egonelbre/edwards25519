package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"os"
	"regexp"
	"strings"
)

func main() {
	var b bytes.Buffer
	translate(&b)

	data, err := format.Source(b.Bytes())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		fmt.Fprintln(os.Stderr, b.String())
	}

	fmt.Println(string(data))
}

func translate(b io.Writer) {
	fmt.Fprintf(b, "package field\n\n")
	fmt.Fprintf(b, `import "math/bits"`+"\n\n")

	fmt.Fprintf(b, `type Element struct { l0, l1, l2, l3, l4 uint64 }`+"\n\n")

	fmt.Fprintf(b, "func Square(v, a *Element) {\n")
	fmt.Fprintf(b, "var R1, R2, R3, R4, R5, R6, R7, R8, R9, R10, R11, R12, R13, R14, R15, R16, R17, c uint64\n")

	defer fmt.Fprintf(b, "}\n")

	for _, line := range strings.Split(asm, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var roff, ra, rb, rout, rout2 string
		switch {
		case match(line, `RET`):
		case match(line, `LDP \(%R\), \(%R, %R\)`, &ra, &rout, &rout2):
			roff = "0"
			fallthrough

		case match(line, `LDP %D\(%R\), \(%R, %R\)`, &roff, &ra, &rout, &rout2):
			switch [2]string{roff, ra} {
			case [2]string{"0", "R1"}:
				fmt.Fprintf(b, "%v, %v = a.l0, a.l1\n", rout, rout2)
			case [2]string{"16", "R1"}:
				fmt.Fprintf(b, "%v, %v = a.l2, a.l3\n", rout, rout2)
			default:
				fmt.Fprintf(b, "%v, %v = ldp(%v, %v)\n", rout, rout2, ra, roff)
			}

		case match(line, `MOVD %D\(%R\), %R`, &roff, &ra, &rout):
			switch [2]string{roff, ra} {
			case [2]string{"32", "R1"}:
				fmt.Fprintf(b, "%v = a.l4\n", rout)
			default:
				fmt.Fprintf(b, "%v = load(%v, %v)\n", rout, ra, roff)
			}

		case match(line, `STP \(%R, %R\), \(%R\)`, &ra, &rb, &rout):
			roff = "0"
			fallthrough
		case match(line, `STP \(%R, %R\), %D\(%R\)`, &ra, &rb, &roff, &rout):
			switch [2]string{roff, rout} {
			case [2]string{"0", "R0"}:
				fmt.Fprintf(b, "v.l0, v.l1 = %v, %v\n", ra, rb)
			case [2]string{"16", "R0"}:
				fmt.Fprintf(b, "v.l2, v.l3 = %v, %v\n", ra, rb)
			default:
				fmt.Fprintf(b, "offset(%v, %v) = %v, %v\n", rout, roff, ra, rb)
			}

		case match(line, `MOVD %R, %D\(%R\)`, &ra, &roff, &rout):
			switch [2]string{roff, rout} {
			case [2]string{"32", "R0"}:
				fmt.Fprintf(b, "v.l4 = %v\n", ra)
			default:
				fmt.Fprintf(b, "offset(%v, %v) = %v\n", rout, roff, ra)
			}

		case match(line, `ADD %R<<2, %R, %R`, &ra, &rb, &rout):
			fmt.Fprintf(b, "%v = %v<<2 + %v\n", rout, ra, rb)

		case match(line, `ADD %R>>51, %R, %R`, &ra, &rb, &rout):
			fmt.Fprintf(b, "%v = %v>>51 + %v\n", rout, ra, rb)

		case match(line, `LSL \$%D, %R, %R`, &roff, &ra, &rout):
			fmt.Fprintf(b, "%v = %v<<%v\n", rout, ra, roff)

		case match(line, `LSR \$%D, %R, %R`, &roff, &ra, &rout):
			fmt.Fprintf(b, "%v = %v>>%v\n", rout, ra, roff)

		case match(line, `MUL %R, %R, %R`, &ra, &rb, &rout):
			// fmt.Fprintf(b,"_, %v = bits.Mul64(%v, %v)\n", rout, ra, rb)
			fmt.Fprintf(b, "%v = %v * %v\n", rout, ra, rb)

		case match(line, `UMULH %R, %R, %R`, &ra, &rb, &rout):
			fmt.Fprintf(b, "%v, _ = bits.Mul64(%v, %v)\n", rout, ra, rb)

		case match(line, `SUB %R, %R, %R`, &ra, &rb, &rout):
			fmt.Fprintf(b, "%v = %v - %v\n", rout, ra, rb)

		case match(line, `ADD %R, %R, %R`, &ra, &rb, &rout):
			fmt.Fprintf(b, "%v = %v + %v\n", rout, ra, rb)

		case match(line, `AND %X, %R, %R`, &ra, &rb, &rout):
			fmt.Fprintf(b, "%v = %v & %v\n", rout, ra, rb)

		case match(line, `ADDS %R, %R, %R`, &ra, &rb, &rout):
			fmt.Fprintf(b, "c, %v = bits.Add64(%v, %v, 0)\n", rout, ra, rb)

		case match(line, `ADC %R, %R, %R`, &ra, &rb, &rout):
			// fmt.Fprintf(b, "_, %v = bits.Add64(%v, %v, c)\n", rout, ra, rb)
			fmt.Fprintf(b, "%v = %v + %v + c\n", rout, ra, rb)

		//  EXTR  $51, R5, R2, R2
		case match(line, `EXTR \$%D, %R, %R, %R`, &roff, &ra, &rb, &rout):
			fmt.Fprintf(b, "%v = (%v << %v) | (%v >> %v)\n", rout, ra, roff, rb, roff)

		default:
			fmt.Fprintf(b, "// ?? %v\n", line)
		}
	}
}

var cache = map[string]*regexp.Regexp{}

func match(line, regex string, args ...*string) bool {
	rx, ok := cache[regex]
	if !ok {
		x := strings.ReplaceAll(regex, "%R", "(R\\d+)")
		x = strings.ReplaceAll(x, "%D", "(\\d+)")
		x = strings.ReplaceAll(x, "%X", "\\$(0x[0-9a-f]+)")
		x = strings.ReplaceAll(x, " ", "\\s*")
		rx = regexp.MustCompile(x)
		cache[regex] = rx
	}

	vals := rx.FindStringSubmatch(line)
	if len(vals) == 0 {
		return false
	}

	if len(vals) != len(args)+1 {
		panic("wrong amount of args")
	}

	for i, v := range vals[1:] {
		*args[i] = v
	}
	return true
}

const asm = `
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
	RET`

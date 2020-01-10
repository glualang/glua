package parser

import (
	"strconv"
	"strings"
)

type instruction uint32

func isConstant(x int) bool   { return 0 != x&bitRK }
func constantIndex(r int) int { return r & ^bitRK }
func asConstant(r int) int    { return r | bitRK }

// creates a mask with 'n' 1 bits at position 'p'
func mask1(n, p uint) instruction { return ^(^instruction(0) << n) << p }

// creates a mask with 'n' 0 bits at position 'p'
func mask0(n, p uint) instruction { return ^mask1(n, p) }

func (i instruction) opCode() opCode         { return opCode(i >> posOp & (1<<sizeOp - 1)) }
func (i instruction) arg(pos, size uint) int { return int(i >> pos & mask1(size, 0)) }
func (i *instruction) setOpCode(op opCode)   { i.setArg(posOp, sizeOp, int(op)) }
func (i *instruction) setArg(pos, size uint, arg int) {
	*i = *i&mask0(size, pos) | instruction(arg)<<pos&mask1(size, pos)
}

// Note: the gc optimizer cannot inline through multiple function calls. Manually inline for now.
// func (i instruction) a() int   { return i.arg(posA, sizeA) }
// func (i instruction) b() int   { return i.arg(posB, sizeB) }
// func (i instruction) c() int   { return i.arg(posC, sizeC) }
// func (i instruction) bx() int  { return i.arg(posBx, sizeBx) }
// func (i instruction) ax() int  { return i.arg(posAx, sizeAx) }
// func (i instruction) sbx() int { return i.bx() - maxArgSBx }

func (i instruction) a() int   { return int(i >> posA & maxArgA) }
func (i instruction) b() int   { return int(i >> posB & maxArgB) }
func (i instruction) c() int   { return int(i >> posC & maxArgC) }
func (i instruction) bx() int  { return int(i >> posBx & maxArgBx) }
func (i instruction) ax() int  { return int(i >> posAx & maxArgAx) }
func (i instruction) sbx() int { return int(i>>posBx&maxArgBx) - maxArgSBx }

func (i *instruction) setA(arg int)   { i.setArg(posA, sizeA, arg) }
func (i *instruction) setB(arg int)   { i.setArg(posB, sizeB, arg) }
func (i *instruction) setC(arg int)   { i.setArg(posC, sizeC, arg) }
func (i *instruction) setBx(arg int)  { i.setArg(posBx, sizeBx, arg) }
func (i *instruction) setAx(arg int)  { i.setArg(posAx, sizeAx, arg) }
func (i *instruction) setSBx(arg int) { i.setArg(posBx, sizeBx, arg+maxArgSBx) }

func createABC(op opCode, a, b, c int) instruction {
	return instruction(op)<<posOp |
		instruction(a)<<posA |
		instruction(b)<<posB |
		instruction(c)<<posC
}

func createABx(op opCode, a, bx int) instruction {
	return instruction(op)<<posOp |
		instruction(a)<<posA |
		instruction(bx)<<posBx
}

func createAx(op opCode, a int) instruction { return instruction(op)<<posOp | instruction(a)<<posAx }

func (i instruction) String(proto *Prototype, indexInProtoCode int) string {
	op := i.opCode()
	opName := opNames[op]
	opName = strings.ToLower(opName)
	s := opName

	opCount := Opcounts[op]
	opInfo := Opinfos[op]
	var operand int
	var useExtended bool = false
	//var extraArg bool = false
	for j:=0;j<opCount;j++ {
		limit := opInfo[j].limit
		switch opInfo[j].pos {
		case OPP_A:
			operand = i.a()
		case OPP_B:
			operand = i.b()
		case OPP_C:
			operand = i.c()
		case OPP_Bx:
			operand = i.bx()
		case OPP_Ax:
			operand = i.ax()
		case OPP_sBx:
			operand = i.sbx()
		case OPP_ARG:
			useExtended = true
		case OPP_C_ARG:
			operand = i.c()
			if operand == 0 {
				useExtended = true //try get next
			}
		}
		if useExtended { //try get next ins
			return "" // TODO
		}

		switch limit {
		case LIMIT_STACKIDX:
			s = s + " %" + strconv.Itoa(int(operand))
		case LIMIT_UPVALUE:
			s = s + " @" + strconv.Itoa(int(operand))
		case LIMIT_EMBED:
			s = s + " " + strconv.Itoa(int(operand))
		case LIMIT_CONSTANT:
			constIdx := constantIndex(operand)
			constVal := proto.constants[constIdx]
			constLiteral, ok := literalValueString(constVal)
			if !ok {
				return "invalid constant literal"
			}
			s = s + " const " + constLiteral
		case LIMIT_LOCATION:
			loclabel := proto.extra.labelLocations[int(operand)+1+indexInProtoCode]
			s = s + " $" + loclabel
		case LIMIT_CONST_STACK:
			if (int(operand) & BITRK) > 0 {
				cconstIdx := constantIndex(operand)
				constVal := proto.constants[cconstIdx]
				constLiteral, ok := literalValueString(constVal)
				if !ok {
					return "invalid constant literal"
				}
				s = s + " const " + constLiteral
			} else {
				s = s + " %" + strconv.Itoa(int(operand))
			}

		case LIMIT_PROTO:
			subProto := proto.prototypes[operand]
			s = s + " " + subProto.name
		}
	}

	return s
}

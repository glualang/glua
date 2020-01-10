package parser

import (
	"fmt"
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
	// TODO: use opcounts and OpInfos to dump to asm
	switch opMode(op) {
	case iABC:
		s = fmt.Sprintf("%s %%%d", s, i.a())
		// TODO: 这里有BUG，比如settabup，settabup %0 @256 const "Person1" 其中%0应该是 第二项@0代币_ENV，@256应该是第3项const ...，最后const "Person1"应该是%0
		if bMode(op) != opArgN {
			if cMode(op) == opArgN {
				// c不存在时
				s = fmt.Sprintf("%s %s", s, proto.opArgToAsmItemString(op, i.b(), bMode(op)))
			} else {
				// c存在时b按register/upvalue处理，而不作为常量
				if op == opGetUpValue || op == opSetUpValue || op == opGetTableUp || op == opSetTableUp {
					s = fmt.Sprintf("%s @%d", s, i.b())
				} else if bMode(op) == opArgU {
					s = fmt.Sprintf("%s %%%d", s, i.b())
				}
			}
		}
		if cMode(op) != opArgN {
			s = fmt.Sprintf("%s %s", s, proto.opArgToAsmItemString(op, i.c(), cMode(op)))
		}

		// TODO: 这里不应该这么硬编码
		switch op {
		case opSetTableUp:
			s = fmt.Sprintf("%s %%%d @%d %s", opName, i.a(), i.c(), proto.opArgToAsmItemString(op, i.b(), bMode(op)))
		case opSetTable:
			s = fmt.Sprintf("%s %%%d %s %s", opName, i.a(), proto.opArgToAsmItemString(op, i.b(), bMode(op)), proto.opArgToAsmItemString(op, i.c(), cMode(op)))
		case opNewTable:
			s = fmt.Sprintf("%s %%%d %d %d", opName, i.a(), i.b(), i.c())
		case opCall:
			s = fmt.Sprintf("%s %%%d %d %d", opName, i.a(), i.b(), i.c())
		case opReturn:
			s = fmt.Sprintf("%s %%%d %d", opName, i.a(), i.b())
		case opTest:
			s = fmt.Sprintf("%s %%%d %d", opName, i.a(), i.b())
		}


	case iAsBx:
		s = fmt.Sprintf("%s %%%d", s, i.a())
		if bMode(op) != opArgN {
			s = fmt.Sprintf("%s %d", s, i.sbx())
		}
		switch op {
		case opJump:
			// TODO: jmp 1 $labelName 这种格式，需要在proto中找到i.sbx() + i所在行数对应的location的label name
			instOffset := i.sbx()
			jmpDest := instOffset + 1 + indexInProtoCode
			label, ok := proto.extra.labelLocations[jmpDest]
			if !ok {
				s = fmt.Sprintf("%s invalid_location", opName)
				return s
			}
			s = fmt.Sprintf("%s %d $%s", opName, i.a(), label)
		}
	case iABx:
		s = fmt.Sprintf("%s %%%d", s, i.a())
		if bMode(op) != opArgN {
			s = fmt.Sprintf("%s %s", s, proto.opArgToAsmItemString(op, i.bx(), bMode(op)))
		}
	case iAx:
		s = fmt.Sprintf("%s %%%d", s, i.ax())
	}
	return s
}

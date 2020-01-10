package parser

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

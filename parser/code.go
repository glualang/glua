package parser

import (
	"fmt"
	"math"
)

const (
	oprMinus = iota
	oprNot
	oprLength
	oprBnot

	oprNoUnary
)

const (
	noJump            = -1
	noRegister        = maxArgA
	maxLocalVariables = 200
)

const (
	// 需要和opcodes中相对顺序一致，因为用到了int(opr - oprAdd)的偏移量
	oprAdd = iota
	oprSub
	oprMul
	oprMod
	oprPow
	oprDiv
	oprIdiv

	oprBand
	oprBor
	oprBxor
	oprShl
	oprShr

	oprConcat
	oprEq
	oprLT
	oprLE
	oprNE
	oprGT
	oprGE
	oprAnd
	oprOr
	oprNoBinary
)

const (
	kindVoid = iota // no value
	kindNil
	kindTrue
	kindFalse
	kindConstant       // info = index of constant
	kindInt            // intValue = integer value
	kindNumber         // value = numerical value
	kindNonRelocatable // info = result register
	kindLocal          // info = local register
	kindUpValue        // info = index of upvalue
	kindIndexed        // table = table register/upvalue, index = register/constant index
	kindJump           // info = instruction pc
	kindRelocatable    // info = instruction pc 表示这个表达式的目标寄存器(第一个instruction op arg)需要之后重定位，暂时可能被设置为其他占位符比如0
	kindCall           // info = instruction pc
	kindVarArg         // info = instruction pc
)

var kinds []string = []string{
	"void",
	"nil",
	"true",
	"false",
	"constant",
	"number",
	"nonrelocatable",
	"local",
	"upvalue",
	"indexed",
	"jump",
	"relocatable",
	"call",
	"vararg",
}

type exprDesc struct {
	kind      int
	index     int // register/constant index
	table     int // register or upvalue
	tableType int // whether 'table' is register (kindLocal) or upvalue (kindUpValue)
	info      int
	// t代表to, f代表from， f表示这个表达式计算完成后目标是回到e.f(比如c = a + b在a+b整个指令块计算完成后都要做赋值给c的操作)
	// t代表此表达式的下一步是否跳转以及跳转目标是哪条指令
	t, f      int // patch lists for 'exit when true/false'
	value     float64
	intValue  int64

	symbol string // 当是单符号变量时用这个
}

func (e *exprDesc) isZero() bool {
	switch e.kind {
	case kindInt:
		return e.intValue == 0
	case kindNumber:
		return e.value == 0.0
	default:
		return false
	}
}

func (e *exprDesc) floatValue() float64 {
	switch e.kind {
	case kindInt:
		return float64(e.intValue)
	case kindNumber:
		return e.value
	default:
		return 0.0
	}
}

type assignmentTarget struct {
	previous *assignmentTarget
	exprDesc
}

type label struct {
	name                string
	pc, line            int
	activeVariableCount int
}

type block struct {
	previous              *block
	firstLabel, firstGoto int
	activeVariableCount   int
	hasUpValue, isLoop    bool
}

type function struct {
	constantLookup      map[value]int
	f                   *Prototype
	previous            *function
	p                   *parser
	block               *block
	jumpPC, lastTarget  int
	freeRegisterCount   int // 使用中的寄存器数量
	activeVariableCount int
	firstLocal          int
}

func (f *function) OpenFunction(line int) {
	f.f.prototypes = append(f.f.prototypes, Prototype{
		source:       f.p.source,
		maxStackSize: 2,
		lineDefined:  line,
		name:         genPrototypeName(),
		extra:        NewPrototypeExtra(),
	})
	f.p.function = &function{f: &f.f.prototypes[len(f.f.prototypes)-1], constantLookup: make(map[value]int), previous: f, p: f.p, jumpPC: noJump, firstLocal: len(f.p.activeVariables)}
	f.p.function.EnterBlock(false)
}

func (f *function) CloseFunction() exprDesc {
	e := f.previous.ExpressionToNextRegister(makeExpression(kindRelocatable, f.previous.encodeABx(opClosure, 0, len(f.previous.f.prototypes)-1)))
	f.ReturnNone()
	f.LeaveBlock()
	f.assert(f.block == nil)
	f.p.function = f.previous
	return e
}

func (f *function) EnterBlock(isLoop bool) {
	// TODO www.lua.org uses a trick here to stack allocate the block, and chain blocks in the stack
	f.block = &block{previous: f.block, firstLabel: len(f.p.activeLabels), firstGoto: len(f.p.pendingGotos), activeVariableCount: f.activeVariableCount, isLoop: isLoop}
	f.assert(f.freeRegisterCount == f.activeVariableCount)
}

func (f *function) undefinedGotoError(g label) {
	if isReserved(g.name) {
		f.semanticError(fmt.Sprintf("<%s> at Line %d not inside a loop", g.name, g.line))
	} else {
		f.semanticError(fmt.Sprintf("no visible label '%s' for <goto> at Line %d", g.name, g.line))
	}
}

func (f *function) LocalVariable(i int) *localVariable {
	index := f.p.activeVariables[f.firstLocal+i]
	return &f.f.localVariables[index]
}

func (f *function) AdjustLocalVariables(n int) {
	for f.activeVariableCount += n; n != 0; n-- {
		f.LocalVariable(f.activeVariableCount - n).startPC = pc(len(f.f.code))
	}
}

func (f *function) removeLocalVariables(level int) {
	for i := level; i < f.activeVariableCount; i++ {
		f.LocalVariable(i).endPC = pc(len(f.f.code))
	}
	f.p.activeVariables = f.p.activeVariables[:len(f.p.activeVariables)-(f.activeVariableCount-level)]
	f.activeVariableCount = level
}

func (f *function) MakeLocalVariable(name string) {
	r := len(f.f.localVariables)
	f.f.localVariables = append(f.f.localVariables, localVariable{name: name})
	f.p.checkLimit(len(f.p.activeVariables)+1-f.firstLocal, maxLocalVariables, "local variables")
	f.p.activeVariables = append(f.p.activeVariables, r)
}

func (f *function) MakeGoto(name string, line, pc int) {
	f.p.pendingGotos = append(f.p.pendingGotos, label{name: name, line: line, pc: pc, activeVariableCount: f.activeVariableCount})
	f.findLabel(len(f.p.pendingGotos) - 1)
}

func (f *function) MakeLabel(name string, line int) int {
	f.p.activeLabels = append(f.p.activeLabels, label{name: name, line: line, pc: len(f.f.code), activeVariableCount: f.activeVariableCount})
	return len(f.p.activeLabels) - 1
}

func (f *function) closeGoto(i int, l label) {
	g := f.p.pendingGotos[i]
	if f.assert(g.name == l.name); g.activeVariableCount < l.activeVariableCount {
		f.semanticError(fmt.Sprintf("<goto %s> at Line %d jumps into the scope of local '%s'", g.name, g.line, f.LocalVariable(g.activeVariableCount).name))
	}
	f.PatchList(g.pc, l.pc)
	copy(f.p.pendingGotos[i:], f.p.pendingGotos[i+1:])
	f.p.pendingGotos = f.p.pendingGotos[:len(f.p.pendingGotos)-1]
}

func (f *function) findLabel(i int) int {
	g, b := f.p.pendingGotos[i], f.block
	for _, l := range f.p.activeLabels[b.firstLabel:] {
		if l.name == g.name {
			if g.activeVariableCount > l.activeVariableCount && (b.hasUpValue || len(f.p.activeLabels) > b.firstLabel) {
				f.PatchClose(g.pc, l.activeVariableCount)
			}
			f.closeGoto(i, l)
			return 0
		}
	}
	return 1
}

func (f *function) CheckRepeatedLabel(name string) {
	for _, l := range f.p.activeLabels[f.block.firstLabel:] {
		if l.name == name {
			f.semanticError(fmt.Sprintf("label '%s' already defined on Line %d", name, l.line))
		}
	}
}

func (f *function) FindGotos(label int) {
	for i, l := f.block.firstGoto, f.p.activeLabels[label]; i < len(f.p.pendingGotos); {
		if f.p.pendingGotos[i].name == l.name {
			f.closeGoto(i, l)
		} else {
			i++
		}
	}
}

func (f *function) moveGotosOut(b block) {
	for i := b.firstGoto; i < len(f.p.pendingGotos); i += f.findLabel(i) {
		if f.p.pendingGotos[i].activeVariableCount > b.activeVariableCount {
			if b.hasUpValue {
				f.PatchClose(f.p.pendingGotos[i].pc, b.activeVariableCount)
			}
			f.p.pendingGotos[i].activeVariableCount = b.activeVariableCount
		}
	}
}

func (f *function) LeaveBlock() {
	b := f.block
	if b.previous != nil && b.hasUpValue { // create a 'jump to here' to close upvalues
		j := f.Jump()
		f.PatchClose(j, b.activeVariableCount)
		f.PatchToHere(j)
	}
	if b.isLoop {
		f.breakLabel() // close pending breaks
	}
	f.block = b.previous
	f.removeLocalVariables(b.activeVariableCount)
	f.assert(b.activeVariableCount == f.activeVariableCount)
	f.freeRegisterCount = f.activeVariableCount
	f.p.activeLabels = f.p.activeLabels[:b.firstLabel]
	if b.previous != nil { // inner block
		f.moveGotosOut(*b) // update pending gotos to outer block
	} else if b.firstGoto < len(f.p.pendingGotos) { // pending gotos in outer block
		f.undefinedGotoError(f.p.pendingGotos[b.firstGoto])
	}
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func not(b int) int {
	if b == 0 {
		return 1
	}
	return 0
}

func makeExpression(kind, info int) exprDesc {
	return exprDesc{f: noJump, t: noJump, kind: kind, info: info}
}

func (f *function) semanticError(message string) {
	f.p.t = 0 // remove "near to" from final message
	f.p.syntaxError(message)
}

func (f *function) breakLabel()                         { f.FindGotos(f.MakeLabel("break", 0)) }
func (f *function) unreachable()                        { f.assert(false) }
func (f *function) assert(cond bool)                    { lua_assert(cond) }
func (f *function) Instruction(e exprDesc) *instruction { return &f.f.code[e.info] }
func (e exprDesc) hasJumps() bool                       { return e.t != e.f }
func (e exprDesc) isNumeral() bool {
	return (e.kind == kindNumber || e.kind == kindInt) && e.t == noJump && e.f == noJump
}                                           // TODO: 需要适配整数 kindInt
func (e exprDesc) isVariable() bool         { return kindLocal <= e.kind && e.kind <= kindIndexed }
func (e exprDesc) hasMultipleReturns() bool { return e.kind == kindCall || e.kind == kindVarArg }

func (f *function) assertEqual(a, b interface{}) {
	if a != b {
		panic(fmt.Sprintf("%v != %v", a, b))
	}
}

func (f *function) encode(i instruction) int {
	f.assert(len(f.f.code) == len(f.f.lineInfo))
	f.dischargeJumpPC()
	f.f.code = append(f.f.code, i)
	f.f.lineInfo = append(f.f.lineInfo, int32(f.p.lastLine))
	return len(f.f.code) - 1
}

func (f *function) dropLastInstruction() {
	f.assert(len(f.f.code) == len(f.f.lineInfo))
	f.f.code = f.f.code[:len(f.f.code)-1]
	f.f.lineInfo = f.f.lineInfo[:len(f.f.lineInfo)-1]
}

func (f *function) EncodeABC(op opCode, a, b, c int) int {
	f.assert(opMode(op) == iABC)
	f.assert(bMode(op) != opArgN || b == 0)
	f.assert(cMode(op) != opArgN || c == 0)
	f.assert(a <= maxArgA && b <= maxArgB && c <= maxArgC)
	return f.encode(createABC(op, a, b, c))
}

func (f *function) encodeABx(op opCode, a, bx int) int {
	f.assert(opMode(op) == iABx || opMode(op) == iAsBx)
	f.assert(cMode(op) == opArgN)
	f.assert(a <= maxArgA && bx <= maxArgBx)
	return f.encode(createABx(op, a, bx))
}

func (f *function) encodeAsBx(op opCode, a, sbx int) int { return f.encodeABx(op, a, sbx+maxArgSBx) }

func (f *function) encodeExtraArg(a int) int {
	f.assert(a <= maxArgAx)
	return f.encode(createAx(opExtraArg, a))
}

func (f *function) EncodeConstant(r, constant int) int {
	if constant <= maxArgBx {
		return f.encodeABx(opLoadConstant, r, constant)
	}
	pc := f.encodeABx(opLoadConstant, r, 0)
	f.encodeExtraArg(constant)
	return pc
}

func (f *function) EncodeString(s string) exprDesc {
	return makeExpression(kindConstant, f.stringConstant(s))
}

func (f *function) loadNil(from, n int) {
	if len(f.f.code) > f.lastTarget { // no jumps to current position
		if previous := &f.f.code[len(f.f.code)-1]; previous.opCode() == opLoadNil {
			if pf, pl, l := previous.a(), previous.a()+previous.b(), from+n-1; pf <= from && from <= pl+1 || from <= pf && pf <= l+1 { // can connect both
				from, l = min(from, pf), max(l, pl)
				previous.setA(from)
				previous.setB(l - from)
				return
			}
		}
	}
	f.EncodeABC(opLoadNil, from, n-1, 0)
}

func (f *function) Jump() int {
	f.assert(f.isJumpListWalkable(f.jumpPC))
	jumpPC := f.jumpPC
	f.jumpPC = noJump
	return f.Concatenate(f.encodeAsBx(opJump, 0, noJump), jumpPC)
}

func (f *function) JumpTo(target int)             { f.PatchList(f.Jump(), target) }
func (f *function) ReturnNone()                   { f.EncodeABC(opReturn, 0, 1, 0) }
func (f *function) SetMultipleReturns(e exprDesc) { f.setReturns(e, MultipleReturns) }

// return语句的指令生成
func (f *function) Return(e exprDesc, resultCount int) {
	if e.hasMultipleReturns() {
		if f.SetMultipleReturns(e); e.kind == kindCall && resultCount == 1 {
			f.Instruction(e).setOpCode(opTailCall)
			f.assert(f.Instruction(e).a() == f.activeVariableCount)
		}
		f.EncodeABC(opReturn, f.activeVariableCount, MultipleReturns+1, 0)
	} else if resultCount == 1 {
		f.EncodeABC(opReturn, f.ExpressionToAnyRegister(e).info, 1+1, 0)
	} else {
		_ = f.ExpressionToNextRegister(e)
		f.assert(resultCount == f.freeRegisterCount-f.activeVariableCount)
		f.EncodeABC(opReturn, f.activeVariableCount, resultCount+1, 0)
	}
}

func (f *function) conditionalJump(op opCode, a, b, c int) int {
	f.EncodeABC(op, a, b, c)
	return f.Jump()
}

func (f *function) fixJump(pc, dest int) {
	f.assert(f.isJumpListWalkable(pc))
	f.assert(dest != noJump)
	offset := dest - (pc + 1)
	if abs(offset) > maxArgSBx {
		f.p.syntaxError("control structure too long")
	}
	f.f.code[pc].setSBx(offset)
}

func (f *function) Label() int {
	// label定义处之前的本proto的指令数量就是本label的开始位置，也就是跳转到这个label对应的目标. goto label1 对应的指令是 jmp (label1.target - pc)
	f.lastTarget = len(f.f.code)
	return f.lastTarget
}

// 返回要跳转到哪条指令(返回值是本proto中指令的索引)
func (f *function) jump(pc int) int {
	f.assert(f.isJumpListWalkable(pc))
	if offset := f.f.code[pc].sbx(); offset != noJump {
		return pc + 1 + offset
	}
	return noJump
}

func (f *function) isJumpListWalkable(list int) bool {
	if list == noJump {
		return true
	}
	if list < 0 || list >= len(f.f.code) {
		return false
	}
	offset := f.f.code[list].sbx()
	return offset == noJump || f.isJumpListWalkable(list+1+offset)
}

// 返回空值这条jmp指令的指令（比如JMP指令前一条的TEST/TESTSET指令，或者JMP指令本身）.
// JMP指令出现一般两种情况，条件比较指令后的跳转（TEST/TESTSET指令后紧跟着JMP指令）或者goto语句（直接生成一个JMP指令）
func (f *function) jumpControl(pc int) *instruction {
	if pc >= 1 && testTMode(f.f.code[pc-1].opCode()) {
		return &f.f.code[pc-1]
	}
	return &f.f.code[pc]
}

func (f *function) needValue(list int) bool {
	f.assert(f.isJumpListWalkable(list))
	for ; list != noJump; list = f.jump(list) {
		if f.jumpControl(list).opCode() != opTestSet {
			return true
		}
	}
	return false
}

func (f *function) patchTestRegister(node, register int) bool {
	if i := f.jumpControl(node); i.opCode() != opTestSet {
		return false
	} else if register != noRegister && register != i.b() {
		i.setA(register)
	} else {
		*i = createABC(opTest, i.b(), 0, i.c())
	}
	return true
}

func (f *function) removeValues(list int) {
	f.assert(f.isJumpListWalkable(list))
	for ; list != noJump; list = f.jump(list) {
		_ = f.patchTestRegister(list, noRegister)
	}
}

func (f *function) patchListHelper(list, target, register, defaultTarget int) {
	f.assert(f.isJumpListWalkable(list))
	for list != noJump {
		next := f.jump(list)
		if f.patchTestRegister(list, register) {
			f.fixJump(list, target)
		} else {
			f.fixJump(list, defaultTarget)
		}
		list = next
	}
}

func (f *function) dischargeJumpPC() {
	f.assert(f.isJumpListWalkable(f.jumpPC))
	f.patchListHelper(f.jumpPC, len(f.f.code), noRegister, len(f.f.code))
	f.jumpPC = noJump
}

func (f *function) PatchList(list, target int) {
	if target == len(f.f.code) {
		f.PatchToHere(list)
	} else {
		f.assert(target < len(f.f.code))
		f.patchListHelper(list, target, noRegister, target)
	}
}

func (f *function) PatchClose(list, level int) {
	f.assert(f.isJumpListWalkable(list))
	for level, next := level+1, 0; list != noJump; list = next {
		next = f.jump(list)
		f.assert(f.f.code[list].opCode() == opJump && f.f.code[list].a() == 0 || f.f.code[list].a() >= level)
		f.f.code[list].setA(level)
	}
}

func (f *function) PatchToHere(list int) {
	f.assert(f.isJumpListWalkable(list))
	f.assert(f.isJumpListWalkable(f.jumpPC))
	f.Label()
	f.jumpPC = f.Concatenate(f.jumpPC, list)
	f.assert(f.isJumpListWalkable(f.jumpPC))
}

// 如果l1和l2都是跳转指令，合并两个跳转指令
func (f *function) Concatenate(l1, l2 int) int {
	f.assert(f.isJumpListWalkable(l1))
	switch {
	case l2 == noJump:
	case l1 == noJump:
		return l2
	default:
		list := l1
		// 合并 jmp label1; label1: jmp label2; label2: ... 这类的连续跳转的指令
		for next := f.jump(list); next != noJump; list, next = next, f.jump(next) {
		}
		// 修改list指令(是一个JMP指令)的跳转目标为l2
		f.fixJump(list, l2)
	}
	return l1
}

// 在proto的常量池中增加一个新常量，返回此常量在常量池中索引
func (f *function) addConstant(k, v value) int {
	if index, ok := f.constantLookup[k]; ok && f.f.constants[index] == v {
		return index
	}
	index := len(f.f.constants)
	f.constantLookup[k] = index
	f.f.constants = append(f.f.constants, v)
	return index
}

func (f *function) IntConstant(i int64) int {
	return f.addConstant(i, i)
}

func (f *function) NumberConstant(n float64) int {
	if n == 0.0 || math.IsNaN(n) {
		return f.addConstant(math.Float64bits(n), n)
	}
	return f.addConstant(n, n)
}

// 检查本proto的栈是否还能增加{n}个寄存器
func (f *function) CheckStack(n int) {
	if n += f.freeRegisterCount; n >= maxStack {
		f.p.syntaxError("function or expression too complex")
	} else if n > f.f.maxStackSize {
		f.f.maxStackSize = n
	}
}

// 分配{n}个寄存器，用来在接下来使用
func (f *function) ReserveRegisters(n int) {
	f.CheckStack(n)
	f.freeRegisterCount += n
}

// 释放一个寄存器（一定从最顶部的寄存器开始释放）
func (f *function) freeRegister(r int) {
	if !isConstant(r) && r >= f.activeVariableCount {
		f.freeRegisterCount--
		f.assertEqual(r, f.freeRegisterCount)
	}
}

// 表达式处理完了，释放占用的(寄存器)资源
func (f *function) freeExpression(e exprDesc) {
	if e.kind == kindNonRelocatable {
		f.freeRegister(e.info)
	}
}

func (f *function) stringConstant(s string) int { return f.addConstant(s, s) }
func (f *function) booleanConstant(b bool) int  { return f.addConstant(b, b) }
func (f *function) nilConstant() int            { return f.addConstant(f, nil) }

func (f *function) setReturns(e exprDesc, resultCount int) {
	if e.kind == kindCall {
		f.Instruction(e).setC(resultCount + 1)
	} else if e.kind == kindVarArg {
		f.Instruction(e).setB(resultCount + 1)
		f.Instruction(e).setA(f.freeRegisterCount)
		f.ReserveRegisters(1)
	}
}

func (f *function) SetReturn(e exprDesc) exprDesc {
	if e.kind == kindCall {
		e.kind, e.info = kindNonRelocatable, f.Instruction(e).a()
	} else if e.kind == kindVarArg {
		f.Instruction(e).setB(2)
		e.kind = kindRelocatable
	}
	return e
}

// 如果表达式是变量（局部变量，全局变量,upvalue变量，...等），把对这个变量的使用转变边不依赖变量名称的使用
func (f *function) DischargeVariables(e exprDesc) exprDesc {
	switch e.kind {
	case kindLocal:
		e.kind = kindNonRelocatable
	case kindUpValue:
		e.kind, e.info = kindRelocatable, f.EncodeABC(opGetUpValue, 0, e.info, 0)
	case kindIndexed:
		if f.freeRegister(e.index); e.tableType == kindLocal {
			f.freeRegister(e.table)
			e.kind, e.info = kindRelocatable, f.EncodeABC(opGetTable, 0, e.table, e.index)
		} else {
			e.kind, e.info = kindRelocatable, f.EncodeABC(opGetTableUp, 0, e.table, e.index)
		}
	case kindVarArg, kindCall:
		e = f.SetReturn(e)
	}
	return e
}

// 把表达式转成结果存入指定寄存器{r}的指令（如果表达式使用了变量或者常量资源的话）
func (f *function) dischargeToRegister(e exprDesc, r int) exprDesc {
	switch e = f.DischargeVariables(e); e.kind {
	case kindNil:
		f.loadNil(r, 1)
	case kindFalse:
		f.EncodeABC(opLoadBool, r, 0, 0)
	case kindTrue:
		f.EncodeABC(opLoadBool, r, 1, 0)
	case kindConstant:
		f.EncodeConstant(r, e.info)
	case kindInt:
		f.EncodeConstant(r, f.IntConstant(e.intValue))
	case kindNumber:
		f.EncodeConstant(r, f.NumberConstant(e.value))
	case kindRelocatable:
		f.Instruction(e).setA(r)
	case kindNonRelocatable:
		if r != e.info {
			f.EncodeABC(opMove, r, e.info, 0)
		}
	default:
		f.assert(e.kind == kindVoid || e.kind == kindJump)
		return e
	}
	e.kind, e.info = kindNonRelocatable, r
	return e
}

// 把表达式转成结果存入新分配寄存器的指令（如果表达式使用了变量或者常量资源的话）
func (f *function) dischargeToAnyRegister(e exprDesc) exprDesc {
	if e.kind != kindNonRelocatable {
		f.ReserveRegisters(1)
		e = f.dischargeToRegister(e, f.freeRegisterCount-1)
	}
	return e
}

func (f *function) encodeLabel(a, b, jump int) int {
	f.Label()
	return f.EncodeABC(opLoadBool, a, b, jump)
}

// 把表达式转成结果存入指定寄存器{r}的指令（如果需要生成指令的话）
func (f *function) expressionToRegister(e exprDesc, r int) exprDesc {
	if e = f.dischargeToRegister(e, r); e.kind == kindJump {
		e.t = f.Concatenate(e.t, e.info)
	}
	if e.hasJumps() {
		loadFalse, loadTrue := noJump, noJump
		if f.needValue(e.t) || f.needValue(e.f) {
			jump := noJump
			if e.kind != kindJump {
				jump = f.Jump()
			}
			loadFalse, loadTrue = f.encodeLabel(r, 0, 1), f.encodeLabel(r, 1, 0)
			f.PatchToHere(jump)
		}
		end := f.Label()
		f.patchListHelper(e.f, end, r, loadFalse)
		f.patchListHelper(e.t, end, r, loadTrue)
	}
	e.f, e.t, e.info, e.kind = noJump, noJump, r, kindNonRelocatable
	return e
}

func (f *function) ExpressionToNextRegister(e exprDesc) exprDesc {
	e = f.DischargeVariables(e)
	f.freeExpression(e)
	f.ReserveRegisters(1)
	return f.expressionToRegister(e, f.freeRegisterCount-1)
}

// 把表达式转成结果存入新分配寄存器的指令（如果需要生成指令的话）
func (f *function) ExpressionToAnyRegister(e exprDesc) exprDesc {
	if e = f.DischargeVariables(e); e.kind == kindNonRelocatable {
		if !e.hasJumps() {
			return e
		}
		if e.info >= f.activeVariableCount {
			return f.expressionToRegister(e, e.info)
		}
	}
	return f.ExpressionToNextRegister(e)
}

func (f *function) ExpressionToAnyRegisterOrUpValue(e exprDesc) exprDesc {
	if e.kind != kindUpValue || e.hasJumps() {
		e = f.ExpressionToAnyRegister(e)
	}
	return e
}

func (f *function) ExpressionToValue(e exprDesc) exprDesc {
	if e.hasJumps() {
		return f.ExpressionToAnyRegister(e)
	}
	return f.DischargeVariables(e)
}

func (f *function) expressionToRegisterOrConstant(e exprDesc) (exprDesc, int) {
	switch e = f.ExpressionToValue(e); e.kind {
	case kindTrue, kindFalse:
		if len(f.f.constants) <= maxIndexRK {
			e.info, e.kind = f.booleanConstant(e.kind == kindTrue), kindConstant
			return e, asConstant(e.info)
		}
	case kindNil:
		if len(f.f.constants) <= maxIndexRK {
			e.info, e.kind = f.nilConstant(), kindConstant
			return e, asConstant(e.info)
		}
	case kindInt, kindNumber:
		if e.kind == kindInt {
			e.info, e.kind = f.IntConstant(e.intValue), kindConstant
		} else if e.kind == kindNumber {
			e.info, e.kind = f.NumberConstant(e.value), kindConstant
		}
		fallthrough
	case kindConstant:
		if e.info <= maxIndexRK {
			return e, asConstant(e.info)
		}
	}
	e = f.ExpressionToAnyRegister(e)
	return e, e.info
}

// 给变量分配新值的语句的指令生成
func (f *function) StoreVariable(v, e exprDesc) {
	switch v.kind {
	case kindLocal:
		f.freeExpression(e)
		f.expressionToRegister(e, v.info)
		return
	case kindUpValue:
		e = f.ExpressionToAnyRegister(e)
		f.EncodeABC(opSetUpValue, e.info, v.info, 0)
	case kindIndexed:
		var r int
		e, r = f.expressionToRegisterOrConstant(e)
		if v.tableType == kindLocal {
			f.EncodeABC(opSetTable, v.table, v.index, r)
		} else {
			f.EncodeABC(opSetTableUp, v.table, v.index, r)
		}
	default:
		f.unreachable()
	}
	f.freeExpression(e)
}

// 把key表达式作为e表达式（比如一个a:b(args)的成员函数调用表达式）的self参数来使用.
// e的指令生成，需要先生成self参数的构造，然后把剩余args参数的压栈指令构造，然后才是构造函数调用的指令
func (f *function) Self(e, key exprDesc) exprDesc {
	e = f.ExpressionToAnyRegister(e)
	r := e.info
	f.freeExpression(e)
	result := exprDesc{info: f.freeRegisterCount, kind: kindNonRelocatable} // base register for opSelf
	f.ReserveRegisters(2)                                                   // function and 'self' produced by opSelf
	key, k := f.expressionToRegisterOrConstant(key)
	f.EncodeABC(opSelf, result.info, r, k)
	f.freeExpression(key)
	return result
}

func (f *function) invertJump(pc int) {
	i := f.jumpControl(pc)
	lua_assert(testTMode(i.opCode()) && i.opCode() != opTestSet && i.opCode() != opTest)
	i.setA(not(i.a()))
}

// 生成如果{e}表达式不匹配cond(一般是1 真或者0 假),就R(A) = e.info，否则跳过下一条指令. 对应TEST/TESTSET指令
func (f *function) jumpOnCondition(e exprDesc, cond int) int {
	if e.kind == kindRelocatable {
		if i := f.Instruction(e); i.opCode() == opNot {
			f.dropLastInstruction() // remove previous opNot
			// 生成TEST指令和接着的JMP指令
			return f.conditionalJump(opTest, i.b(), 0, not(cond))
		}
	}
	e = f.dischargeToAnyRegister(e)
	f.freeExpression(e)
	// 生成 TESTSET指令和接着的JMP指令
	return f.conditionalJump(opTestSet, noRegister, e.info, cond)
}

// 生成如果{e}表达式是真值，就跳转到目标位置(根据e类型和e.info来确定)的指令
// 生成的指令对应伪代码类似. if e then e.info else pc++
func (f *function) GoIfTrue(e exprDesc) exprDesc {
	pc := noJump // pc保存在{e}为真时下一步指令的位置. 也就是{e}为真就跳转到这个{pc}位置
	switch e = f.DischargeVariables(e); e.kind {
	case kindJump:
		f.invertJump(e.info)
		pc = e.info
	case kindConstant, kindInt, kindNumber, kindTrue:
	default:
		// 如果{e}是真值，则R(A) = e.info,否则跳过下一条指令
		pc = f.jumpOnCondition(e, 0) // pc是生成的最后一个JMP指令的索引. 而最后一个JMP指令的目标是这块指令在{e}为真后跳转的目标位置
	}
	e.f = f.Concatenate(e.f, pc) // 本表达式被使用时，如果是跳转来的，跳转来源是 e.f和pc合并JMP后的结果
	f.PatchToHere(e.t) // 生成一个label，用来作为刚生成的JMP指令的跳转目标. 因为刚生成的JMP指令的目标还没填
	e.t = noJump
	return e
}

// 生成如果{e}表达式是假值，就跳转到目标位置(根据e类型和e.info来确定)的指令
// 生成的指令对应伪代码类似:
// 指令1: 检查 {e} 是否为假，如果{e}是假值，则R(A)=e.info，否则跳过下一条指令
// 指令2: JMP {占位符}, 跳转目标是{e}为假时的处理逻辑
// 指令3: {e}为真时的处理逻辑
func (f *function) GoIfFalse(e exprDesc) exprDesc {
	pc := noJump // pc指向注释中的指令2
	switch e = f.DischargeVariables(e); e.kind {
	case kindJump:
		pc = e.info
	case kindNil, kindFalse:
	default:
		pc = f.jumpOnCondition(e, 1)
	}
	// TODO: 这里需要加上注释
	e.t = f.Concatenate(e.t, pc)
	f.PatchToHere(e.f)
	e.f = noJump
	return e
}

func (f *function) encodeLikeNot(e exprDesc, op opCode) exprDesc {
	switch e = f.DischargeVariables(e); e.kind {
	case kindNil, kindFalse:
		e.kind = kindTrue
	case kindConstant, kindInt, kindNumber, kindTrue:
		e.kind = kindFalse
	case kindJump:
		f.invertJump(e.info)
	case kindRelocatable, kindNonRelocatable:
		e = f.dischargeToAnyRegister(e)
		f.freeExpression(e)
		e.info, e.kind = f.EncodeABC(op, 0, e.info, 0), kindRelocatable
	default:
		f.unreachable()
	}
	e.f, e.t = e.t, e.f
	f.removeValues(e.f)
	f.removeValues(e.t)
	return e
}

func (f *function) encodeNot(e exprDesc) exprDesc {
	return f.encodeLikeNot(e, opNot)
}

func (f *function) encodeBnot(e exprDesc) exprDesc {
	return f.encodeLikeNot(e, opBnot)
}

func (f *function) Indexed(t, k exprDesc) (r exprDesc) {
	f.assert(!t.hasJumps())
	r = makeExpression(kindIndexed, 0)
	r.table = t.info
	_, r.index = f.expressionToRegisterOrConstant(k)
	if t.kind == kindUpValue {
		r.tableType = kindUpValue
	} else {
		f.assert(t.kind == kindNonRelocatable || t.kind == kindLocal)
		r.tableType = kindLocal
	}
	return
}

// 编译器可以处理的常量表达式，比如 1 + 2这样的，编译器合并成改成单个常量
func foldConstants(op opCode, e1, e2 exprDesc) (exprDesc, bool) {
	if !e1.isNumeral() || !e2.isNumeral() {
		return e1, false
	} else if (op == opDiv || op == opIDiv || op == opMod) && e2.isZero() {
		return e1, false
	}
	arithOp := Operator(op-opAdd) + OpAdd
	if e1.kind == kindInt && e2.kind == kindInt {
		e1.intValue = intArith(arithOp, e1.intValue, e2.intValue)
	} else if e1.kind == kindInt {
		e1.intValue = int64(arith(arithOp, e1.floatValue(), e2.floatValue()))
	} else {
		e1.value = arith(arithOp, e1.floatValue(), e2.floatValue())
	}
	return e1, true
}

// 数学运算表达式的指令生成
func (f *function) encodeArithmetic(op opCode, e1, e2 exprDesc, line int) exprDesc {
	if e, folded := foldConstants(op, e1, e2); folded {
		return e
	}
	o2 := 0
	if op != opUnaryMinus && op != opLength {
		e2, o2 = f.expressionToRegisterOrConstant(e2)
	}
	e1, o1 := f.expressionToRegisterOrConstant(e1)
	if o1 > o2 {
		f.freeExpression(e1)
		f.freeExpression(e2)
	} else {
		f.freeExpression(e2)
		f.freeExpression(e1)
	}
	e1.info, e1.kind = f.EncodeABC(op, 0, o1, o2), kindRelocatable
	f.FixLine(line)
	return e1
}

// 单元运算符的表达式的指令的生成，比如: -a, not a, #a, ~a
func (f *function) Prefix(op int, e exprDesc, line int) exprDesc {
	switch op {
	case oprMinus:
		if e.kind == kindInt {
			e.intValue = -e.intValue
			return e
		}
		if e.isNumeral() {
			e.value = -e.value
			return e
		}
		return f.encodeArithmetic(opUnaryMinus, f.ExpressionToAnyRegister(e), makeExpression(kindNumber, 0), line)
	case oprNot:
		return f.encodeNot(e)
	case oprLength:
		return f.encodeArithmetic(opLength, f.ExpressionToAnyRegister(e), makeExpression(kindNumber, 0), line) // TODO: 考虑 kindInt
	case oprBnot:
		return f.encodeBnot(e)
	}
	panic("unreachable")
}

// 二元运算符的表达式的第一个操作数表达式的指令生成，比如 a and b 相当于 if a then <result> = a else <result> = b end. 实际生成时用TEST/TEST+JMP指令来生成
func (f *function) Infix(op int, e exprDesc) exprDesc {
	switch op {
	case oprAnd:
		e = f.GoIfTrue(e)
	case oprOr:
		e = f.GoIfFalse(e)
	case oprConcat:
		e = f.ExpressionToNextRegister(e)
	case oprAdd, oprSub, oprMul, oprDiv, oprMod, oprPow:
		if !e.isNumeral() {
			e, _ = f.expressionToRegisterOrConstant(e)
		}
	default:
		e, _ = f.expressionToRegisterOrConstant(e)
	}
	return e
}

// 比较表达式的指令生成
func (f *function) encodeComparison(op opCode, cond int, e1, e2 exprDesc) exprDesc {
	e1, o1 := f.expressionToRegisterOrConstant(e1)
	e2, o2 := f.expressionToRegisterOrConstant(e2)
	f.freeExpression(e2)
	f.freeExpression(e1)
	if cond == 0 && op != opEqual {
		o1, o2, cond = o2, o1, 1
	}
	return makeExpression(kindJump, f.conditionalJump(op, cond, o1, o2))
}

// 表达式在处理完参数后的处理，比如 a + b在加载完参数后调用opcode计算结果存入寄存器
func (f *function) Postfix(op int, e1, e2 exprDesc, line int) exprDesc {
	switch op {
	case oprAnd:
		f.assert(e1.t == noJump)
		e2 = f.DischargeVariables(e2)
		e2.f = f.Concatenate(e2.f, e1.f) // a and b 表达式，b部分的结果是给a的来源的.也就是把e2.f也设置为e1.f，也就是e.f。表达式作为一个整体，计算完成后结果交给e.f
		return e2
	case oprOr:
		f.assert(e1.f == noJump)
		e2 = f.DischargeVariables(e2)
		e2.t = f.Concatenate(e2.t, e1.t)
		return e2
	case oprConcat:
		if e2 = f.ExpressionToValue(e2); e2.kind == kindRelocatable && f.Instruction(e2).opCode() == opConcat {
			f.assert(e1.info == f.Instruction(e2).b()-1)
			f.freeExpression(e1)
			f.Instruction(e2).setB(e1.info)
			return makeExpression(kindRelocatable, e2.info)
		}
		return f.encodeArithmetic(opConcat, e1, f.ExpressionToNextRegister(e2), line)
	case oprAdd, oprSub, oprMul, oprDiv, oprIdiv, oprMod, oprPow, oprBand, oprBor, oprBxor, oprShl, oprShr:
		return f.encodeArithmetic(opCode(op-oprAdd)+opAdd, e1, e2, line)
	case oprEq, oprLT, oprLE:
		return f.encodeComparison(opCode(op-oprEq)+opEqual, 1, e1, e2)
	case oprNE, oprGT, oprGE:
		return f.encodeComparison(opCode(op-oprNE)+opEqual, 0, e1, e2)
	}
	panic("unreachable")
}

func (f *function) FixLine(line int) { f.f.lineInfo[len(f.f.code)-1] = int32(line) }

func (f *function) setList(base, elementCount, storeCount int) {
	if f.assert(storeCount != 0); storeCount == MultipleReturns {
		storeCount = 0
	}
	if c := (elementCount-1)/listItemsPerFlush + 1; c <= maxArgC {
		f.EncodeABC(opSetList, base, storeCount, c)
	} else if c <= maxArgAx {
		f.EncodeABC(opSetList, base, storeCount, 0)
		f.encodeExtraArg(c)
	} else {
		f.p.syntaxError("constructor too long")
	}
	f.freeRegisterCount = base + 1
}

func (f *function) CheckConflict(t *assignmentTarget, e exprDesc) {
	extra, conflict := f.freeRegisterCount, false
	for ; t != nil; t = t.previous {
		if t.kind == kindIndexed {
			if t.tableType == e.kind && t.table == e.info {
				conflict = true
				t.table, t.tableType = extra, kindLocal
			}
			if e.kind == kindLocal && t.index == e.info {
				conflict = true
				t.index = extra
			}
		}
	}
	if conflict {
		if e.kind == kindLocal {
			f.EncodeABC(opMove, extra, e.info, 0)
		} else {
			f.EncodeABC(opGetUpValue, extra, e.info, 0)
		}
		f.ReserveRegisters(1)
	}
}

// 变量赋值语句的后续处理（比如变量数量比赋值的值的数量多，多出来的变量用nil值填充
func (f *function) AdjustAssignment(variableCount, expressionCount int, e exprDesc) {
	if extra := variableCount - expressionCount; e.hasMultipleReturns() {
		if extra++; extra < 0 {
			extra = 0
		}
		if f.setReturns(e, extra); extra > 1 {
			f.ReserveRegisters(extra - 1)
		}
	} else {
		if expressionCount > 0 {
			_ = f.ExpressionToNextRegister(e)
		}
		if extra > 0 {
			r := f.freeRegisterCount
			f.ReserveRegisters(extra)
			f.loadNil(r, extra)
		}
	}
}

func (f *function) makeUpValue(name string, e exprDesc) int {
	f.p.checkLimit(len(f.f.upValues)+1, maxUpValue, "upvalues")
	f.f.upValues = append(f.f.upValues, upValueDesc{name: name, isLocal: e.kind == kindLocal, index: e.info})
	return len(f.f.upValues) - 1
}

func singleVariableHelper(f *function, name string, base bool) (e exprDesc, found bool) {
	owningBlock := func(b *block, level int) *block {
		for b.activeVariableCount > level {
			b = b.previous
		}
		return b
	}
	find := func() (int, bool) {
		for i := f.activeVariableCount - 1; i >= 0; i-- {
			if name == f.LocalVariable(i).name {
				return i, true
			}
		}
		return 0, false
	}
	findUpValue := func() (int, bool) {
		for i, u := range f.f.upValues {
			if u.name == name {
				return i, true
			}
		}
		return 0, false
	}
	if f == nil {
		return
	}
	var v int
	if v, found = find(); found {
		if e = makeExpression(kindLocal, v); !base {
			owningBlock(f.block, v).hasUpValue = true
		}
		e.symbol = name
		return
	}
	if v, found = findUpValue(); found {
		e = makeExpression(kindUpValue, v)
		e.symbol = name
		return e, true
	}
	if e, found = singleVariableHelper(f.previous, name, false); !found {
		return
	}
	e = makeExpression(kindUpValue, f.makeUpValue(name, e))
	e.symbol = name
	return e, true
}

func (f *function) SingleVariable(name string) (e exprDesc) {
	var found bool
	if e, found = singleVariableHelper(f, name, true); !found {
		e, found = singleVariableHelper(f, "_ENV", true)
		f.assert(found && (e.kind == kindLocal || e.kind == kindUpValue))
		e = f.Indexed(e, f.EncodeString(name))
	}
	return
}

func (f *function) OpenConstructor() (pc int, t exprDesc) {
	pc = f.EncodeABC(opNewTable, 0, 0, 0)
	t = f.ExpressionToNextRegister(makeExpression(kindRelocatable, pc))
	return
}

func (f *function) FlushFieldToConstructor(tableRegister, freeRegisterCount int, k exprDesc, v func() exprDesc) {
	_, rk := f.expressionToRegisterOrConstant(k)
	_, rv := f.expressionToRegisterOrConstant(v())
	f.EncodeABC(opSetTable, tableRegister, rk, rv)
	f.freeRegisterCount = freeRegisterCount
}

func (f *function) FlushToConstructor(tableRegister, pending, arrayCount int, e exprDesc) int {
	f.ExpressionToNextRegister(e)
	if pending == listItemsPerFlush {
		f.setList(tableRegister, arrayCount, listItemsPerFlush)
		pending = 0
	}
	return pending
}

func (f *function) CloseConstructor(pc, tableRegister, pending, arrayCount, hashCount int, e exprDesc) {
	if pending != 0 {
		if e.hasMultipleReturns() {
			f.SetMultipleReturns(e)
			f.setList(tableRegister, arrayCount, MultipleReturns)
			arrayCount--
		} else {
			if e.kind != kindVoid {
				f.ExpressionToNextRegister(e)
			}
			f.setList(tableRegister, arrayCount, pending)
		}
	}
	f.f.code[pc].setB(int(float8FromInt(arrayCount)))
	f.f.code[pc].setC(int(float8FromInt(hashCount)))
}

func (f *function) OpenForBody(base, n int, isNumeric bool) (prep int) {
	if isNumeric {
		prep = f.encodeAsBx(opForPrep, base, noJump)
	} else {
		prep = f.Jump()
	}
	f.EnterBlock(false)
	f.AdjustLocalVariables(n)
	f.ReserveRegisters(n)
	return
}

func (f *function) CloseForBody(prep, base, line, n int, isNumeric bool) {
	f.LeaveBlock()
	f.PatchToHere(prep)
	var end int
	if isNumeric {
		end = f.encodeAsBx(opForLoop, base, noJump)
	} else {
		f.EncodeABC(opTForCall, base, 0, n)
		f.FixLine(line)
		end = f.encodeAsBx(opTForLoop, base+2, noJump)
	}
	f.PatchList(end, prep+1)
	f.FixLine(line)
}

func (f *function) OpenMainFunction() {
	f.EnterBlock(false)
	f.makeUpValue("_ENV", makeExpression(kindLocal, 0))
}

func (f *function) CloseMainFunction() *function {
	f.ReturnNone()
	f.LeaveBlock()
	f.assert(f.block == nil)
	return f.previous
}

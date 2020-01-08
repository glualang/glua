package parser

func (p *parser) genEmptyTable() exprDesc {
	pc, t := p.function.OpenConstructor()
	_, a, h, pending := p.lineNumber, 0, 0, 0
	var e exprDesc
	p.function.CloseConstructor(pc, t.info, pending, a, h, e)
	return t
}

// 生成一个匿名的record的构造函数，函数体逻辑是如果提高了table作为参数则直接返回，否则返回一个新的空table
func (p *parser) genAnnoyRecordFunc(protoName string, line int) exprDesc {
	p.function.OpenFunction(line)
	// 增加一个可选的参数, table类型，作为默认实现
	p.function.MakeLocalVariable("props")
	p.function.AdjustLocalVariables(1)

	p.function.f.parameterCount = 1
	p.function.ReserveRegisters(p.function.f.parameterCount)

	// 函数体的生成
	p.enterLevel() // enter record func body
	f := p.function
	f.f.name = protoName
	// 函数体目前逻辑是 if props then return props else return {} end
	escapes := noJump
	propsCheckE := p.function.SingleVariable("props")
	propsCheckE = p.function.GoIfTrue(propsCheckE)
	p.function.EnterBlock(false)
	jumpFalse := propsCheckE.f
	// statementList
	// return props
	propsE := p.function.SingleVariable("props")
	p.function.ExpressionToNextRegister(propsE)
	f.Return(propsE, 1)

	p.function.LeaveBlock()
	escapes = p.function.Concatenate(escapes, p.function.Jump())
	p.function.PatchToHere(jumpFalse) // end if then

	// else body
	p.function.EnterBlock(false)
	p.enterLevel()
	// new table {}
	newTableE := p.genEmptyTable()
	f.Return(newTableE, 1)
	p.leaveLevel()
	p.function.LeaveBlock()

	// end if
	p.function.PatchToHere(escapes)

	p.leaveLevel() // end record func body

	p.function.f.lastLineDefined = p.lineNumber
	return p.function.CloseFunction()
}

// 产生record的构造函数的指令
func (p *parser) genRecordFunc(recordInfo *recordTypeInfo, line int) exprDesc {
	return p.genAnnoyRecordFunc(recordInfo.name, line)
}

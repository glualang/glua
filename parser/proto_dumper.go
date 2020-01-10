package parser

import (
	"errors"
	"os"
	"strconv"
)

var protoNameIdGen = 0
func genPrototypeName() string {
	defer func() {
		protoNameIdGen++
	}()
	return "proto_" + strconv.Itoa(protoNameIdGen)
}

// 把prototype转成lua asm的字符串格式
func (p *Prototype) ToFuncAsm(outfile *os.File, isTop bool) (err error) {
	if isTop {
		outfile.WriteString(".upvalues " + strconv.Itoa(len(p.upValues)) + "\r\n")
	}

	protoName := p.name
	if len(p.name) < 1 {
		protoName = genPrototypeName()
	}

	outfile.WriteString(".func " + protoName + " " + strconv.Itoa(int(p.maxStackSize)) +
		" " + strconv.Itoa(int(p.parameterCount)) + " " + strconv.Itoa(len(p.localVariables)) + "\r\n")

	//write constants  --------------------------------------
	outfile.WriteString(".begin_const\r\n")
	//var valtype int
	//var val string
	for i := 0; i < len(p.constants); i++ {
		constantValue := p.constants[i]
		outfile.WriteString("\t")
		valStr, ok := literalValueString(constantValue)
		if !ok {
			err = errors.New("non-literal constant value " + debugValue(constantValue))
			return
		}
		outfile.WriteString(valStr + "\r\n")
	}
	outfile.WriteString(".end_const\r\n")

	//write upvalue -------------------------------------------
	outfile.WriteString(".begin_upvalue\r\n")
	for i := 0; i < len(p.upValues); i++ {
		upvalue := p.upValues[i]
		var instack int
		if upvalue.isLocal {
			instack = 1
		} else {
			instack = 0
		}
		outfile.WriteString("\t" + strconv.Itoa(instack) + " " + strconv.Itoa(int(upvalue.index)) + " \"" + upvalue.name + "\"\r\n")
	}
	outfile.WriteString(".end_upvalue\r\n")

	//write locals  ---------------------------------------------
	var local *localVariable
	outfile.WriteString(".begin_local\r\n")
	for i := 0; i < len(p.localVariables); i++ {
		local = &p.localVariables[i]
		outfile.WriteString("\t\"" + local.name + "\" " + strconv.Itoa(int(local.startPC)) + " " + strconv.Itoa(int(local.endPC)) + "\r\n")
	}
	outfile.WriteString(".end_local\r\n")

	//pre parse location  ----------------------------------------
	err = p.preParseLabelLocations()
	if err != nil {
		return
	}
	locationinfos := p.extra.labelLocations // line -> label mapping

	//write code  ---------------------------------------------
	outfile.WriteString(".begin_code\r\n")
	lineinfoNum := len(p.lineInfo)
	var sourceCodeLine int32
	for i := 0; i < len(p.code); i++ {
		//add location label
		if inslabel, ok := locationinfos[i]; ok {
			outfile.WriteString(inslabel + ":\r\n")
		}

		instruction := p.code[i]

		insParseAsmEr, insStr, useExtended, extraArg := p.ParseInstructionToAsmLine(instruction, i)
		if insParseAsmEr != nil {
			err = insParseAsmEr
			return
		}

		if extraArg {
			outfile.WriteString(insStr)
		} else {
			outfile.WriteString("\t" + insStr)
		}

		if !useExtended {
			if i < lineinfoNum {
				sourceCodeLine = p.lineInfo[i]
				outfile.WriteString(";L" + strconv.Itoa(int(sourceCodeLine)) + ";")
			}
			outfile.WriteString("\r\n")
		}

	}
	outfile.WriteString(".end_code\r\n")

	//write subprotos   -----------------------------------------
	for i := 0; i < len(p.prototypes); i++ {
		outfile.WriteString("\r\n")
		subpf := p.prototypes[i]
		err = subpf.ToFuncAsm(outfile, false)
		if err != nil {
			return
		}
	}
	outfile.WriteString("\r\n")
	return
}


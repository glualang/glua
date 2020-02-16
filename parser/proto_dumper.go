package parser

import (
	"errors"
	"github.com/glualang/gluac/utils"
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
func (p *Prototype) ToFuncAsm(outStream utils.ByteStream, isTop bool) (err error) {
	if isTop {
		outStream.WriteString(".upvalues " + strconv.Itoa(len(p.upValues)) + "\r\n")
	}

	protoName := p.name
	if len(p.name) < 1 {
		protoName = genPrototypeName()
	}

	outStream.WriteString(".func " + protoName + " " + strconv.Itoa(int(p.maxStackSize)) +
		" " + strconv.Itoa(int(p.parameterCount)) + " " + strconv.Itoa(len(p.localVariables)) + "\r\n")

	//write constants  --------------------------------------
	outStream.WriteString(".begin_const\r\n")
	//var valtype int
	//var val string
	for i := 0; i < len(p.constants); i++ {
		constantValue := p.constants[i]
		outStream.WriteString("\t")
		valStr, ok := literalValueString(constantValue)
		if !ok {
			err = errors.New("non-literal constant value " + debugValue(constantValue))
			return
		}
		outStream.WriteString(valStr + "\r\n")
	}
	outStream.WriteString(".end_const\r\n")

	//write upvalue -------------------------------------------
	outStream.WriteString(".begin_upvalue\r\n")
	for i := 0; i < len(p.upValues); i++ {
		upvalue := p.upValues[i]
		var instack int
		if upvalue.isLocal {
			instack = 1
		} else {
			instack = 0
		}
		outStream.WriteString("\t" + strconv.Itoa(instack) + " " + strconv.Itoa(int(upvalue.index)) + " \"" + upvalue.name + "\"\r\n")
	}
	outStream.WriteString(".end_upvalue\r\n")

	//write locals  ---------------------------------------------
	var local *localVariable
	outStream.WriteString(".begin_local\r\n")
	for i := 0; i < len(p.localVariables); i++ {
		local = &p.localVariables[i]
		outStream.WriteString("\t\"" + local.name + "\" " + strconv.Itoa(int(local.startPC)) + " " + strconv.Itoa(int(local.endPC)) + "\r\n")
	}
	outStream.WriteString(".end_local\r\n")

	//pre parse location  ----------------------------------------
	err = p.preParseLabelLocations()
	if err != nil {
		return
	}
	locationinfos := p.extra.labelLocations // Line -> label mapping

	//write code  ---------------------------------------------
	outStream.WriteString(".begin_code\r\n")
	lineinfoNum := len(p.lineInfo)
	var sourceCodeLine int32
	for i := 0; i < len(p.code); i++ {
		//add location label
		if inslabel, ok := locationinfos[i]; ok {
			outStream.WriteString(inslabel + ":\r\n")
		}

		instruction := p.code[i]

		insParseAsmEr, insStr, useExtended, extraArg := p.ParseInstructionToAsmLine(instruction, i)
		if insParseAsmEr != nil {
			err = insParseAsmEr
			return
		}

		if extraArg {
			outStream.WriteString(insStr)
		} else {
			outStream.WriteString("\t" + insStr)
		}

		if !useExtended {
			if i < lineinfoNum {
				sourceCodeLine = p.lineInfo[i]
				outStream.WriteString(";L" + strconv.Itoa(int(sourceCodeLine)) + ";")
			}
			outStream.WriteString("\r\n")
		}

	}
	outStream.WriteString(".end_code\r\n")

	//write subprotos   -----------------------------------------
	for i := 0; i < len(p.prototypes); i++ {
		outStream.WriteString("\r\n")
		subpf := p.prototypes[i]
		err = subpf.ToFuncAsm(outStream, false)
		if err != nil {
			return
		}
	}
	outStream.WriteString("\r\n")
	return
}

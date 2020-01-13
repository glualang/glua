package parser

import (
	"errors"
	"strconv"
	"strings"
)

func (p *Prototype) preParseLabelLocations() (err error) {
	extra := p.extra
	var ins instruction
	var operand int
	var jmpDest int
	for i:=0;i<len(p.code);i++ {
		ins = p.code[i]
		opcode := ins.opCode()
		if int(opcode) >= int(NUM_OPCODES) {
			return errors.New("unknown opcode " + strconv.Itoa(int(opcode)))
		}
		count := Opcounts[opcode]
		info := Opinfos[opcode]

		for j:=0;j<count;j++ {
			if info[j].limit == LIMIT_LOCATION {
				operand = ins.sbx()
				jmpDest = operand + 1 + i
				if (jmpDest >= len(p.code)) || (jmpDest < 0) {
					return errors.New("jmp dest exceed")
				}
				insLabel := "label_" + strconv.Itoa(jmpDest)
				extra.labelLocations[jmpDest] = insLabel
			}
		}
	}
	return
}

// ParseInstructionToAsmLine transform instruction to asm Line
// @return : err, insStr, useExtended, extraArg
func (proto *Prototype) ParseInstructionToAsmLine(i instruction, indexInProtoCode int) (err error, insStr string, useExtended bool, extraArg bool) {
	op := i.opCode()
	opName := opNames[op]
	opName = strings.ToLower(opName)
	insStr = opName

	opCount := Opcounts[op]
	opInfo := Opinfos[op]
	var operand int

	//{{OPP_Ax, LIMIT_EMBED}}
	if op == opExtraArg {
		extraArg = true
		insStr = ""
	}

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
			return
		}

		switch limit {
		case LIMIT_STACKIDX:
			insStr = insStr + " %" + strconv.Itoa(int(operand))
		case LIMIT_UPVALUE:
			insStr = insStr + " @" + strconv.Itoa(int(operand))
		case LIMIT_EMBED:
			insStr = insStr + " " + strconv.Itoa(int(operand))
		case LIMIT_CONSTANT:
			constIdx := constantIndex(operand)
			constVal := proto.constants[constIdx]
			constLiteral, ok := literalValueString(constVal)
			if !ok {
				err = errors.New("invalid constant literal")
				return
			}
			insStr = insStr + " const " + constLiteral
		case LIMIT_LOCATION:
			loclabel := proto.extra.labelLocations[int(operand)+1+indexInProtoCode]
			insStr = insStr + " $" + loclabel
		case LIMIT_CONST_STACK:
			if (int(operand) & BITRK) > 0 {
				cconstIdx := constantIndex(operand)
				constVal := proto.constants[cconstIdx]
				constLiteral, ok := literalValueString(constVal)
				if !ok {
					err = errors.New("invalid constant literal")
					return
				}
				insStr = insStr + " const " + constLiteral
			} else {
				insStr = insStr + " %" + strconv.Itoa(int(operand))
			}

		case LIMIT_PROTO:
			subProto := proto.prototypes[operand]
			insStr = insStr + " " + subProto.name
		}
	}

	return
}

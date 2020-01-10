package parser

import (
	"errors"
	"strconv"
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
				insLabel := p.name + "_to_dest_br_" + strconv.Itoa(jmpDest)
				extra.labelLocations[jmpDest] = insLabel
			}
		}
	}
	return
}

// TODO
//
//func PreParseInstructionLocation(pf *Prototype) (bool, string, map[int]string) {
//	var ins instruction
//	var jmpdest int
//	var operand uint
//	locationsinfo := make(map[int]string)
//	for i := 0; i < len(pf.code); i++ {
//		ins = pf.code[i]
//		opcode := ins.opCode()
//
//		if int(opcode) >= int(NUM_OPCODES) {
//			return false, "unkown opcode", nil
//		}
//		count := Opcounts[opcode]
//		info := Opinfos[opcode]
//		for j := 0; j < count; j++ {
//			if info[j].limit == LIMIT_LOCATION {
//				operand = GETARG_sBx(ins)
//				jmpdest = int(operand) + 1 + i
//				if (jmpdest >= len(pf.instructions)) || (jmpdest < 0) {
//					return false, "jmp dest exceed", nil
//				}
//				insLabel := pf.name + "_to_dest_br_" + strconv.Itoa(jmpdest)
//				locationsinfo[jmpdest] = insLabel
//			}
//		}
//	}
//
//	return true, "", locationsinfo
//
//}
//
////returns : res, err, insstr, useExtended, extraArg
//func ParseInstruction(pf *Prototype, insIdx int, ins Instruction, locationifos map[int]string) (bool, string, string, bool, bool) {
//	opcode := GET_OPCODE(ins)
//	if int(opcode) >= NUM_OPCODES {
//		return false, "unkown opcode", "", false, false
//	}
//	opname := UvmPOpnames[opcode]
//	count := Opcounts[opcode]
//	info := Opinfos[opcode]
//
//	var insstr string = opname
//	var operand uint
//	var limit int
//	//var opValue uint
//	var constval string
//	var consttype int
//	var useExtended bool = false
//	var extraArg bool = false
//
//	//{{OPP_Ax, LIMIT_EMBED}}
//	if opcode == OP_EXTRAARG {
//		extraArg = true
//		insstr = ""
//	}
//
//	for i := 0; i < count; i++ {
//		limit = info[i].limit
//		switch info[i].pos {
//		case OPP_A:
//			operand = GETARG_A(ins)
//		case OPP_B:
//			operand = GETARG_B(ins)
//		case OPP_C:
//			operand = GETARG_C(ins)
//		case OPP_Bx:
//			operand = GETARG_Bx(ins)
//		case OPP_Ax:
//			operand = GETARG_Ax(ins)
//		case OPP_sBx:
//			operand = GETARG_sBx(ins)
//		case OPP_ARG:
//			useExtended = true
//		case OPP_C_ARG:
//			operand = GETARG_C(ins)
//			if operand == 0 {
//				useExtended = true //try get next
//			}
//		}
//
//		if useExtended { //try get next ins
//			return true, "", insstr, useExtended, false
//		}
//
//		switch limit {
//		case LIMIT_STACKIDX:
//			insstr = insstr + " %" + strconv.Itoa(int(operand))
//		case LIMIT_UPVALUE:
//			insstr = insstr + " @" + strconv.Itoa(int(operand))
//		case LIMIT_EMBED:
//			insstr = insstr + " " + strconv.Itoa(int(operand))
//		case LIMIT_CONSTANT:
//			constval = (*pf.constants[operand]).str()
//			consttype = (*pf.constants[operand]).valueType()
//			if consttype == LUA_TSTRING || consttype == LUA_TLNGSTR {
//				constval = "\"" + constval + "\""
//			}
//			insstr = insstr + " const " + constval
//		case LIMIT_LOCATION:
//			loclabel := locationifos[int(operand)+1+insIdx]
//			insstr = insstr + " $" + loclabel
//		case LIMIT_CONST_STACK:
//			if (int(operand) & BITRK) > 0 {
//				constval = (*pf.constants[int(operand)-BITRK]).str()
//				consttype = (*pf.constants[int(operand)-BITRK]).valueType()
//				if consttype == LUA_TSTRING || consttype == LUA_TLNGSTR {
//					constval = "\"" + constval + "\""
//				}
//				insstr = insstr + " const " + constval
//			} else {
//				insstr = insstr + " %" + strconv.Itoa(int(operand))
//			}
//
//		case LIMIT_PROTO:
//			propname := pf.usedSubroutines[operand]
//			insstr = insstr + " " + propname
//		}
//
//	}
//
//	return true, "", insstr, false, extraArg
//
//}

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
	if(len(p.extra.labelLocations)==0){
		err = p.preParseLabelLocations()
		if err != nil {
			return
		}
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

///////////////////////////////////////////////
type InsSegment struct {
	fromSegIdxs       []int
	toSegIdxs 		  []int
	instructions      []instruction
	endOp opCode
	isTest bool   //  opTest opTestSet opEqual opLessThan opLessOrEqual   (test类型指令后必跟随jmp, jmp指令不计入gas)
	isSubSegment bool  //子分片
	origStartCodeIdx int
	newStartCodeIdx int
	needModify bool
	labelLocation string
}

func newSegment()(*InsSegment){
	seg := new(InsSegment)
	seg.fromSegIdxs = []int{}
	seg.toSegIdxs = []int{}
	seg.instructions = []instruction{}
	seg.isTest = false
	seg.isSubSegment = false
	return seg
}

var branch_ops = [...]opCode{opTest,opTestSet,opJump,opEqual,opLessThan,opLessOrEqual,opForLoop,opForPrep,opTForLoop,opLoadBool,opReturn,opCall,opTailCall}

func getSegmentIdx(segments []*InsSegment, codeIdx int)(int){
	for j := 0; j < len(segments); j++ {
		if(codeIdx == segments[j].origStartCodeIdx){
			return j
		}
	}
	return -1
}

func setFromTo(segments []*InsSegment, segFromIdx int, segToIdx int)(){
	segments[segFromIdx].toSegIdxs = append(segments[segFromIdx].toSegIdxs,segToIdx)
	segments[segToIdx].fromSegIdxs = append(segments[segToIdx].fromSegIdxs,segFromIdx)
}

func isBranchOp(ins instruction)(bool){
	op := ins.opCode()
	//UOP_LOADBOOL,/*	A B C	R(A) := (Bool)B; if (C) pc++			*/
	if((op==opLoadBool && ins.c()!=0)){
		return true
	}
	for j:=0;j<len(branch_ops);j++ {
		if(op==branch_ops[j]){
			return true
		}
	}
	return false
}

//计算gas 目前每条指令1gas
func caculateGas(segment *InsSegment)(int){
	insGasSum := len(segment.instructions)
	if(segment.instructions[insGasSum-1].opCode()==opTForLoop){//UOP_TFORCALL指令后面总是紧随着UOP_TFORLOOP，uvm运行时只计算一次指令gas
		insGasSum = insGasSum - 1
		return insGasSum
	}
	if(segment.isTest){ //test op随后必然跟随jmp指令,uvm运行时只计算一次指令gas
		insGasSum = len(segment.instructions) - 1
	}
	return insGasSum
}

func (p *Prototype)meteringProto() (err error) {
	//segments:= assembler.Instruction()[]
	//segment := []
	segments := []*InsSegment{}
	segment := newSegment()
	err = p.preParseLabelLocations()
	if err!=nil{
		return
	}
	locationinfos := p.extra.labelLocations //  code index => label
	future_seg_idx := -1
	//locations_idxs := []int{}
	for i := 0; i < len(p.code); i++ {
		ins := p.code[i]
		if(i==0 && ins.opCode()==opMeter){
			print("already meter proto\n")
			return
		}
		if(segment.isTest) && ins.opCode()==opJump{ //is branch
			//add to seg
			segment.instructions = append(segment.instructions,ins)
			segment.endOp = segment.instructions[len(segment.instructions)-1].opCode()
			segments = append(segments,segment)
			segment = newSegment()
			continue
		}
		if(segment.isTest) && ins.opCode()!=opJump{ //is branch
			return errors.New("tesp op must from jmp")
		}

		if _, ok := locationinfos[i]; ok {
			//locations_idxs = append(locations_idxs,i)
			//add segments ， 再清空seg
			if(len(segment.instructions)>0){
				segment.endOp = segment.instructions[len(segment.instructions)-1].opCode()
				segments = append(segments,segment)
				segment = newSegment()
			}
			segment.labelLocation = locationinfos[i]
		}
		if(i == future_seg_idx){
			if(len(segment.instructions)>0){
				segment.endOp = segment.instructions[len(segment.instructions)-1].opCode()
				segments = append(segments,segment)
				segment = newSegment()
			}
		}


		if(len(segment.instructions)==0){
			segment.origStartCodeIdx = i
		}
		//add to seg
		segment.instructions = append(segment.instructions,ins)

		//check branch
		ins_op := ins.opCode()
		//UOP_LOADBOOL,/*	A B C	R(A) := (Bool)B; if (C) pc++			*/
		if isBranchOp(ins) {
			if(ins_op==opTest || ins_op==opTestSet || ins_op==opEqual || ins_op==opLessThan || ins_op==opLessOrEqual ){ //op PC++
				segment.isTest = true
				future_seg_idx = i+2
			}else if (ins_op == opLoadBool && ins.c()!=0) {
				future_seg_idx = i + 2
			}else if(len(segment.instructions)>0){ //is branch , add segments ， 再清空seg
				segment.endOp = segment.instructions[len(segment.instructions)-1].opCode()
				segments = append(segments,segment)
				segment = newSegment()
			}
		}
	}
	if(len(segment.instructions)>0){
		segment.endOp = segment.instructions[len(segment.instructions)-1].opCode()
		segments = append(segments,segment)
		segment = newSegment()
	}

	for i := 0; i < len(segments); i++ {
		segment := segments[i]
		lastIns := segment.instructions[len(segment.instructions)-1]
		ins_op := lastIns.opCode()
		if(ins_op==opJump || ins_op==opForLoop || ins_op==opTForLoop || ins_op == opForPrep ){
			//test 类型op 其后都会跟随jmp
			//jmp to des
			operand := lastIns.sbx()
			desInsIdx := segment.origStartCodeIdx + len(segment.instructions)  + operand //从下条指令起跳
			segToIdx := getSegmentIdx(segments,desInsIdx)
			if(segToIdx<0){
				return errors.New("can't find dest segment ")
			}
			segment.needModify = true
			setFromTo(segments,i,segToIdx)

			//may go to next segment
			if(segment.isTest || ins_op==opTForLoop  || ins_op == opForLoop){
				desInsIdx = segment.origStartCodeIdx + len(segment.instructions)
				segToIdx1 := getSegmentIdx(segments, desInsIdx)
				if(segToIdx1<0){
					return errors.New("can't find dest segment ")
				}
				setFromTo(segments,i,segToIdx1)
			}
		}else if(ins_op==opLoadBool && lastIns.c()!=0){ //pc++
			desInsIdx := segment.origStartCodeIdx + len(segment.instructions) + 1
			//next and next
			segToIdx2 := getSegmentIdx(segments,desInsIdx )
			if(segToIdx2<0){
				return errors.New("can't find dest segment ")
			}
			setFromTo(segments,i,segToIdx2)
		}else if(ins_op==opReturn){
			segment.toSegIdxs = []int{}
		}else{ // no jump , just go next
			desInsIdx := segment.origStartCodeIdx + len(segment.instructions)
			segToIdx1 := getSegmentIdx(segments,desInsIdx)
			if(segToIdx1<0){
				return errors.New("can't find dest segment ")
			}
			setFromTo(segments,i,segToIdx1)
		}
	}

	//mark subsegment
	for i := 0; i < len(segments); i++ {
		segment := segments[i]
		if(len(segment.fromSegIdxs) == 1) && segments[segment.fromSegIdxs[0]].endOp != opCall && segments[segment.fromSegIdxs[0]].endOp != opTailCall&& len(segments[segment.fromSegIdxs[0]].toSegIdxs)==1 {
			segment.isSubSegment = true
		}
	}

	//add meter ,
	for i := 0; i < len(segments); i++ {
		segment = segments[i]
		if (!segment.isSubSegment) {
			//add meter , meter op 没有gas， 生产环境不允许meter op
			insGasSum := caculateGas(segment)
			tempSegment := segment
			for ; ; {
				if (len(tempSegment.toSegIdxs) == 1) && segments[tempSegment.toSegIdxs[0]].isSubSegment {
					tempSegment = segments[tempSegment.toSegIdxs[0]]
					seg_gas := caculateGas(tempSegment)
					insGasSum = insGasSum + seg_gas
				} else {
					break
				}
			}
			meterIns := createAx(opMeter, insGasSum)
			segment.instructions = append([]instruction{meterIns}, segment.instructions...)
		}
		//modify start code idx
		if (i > 0) {
			segment.newStartCodeIdx = segments[i-1].newStartCodeIdx + len(segments[i-1].instructions)
		}

	}

	newlineInfo := []int32{}
	newPcode := []instruction{}
	new_locations_map := map[int]string{}
	//modify jmp instruction and reset locationlabel ,  set new code instructions
	for i := 0; i < len(segments); i++ {
		segment = segments[i]
		if segment.needModify{
			lastIns := segment.instructions[len(segment.instructions)-1]
			toSegIdx := segment.toSegIdxs[0]
			toInsIdx := segments[toSegIdx].newStartCodeIdx
			savedPc := segment.newStartCodeIdx + len(segment.instructions) // 从下条指令起跳
			operand := toInsIdx - savedPc //
			lastIns.setSBx(operand)
			segment.instructions[len(segment.instructions)-1] = lastIns
		}
		if(len(segment.labelLocation)>0){
			new_locations_map[segment.newStartCodeIdx] = segment.labelLocation
		}

		if(segment.isSubSegment){
			newlineInfo = append(newlineInfo,p.lineInfo[segment.origStartCodeIdx:(segment.origStartCodeIdx + len(segment.instructions))]...)
		}else{
			newlineInfo = append(newlineInfo,p.lineInfo[segment.origStartCodeIdx]) // set meter op line
			newlineInfo = append(newlineInfo,p.lineInfo[segment.origStartCodeIdx:(segment.origStartCodeIdx + len(segment.instructions)-1)]...)
		}
		newPcode = append(newPcode,segment.instructions...)

	}
	if(len(newlineInfo)!=len(newPcode)){
		return errors.New("line info not match code err")
	}
	p.lineInfo = newlineInfo
	p.code = newPcode
	if(len(p.extra.labelLocations)!=len(new_locations_map)){
		return errors.New("location err")
	}
	p.extra.labelLocations = new_locations_map
	return nil
}


// 把prototype转成lua asm的字符串格式
func (p *Prototype) AddMeter(isRecurse bool) (err error) {
	err = p.meteringProto()
	if err != nil {
		return
	}
	if(isRecurse){
		//write subprotos   -----------------------------------------
		for i := 0; i < len(p.prototypes); i++ {
			err = p.prototypes[i].meteringProto()
			if err != nil {
				return
			}
		}
	}
	return nil
}




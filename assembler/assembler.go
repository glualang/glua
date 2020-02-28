package assembler

import (
	"errors"
	"github.com/glualang/gluac/parser"
	"strconv"
	"strings"
)

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"unicode"
)

// OperandTypes
const (
	STACKIDX = iota
	CONSTANT = iota
	LOCATION = iota
	UPVALUE  = iota
	EMBEDDED = iota
)

const LUAI_BITSINT = 32

// OpModes, basic instruction format
const (
	iABC  = iota
	iABx  = iota
	iAsBx = iota
	iAx   = iota
)

type OpMode int

/*
** size and position of opcode arguments.
 */
const (
	SIZE_C  = 9
	SIZE_B  = 9
	SIZE_Bx = (SIZE_C + SIZE_B)
	SIZE_A  = 8
	SIZE_Ax = (SIZE_C + SIZE_B + SIZE_A)

	SIZE_OP = 6

	POS_OP = 0
	POS_A  = (POS_OP + SIZE_OP)
	POS_C  = (POS_A + SIZE_A)
	POS_B  = (POS_C + SIZE_C)
	POS_Bx = POS_C
	POS_Ax = POS_A
)

// max instruction length
const MAX_INT = 32

const MAXARG_Bx = ((1 << SIZE_Bx) - 1)
const MAXARG_sBx = (MAXARG_Bx >> 1) /* 'sBx' is signed */
const MAXARG_Ax = ((1 << SIZE_Ax) - 1)

const MAXARG_C = ((1 << SIZE_C) - 1)

/* creates a mask with 'n' 1 bits at position 'p' */
// #define MASK1(n,p)	((~((~(Instruction)0)<<(n)))<<(p))
func MASK1(n uint, p uint) uint {
	return uint((^((^(Instruction(0))) << n)) << p)
}

/* creates a mask with 'n' 0 bits at position 'p' */
// #define MASK0(n,p)	(~MASK1(n,p))
func MASK0(n uint, p uint) uint {
	return ^MASK1(n, p)
}

type Instruction uint

type OperandType int

type Operand struct {
	opType OperandType
	value  int // stack index, const id, location id (or -1 if unknown), upvalue index, embedded integer
}

type Upvalue struct {
	instack uint8
	idx     uint8
	name    string
}

type LocVar struct {
	varname string
	startpc int
	endpc   int
}

type StringIntPair struct {
	left  string
	right int
}

type TValue interface {
	str() string
	valueType() int
}

type TString struct {
	TValue
	string_value string
}

func (str *TString) str() string {
	return str.string_value
}

func (str *TString) valueType() int {
	return LUA_TSTRING
}

type TBool struct {
	TValue
	bool_value bool
}

func (b *TBool) str() string {
	return strconv.FormatBool(b.bool_value)
}

func (b *TBool) valueType() int {
	return LUA_TBOOLEAN
}

type TNumber struct {
	TValue
	number_value float64
}

func (n *TNumber) str() string {
	return strconv.FormatFloat(n.number_value, 'f', 20, 64)
}

func (n *TNumber) valueType() int {
	return LUA_TNUMBER
}

type TInteger struct {
	TValue
	int_value int64
}

func (n *TInteger) str() string {
	return strconv.FormatInt(n.int_value, 10)
}

func (n *TInteger) valueType() int {
	return LUA_TNUMINT
}

type TNil struct {
	TValue
}

func (n *TNil) str() string {
	return "nil"
}

func (n *TNil) valueType() int {
	return LUA_TNIL
}

type ParsedFunction struct {
	name              string
	instructions      []Instruction
	upvalues          []Upvalue
	usedSubroutines   []string
	lineinfos         []int
	neededSubroutines []StringIntPair
	constants         []*TValue
	maxstacksize      uint
	params            uint
	vararg            uint
	linedefined       uint
	lastlinedefined   uint
	//add locals
	locals []LocVar
}

// ParseStates
const (
	PARSE_FUNC    = iota
	PARSE_CODE    = iota
	PARSE_CONST   = iota
	PARSE_UPVALUE = iota
	PARSE_LOCAL   = iota
	PARSE_NONE    = iota
)

type ParseState int

type Assembler struct {
	parseStatus     ParseState
	nUpvalues       int
	bUpvalues       bool
	functions       map[string]*ParsedFunction
	subroutines     map[string]int
	usedSubroutines map[string]int // subroutine name, function id

	funcid int
	/* The following should be save on a new function declaration or end of file */
	fUsedSubroutines []string
	upvalues         []Upvalue
	locals           []LocVar
	funcname         string

	instructions      []Instruction
	neededSubroutines []StringIntPair
	neededLocations   []StringIntPair // for jmps to the future
	locations         map[string]int

	lineinfos []int

	fMaxstacksize, fParams, fVararg uint
	constants                       []*TValue

	wBuffer *bytes.Buffer
}

func (assembler *Assembler) Assemble() (bool, string) {
	return true, ""
}

type AsmValue struct {
}

type AsmBlock struct {
	NodeType string
	Body     []AsmBlock
	Args     []AsmValue
}

type LuaProto struct {
}

/**
 * Remove comments from a line of asm code (comment line is a line beginning with ';')
 */
func (assembler *Assembler) getLineCommentFromAsmLineCode(line string, lineLen int) string {
	var commentStartPos = -1
	for i := 0; i < len(line); i++ {
		if line[i] == ';' {
			commentStartPos = i
			break
		}
	}
	if commentStartPos >= 0 {
		return line[commentStartPos:]
	}
	return ""
}

/**
 * get the line number from comment with format ";<line_number>;<other_line_comment", or return -1 if linenumber does not exist in this comment
 */
func (assembler *Assembler) getLinenumberFromAsmLineComment(lineComment string) int {
	var lineNumber = -1
	if len(lineComment) < 1 {
		return lineNumber
	}
	if len(lineComment) >= 3 && lineComment[0:2] == ";L" {
		lineNumberChars := ""
		nextPosAfterNumber := -1
		for i := 2; i < len(lineComment); i++ {
			iChar := lineComment[i]
			if unicode.IsDigit(rune(iChar)) {
				lineNumberChars += string(iChar)
			} else {
				nextPosAfterNumber = i
				break
			}
		}
		if nextPosAfterNumber <= 0 || nextPosAfterNumber >= len(lineComment) {
			return lineNumber
		}
		var err error
		lineNumber, err = strconv.Atoi(lineNumberChars)
		if err != nil {
			return -1
		}
		return lineNumber
	}
	return lineNumber
}

func (assembler *Assembler) finalizeFunction() (bool, string) {
	if len(assembler.neededLocations) > 0 {
		return false, fmt.Sprintf("undeclared locations: %v", assembler.neededLocations)
	}
	fn := new(ParsedFunction)
	fn.name = assembler.funcname
	fn.instructions = make([]Instruction, len(assembler.instructions))
	copy(fn.instructions, assembler.instructions)
	assembler.instructions = assembler.instructions[:0]
	fn.upvalues = make([]Upvalue, len(assembler.upvalues))
	copy(fn.upvalues, assembler.upvalues)
	assembler.upvalues = assembler.upvalues[:0]
	fn.locals = make([]LocVar, len(assembler.locals))
	copy(fn.locals, assembler.locals)
	assembler.locals = assembler.locals[:0]
	fn.constants = make([]*TValue, len(assembler.constants))
	copy(fn.constants, assembler.constants)
	assembler.constants = assembler.constants[:0]
	fn.neededSubroutines = make([]StringIntPair, len(assembler.neededSubroutines))
	copy(fn.neededSubroutines, assembler.neededSubroutines)
	assembler.neededSubroutines = assembler.neededSubroutines[:0]
	fn.usedSubroutines = make([]string, len(assembler.fUsedSubroutines))
	copy(fn.usedSubroutines, assembler.fUsedSubroutines)
	assembler.fUsedSubroutines = assembler.fUsedSubroutines[:0]
	fn.maxstacksize = assembler.fMaxstacksize
	assembler.fMaxstacksize = 0
	fn.params = assembler.fParams
	assembler.fParams = 0
	fn.vararg = assembler.fVararg
	assembler.fVararg = 0
	fn.lineinfos = make([]int, len(assembler.lineinfos))
	copy(fn.lineinfos, assembler.lineinfos)
	assembler.lineinfos = assembler.lineinfos[:0]
	assembler.functions[fn.name] = fn
	return true, ""
}

// extract symbol from line
func ParseLabel(line string, start int, end int) (bool, int, string) {
	for i := start; i < end && i < len(line); i++ {
		if !unicode.IsSpace(rune(line[i])) {
			start = i
			break
		}
	}
	lend1 := strings.IndexFunc(line[start:end], func(c rune) bool {
		return c != '_' && !unicode.IsLetter(c) && !unicode.IsDigit(c)
	})
	var lend int
	if lend1 < 0 {
		lend = end
	} else {
		lend = lend1 + start
	}
	return true, lend, line[start:lend]
}

func ParseInt(line string, start int, end int) (bool, int, int) {
	for i := start; i < end && i < len(line); i++ {
		if !unicode.IsSpace(rune(line[i])) {
			start = i
			break
		}
	}
	lend1 := strings.IndexFunc(line[start:end], func(c rune) bool {
		return !unicode.IsDigit(c) && c != '+' && c != '-'
	})
	var lend int
	if lend1 < 0 {
		lend = end
	} else {
		lend = lend1 + start
	}
	n, err := strconv.Atoi(line[start:lend])
	if err != nil {
		return false, 0, 0
	}
	return true, lend, n
}

//add parse string  format "name"
func ParseStr(line string, start int, end int) (bool, int, string) {
	for i := start; i < end && i < len(line); i++ {
		if !unicode.IsSpace(rune(line[i])) {
			start = i
			break
		}
	}

	lend1 := strings.IndexFunc(line[start:end], func(c rune) bool {
		return c == '"'
	})

	var lend int
	var strbegin int
	if lend1 < 0 {
		return false, start, ""
	} else {
		strbegin = lend1 + start + 1
	}

	lend2 := strings.IndexFunc(line[strbegin:end], func(c rune) bool {
		return c == '"'
	})
	var strend int
	if lend2 < 0 {
		return false, start, ""
	} else {
		strend = lend2 + strbegin
	}

	if strbegin >= strend || strend > end {
		return false, start, ""
	}
	if strend < end {
		lend = strend + 1
	} else {
		lend = end
	}
	return true, lend, string(line[strbegin:strend])
}

//end
func CheckName(name string) (bool, string) {
	if len(name) <= 0 {
		return false, "invalid name:" + name
	}
	if !IsSymbolStartChar(rune(name[0])) {
		return false, "invalid name:" + name
	}

	cpos := strings.IndexFunc(name[0:len(name)], func(c rune) bool {
		return unicode.IsSpace(c) || !(unicode.IsDigit(c) || unicode.IsLetter(c) || c == '_')
	})
	if cpos < 0 {
		return true, ""
	} else {
		return false, "invalid name:" + name
	}
}

func (assembler *Assembler) parseDirective(line string, lineLen int) (bool, string) {
	var nameEnd = len(line)
	var cmdEnd = len(line)
	for i := 1; i < len(line); i++ {
		iChar := line[i]
		iRune := rune(iChar)
		if !unicode.IsLetter(iRune) && !unicode.IsDigit(iRune) && iChar != '_' {
			if unicode.IsSpace(iRune) || iChar == ';' {
				nameEnd = i
				break
			}
			return false, "could not parse directive: illegal character '" + string(iChar) + "'"
		}
	}
	for i := 0; i < len(line); i++ {
		iChar := line[i]
		if iChar == ';' {
			cmdEnd = i
			break
		}
	}
	name := strings.ToLower(strings.Trim(line[1:nameEnd], " \t")) // ignore first because directive starts with '.'
	lineWithoutComment := strings.Trim(line[0:cmdEnd], " \t")
	if len(name) < 1 {
		return false, "could not parse directive " + line
	}
	argsAfterDirectiveName := strings.Trim(lineWithoutComment[len(name)+1:], " \t") // +1 when starts with '.'
	// var curPosOfArgs = 0 // which pos when parsing to argsAfterDirectiveName
	// lineComment := assembler.getLineCommentFromAsmLineCode(line, lineLen)
	// lineNumber := assembler.getLinenumberFromAsmLineComment(lineComment)
	if name == "upvalues" {
		if assembler.bUpvalues {
			return false, "already declared amount of upvalues"
		}
		var argErr error
		assembler.nUpvalues, argErr = strconv.Atoi(argsAfterDirectiveName)
		if argErr != nil {
			return false, "invalid args for directive .upvalues"
		}
		assembler.bUpvalues = true
	} else if name == "func" {
		if assembler.parseStatus != PARSE_FUNC && assembler.parseStatus != PARSE_NONE {
			return false, "func declaration cannot be inside a code or const segment"
		}
		// even empty func can finalize
		finalizeFuncRes, err := assembler.finalizeFunction()
		if !finalizeFuncRes {
			return false, err
		}
		parseRes, _, funcname := ParseLabel(argsAfterDirectiveName, 0, len(argsAfterDirectiveName))
		if !parseRes {
			return false, "parse funcname error"
		}
		assembler.funcname = strings.Trim(funcname, " ")
		argsAfterDirectiveName = strings.Trim(argsAfterDirectiveName[len(funcname):], " ")

		parseRes, maxStacksizeLend, maxStacksize := ParseInt(argsAfterDirectiveName, 0, len(argsAfterDirectiveName))
		if !parseRes {
			return false, "parse maxstacksize int error"
		}
		assembler.fMaxstacksize = uint(maxStacksize)
		argsAfterDirectiveName = strings.Trim(argsAfterDirectiveName[maxStacksizeLend:], " ")

		parseRes, lend, params := ParseInt(argsAfterDirectiveName, 0, len(argsAfterDirectiveName))
		if !parseRes {
			return false, "parse params int error"
		}
		assembler.fParams = uint(params)
		argsAfterDirectiveName = strings.Trim(argsAfterDirectiveName[lend:], " ")

		parseRes, lend, vararg := ParseInt(argsAfterDirectiveName, 0, len(argsAfterDirectiveName))
		if !parseRes {
			return false, "parse vararg int error"
		}
		assembler.fVararg = uint(vararg)
		argsAfterDirectiveName = strings.Trim(argsAfterDirectiveName[lend:], " ")
		assembler.funcid++
		assembler.parseStatus = PARSE_FUNC
		//fmt.Printf("func %s %d %d %d\n", funcname, maxStacksize, params, vararg)
	} else if name == "begin_const" {
		if assembler.parseStatus != PARSE_FUNC {
			return false, "const declaration must be inside function"
		}
		assembler.parseStatus = PARSE_CONST
	} else if name == "end_const" {
		if assembler.parseStatus != PARSE_CONST {
			return false, "end_const must be inside const segment"
		}
		assembler.parseStatus = PARSE_FUNC
	} else if name == "begin_code" {
		if assembler.parseStatus != PARSE_FUNC {
			return false, "code declaration must be inside function"
		}
		assembler.parseStatus = PARSE_CODE
	} else if name == "end_code" {
		if assembler.parseStatus != PARSE_CODE {
			return false, "end_code must be inside code segment"
		}
		assembler.parseStatus = PARSE_FUNC
	} else if name == "begin_upvalue" {
		if assembler.parseStatus != PARSE_FUNC {
			return false, "upvalue declaration must be inside function"
		}
		assembler.parseStatus = PARSE_UPVALUE
	} else if name == "end_upvalue" {
		if assembler.parseStatus != PARSE_UPVALUE {
			return false, "end_upvalue must be inside upvalue segment"
		}
		assembler.parseStatus = PARSE_FUNC
		//add local>>>>>>>>>>>>>>>>
	} else if name == "begin_local" {
		if assembler.parseStatus != PARSE_FUNC {
			return false, "local declaration must be inside function"
		}
		assembler.parseStatus = PARSE_LOCAL
	} else if name == "end_local" {
		if assembler.parseStatus != PARSE_LOCAL {
			return false, "end_local must be inside local segment"
		}
		assembler.parseStatus = PARSE_FUNC
		//add local end <<<<<<<<<<<<<<<<<<<
	} else {
		return false, "Unsupported directive name " + name
	}
	return true, ""
}

//parse const string
func (assembler *Assembler) parseString(line string,hasComment bool) (bool, int, string) {
	buf := bytes.NewBuffer([]byte{})
	current := 1
	end := len(line)-1
	c := line[0]
	success := false
	if line[0] == '"' && end >=1 {
		for current<=end{
			c = line[current]
			if(c == '\\'){
				current++
				c = line[current]
			}else if(c == '"'){
				success = true
				break
			}
			buf.WriteByte(c)
			current++
		}

		if(success){
			result := string(buf.Bytes())
			remainStr := line[(current+1):]
			Trim(remainStr)
			if(len(remainStr)>0) && (remainStr[0]==';'){
				return true, len(line), result
			}
			return true, len(result)+2, result
		}

	}
	return false, len(line), "truncated constant string value " + line

}

func (assembler *Assembler) parseConstant(line string, lineLen int, id *int) (bool, int, string) {
	var tval TValue
	c := 0
	// lineComment := assembler.getLineCommentFromAsmLineCode(line, lineLen)
	// lineNumber := assembler.getLinenumberFromAsmLineComment(lineComment)
	line = Trim(line)

	cf := strings.ToLower(line[c : c+1])[0]
	cfStr := string(cf)
	if cfStr != "\"" {  // /////
		if strings.Index(line, ";") >= 0 {
			line = line[:strings.Index(line, ";")]
		}
	}
	lineBytes := []byte(line)
	bend := StringFirstIndexOf(line, func(c byte) bool {
		return IsEmptyChar(c)
	}) // read constant value string end index
	// TODO: read next value or using tokenizer to parse line tokens
	if cfStr == "\"" { // parse string
		ok,end,result := assembler.parseString(line,true)
		if(!ok){
			return false, bend, result
		}
		bend = end
		tstr := new(TString)
		tstr.string_value = result
		tval = tstr
	} else if cfStr == "+" || cfStr == "-" || cfStr == "." || unicode.IsDigit(rune(cf)) {
		if strings.Index(line, ".") >= 0 {
			// parse float
			var fltVal float64
			json.Unmarshal(lineBytes, &fltVal)
			if math.Abs(fltVal-math.Trunc(fltVal)) > 0.000001 {
				tnum := new(TNumber)
				tnum.number_value = fltVal
				tval = tnum
			} else {
				tnum := new(TInteger)
				tnum.int_value = int64(fltVal)
				tval = tnum
			}
		} else {
			// parse int
			var intVal int64
			json.Unmarshal(lineBytes, &intVal)
			tnum := new(TInteger)
			tnum.int_value = int64(intVal)
			tval = tnum
		}
	} else if line == "true" {
		tbool := new(TBool)
		tbool.bool_value = true
		tval = tbool
	} else if line == "false" {
		tbool := new(TBool)
		tbool.bool_value = false
		tval = tbool
	} else if line == "nil" {
		tnilval := new(TNil)
		tval = tnilval
	} else {
		return false, bend, "unexpected constant value " + line
	}
	found := false
	for i := 0; i < len(assembler.constants); i++ {
		constVal := *assembler.constants[i]
		if constVal.valueType() == tval.valueType() && constVal.str() == tval.str() {
			found = true
			if id != nil {
				*id = i
			}
			break
		}
	}
	if !found {
		assembler.constants = append(assembler.constants, &tval)
		if id != nil {
			*id = len(assembler.constants) - 1
		}
	}
	return true, bend, ""
}

type OpCode uint8

// #define GET_OPCODE(i)	(cast(OpCode, ((i)>>POS_OP) & MASK1(SIZE_OP,0)))
func GET_OPCODE(inst Instruction) OpCode {
	return OpCode(uint(inst>>POS_OP) & MASK1(SIZE_OP, 0))
}

// #define SET_OPCODE(i,o)	((i) = (((i)&MASK0(SIZE_OP,POS_OP)) | ((cast(Instruction, o)<<POS_OP)&MASK1(SIZE_OP,POS_OP))))
func SET_OPCODE(inst *Instruction, o OpCode) {
	*inst = Instruction((uint(*inst) & MASK0(SIZE_OP, POS_OP)) | (uint(Instruction(o)<<POS_OP) & MASK1(SIZE_OP, POS_OP)))
}

// #define getarg(i,pos,size)	(cast(int, ((i)>>pos) & MASK1(size,0)))
func getarg(inst Instruction, pos uint, size uint) uint {
	return uint(inst>>pos) & MASK1(size, 0)
}

// #define setarg(i,v,pos,size)	((i) = (((i)&MASK0(size,pos)) | ((cast(Instruction, v)<<pos)&MASK1(size,pos))))
func setarg(inst *Instruction, v uint, pos uint, size uint) {
	*inst = Instruction((uint(*inst) & MASK0(size, pos)) | ((uint(v) << pos) & MASK1(size, pos)))
}

// #define GETARG_A(i)	getarg(i, POS_A, SIZE_A)
func GETARG_A(inst Instruction) uint {
	return getarg(inst, POS_A, SIZE_A)
}

// #define SETARG_A(i,v)	setarg(i, v, POS_A, SIZE_A)
func SETARG_A(inst *Instruction, v uint) {
	setarg(inst, v, POS_A, SIZE_A)
}

// #define GETARG_B(i)	getarg(i, POS_B, SIZE_B)
func GETARG_B(inst Instruction) uint {
	return getarg(inst, POS_B, SIZE_B)
}

// #define SETARG_B(i,v)	setarg(i, v, POS_B, SIZE_B)
func SETARG_B(inst *Instruction, v uint) {
	setarg(inst, v, POS_B, SIZE_B)
}

// #define GETARG_C(i)	getarg(i, POS_C, SIZE_C)
func GETARG_C(inst Instruction) uint {
	return getarg(inst, POS_C, SIZE_C)
}

// #define SETARG_C(i,v)	setarg(i, v, POS_C, SIZE_C)
func SETARG_C(inst *Instruction, v uint) {
	setarg(inst, v, POS_C, SIZE_C)
}

// #define GETARG_Bx(i)	getarg(i, POS_Bx, SIZE_Bx)
func GETARG_Bx(inst Instruction) uint {
	return getarg(inst, POS_Bx, SIZE_Bx)
}

// #define SETARG_Bx(i,v)	setarg(i, v, POS_Bx, SIZE_Bx)
func SETARG_Bx(inst *Instruction, v uint) {
	setarg(inst, v, POS_Bx, SIZE_Bx)
}

// #define GETARG_Ax(i)	getarg(i, POS_Ax, SIZE_Ax)
func GETARG_Ax(inst Instruction) uint {
	return getarg(inst, POS_Ax, SIZE_Ax)
}

// #define SETARG_Ax(i,v)	setarg(i, v, POS_Ax, SIZE_Ax)
func SETARG_Ax(inst *Instruction, v uint) {
	setarg(inst, v, POS_Ax, SIZE_Ax)
}

// #define GETARG_sBx(i)	(GETARG_Bx(i)-MAXARG_sBx)
func GETARG_sBx(inst Instruction) uint {
	return GETARG_Bx(inst) - MAXARG_sBx
}

// #define SETARG_sBx(i,b)	SETARG_Bx((i),cast(unsigned int, (b)+MAXARG_sBx))
func SETARG_sBx(inst *Instruction, b uint) {
	SETARG_Bx(inst, uint((b)+MAXARG_sBx))
}

// #define CREATE_ABC(o,a,b,c)	((cast(Instruction, o)<<POS_OP) \
// 			| (cast(Instruction, a)<<POS_A) \
// 			| (cast(Instruction, b)<<POS_B) \
// 			| (cast(Instruction, c)<<POS_C))
func CREATE_ABC(o uint, a uint, b uint, c uint) Instruction {
	return (Instruction(o) << POS_OP) | (Instruction(a) << POS_A) | (Instruction(b) << POS_B) | (Instruction(c) << POS_C)
}

// #define CREATE_ABx(o,a,bc)	((cast(Instruction, o)<<POS_OP) \
// 			| (cast(Instruction, a)<<POS_A) \
// 			| (cast(Instruction, bc)<<POS_Bx))
func CREATE_ABx(o int, a int, bc int) Instruction {
	return (Instruction(o) << POS_OP) | (Instruction(a) << POS_A) | (Instruction(bc) << POS_Bx)
}

// #define CREATE_Ax(o,a)		((cast(Instruction, o)<<POS_OP) \
// 			| (cast(Instruction, a)<<POS_Ax))
func CREATE_Ax(o int, a int) Instruction {
	return (Instruction(o) << POS_OP) | (Instruction(a) << POS_Ax)
}

/*
** Macros to operate RK indices
 */

/* this bit 1 means constant (0 means register) */
const BITRK = 1 << (SIZE_B - 1)

/* test whether value is a constant */
// #define ISK(x)		((x) & BITRK)
func ISK(x uint) bool {
	return x&BITRK != 0
}

/* gets the index of the constant */
// #define INDEXK(r)	((int)(r) & ~BITRK)
func INDEXK(r uint) uint {
	return uint((int(r) & int(^BITRK)))
}

const MAXINDEXRK = BITRK - 1

/* code a constant index as a RK value */
// #define RKASK(x)	((x) | BITRK)
func RKASK(x uint) uint {
	return uint(x | BITRK)
}

const ParseOperandDefaultLimit = 0xFFFFFFFF

/**
@return success, end, value
*/
func (assembler *Assembler) ParseOperand(operand *Operand, line string, limit int) (bool, int, string) {
	cpos := strings.IndexFunc(line, func(c rune) bool {
		return !unicode.IsSpace(c)
	})
	if cpos < 0 {
		return false, 0, "empty line"
	}
	switch line[cpos] {
	case '%': // stack index
		if (limit&parser.LIMIT_STACKIDX) == 0 || cpos >= len(line) {
			return false, 0, "error stack index in " + line
		}
		cpos++
		parseRes, lend, val := ParseInt(line, cpos, len(line))
		if !parseRes {
			return false, cpos, "parse stack index error " + line
		}
		operand.opType = STACKIDX
		operand.value = val
		return true, lend, ""
	case '@': // upvalue index.
		if (limit&parser.LIMIT_UPVALUE) == 0 || cpos >= len(line) {
			return false, 0, "error upvalue index in " + line
		}
		cpos++
		parseRes, lend, val := ParseInt(line[cpos:], 0, len(line[cpos:]))
		if !parseRes {
			return false, cpos, "parse upvalue index error " + line
		}
		operand.opType = UPVALUE
		operand.value = val
		return true, cpos + lend, ""
	case '$': // jmp location
		if (limit&parser.LIMIT_LOCATION) == 0 || cpos >= len(line) {
			return false, 0, "error location in " + line
		}
		cpos++
		parseRes, lend, label := ParseLabel(line[cpos:], 0, len(line[cpos:]))
		if !parseRes {
			return false, cpos, "parse location error " + line
		}
		operand.opType = LOCATION
		if it, ok := assembler.locations[label]; ok {
			operand.value = it - len(assembler.instructions) - 1
		} else {
			operand.value = -1
			var pair StringIntPair
			pair.left = label
			pair.right = len(assembler.instructions)
			assembler.neededLocations = append(assembler.neededLocations, pair)
		}
		return true, cpos + lend, ""
	default:
		if line[cpos] == 'c' { // could be const
			firstNotAlpha := strings.IndexFunc(line[cpos:], func(c rune) bool {
				return !unicode.IsLetter(c)
			})
			if firstNotAlpha >= 0 && strings.Index(line[cpos:], "const") == 0 {
				if (limit & parser.LIMIT_CONSTANT) == 0 {
					return false, 0, "error constant in " + line
				}
				remainingLine := line[firstNotAlpha:]
				bendOfConstant := StringFirstIndexOfNot(remainingLine, func(c byte) bool {
					return IsEmptyChar(c)
				})
				// bend := len(line)
				// for i := firstNotAlpha; i < bend; i++ {
				// 	if !unicode.IsSpace(rune(line[i])) {
				// 		bend = i
				// 		break
				// 	}
				// }
				bend := bendOfConstant + firstNotAlpha
				var id int
				parseRes, bend2, _ := assembler.parseConstant(line[bend:], len(line[bend:]), &id)
				if !parseRes {
					return false, cpos, "parse constant error " + line
				}
				bend += bend2
				operand.opType = CONSTANT
				if (limit & parser.LIMIT_STACKIDX) != 0 {
					operand.value = int(RKASK(uint(id)))
				} else {
					operand.value = id
				}
				return true, bend, ""
			}
		}

		if (limit & parser.LIMIT_PROTO) != 0 {
			parseRes, bend, label := ParseLabel(line[cpos:], 0, len(line[cpos:]))
			if !parseRes {
				return false, 0, "parse label error " + line
			}
			// FIXME
			//if u, ok := assembler.usedSubroutines[label]; ok && u != assembler.funcid {
			//	return false, 0, "used subproutine " + label
			//}
			bend += cpos
			_, subroutineFound := assembler.subroutines[label]
			operand.value = -1
			var pair StringIntPair
			pair.left = label
			pair.right = len(assembler.instructions)
			assembler.neededSubroutines = append(assembler.neededSubroutines, pair)
			if !subroutineFound {
				assembler.usedSubroutines[label] = assembler.funcid
				assembler.fUsedSubroutines = append(assembler.fUsedSubroutines, label)
			}
			return true, bend, ""
		}

		if (limit & (parser.LIMIT_EMBED | parser.LIMIT_CONSTANT)) == 0 {
			return false, 0, "wrong limit " + string(limit)
		}
		cf := strings.ToLower(string(line[cpos]))[0]
		var val int
		var bend int
		if cf == 't' || cf == 'f' { // true or false
			bend = strings.IndexFunc(line[cpos:], func(c rune) bool {
				return !unicode.IsLetter(c)
			})
			if bend < 0 {
				return false, 0, "error parse in " + line
			}
			s := line[cpos:bend]
			if s == "true" {
				val = 1
			} else if s == "false" {
				val = 0
			} else {
				return false, cpos, "error parse bool value in " + line
			}
		} else {
			parseRes, bendOfN, n := ParseInt(line[cpos:], 0, len(line[cpos:]))
			if !parseRes {
				return false, 0, "parse int error in " + line[cpos:]
			}
			bend = bendOfN
			val = n
		}
		operand.opType = EMBEDDED
		operand.value = val
		return true, bend, ""
	}
	return true, cpos, "not supported operand " + line
	// return false, cpos, "not supported operand"
}

func (assembler *Assembler) parseCode(line string, lineLen int) (bool, string) {
	lineComment := assembler.getLineCommentFromAsmLineCode(line, lineLen)
	lineNumber := assembler.getLinenumberFromAsmLineComment(lineComment)
	if lineNumber >= 0 {
		assembler.lineinfos = append(assembler.lineinfos, lineNumber)
	}
	line = Trim(line)

	remainingLine := line
	parseRes, lend, opcodestr := ParseLabel(remainingLine, 0, len(remainingLine))
	if !parseRes {
		return false, "parse code opcode error " + line
	}
	remainingLine = strings.Trim(remainingLine[lend:], " \t")


	if len(remainingLine) > 0 && remainingLine[0] == ':' {
		assembler.locations[opcodestr] = len(assembler.instructions)
		// 一些指令用到了sBx参数，所以需要记录指令位置
		for i := 0; i < len(assembler.neededLocations); {
			if assembler.neededLocations[i].left == opcodestr {
				// Fix entry in assembler.instructions[*i.right]
				ins := &(assembler.instructions[assembler.neededLocations[i].right])
				SETARG_sBx(ins, uint(len(assembler.instructions)-assembler.neededLocations[i].right-1))
				assembler.neededLocations = append(assembler.neededLocations[:i], assembler.neededLocations[i+1:]...)
			} else {
				i++
			}
		}
		return true, ""
	}
	opcodestr = strings.ToLower(opcodestr)
	var opcode OpCode
	for i := 0; i < int(parser.NUM_OPCODES); i++ {
		if IsSameStringIgnoreCase(opcodestr, parser.OpNames[i]) {
			opcode = OpCode(i)
			break
		}
	}
	var ins Instruction
	var extended int
	useExtended := false
	SET_OPCODE(&ins, opcode)

	count := parser.Opcounts[opcode]
	info := parser.Opinfos[opcode]

	for i := 0; i < count; i++ {
		var op Operand
		res, bend, err := assembler.ParseOperand(&op, remainingLine, info[i].Limit)
		if !res {
			return false, "invalid operand(s) " + err + " in line " + line
		}
		remainingLine = Trim(remainingLine[bend:])
		opValue := uint(op.value)
		switch info[i].Pos {
		case parser.OPP_A:
			SETARG_A(&ins, opValue)
		case parser.OPP_B:
			SETARG_B(&ins, opValue)
		case parser.OPP_C:
			SETARG_C(&ins, opValue)
		case parser.OPP_Bx:
			SETARG_Bx(&ins, opValue)
		case parser.OPP_Ax:
			SETARG_Ax(&ins, opValue)
		case parser.OPP_sBx:
			SETARG_sBx(&ins, opValue)
		case parser.OPP_ARG:
			extended = op.value
			useExtended = true
		case parser.OPP_C_ARG:
			if opValue > MAXARG_C {
				SETARG_C(&ins, 0)
				extended = op.value
				useExtended = true
			} else {
				SETARG_C(&ins, opValue)
			}
		}
	}
	remainingLine = Trim(remainingLine)
	if len(remainingLine) > 0 {
		return false, "too many operands in instruction " + line
	}
	assembler.instructions = append(assembler.instructions, ins)

	if useExtended {
		extendedInst := Instruction(0)
		SET_OPCODE(&extendedInst, OpCode(parser.OpExtraArg))
		SETARG_Ax(&extendedInst, uint(extended))
		assembler.instructions = append(assembler.instructions, extendedInst)
	}
	return true, ""
}

func (assembler *Assembler) parseUpvalue(line string, lineLen int) (bool, string) {
	cpos := strings.IndexFunc(line, func(c rune) bool {
		return !unicode.IsSpace(c)
	})
	if cpos < 0 {
		cpos = 0
	}
	end := len(line)
	var instack, idx int
	res, nc, n1 := ParseInt(line, cpos, end)
	if !res {
		return false, "could not parse instack " + line[cpos:end]
	}
	instack = n1
	cpos = nc
	cpos = strings.IndexFunc(line[cpos:], func(c rune) bool {
		return !unicode.IsSpace(c)
	}) + cpos
	res, nc, n2 := ParseInt(line, cpos, end)
	if !res {
		return false, "could not parse idx " + line[cpos:end]
	}
	idx = n2
	cpos = nc
	//add parse name  format "name"
	var name string = string("")
	spos := strings.IndexFunc(line[cpos:], func(c rune) bool {
		return !unicode.IsSpace(c)
	})
	if spos > 0 {
		cpos = spos + cpos
		res, nc, varname := ParseStr(line, cpos, end)
		if res { //find str
			if varname != "_ENV" {
				isValid, errmsg := CheckName(varname)
				if !isValid {
					return false, errmsg
				}
			}
			name = varname
			cpos = nc

			apos := strings.IndexFunc(line[cpos:], func(c rune) bool {
				return !unicode.IsSpace(c)
			})

			if apos >= 0 && cpos+apos < end && line[cpos+apos] != ';' {
				return false, "invalid upvalue:" + line
			}

		} else {
			if cpos < end && line[cpos] != ';' {
				return false, "invalid upvalue:" + line
			}
		}

	} else {
		if cpos < end && line[cpos] != ';' {
			return false, "invalid upvalue:" + line
		}

	}

	var upvalue Upvalue
	upvalue.idx = uint8(idx)
	upvalue.instack = uint8(instack)
	upvalue.name = name
	assembler.upvalues = append(assembler.upvalues, upvalue)
	return true, ""
}

//add parse local func>>>>>>>>>>>>>>>>>
func (assembler *Assembler) parseLocal(line string, lineLen int) (bool, string) {
	cpos := strings.IndexFunc(line, func(c rune) bool {
		return !unicode.IsSpace(c)
	})
	if cpos < 0 {
		cpos = 0
	}
	end := len(line)
	var startpc, endpc int
	res, nc, name := ParseStr(line, cpos, end)
	if !res {
		return false, "could not parse local varname " + line[cpos:end]
	}

	//	isValid, errmsg := CheckName(name)
	//	if !isValid {
	//		return false, errmsg
	//	}
	cpos = nc
	cpos = strings.IndexFunc(line[cpos:], func(c rune) bool {
		return !unicode.IsSpace(c)
	}) + cpos
	res, nc, n1 := ParseInt(line, cpos, end)
	if !res {
		return false, "could not parse local startpc " + line[cpos:end]
	}
	startpc = n1
	cpos = nc
	cpos = strings.IndexFunc(line[cpos:], func(c rune) bool {
		return !unicode.IsSpace(c)
	}) + cpos
	res, nc, n2 := ParseInt(line, cpos, end)
	if !res {
		return false, "could not parse local endpc " + line[cpos:end]
	}
	endpc = n2
	cpos = nc
	foundSpace := strings.IndexFunc(line[cpos:end], func(c rune) bool {
		return !unicode.IsSpace(c)
	})
	if foundSpace >= 0 && cpos+foundSpace < end && line[cpos+foundSpace] != ';' {
		return false, "invalid local:" + line
	}
	var local LocVar
	local.varname = name
	local.startpc = startpc
	local.endpc = endpc
	assembler.locals = append(assembler.locals, local)
	return true, ""
}

//parse local func end

func (assembler *Assembler) ParseLine(line string, lineLen int) (bool, string) {
	if len(Trim(line)) < 1 {
		return true, ""
	}
	if Trim(line)[0] == ';' {
		return true, ""
	}
	if line[0] == '.' {
		return assembler.parseDirective(line, lineLen)
	}
	switch assembler.parseStatus {
	case PARSE_CONST:
		var id int
		res, _, err := assembler.parseConstant(line, lineLen, &id)
		return res, err
	case PARSE_CODE:
		return assembler.parseCode(line, lineLen)
	case PARSE_UPVALUE:
		return assembler.parseUpvalue(line, lineLen)
		//add parse local
	case PARSE_LOCAL:
		return assembler.parseLocal(line, lineLen)
		//parse local end
	case PARSE_FUNC:
		return false, "unimplemented syntax"
	case PARSE_NONE:
		return false, "unimplemented syntax"
	default:
		return false, "unknown parse status"
	}
	return true, ""
}

func NewAssembler() *Assembler {
	instance := new(Assembler)
	instance.locations = make(map[string]int)
	instance.functions = make(map[string]*ParsedFunction)
	instance.subroutines = make(map[string]int)
	instance.usedSubroutines = make(map[string]int)
	instance.wBuffer = bytes.NewBuffer([]byte(""))
	return instance
}

/**
 * write lua bytecode header
 */
func (assembler *Assembler) writeHeader() (bool, string) {
	if BufferWriteCharArray(assembler.wBuffer, CurrentLuaConfig.LuaSignature) != nil {
		return false, "failed to write signature"
	}
	if BufferWriteInt8(assembler.wBuffer, CurrentLuaConfig.CompilerVersion) != nil {
		return false, "failed to write version"
	}
	if BufferWriteInt8(assembler.wBuffer, CurrentLuaConfig.CompilerFormat) != nil {
		return false, "failed to write format"
	}
	if BufferWriteCharArray(assembler.wBuffer, CurrentLuaConfig.MagicData) != nil {
		return false, "failed to write LUAC_DATA"
	}
	// write int32 size
	if BufferWriteInt8(assembler.wBuffer, 4) != nil {
		return false, "failed to write int size"
	}
	if BufferWriteInt8(assembler.wBuffer, CurrentLuaConfig.SizeTypeSize) != nil {
		return false, "failed to write size_t size"
	}
	// write instruction size
	if BufferWriteInt8(assembler.wBuffer, 4) != nil {
		return false, "failed to write instruction size"
	}
	if BufferWriteInt8(assembler.wBuffer, CurrentLuaConfig.IntegerTypeSize) != nil {
		return false, "failed to write integer size"
	}
	if BufferWriteInt8(assembler.wBuffer, CurrentLuaConfig.NumberTypeSize) != nil {
		return false, "failed to write number size"
	}
	if CurrentLuaConfig.IntegerTypeSize == 4 {
		if BufferWriteUInt32(assembler.wBuffer, CurrentLuaConfig.MagicInt32) != nil {
			return false, "failed to write LUAC_INT"
		}
	} else if CurrentLuaConfig.IntegerTypeSize == 8 {
		if BufferWriteUInt64(assembler.wBuffer, CurrentLuaConfig.MagicInt64) != nil {
			return false, "failed to write LUAC_INT"
		}
	} else {
		return false, "unsupported lua_Integer size " + string(CurrentLuaConfig.IntegerTypeSize)
	}
	if CurrentLuaConfig.NumberTypeSize == 4 {
		if BufferWriteFloat32(assembler.wBuffer, CurrentLuaConfig.MagicFloat32) != nil {
			return false, "failed to write LUAC_NUM"
		}
	} else if CurrentLuaConfig.NumberTypeSize == 8 {
		if BufferWriteFloat64(assembler.wBuffer, CurrentLuaConfig.MagicFloat64) != nil {
			return false, "failed to write LUAC_NUM"
		}
	} else {
		return false, "unsupported lua_Number size " + string(CurrentLuaConfig.NumberTypeSize)
	}
	return true, ""
}

/**
 * write proto to bytecode buffer
 */
func (assembler *Assembler) writeFunction(fn *ParsedFunction) (bool, string) {
	wBuffer := assembler.wBuffer
	//fmt.Printf("proto name: %s\n", fn.name)
	funcname := fn.name
	if strings.HasSuffix(funcname, "fake") {
		funcname = ""
	}
	if BufferWriteString(wBuffer, funcname) != nil {
		return false, "write proto name error"
	}
	linedefined, lastlinedefined := 0, 0
	lines := len(fn.lineinfos)
	if lines > 0 {
		linedefined = fn.lineinfos[0]
		lastlinedefined = fn.lineinfos[lines-1]
		for i := 0; i < lines; i++ {
			if fn.lineinfos[i] < linedefined {
				linedefined = fn.lineinfos[i]
			}

			if fn.lineinfos[i] > lastlinedefined {
				lastlinedefined = fn.lineinfos[i]
			}
		}

	}
	//  linedefined, maybe use first instruction's linenumber or use .begin_code directive's linenumber comment
	if BufferWriteUInt32(wBuffer, uint32(linedefined)) != nil {
		return false, "write proto linedefined error"
	}
	//  lastlinedefined
	if BufferWriteUInt32(wBuffer, uint32(lastlinedefined)) != nil {
		return false, "write proto lastlinedefined error"
	}
	if BufferWriteInt8(wBuffer, uint8(fn.params)) != nil {
		return false, "write proto params num error"
	}
	if BufferWriteInt8(wBuffer, uint8(fn.vararg)) != nil {
		return false, "write proto vararg error"
	}
	if BufferWriteInt8(wBuffer, uint8(fn.maxstacksize)) != nil {
		return false, "write proto maxstacksize error"
	}

	var protos []*ParsedFunction
	// fix missing closure pointers
	for i := 0; i < len(fn.usedSubroutines); i++ {
		pName := fn.usedSubroutines[i]
		sub, ok := assembler.functions[pName]
		if !ok {
			return false, "no such function " + pName
		}
		protos = append(protos, sub)
		for j := 0; j < len(fn.neededSubroutines); {
			if j >= len(fn.neededSubroutines) {
				break
			}
			p := fn.neededSubroutines[j]
			if p.left == pName {
				SETARG_Bx(&fn.instructions[p.right], uint(i))
				fn.neededSubroutines = append(fn.neededSubroutines[:j], fn.neededSubroutines[j+1:]...)
				// keep j so j now point to next item
			} else {
				j++
			}
		}
	}

	if BufferWriteUInt32(wBuffer, uint32(len(fn.instructions))) != nil {
		return false, "write function instructions length error"
	}
	for i := 0; i < len(fn.instructions); i++ {
		if BufferWriteUInt32(wBuffer, uint32(fn.instructions[i])) != nil {
			return false, "failed to write instructions"
		}
	}

	if BufferWriteUInt32(wBuffer, uint32(len(fn.constants))) != nil {
		return false, "write function constants length error"
	}
	for i := 0; i < len(fn.constants); i++ {
		constantValue := *fn.constants[i]
		valtype := constantValue.valueType()

		if valtype == LUA_TSTRING && len([]byte(constantValue.str())) > CurrentLuaConfig.MaxShortLen {
			valtype = LUA_TLNGSTR
		}
		if BufferWriteInt8(wBuffer, uint8(valtype)) != nil {
			return false, "write constant value type error " + constantValue.str()
		}
		switch constantValue.valueType() {
		case LUA_TNIL:
			{
				continue
			}
		case LUA_TSTRING:
			strVal, _ := constantValue.(*TString)
			if BufferWriteString(wBuffer, strVal.string_value) != nil {
				return false, "write constant value error " + constantValue.str()
			}
		case LUA_TNUMINT:
			numVal, _ := constantValue.(*TInteger)
			if BufferWriteInt64(wBuffer, numVal.int_value) != nil {
				return false, "write constant value error " + constantValue.str()
			}
			//case LUA_TNUMFLT:
			//  numVal, _ := constantValue.(*TNumber)
			//  if BufferWriteFloat64(wBuffer, numVal.number_value) != nil {
			//    return false, "write constant value error " + constantValue.str()
			//  }
		case LUA_TNUMBER:
			numVal, _ := constantValue.(*TNumber)
			if BufferWriteFloat64(wBuffer, numVal.number_value) != nil {
				return false, "write constant value error " + constantValue.str()
			}
		case LUA_TBOOLEAN:
			boolVal, _ := constantValue.(*TBool)
			if BufferWriteBool(wBuffer, boolVal.bool_value) != nil {
				return false, "write constant value error " + constantValue.str()
			}
		default:
			return false, "unknown constant value type " + strconv.Itoa(constantValue.valueType())
		}
	}

	if BufferWriteUInt32(wBuffer, uint32(len(fn.upvalues))) != nil {
		return false, "write function upvalues length error"
	}

	for i := 0; i < len(fn.upvalues); i++ {
		upvalue := fn.upvalues[i]
		if BufferWriteInt8(wBuffer, upvalue.instack) != nil {
			return false, "write function's upvalue instack error"
		}
		if BufferWriteInt8(wBuffer, upvalue.idx) != nil {
			return false, "write function's upvalue idx error"
		}
	}

	if BufferWriteUInt32(wBuffer, uint32(len(protos))) != nil {
		return false, "write function's sub protos length error"
	}
	for i := 0; i < len(protos); i++ {
		proto := protos[i]
		if ok, err := assembler.writeFunction(proto); !ok {
			return false, err
		}
	}

	// line info size
	if BufferWriteUInt32(wBuffer, uint32(len(fn.lineinfos))) != nil {
		return false, "write function's lineinfos length error"
	}
	for i := 0; i < len(fn.lineinfos); i++ {
		lineinfo := fn.lineinfos[i]
		if BufferWriteUInt32(wBuffer, uint32(lineinfo)) != nil {
			return false, "write function's lineinfos error"
		}
	}

	// locals
	if BufferWriteUInt32(wBuffer, uint32(len(fn.locals))) != nil {
		return false, "write function's local var size error"
	}
	for i := 0; i < len(fn.locals); i++ {
		if BufferWriteString(wBuffer, fn.locals[i].varname) != nil {
			return false, "write function's locals name error"
		}
		if BufferWriteUInt32(wBuffer, uint32(fn.locals[i].startpc)) != nil {
			return false, "write function's local var startpc error"
		}
		if BufferWriteUInt32(wBuffer, uint32(fn.locals[i].endpc)) != nil {
			return false, "write function's local var endpc error"
		}
	}

	//upvalues name
	if BufferWriteUInt32(wBuffer, uint32(len(fn.upvalues))) != nil {
		return false, "write function's upvalues length error"
	}
	for i := 0; i < len(fn.upvalues); i++ {
		if BufferWriteString(wBuffer, fn.upvalues[i].name) != nil {
			return false, "write function's upvalues name error"
		}
	}

	return true, ""
}
func (assembler *Assembler) ParseAsmContent(asmContent string) (result []byte, err error) {
	lines := strings.Split(asmContent, "\n")
	var line string
	for i := 0; i < len(lines); i++ {
		line = strings.Trim(lines[i], "\r\n")
		if len(line) < 1 {
			continue
		}
		res, ParseLineError := assembler.ParseLine(line, len(line))
		if !res {
			err = errors.New("L:" + strconv.Itoa(i+1) + ": " + ParseLineError)
			return
		}
	}
	if len(assembler.instructions) > 0 {
		res, finalizeFunctionErr := assembler.finalizeFunction()
		if !res {
			err = errors.New(finalizeFunctionErr)
			return
		}
	}

	if !assembler.bUpvalues {
		err = errors.New("amount of upvalues never declared")
		return
	}

	if ok, writeHeaderErr := assembler.writeHeader(); !ok {
		err = errors.New(writeHeaderErr)
		return
	}
	if BufferWriteInt8(assembler.wBuffer, uint8(assembler.nUpvalues)) != nil {
		err = errors.New("failed to write num upvalues")
		return
	}
	mainFn, ok := assembler.functions["main"]
	if !ok {
		err = errors.New("no main function")
		return
	}

	if ok, writeFunctionErr := assembler.writeFunction(mainFn); !ok {
		err = errors.New(writeFunctionErr)
		return
	}
	result = assembler.wBuffer.Bytes()
	return
}

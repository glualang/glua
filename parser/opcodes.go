package parser

type opCode uint

const (
	iABC int = iota
	iABx
	iAsBx
	iAx
)

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

const (
	/*----------------------------------------------------------------------
	  name		args	description
	  ------------------------------------------------------------------------*/
	opMove opCode = iota       /*	A B	R(A) := R(B)					*/
	opLoadConstant             /*	A Bx	R(A) := Kst(Bx)					*/
	opLoadConstantEx           /*	A 	R(A) := Kst(extra arg)				*/
	opLoadBool                 /*	A B C	R(A) := (Bool)B; if (C) pc++			*/
	opLoadNil                  /*	A B	R(A), R(A+1), ..., R(A+B) := nil		*/
	opGetUpValue               /*	A B	R(A) := UpValue[B]				*/
	opGetTableUp               /*	A B C	R(A) := UpValue[B][RK(C)]			*/
	opGetTable                 /*	A B C	R(A) := R(B)[RK(C)]				*/
	opSetTableUp               /*	A B C	UpValue[A][RK(B)] := RK(C)			*/
	opSetUpValue               /*	A B	UpValue[B] := R(A)				*/
	opSetTable                 /*	A B C	R(A)[RK(B)] := RK(C)				*/
	opNewTable                 /*	A B C	R(A) := {} (size = B,C)				*/
	opSelf                     /*	A B C	R(A+1) := R(B); R(A) := R(B)[RK(C)]		*/
	opAdd                      /*	A B C	R(A) := RK(B) + RK(C)				*/
	opSub                      /*	A B C	R(A) := RK(B) - RK(C)				*/
	opMul                      /*	A B C	R(A) := RK(B) * RK(C)				*/
	opMod                      /*	A B C	R(A) := RK(B) % RK(C)				*/
	opPow                      /*	A B C	R(A) := RK(B) ^ RK(C)				*/
	opDiv                      /*	A B C	R(A) := RK(B) / RK(C)				*/
	opIDiv                     /*	A B C	R(A) := RK(B) // RK(C)				*/
	opBand                     /*	A B C	R(A) := RK(B) & RK(C)				*/
	opBor                      /*	A B C	R(A) := RK(B) | RK(C)				*/
	opBxor                     /*	A B C	R(A) := RK(B) ~ RK(C)				*/
	opShl                      /*	A B C	R(A) := RK(B) << RK(C)				*/
	opShr                      /*	A B C	R(A) := RK(B) >> RK(C)				*/
	opUnaryMinus               /*	A B	R(A) := -R(B)					*/
	opBnot                     /*	A B	R(A) := ~R(B)					*/
	opNot                      /*	A B	R(A) := not R(B)				*/
	opLength                   /*	A B	R(A) := length of R(B)				*/
	opConcat                   /*	A B C	R(A) := R(B).. ... ..R(C)			*/
	opJump                     /*	A sBx	pc+=sBx; if (A) close all upvalues >= R(A - 1)	*/
	opEqual                    /*	A B C	if ((RK(B) == RK(C)) ~= A) then pc++		*/
	opLessThan                 /*	A B C	if ((RK(B) <  RK(C)) ~= A) then pc++		*/
	opLessOrEqual              /*	A B C	if ((RK(B) <= RK(C)) ~= A) then pc++		*/

	opTest                     /*	A C	if not (R(A) <=> C) then pc++			*/
	opTestSet                  /*	A B C	if (R(B) <=> C) then R(A) := R(B) else pc++	*/
	opCall                     /*	A B C	R(A), ... ,R(A+C-2) := R(A)(R(A+1), ... ,R(A+B-1)) */
	opTailCall                 /*	A B C	return R(A)(R(A+1), ... ,R(A+B-1))		*/
	opReturn                   /*	A B	return R(A), ... ,R(A+B-2)	(see note)	*/
	opForLoop                  /*	A sBx	R(A)+=R(A+2);
		                                         if R(A) <?= R(A+1) then { pc+=sBx; R(A+3)=R(A) }*/
	opForPrep                  /*	A sBx	R(A)-=R(A+2); pc+=sBx				*/
	opTForCall                 /*	A C	R(A+3), ... ,R(A+2+C) := R(A)(R(A+1), R(A+2));	*/
	opTForLoop                 /*	A sBx	if R(A+1) ~= nil then { R(A)=R(A+1); pc += sBx }*/
	opSetList                  /*	A B C	R(A)[(C-1)*FPF+i] := R(A+i), 1 <= i <= B	*/
	opClosure                  /*	A Bx	R(A) := closure(KPROTO[Bx])			*/
	opVarArg                   /*	A B	R(A), R(A+1), ..., R(A+B-2) = vararg		*/
	opExtraArg                 /*	Ax	extra (larger) argument for previous opcode	*/

	// added opcodes for other languages trans to lua bytecode
	opPush                     /* A   top++, evalstack(top) = R(A)  */
	opPop                      /* A   R(A) := evalstack(top), top-- */
	opGetTop                   /* A   R(A) := evalstack(top) */
	opCmp                      /* A B C   R(A) = 1 if RK(B) > RK(C), 0 if RK(B) == RK(C), -1 if RK(B) < RK(C) */
	opCmpEq                    /* A B C R(A) = 1 if RK(B) == RK(C), else 0 */
	opCmpNe                    /* A B C R(A) = 1 if RK(B) != RK(C), else 0 */
	opCmpGt                    /* A B C R(A) = 1 if RK(B) > RK(C), else 0 */
	opCmpLt                    /* A B C R(A) = 1 if RK(B) < RK(C), else 0 */

	NUM_OPCODES
)

var opNames = []string{
	"MOVE",
	"LOADK",
	"LOADKX",
	"LOADBOOL",
	"LOADNIL",
	"GETUPVAL",
	"GETTABUP",
	"GETTABLE",
	"SETTABUP",
	"SETUPVAL",
	"SETTABLE",
	"NEWTABLE",
	"SELF",
	"ADD",
	"SUB",
	"MUL",

	"MOD",
	"POW",
	"DIV",
	"IDIV",
	"BAND",
	"BOR",
	"BXOR",
	"SHL",
	"SHR",
	"UNM",
	"BNOT",
	"NOT",
	"LEN",
	"CONCAT",
	"JMP",
	"EQ",
	"LT",
	"LE",

	"TEST",
	"TESTSET",
	"CALL",
	"TAILCALL",
	"RETURN",
	"FORLOOP",
	"FORPREP",
	"TFORCALL",
	"TFORLOOP",
	"SETLIST",
	"CLOSURE",
	"VARARG",
	"EXTRAARG",

	"PUSH",
	"POP",
	"GETTOP",
	"CMP",

	"CMP_EQ",
	"CMP_NE",
	"CMP_GT",
	"CMP_LT",
}

const (
	sizeC             = 9
	sizeB             = 9
	sizeBx            = sizeC + sizeB
	sizeA             = 8
	sizeAx            = sizeC + sizeB + sizeA
	sizeOp            = 6
	posOp             = 0
	posA              = posOp + sizeOp
	posC              = posA + sizeA
	posB              = posC + sizeC
	posBx             = posC
	posAx             = posA
	bitRK             = 1 << (sizeB - 1)
	maxIndexRK        = bitRK - 1
	maxArgAx          = 1<<sizeAx - 1
	maxArgBx          = 1<<sizeBx - 1
	maxArgSBx         = maxArgBx >> 1 // sBx is signed
	maxArgA           = 1<<sizeA - 1
	maxArgB           = 1<<sizeB - 1
	maxArgC           = 1<<sizeC - 1
	listItemsPerFlush = 50 // # list items to accumulate before a setList instruction
)



func opmode(t, a, b, c, m int) byte { return byte(t<<7 | a<<6 | b<<4 | c<<2 | m) }

const (
	opArgN = iota // argument is not used
	opArgU        // argument is used
	opArgR        // argument is a register or a jump offset
	opArgK        // argument is a constant or register/constant
)

func opMode(m opCode) int     { return int(opModes[m] & 3) }
func bMode(m opCode) byte     { return (opModes[m] >> 4) & 3 }
func cMode(m opCode) byte     { return (opModes[m] >> 2) & 3 }
func testAMode(m opCode) bool { return opModes[m]&(1<<6) != 0 }
func testTMode(m opCode) bool { return opModes[m]&(1<<7) != 0 }

/* this bit 1 means constant (0 means register) */
const BITRK = 1 << (SIZE_B - 1)

// t means is-test-opcode
var opModes []byte = []byte{
	//     T  A    B       C     mode		    opcode
	opmode(0, 1, opArgR, opArgN, iABC),  // opMove
	opmode(0, 1, opArgK, opArgN, iABx),  // opLoadConstant
	opmode(0, 1, opArgN, opArgN, iABx),  // opLoadConstantEx
	opmode(0, 1, opArgU, opArgU, iABC),  // opLoadBool
	opmode(0, 1, opArgU, opArgN, iABC),  // opLoadNil
	opmode(0, 1, opArgU, opArgN, iABC),  // opGetUpValue
	opmode(0, 1, opArgU, opArgK, iABC),  // opGetTableUp
	opmode(0, 1, opArgR, opArgK, iABC),  // opGetTable
	opmode(0, 0, opArgK, opArgK, iABC),  // opSetTableUp
	opmode(0, 0, opArgU, opArgN, iABC),  // opSetUpValue
	opmode(0, 0, opArgK, opArgK, iABC),  // opSetTable
	opmode(0, 1, opArgU, opArgU, iABC),  // opNewTable
	opmode(0, 1, opArgR, opArgK, iABC),  // opSelf
	opmode(0, 1, opArgK, opArgK, iABC),  // opAdd
	opmode(0, 1, opArgK, opArgK, iABC),  // opSub
	opmode(0, 1, opArgK, opArgK, iABC),  // opMul
	opmode(0, 1, opArgK, opArgK, iABC),  // opMod
	opmode(0, 1, opArgK, opArgK, iABC),  // opPow
	opmode(0, 1, opArgK, opArgK, iABC),  // opDiv
	opmode(0, 1, opArgK, opArgK, iABC),  // opIdiv
	opmode(0, 1, opArgK, opArgK, iABC), // opBand
	opmode(0, 1, opArgK, opArgK, iABC), // opBor
	opmode(0, 1, opArgK, opArgK, iABC), // opBxor
	opmode(0, 1, opArgK, opArgK, iABC), // opShl
	opmode(0, 1, opArgK, opArgK, iABC), // opShr
	opmode(0, 1, opArgR, opArgN, iABC),  // opUnaryMinus
	opmode(0, 1, opArgR, opArgN, iABC),  // opBnot
	opmode(0, 1, opArgR, opArgN, iABC),  // opNot

	opmode(0, 1, opArgR, opArgN, iABC),  // opLength
	opmode(0, 1, opArgR, opArgR, iABC),  // opConcat
	opmode(0, 0, opArgR, opArgN, iAsBx), // opJump
	opmode(1, 0, opArgK, opArgK, iABC),  // opEqual
	opmode(1, 0, opArgK, opArgK, iABC),  // opLessThan
	opmode(1, 0, opArgK, opArgK, iABC),  // opLessOrEqual
	opmode(1, 0, opArgN, opArgU, iABC),  // opTest
	opmode(1, 1, opArgR, opArgU, iABC),  // opTestSet
	opmode(0, 1, opArgU, opArgU, iABC),  // opCall
	opmode(0, 1, opArgU, opArgU, iABC),  // opTailCall
	opmode(0, 0, opArgU, opArgN, iABC),  // opReturn
	opmode(0, 1, opArgR, opArgN, iAsBx), // opForLoop
	opmode(0, 1, opArgR, opArgN, iAsBx), // opForPrep
	opmode(0, 0, opArgN, opArgU, iABC),  // opTForCall
	opmode(0, 1, opArgR, opArgN, iAsBx), // opTForLoop
	opmode(0, 0, opArgU, opArgU, iABC),  // opSetList
	opmode(0, 1, opArgU, opArgN, iABx),  // opClosure
	opmode(0, 1, opArgU, opArgN, iABC),  // opVarArg
	opmode(0, 0, opArgU, opArgU, iAx),   // opExtraArg
}

// count of parameters for each instruction
var Opcounts = []int{
	2, // MOVE
	2, // LOADK
	2, // LOADKX
	3, // LOADBOOL
	2, // LOADNIL
	2, // GETUPVAL
	3, // GETTABUP
	3, // GETTABLE
	3, // SETTABUP
	2, // SETUPVAL
	3, // SETTABLE
	3, // NEWTABLE
	3, // SELF
	3, // ADD
	3, // SUB
	3, // MUL
	3, // DIV
	3, // BAND
	3, // BOR
	3, // BXOR
	3, // SHL
	3, // SHR
	3, // MOD
	3, // IDIV
	3, // POW
	2, // UNM
	2, // BNOT
	2, // NOT
	2, // LEN
	3, // CONCAT
	2, // JMP
	3, // EQ
	3, // LT
	3, // LE
	2, // TEST
	3, // TESTSET
	3, // CALL
	3, // TAILCALL
	2, // RETURN
	2, // FORLOOP
	2, // FORPREP
	2, // TFORCALL
	2, // TFORLOOP
	3, // SETLIST
	2, // CLOSURE
	2, // VARARG
	1, // EXTRAARG
	1, // PUSH
	1, // POP
	1, // GETTOP
	3, // CMP
	3, // CMP_EQ
	3, // CMP_NE
	3, // CMP_GT
	3, // CMP_LT
}


const (
	LIMIT_STACKIDX    = 1
	LIMIT_UPVALUE     = 2
	LIMIT_LOCATION    = 4
	LIMIT_CONSTANT    = 8
	LIMIT_EMBED       = 0x10
	LIMIT_PROTO       = 0x20
	LIMIT_CONST_STACK = LIMIT_CONSTANT | LIMIT_STACKIDX
)

// OpPos
const (
	OPP_A     = 0
	OPP_B     = 1
	OPP_C     = 2
	OPP_Ax    = 3
	OPP_Bx    = 4
	OPP_sBx   = 5
	OPP_ARG   = 6
	OPP_C_ARG = 7
)


type OpPos int

type OpInfo struct {
	pos   OpPos
	limit int
}

var Opinfos = [][]OpInfo{ // Maximum of 3 operands
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_STACKIDX}},                                // MOVE
	{{OPP_A, LIMIT_STACKIDX}, {OPP_Bx, LIMIT_CONSTANT}},                               // LOADK
	{{OPP_A, LIMIT_STACKIDX}, {OPP_ARG, LIMIT_CONSTANT}},                              // LOADKX
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_EMBED}, {OPP_C, LIMIT_EMBED}},             // LOADBOOL
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_EMBED}},                                   // LOADNIL
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_UPVALUE}},                                 // GETUPVAL
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_UPVALUE}, {OPP_C, LIMIT_CONST_STACK}},     // GETTABUP
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_STACKIDX}, {OPP_C, LIMIT_CONST_STACK}},    // GETTABLE
	{{OPP_A, LIMIT_UPVALUE}, {OPP_B, LIMIT_CONST_STACK}, {OPP_C, LIMIT_CONST_STACK}},  // SETTABUP
	{{OPP_B, LIMIT_UPVALUE}, {OPP_A, LIMIT_STACKIDX}},                                 /// SETUPVAL
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_CONST_STACK}, {OPP_C, LIMIT_CONST_STACK}}, // SETTABLE
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_EMBED}, {OPP_C, LIMIT_EMBED}},             // NEWTABLE
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_STACKIDX}, {OPP_C, LIMIT_CONST_STACK}},    // SELF
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_CONST_STACK}, {OPP_C, LIMIT_CONST_STACK}}, // ADD
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_CONST_STACK}, {OPP_C, LIMIT_CONST_STACK}}, // SUB
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_CONST_STACK}, {OPP_C, LIMIT_CONST_STACK}}, // MUL
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_CONST_STACK}, {OPP_C, LIMIT_CONST_STACK}}, // DIV
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_CONST_STACK}, {OPP_C, LIMIT_CONST_STACK}}, // BAND
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_CONST_STACK}, {OPP_C, LIMIT_CONST_STACK}}, // BOR
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_CONST_STACK}, {OPP_C, LIMIT_CONST_STACK}}, // BXOR
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_CONST_STACK}, {OPP_C, LIMIT_CONST_STACK}}, // SHL
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_CONST_STACK}, {OPP_C, LIMIT_CONST_STACK}}, // SHR
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_CONST_STACK}, {OPP_C, LIMIT_CONST_STACK}}, // MOD
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_CONST_STACK}, {OPP_C, LIMIT_CONST_STACK}}, // IDIV
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_CONST_STACK}, {OPP_C, LIMIT_CONST_STACK}}, // POW
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_STACKIDX}},                                // UNM
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_STACKIDX}},                                // BNOT
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_STACKIDX}},                                // NOT
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_STACKIDX}},                                // LEN
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_STACKIDX}, {OPP_C, LIMIT_STACKIDX}},       // CONCAT
	{{OPP_A, LIMIT_EMBED}, {OPP_sBx, LIMIT_LOCATION}},                                 // JMP
	{{OPP_A, LIMIT_EMBED}, {OPP_B, LIMIT_CONST_STACK}, {OPP_C, LIMIT_CONST_STACK}},    // EQ
	{{OPP_A, LIMIT_EMBED}, {OPP_B, LIMIT_CONST_STACK}, {OPP_C, LIMIT_CONST_STACK}},    // LT
	{{OPP_A, LIMIT_EMBED}, {OPP_B, LIMIT_CONST_STACK}, {OPP_C, LIMIT_CONST_STACK}},    // LE
	{{OPP_A, LIMIT_STACKIDX}, {OPP_C, LIMIT_EMBED}},                                   // TEST
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_STACKIDX}, {OPP_C, LIMIT_EMBED}},          // TESTSET
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_EMBED}, {OPP_C, LIMIT_EMBED}},             // CALL
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_EMBED}, {OPP_C, LIMIT_EMBED}},             // TAILCALL
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_EMBED}},                                   // RETURN
	{{OPP_A, LIMIT_STACKIDX}, {OPP_sBx, LIMIT_LOCATION}},                              // FORLOOP
	{{OPP_A, LIMIT_STACKIDX}, {OPP_sBx, LIMIT_LOCATION}},                              // FORPREP
	{{OPP_A, LIMIT_STACKIDX}, {OPP_C, LIMIT_EMBED}},                                   // TFORCALL
	{{OPP_A, LIMIT_STACKIDX}, {OPP_sBx, LIMIT_LOCATION}},                              // TFORLOOP
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_EMBED}, {OPP_C_ARG, LIMIT_EMBED}},         // SETLIST
	{{OPP_A, LIMIT_STACKIDX}, {OPP_Bx, LIMIT_PROTO}},                                  // CLOSURE
	{{OPP_A, LIMIT_STACKIDX}, {OPP_Bx, LIMIT_EMBED}},                                  // VARARG

	{{OPP_Ax, LIMIT_EMBED}}, // EXTRAARG

	{{OPP_A, LIMIT_STACKIDX}}, // PUSH
	{{OPP_A, LIMIT_STACKIDX}}, // POP
	{{OPP_A, LIMIT_STACKIDX}}, // GETTOP
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_CONST_STACK}, {OPP_C, LIMIT_CONST_STACK}}, // CMP

	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_CONST_STACK}, {OPP_C, LIMIT_CONST_STACK}}, // CMP_EQ
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_CONST_STACK}, {OPP_C, LIMIT_CONST_STACK}}, // CMP_NE
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_CONST_STACK}, {OPP_C, LIMIT_CONST_STACK}}, // CMP_GT
	{{OPP_A, LIMIT_STACKIDX}, {OPP_B, LIMIT_CONST_STACK}, {OPP_C, LIMIT_CONST_STACK}}, // CMP_LT
}

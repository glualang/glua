package parser

import (
	"fmt"
	"math"
	"reflect"
)

type value interface{}
type float8 int

type pc int

var none value = &struct{}{}

func literalValueString(v value) (result string, ok bool) {
	switch v := v.(type) {
	case string:
		return "\"" + v + "\"", true // TODO: escape string
	case int64:
		return fmt.Sprintf("%d", v), true
	case float64:
		return fmt.Sprintf("%f", v), true
	case nil:
		return "nil", true
	case bool:
		return fmt.Sprintf("%#v", v), true
	}
	return "", false
}

func debugValue(v value) string {
	switch v := v.(type) {
	case string:
		return "'" + v + "'"
	case float64:
		return fmt.Sprintf("%f", v)
	case nil:
		return "nil"
	case bool:
		return fmt.Sprintf("%#v", v)
	}
	return fmt.Sprintf("unknown %#v %s", v, reflect.TypeOf(v).Name())
}

func stack(s []value) string {
	r := fmt.Sprintf("stack (len: %d, cap: %d):\n", len(s), cap(s))
	for i, v := range s {
		r = fmt.Sprintf("%s %d: %s\n", r, i, debugValue(v))
	}
	return r
}

func isFalse(s value) bool {
	if s == nil || s == none {
		return true
	}
	b, isBool := s.(bool)
	return isBool && !b
}

type localVariable struct {
	name           string
	startPC, endPC pc
}

type upValueDesc struct {
	name    string
	isLocal bool
	index   int
}

// prototype extra info for compile
type PrototypeExtra struct {
	labelLocations map[int]string // label locations, code index => label
}

func NewPrototypeExtra() *PrototypeExtra {
	return &PrototypeExtra{
		labelLocations: make(map[int]string),
	}
}

type Prototype struct {
	constants                    []value
	code                         []instruction
	prototypes                   []Prototype
	lineInfo                     []int32
	localVariables               []localVariable
	upValues                     []upValueDesc
	source                       string
	lineDefined, lastLineDefined int
	parameterCount, maxStackSize int
	isVarArg                     bool
	name                         string

	extra *PrototypeExtra
}

func (p *Prototype) upValueName(index int) string {
	if s := p.upValues[index].name; s != "" {
		return s
	}
	return "?"
}

func (p *Prototype) lastLoad(reg int, lastPC pc) (loadPC pc, found bool) {
	var ip, jumpTarget pc
	for ; ip < lastPC; ip++ {
		i, maybe := p.code[ip], false
		switch i.opCode() {
		case opLoadNil:
			maybe = i.a() <= reg && reg <= i.a()+i.b()
		case opTForCall:
			maybe = reg >= i.a()+2
		case opCall, opTailCall:
			maybe = reg >= i.a()
		case opJump:
			if dest := ip + 1 + pc(i.sbx()); ip < dest && dest <= lastPC && dest > jumpTarget {
				jumpTarget = dest
			}
		case opTest:
			maybe = reg == i.a()
		default:
			maybe = testAMode(i.opCode()) && reg == i.a()
		}
		if maybe {
			if ip < jumpTarget { // Can't know loading instruction because code is conditional.
				found = false
			} else {
				loadPC, found = ip, true
			}
		}
	}
	return
}

func (p *Prototype) objectName(reg int, lastPC pc) (name, kind string) {
	if name, isLocal := p.localName(reg+1, lastPC); isLocal {
		return name, "local"
	}
	if pc, found := p.lastLoad(reg, lastPC); found {
		i := p.code[pc]
		switch op := i.opCode(); op {
		case opMove:
			if b := i.b(); b < i.a() {
				return p.objectName(b, pc)
			}
		case opGetTableUp:
			name, kind = p.constantName(i.c(), pc), "local"
			if p.upValueName(i.b()) == "_ENV" {
				kind = "global"
			}
			return
		case opGetTable:
			name, kind = p.constantName(i.c(), pc), "local"
			if v, ok := p.localName(i.b()+1, pc); ok && v == "_ENV" {
				kind = "global"
			}
			return
		case opGetUpValue:
			return p.upValueName(i.b()), "upvalue"
		case opLoadConstant:
			if s, ok := p.constants[i.bx()].(string); ok {
				return s, "constant"
			}
		case opLoadConstantEx:
			if s, ok := p.constants[p.code[pc+1].ax()].(string); ok {
				return s, "constant"
			}
		case opSelf:
			return p.constantName(i.c(), pc), "method"
		}
	}
	return
}

func (p *Prototype) constantName(k int, pc pc) string {
	if isConstant(k) {
		if s, ok := p.constants[constantIndex(k)].(string); ok {
			return s
		}
	} else if name, kind := p.objectName(k, pc); kind == "c" {
		return name
	}
	return "?"
}

func (p *Prototype) localName(index int, pc pc) (string, bool) {
	for i := 0; i < len(p.localVariables) && p.localVariables[i].startPC <= pc; i++ {
		if pc < p.localVariables[i].endPC {
			if index--; index == 0 {
				return p.localVariables[i].name, true
			}
		}
	}
	return "", false
}

// Converts an integer to a "floating point byte", represented as
// (eeeeexxx), where the real value is (1xxx) * 2^(eeeee - 1) if
// eeeee != 0 and (xxx) otherwise.
func float8FromInt(x int) float8 {
	if x < 8 {
		return float8(x)
	}
	e := 0
	for ; x >= 0x10; e++ {
		x = (x + 1) >> 1
	}
	return float8(((e + 1) << 3) | (x - 8))
}

func intFromFloat8(x float8) int {
	e := x >> 3 & 0x1f
	if e == 0 {
		return int(x)
	}
	return int(x&7+8) << uint(e-1)
}


func numberToString(f float64) string {
	return fmt.Sprintf("%.14g", f)
}

func intArith(op Operator, v1, v2 int64) int64 {
	switch op {
	case OpAdd:
		return v1 + v2
	case OpSub:
		return v1 - v2
	case OpMul:
		return v1 * v2
	case OpDiv:
		return v1 / v2
	case OpMod:
		return v1 - v1/v2*v2
	case OpPow:
		// Golang bug: math.Pow(10.0, 33.0) is incorrect by 1 bit.
		if v1 == 10.0 && int64(int(v2)) == v2 {
			return int64(math.Pow10(int(v2)))
		}
		return int64(math.Pow(float64(v1), float64(v2)))
	case OpUnaryMinus:
		return -v1
	}
	panic(fmt.Sprintf("not an arithmetic op code (%d)", op))
}

func arith(op Operator, v1, v2 float64) float64 {
	switch op {
	case OpAdd:
		return v1 + v2
	case OpSub:
		return v1 - v2
	case OpMul:
		return v1 * v2
	case OpDiv:
		return v1 / v2
	case OpMod:
		return v1 - math.Floor(v1/v2)*v2
	case OpPow:
		// Golang bug: math.Pow(10.0, 33.0) is incorrect by 1 bit.
		if v1 == 10.0 && float64(int(v2)) == v2 {
			return math.Pow10(int(v2))
		}
		return math.Pow(v1, v2)
	case OpUnaryMinus:
		return -v1
	}
	panic(fmt.Sprintf("not an arithmetic op code (%d)", op))
}

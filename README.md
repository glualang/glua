gluac
============

glua(lua + static type system) compiler implemented by golang

# Features

* Compatible with Lua5.3 syntax(完全兼容Lua 5.3语法，是Lua 5.3语法的超集)
* Compile-time static type system(在Lua5.3之上增加了编译期的静态类型系统)
* Support for generating pseudo-assembly and bytecode formats(支持生成伪汇编代码和直接生成字节码)
* Supports generating bytecode in Lua5.3 format and bytecode in glua format(支持生成Lua5.3格式的字节码和glua格式的字节码)

# Usage

* `gluac -target binary -vm lua5.3 example/record.lua` 把源码编译并生成Lua5.3格式的字节码
* `gluac -target asm example/record.lua` 把源码编译生成伪汇编文本代码(方便字节码级调试和其他语言开发)

# Example

* sample of type hint, record type and generic type
```lua

type Person1 <T1, T2> = {
    name: string,
    age: T1,
    country: T2,
}

print("Person1 defined")

type Person12<T> = Person1<int, T>

print("Person12 defined")

type Person2 = Person12<string>

print("Person2 defined")

type State = {
    name: string,
    age: int default 1
}

print("State defined")

var a: int = 3
var a_err: int = "error"
let b: string = 'hello'
local c: Person2 = Person2({})
c.name = 'person2[c]'

print('a', a, 'b', b, 'c', c)

-- arith example
print('a'..'b')
print('3//2=', 3//2)
print('3 & 2=', 3&2)
print('3 | 2=', 3|2)
print('3 ~ 2=', 3 ~ 2)
print('~3=', ~a)
print('1<<2=', 1 << 2)
print('7>>2=', 7 >> 2)

let f1: string = 'hello'
let f2: Person2 = Person2()
let f3: Person1 = f2
let f4: Person1<int, string> = Person2()
let f5: (a: int, b: Person2) => int = function(a: int, b: Person2)
    return a
end
let f6: State = Person2()

let b1 = a and 5
print('b1=', b1)
let b2 = a or 5
print('b2=', b2)

var c1 = 1
var c2: int = 2
c1, c2 = 3, '4'
print('c1=', c1, ', c2=', c2)

let d1 = {1, 2}
let d2 = #d1
print("d1=", d1, ", d2=", d2)

let e1 = Person1<string, int>()
e1.name = 'hello, e1'
print("e1=", e1)

```

output asm file(以上源码生成的伪汇编代码):

```
.upvalues 1
.func main 25 0 21
.begin_const
	"print"
	"record.lua started"
	"Person1 defined"
	"Person12 defined"
	"Person2 defined"
	"State defined"
	3
	"error"
	"hello"
	"name"
	"person2[c]"
	"a"
	"b"
	"c"
	"3//2="
	2
	"3 & 2="
	"3 | 2="
	"3 ~ 2="
	"~3="
	"1<<2="
	1
	"7>>2="
	7
	5
	"b1="
	"b2="
	"4"
	"c1="
	", c2="
	"d1="
	", d2="
	"hello, e1"
	"pprint"
	"e1="
.end_const
.begin_upvalue
	1 0 "_ENV"
.end_upvalue
.begin_local
	"Person1" 3 120
	"Person12" 7 120
	"Person2" 11 120
	"State" 15 120
	"a" 20 120
	"a_err" 21 120
	"b" 22 120
	"c" 25 120
	"f1" 68 120
	"f2" 70 120
	"f3" 71 120
	"f4" 73 120
	"f5" 74 120
	"f6" 76 120
	"b1" 79 120
	"b2" 86 120
	"c1" 91 120
	"c2" 92 120
	"d1" 105 120
	"d2" 106 120
	"e1" 114 120
.end_local
.begin_code
	gettabup %0 @0 const "print";L1;
	loadk %1 const "record.lua started";L1;
	call %0 2 1;L1;
	closure %0 Person1;L3;
	gettabup %1 @0 const "print";L9;
	loadk %2 const "Person1 defined";L9;
	call %1 2 1;L9;
	closure %1 Person12;L11;
	gettabup %2 @0 const "print";L13;
	loadk %3 const "Person12 defined";L13;
	call %2 2 1;L13;
	closure %2 Person2;L15;
	gettabup %3 @0 const "print";L17;
	loadk %4 const "Person2 defined";L17;
	call %3 2 1;L17;
	closure %3 State;L19;
	gettabup %4 @0 const "print";L24;
	loadk %5 const "State defined";L24;
	call %4 2 1;L24;
	loadk %4 const 3;L26;
	loadk %5 const "error";L27;
	loadk %6 const "hello";L28;
	move %7 %2;L29;
	newtable %8 0 0;L29;
	call %7 2 2;L29;
	settable %7 const "name" const "person2[c]";L30;
	gettabup %8 @0 const "print";L32;
	loadk %9 const "a";L32;
	move %10 %4;L32;
	loadk %11 const "b";L32;
	move %12 %6;L32;
	loadk %13 const "c";L32;
	move %14 %7;L32;
	call %8 7 1;L32;
	gettabup %8 @0 const "print";L35;
	loadk %9 const "a";L35;
	loadk %10 const "b";L35;
	concat %9 %9 %10;L35;
	call %8 2 1;L35;
	gettabup %8 @0 const "print";L36;
	loadk %9 const "3//2=";L36;
	idiv %10 const 3 const 2;L36;
	call %8 3 1;L36;
	gettabup %8 @0 const "print";L37;
	loadk %9 const "3 & 2=";L37;
	band %10 const 3 const 2;L37;
	call %8 3 1;L37;
	gettabup %8 @0 const "print";L38;
	loadk %9 const "3 | 2=";L38;
	bor %10 const 3 const 2;L38;
	call %8 3 1;L38;
	gettabup %8 @0 const "print";L39;
	loadk %9 const "3 ~ 2=";L39;
	bxor %10 const 3 const 2;L39;
	call %8 3 1;L39;
	gettabup %8 @0 const "print";L40;
	loadk %9 const "~3=";L40;
	bnot %10 %4;L40;
	call %8 3 1;L40;
	gettabup %8 @0 const "print";L41;
	loadk %9 const "1<<2=";L41;
	shl %10 const 1 const 2;L41;
	call %8 3 1;L41;
	gettabup %8 @0 const "print";L42;
	loadk %9 const "7>>2=";L42;
	shr %10 const 7 const 2;L42;
	call %8 3 1;L42;
	loadk %8 const "hello";L44;
	move %9 %2;L45;
	call %9 1 2;L45;
	move %10 %9;L46;
	move %11 %2;L47;
	call %11 1 2;L47;
	closure %12 proto_4;L50;
	move %13 %2;L51;
	call %13 1 2;L51;
	testset %14 %4 0;L53;
	jmp 0 $label_79;L53;
	loadk %14 const 5;L53;
label_79:
	gettabup %15 @0 const "print";L54;
	loadk %16 const "b1=";L54;
	move %17 %14;L54;
	call %15 3 1;L54;
	testset %15 %4 1;L55;
	jmp 0 $label_86;L55;
	loadk %15 const 5;L55;
label_86:
	gettabup %16 @0 const "print";L56;
	loadk %17 const "b2=";L56;
	move %18 %15;L56;
	call %16 3 1;L56;
	loadk %16 const 1;L58;
	loadk %17 const 2;L59;
	loadk %18 const 3;L60;
	loadk %17 const "4";L60;
	move %16 %18;L60;
	gettabup %18 @0 const "print";L61;
	loadk %19 const "c1=";L61;
	move %20 %16;L61;
	loadk %21 const ", c2=";L61;
	move %22 %17;L61;
	call %18 5 1;L61;
	newtable %18 2 0;L63;
	loadk %19 const 1;L63;
	loadk %20 const 2;L63;
	setlist %18 2 1;L63;
	len %19 %18;L64;
	gettabup %20 @0 const "print";L65;
	loadk %21 const "d1=";L65;
	move %22 %18;L65;
	loadk %23 const ", d2=";L65;
	move %24 %19;L65;
	call %20 5 1;L65;
	move %20 %0;L67;
	call %20 1 2;L67;
	settable %20 const "name" const "hello, e1";L68;
	gettabup %21 @0 const "pprint";L69;
	loadk %22 const "e1=";L69;
	move %23 %20;L69;
	call %21 3 1;L69;
	return %0 1;L69;
.end_code

.func Person1 2 1 1
.begin_const
.end_const
.begin_upvalue
.end_upvalue
.begin_local
	"props" 0 8
.end_local
.begin_code
	test %0 0;L7;
	jmp 0 $label_5;L7;
	move %1 %0;L7;
	return %0 2;L7;
	jmp 0 $label_7;L7;
label_5:
	newtable %1 0 0;L7;
	return %1 2;L7;
label_7:
	return %0 1;L7;
.end_code


.func Person12 2 1 1
.begin_const
.end_const
.begin_upvalue
.end_upvalue
.begin_local
	"props" 0 8
.end_local
.begin_code
	test %0 0;L11;
	jmp 0 $label_5;L11;
	move %1 %0;L11;
	return %0 2;L11;
	jmp 0 $label_7;L11;
label_5:
	newtable %1 0 0;L11;
	return %1 2;L11;
label_7:
	return %0 1;L11;
.end_code


.func Person2 2 1 1
.begin_const
.end_const
.begin_upvalue
.end_upvalue
.begin_local
	"props" 0 8
.end_local
.begin_code
	test %0 0;L15;
	jmp 0 $label_5;L15;
	move %1 %0;L15;
	return %0 2;L15;
	jmp 0 $label_7;L15;
label_5:
	newtable %1 0 0;L15;
	return %1 2;L15;
label_7:
	return %0 1;L15;
.end_code


.func State 2 1 1
.begin_const
.end_const
.begin_upvalue
.end_upvalue
.begin_local
	"props" 0 8
.end_local
.begin_code
	test %0 0;L22;
	jmp 0 $label_5;L22;
	move %1 %0;L22;
	return %0 2;L22;
	jmp 0 $label_7;L22;
label_5:
	newtable %1 0 0;L22;
	return %1 2;L22;
label_7:
	return %0 1;L22;
.end_code


.func proto_4 2 2 2
.begin_const
.end_const
.begin_upvalue
.end_upvalue
.begin_local
	"a" 0 2
	"b" 0 2
.end_local
.begin_code
	return %0 2;L49;
	return %0 1;L50;
.end_code
```
print("record.lua started")

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
print('a < a=', a < a) -- 测试小于号不会和泛型参数 G<T1>这种情况冲突

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
let d3 = [1, 2] -- 类似JSON的[a, b, ...]的数组语法
let d4 = { name: 'hello', age: 18 } -- 类似JSON的object语法
print("d1=", d1, ", d2=", d2, ", d3=", d3, ", d4=", d4)

let e1 = Person1<string, int>()
e1.name = 'hello, e1'
pprint("e1=", e1)

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

let f1: string = 'hello'
let f2: Person2 = Person2()
let f3: Person1 = f2
let f4: Person1<int, string> = Person2()
let f5: (a: int, b: Person2) => int = function(a: int, b: Person2)
    return a
end
let f6: State = Person2()

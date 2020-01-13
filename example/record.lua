type Person1 <T1, T2> = {
    name: string,
    age: T1,
    country: T2,
}

type Person2<int, string> = Person1<int, string>

type State = {
    name: string,
    age: int default 1
}

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
let f3: Person1<int, string> = Person2()
let f4: (a: int, b: Person2) => int = function(a: int, b: Person2)
    return a
end
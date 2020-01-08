type Person1 <T1, T2> = {
    name: string,
    age: T1,
    country: T2,
}

type Person2<int, string> = Person1<int, string>

type State = {
    name: string,
    age: int
}

var a: int = 3
let b: string = 'hello'
local c: Person2 = Person2({})
c.name = 'person2[c]'

print('a', a, 'b', b, 'c', c)

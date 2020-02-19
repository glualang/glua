-- add this Contract type when only compile by gluac
type Contract<T> = {
    storage: T
}

type Storage = {
    name: string
}

var M = Contract<Storage>()

function M:init()
    self.storage.name = "test"
end

function M:setName(data: string)
    self.storage.name = data
    emit setName(data)
end

offline function M:query(_: string)
    print("call query")
    return self.storage.name
end

return M
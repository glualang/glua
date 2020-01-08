package parser

import "math"

const (
	maxStack          = 1000000
	maxCallCount      = 200
	maxUpValue        = math.MaxUint8
	idSize            = 60
)

var defaultPath = "./?.lua" // TODO "${LUA_LDIR}?.lua;${LUA_LDIR}?/init.lua;./?.lua"

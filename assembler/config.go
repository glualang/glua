package assembler

// TODO: lua 5.3和glua需要2套config,可以运行时切换

const (
	LUA_VERSION_MAJOR   = 1
	LUA_VERSION_MINOR   = 0
	LUA_VERSION_NUM     = 10
	LUA_VERSION_RELEASE = "0"

	LUA_SIGNATURE = "\x1bGlua"

	LUAC_VERSION = (LUA_VERSION_MAJOR * 16) + LUA_VERSION_MINOR

	LUA_VERSION = "glua " + string(LUA_VERSION_MAJOR) + "." + string(LUA_VERSION_MINOR)
	LUA_RELEASE = string(LUA_VERSION) + "." + (LUA_VERSION_RELEASE)

	LUAC_FORMAT = 0

	LUAC_DATA = "\x19\x93\r\n\x1a\n"

	LUA_COPYRIGHT = LUA_RELEASE + "  Copyright (C) 2020- glua"
	LUA_AUTHORS   = "glua"

	LUAC_INT = 0x5678
	LUAC_NUM = 370.5

	// bytecode headers' size_t type
	LUA_SIZE_T_TYPE_SIZE = 8

	// lua_Integer类型的长度
	LUA_INTEGER_TYPE_SIZE = 8
	// lua_Number类型的长度
	LUA_NUMBER_TYPE_SIZE = 8

	LUAI_MAXSHORTLEN = 40
)

// lua types
const (
	LUA_TNONE          = (-1)
	LUA_TNIL           = 0
	LUA_TBOOLEAN       = 1
	LUA_TLIGHTUSERDATA = 2
	LUA_TNUMBER        = 3
	LUA_TSTRING        = 4
	LUA_TTABLE         = 5
	LUA_TFUNCTION      = 6
	LUA_TUSERDATA      = 7
	LUA_TTHREAD        = 8

	LUA_NUMTAGS = 9

	LUA_TSHRSTR = (LUA_TSTRING | (0 << 4)) /* short strings */
	LUA_TLNGSTR = (LUA_TSTRING | (1 << 4)) /* long strings */

	LUA_TNUMFLT = (LUA_TNUMBER | (0 << 4)) /* float numbers */
	LUA_TNUMINT = (LUA_TNUMBER | (1 << 4)) /* integer numbers */
)

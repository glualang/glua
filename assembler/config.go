package assembler

// TODO: lua 5.3和glua需要2套config,可以运行时切换

type LuaConfig struct {
	VersionMajor int // 主版本号
	VersionMinor int // 次版本号
	VersionRelease string // 次次版本号

	LuaSignature string // 字节码头中的格式签名，用来快速区分是否是错误的格式

	CompilerVersion uint8 // 编译器的版本号
	VersionString string // 版本字符串
	CompilerFormat uint8 // 格式版本
	MagicData string // 字节码头中的魔树，用来快速区分是否是错误的字节码文件

	MagicInt32 uint32 // 字节码头中一个特殊的整数，用来快速检查字节码中32位整数的格式是否错误
	MagicInt64 uint64 // 字节码头中一个特殊的整数，用来快速检查字节码中64位整数的格式是否错误
	MagicFloat32 float32 // 字节码头中一个特殊的整数，用来快速检查字节码中32位浮点数的格式是否错误
	MagicFloat64 float64 // 字节码头中一个特殊的整数，用来快速检查字节码中64位浮点数的格式是否错误

	SizeTypeSize uint8 // bytecode headers' size_t type
	IntegerTypeSize uint8 // lua_Integer类型的长度
	NumberTypeSize uint8 // lua_Number类型的长度

	MaxShortLen int  // lua中短字符串的最大长度，Lua中长字符串和短字符串是两个不同的子类型
}

var Lua53Config = &LuaConfig{
	VersionMajor:    5,
	VersionMinor:    3,
	VersionRelease:  "0",
	LuaSignature:    "\x1bLua",
	CompilerVersion: 5*16+3, // (LUA_VERSION_MAJOR * 16) + LUA_VERSION_MINOR
	VersionString:   "Lua5.3",
	CompilerFormat:  0, // 0 代表lua官方format
	MagicData:       "\x19\x93\r\n\x1a\n",
	MagicInt32:        0x5678,
	MagicInt64:        0x5678,
	MagicFloat32:    370.5,
	MagicFloat64:    370.5,
	SizeTypeSize:    8,
	IntegerTypeSize: 8,
	NumberTypeSize:  8,
	MaxShortLen:     40,
}

var GluaConfig = &LuaConfig{
	VersionMajor:    1,
	VersionMinor:    0,
	VersionRelease:  "0",
	LuaSignature:    "\x1bGlua",
	CompilerVersion: 1*16+0, // (LUA_VERSION_MAJOR * 16) + LUA_VERSION_MINOR
	VersionString:   "glua1.0",
	CompilerFormat:  0,
	MagicData:       "\x19\x93\r\n\x1a\n",
	MagicInt32:        0x5678,
	MagicInt64:        0x5678,
	MagicFloat32:    370.5,
	MagicFloat64:    370.5,
	SizeTypeSize:    8,
	IntegerTypeSize: 8,
	NumberTypeSize:  8,
	MaxShortLen:     40,
}


var CurrentLuaConfig *LuaConfig = Lua53Config

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

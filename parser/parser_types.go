package parser

// 编译期的类型系统

type typeItemType int

const (
	simpleNameType typeItemType = iota // 单独一个符号指向的类型
	simpleNameWithGenericTypesType     // P<T1, T2> 这类类型符号加泛型参数的类型，属于还没resolve过程中的类型
	simpleRecordType                   // record类型
	simpleFuncType                     // 函数类型
	simpleInnerType                    // 内置类型，比如int, string, table, Array, Map等
)

type recordTypePropInfo struct {
	propName string
	propType *typeTreeItem
}

type recordTypeInfo struct {
	name string
	props []*recordTypePropInfo
}

type funcTypeParamInfo struct {
	name string
	typeInfo *typeTreeItem
	isDynamicParams bool // 是否是 ... 参数
}

type typeTreeItem struct {
	itemType typeItemType
	name string
	genericTypeParams []string
	aliasTypeName string
	aliasTypeParams []string

	recordType *recordTypeInfo

	funcTypeParams []*funcTypeParamInfo
	funcReturnType *typeTreeItem
}

var (
	objectTypeTreeItem = &typeTreeItem{itemType:simpleInnerType, name:"object"}
)
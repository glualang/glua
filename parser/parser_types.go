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

type RecordTypePropInfo struct {
	PropName string
	PropType *TypeTreeItem
}

type RecordTypeInfo struct {
	Name  string
	Props []*RecordTypePropInfo
}

type FuncTypeParamInfo struct {
	Name            string
	TypeInfo        *TypeTreeItem
	IsDynamicParams bool // 是否是 ... 参数
}

type TypeTreeItem struct {
	ItemType          typeItemType
	Name              string
	GenericTypeParams []string
	AliasTypeName     string
	AliasTypeParams   []string

	RecordType *RecordTypeInfo

	FuncTypeParams []*FuncTypeParamInfo
	FuncReturnType *TypeTreeItem
}

var (
	objectTypeTreeItem = &TypeTreeItem{ItemType: simpleInnerType, Name:"object"}
)
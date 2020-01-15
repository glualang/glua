package parser

import "fmt"

// 编译期的类型系统

type typeItemType int

const (
	simpleNameType typeItemType = iota // 单独一个符号指向的类型
	simpleNameWithGenericTypesType     // P<T1, T2> 这类类型符号加泛型参数的类型，属于还没resolve过程中的类型
	simpleRecordType                   // record类型
	simpleAliasType                    // 类型重命名或者带泛型参数的类型重命名
	simpleFuncType                     // 函数类型
	simpleInnerType                    // 内置类型，比如int, string, table, Array, Map等

	simpleNotDerivedType               // 暂未推导出的类型
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
	GenericTypeParams []string `json:"GenericTypeParams,omitempty"`
	AliasTypeName     string `json:"AliasTypeName,omitempty"`
	AliasTypeParams   []string `json:"AliasTypeParams,omitempty"`

	RecordType *RecordTypeInfo `json:"RecordType,omitempty"`

	FuncTypeParams []*FuncTypeParamInfo `json:"FuncTypeParams,omitempty"`
	FuncReturnType *TypeTreeItem `json:"FuncReturnType,omitempty"`
}

func (item *TypeTreeItem) String() string {
	switch item.ItemType {
	case simpleNameType:
		return item.Name
	case simpleInnerType:
		return item.Name
	case simpleFuncType:
		return fmt.Sprintf("<func %s(%a)%s>", item.Name, item.FuncTypeParams, item.FuncReturnType.String())
	case simpleRecordType:
		return fmt.Sprintf("<record (%a)>", item.RecordType)
	case simpleNameWithGenericTypesType:
		return fmt.Sprintf("<record %s<%a>>", item.Name, item.GenericTypeParams)
	case simpleNotDerivedType:
		return "<not_derived>"
	default:
		return "uknown type"
	}
}

var (
	objectTypeTreeItem = &TypeTreeItem{ItemType: simpleInnerType, Name:"object"}
	nilTypeTreeItem = &TypeTreeItem{ItemType: simpleInnerType, Name:"nil"}
	boolTypeTreeItem = &TypeTreeItem{ItemType: simpleInnerType, Name:"bool"}
	intTypeTreeItem = &TypeTreeItem{ItemType: simpleInnerType, Name:"int"}
	numberTypeTreeItem = &TypeTreeItem{ItemType: simpleInnerType, Name:"number"}
	stringTypeTreeItem = &TypeTreeItem{ItemType: simpleInnerType, Name:"string"}
	arrayTypeTreeItem = &TypeTreeItem{ItemType: simpleInnerType, Name:"Array"}
	mapTypeTreeItem = &TypeTreeItem{ItemType: simpleInnerType, Name:"Map"}
	tableTypeTreeItem = &TypeTreeItem{ItemType: simpleInnerType, Name:"table"}

	notDerivedTypeTreeItem = &TypeTreeItem{ItemType: simpleNotDerivedType, Name:"not_derived"}
)
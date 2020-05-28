package parser

import (
	"errors"
	"fmt"
	"strings"
)

// 编译期的类型系统

type typeItemType int

const (
	simpleNameType                 typeItemType = iota // 单独一个符号指向的类型
	simpleNameWithGenericTypesType                     // P<T1, T2> 这类类型符号加泛型参数的类型，属于还没resolve过程中的类型
	simpleRecordType                                   // record类型
	simpleAliasType                                    // 类型重命名或者带泛型参数的类型重命名
	simpleFuncType                                     // 函数类型
	simpleInnerType                                    // 内置类型，比如int, string, table, Array, Map等
	simpleNilType                                      // nil类型

	simpleNotDerivedType // 暂未推导出的类型
)

type RecordTypePropInfo struct {
	PropName string
	PropType *TypeTreeItem
	Offline  bool
}

type RecordTypeInfo struct {
	Name  string
	Props []*RecordTypePropInfo
}

func (r *RecordTypeInfo) FindProp(propName string) (result *TypeTreeItem, ok bool) {
	for _, p := range r.Props {
		if p.PropName == propName {
			result = p.PropType
			ok = true
			return
		}
	}
	return
}

func (r *RecordTypeInfo) AddProp(propName string, propType *TypeTreeItem, offline bool) {
	for _, p := range r.Props {
		if p.PropName == propName {
			p.PropType = propType
			p.Offline = offline
			return
		}
	}
	r.Props = append(r.Props, &RecordTypePropInfo{
		PropName: propName,
		PropType: propType,
		Offline:  offline,
	})
}

func (info *RecordTypeInfo) String() string {
	var propsStrs []string
	for _, prop := range info.Props {
		propsStrs = append(propsStrs, fmt.Sprintf("%s %s", prop.PropName, prop.PropType.String()))
	}
	return fmt.Sprintf("record %s<%s>", info.Name, strings.Join(propsStrs, ","))
}

type FuncTypeParamInfo struct {
	Name            string
	TypeInfo        *TypeTreeItem
	IsDynamicParams bool // 是否是 ... 参数
}

func (info *FuncTypeParamInfo) String() string {
	dotsStr := ""
	if info.IsDynamicParams {
		dotsStr = ", ..."
	}
	return fmt.Sprintf("func %s(%s%s)", info.Name, info.TypeInfo.String(), dotsStr)
}

type TypeTreeItem struct {
	ItemType          typeItemType
	Name              string
	GenericTypeParams []*TypeTreeItem `json:"GenericTypeParams,omitempty"`
	AliasTypeName     string          `json:"AliasTypeName,omitempty"`
	AliasTypeParams   []string        `json:"AliasTypeParams,omitempty"`

	RecordType *RecordTypeInfo `json:"RecordType,omitempty"`

	FuncTypeParams []*FuncTypeParamInfo `json:"FuncTypeParams,omitempty"`
	FuncReturnType *TypeTreeItem        `json:"FuncReturnType,omitempty"`
}

// 某个TypeTreeItem在某个name-type binding(链表，多层)下apply得到实际类型的函数
func (item *TypeTreeItem) ApplyBinding(binding *Binding) (result *TypeTreeItem, err error) {
	switch item.ItemType {
	case simpleNameType:
		{
			result = binding.getOrElse(item.Name, item)
			return
		}

	}
	// TODO: 其他itemType类型的处理，需要递归处理成员的类型
	result = item
	return
}

// 对有泛型参数的类型用泛型参数实例化类型产生新的类型
func (item *TypeTreeItem) ApplyRecordGenericTypeParams(genericTypeParams []*TypeTreeItem) (result *TypeTreeItem, err error) {
	if !item.IsRecordType() {
		err = errors.New("can't extend non-record type with generic type params")
		return
	}
	if len(item.GenericTypeParams) < 1 {
		result = item
		return
	}
	result = new(TypeTreeItem)
	*result = *item
	result.RecordType = new(RecordTypeInfo)
	result.GenericTypeParams = item.GenericTypeParams[:]
	*(result.RecordType) = *item.RecordType
	result.RecordType.Props = item.RecordType.Props[:]
	// TODO: 在result.genericTypeParams中应用前len(genericTypeParams)个.需要递归apply类型的props中的类型
	// 目前简单处理只处理record prop的声明的类型
	b := newBinding(nil)
	applyGenericCount := len(result.GenericTypeParams)
	if len(genericTypeParams) < applyGenericCount {
		applyGenericCount = len(genericTypeParams)
	}
	for i := 0; i < applyGenericCount; i++ {
		name := result.GenericTypeParams[i].Name
		valueType := genericTypeParams[i]
		b.bind(name, valueType)
	}
	// 替换成员的类型
	for i, p := range result.RecordType.Props {
		if p.PropType.ItemType == simpleNameType && b.get(p.PropType.Name) != nil {
			newPType := b.get(p.PropType.Name)
			result.RecordType.Props[i] = &RecordTypePropInfo{
				PropName: p.PropName,
				PropType: newPType,
				Offline:  p.Offline,
			}
		}
	}
	// 移除已经被apply的泛型类型
	result.GenericTypeParams = result.GenericTypeParams[applyGenericCount:]
	return
}

func (item *TypeTreeItem) IsRecordType() bool {
	return item.ItemType == simpleRecordType
}

func (item *TypeTreeItem) IsFuncType() bool {
	return item.ItemType == simpleFuncType
}

func (item *TypeTreeItem) IsInnerType() bool {
	return item.ItemType == simpleInnerType
}

func (item *TypeTreeItem) IsSimpleNameType() bool {
	return item.ItemType == simpleNameType
}

func (item *TypeTreeItem) IsSimpleNameWithGenericTypesType() bool {
	return item.ItemType == simpleNameWithGenericTypesType
}

func (item *TypeTreeItem) String() string {
	switch item.ItemType {
	case simpleNameType:
		return item.Name
	case simpleInnerType:
		return item.Name
	case simpleFuncType:
		var paramsStr []string
		for _, p := range item.FuncTypeParams {
			paramsStr = append(paramsStr, p.String())
		}
		return fmt.Sprintf("<func %s(%s)%s>", item.Name, strings.Join(paramsStr, ","), item.FuncReturnType.String())
	case simpleRecordType:
		return fmt.Sprintf("<record (%s)>", item.RecordType.String())
	case simpleNameWithGenericTypesType:
		paramsStrs := make([]string, 0)
		for _, param := range item.GenericTypeParams {
			paramsStrs = append(paramsStrs, param.String())
		}
		return fmt.Sprintf("<record %s<%s>>", item.Name, strings.Join(paramsStrs, ","))
	case simpleNotDerivedType:
		return "<not_derived>"
	case simpleNilType:
		return "nil"
	default:
		return "uknown type"
	}
}

var (
	objectTypeTreeItem = &TypeTreeItem{ItemType: simpleInnerType, Name: "object"}
	nilTypeTreeItem    = &TypeTreeItem{ItemType: simpleNilType}
	boolTypeTreeItem   = &TypeTreeItem{ItemType: simpleInnerType, Name: "bool"}
	intTypeTreeItem    = &TypeTreeItem{ItemType: simpleInnerType, Name: "int"}
	numberTypeTreeItem = &TypeTreeItem{ItemType: simpleInnerType, Name: "number"}
	stringTypeTreeItem = &TypeTreeItem{ItemType: simpleInnerType, Name: "string"}
	arrayTypeTreeItem  = &TypeTreeItem{ItemType: simpleInnerType, Name: "Array"}
	mapTypeTreeItem    = &TypeTreeItem{ItemType: simpleInnerType, Name: "Map"}
	tableTypeTreeItem  = &TypeTreeItem{ItemType: simpleInnerType, Name: "table"}

	notDerivedTypeTreeItem = &TypeTreeItem{ItemType: simpleNotDerivedType, Name: "not_derived"}
)

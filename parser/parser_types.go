package parser

// 编译期的类型系统

type typeItemType int

const (
	simpleNameType typeItemType = iota
	simpleRecordType
	simpleFuncType
	simpleInnerType
)

type recordTypePropInfo struct {
	propName string
	propType *typeTreeItem
}

type recordTypeInfo struct {
	name string
	props []*recordTypePropInfo
}

type typeTreeItem struct {
	itemType typeItemType
	name string
	genericTypeParams []string
	aliasTypeName string
	aliasTypeParams []string

	recordType *recordTypeInfo
}

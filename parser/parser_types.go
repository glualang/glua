package parser

// 编译期的类型系统

type typeItemType int

const (
	simpleNameType typeItemType = 0
	simpleRecordType typeItemType = 1
	simpleFuncType typeItemType = 2
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

type typeTree map[string]*typeTreeItem

func (tree *typeTree) AddItem(item *typeTreeItem) {
	(*tree)[item.name] = item
}

func (tree *typeTree) Contains(typeName string) bool {
	_, ok := (*tree)[typeName]
	return ok
}

func (tree *typeTree) IsRecordType(typeName string) bool {
	info, ok := (*tree)[typeName]
	if !ok {
		return false
	}
	return info.itemType == simpleNameType || info.itemType == simpleRecordType // TODO: 暂时因为没有向上resolve，所以可能是simpleName/record类型
}
package parser

import "log"

// glua compile-time static type system

// 类型信息的约束
type TypeInfoConstraint struct {
	name string // 变量名称
	line int // 使用地方在proto中的行数
	usingAsTypeInfo *typeTreeItem // name被当成什么类型来使用。要求name的实际类型能和这个类型兼容，也就是需要name的类型是usingAsTypeInfo的子类型或者本身
}

// 类型信息作用域
type TypeInfoScope struct {
	names []string
	nameLines map[string]int // 变量申明时所在的proto的函数
	variableTypeInfos map[string]*typeTreeItem
	constraints []*TypeInfoConstraint // 本词法作用域中的类型约束

	children []*TypeInfoScope // 子作用域
	parent *TypeInfoScope // 上级作用域
}

func NewTypeInfoScope() *TypeInfoScope {
	return &TypeInfoScope{
		names:     nil,
		nameLines: make(map[string]int),
		variableTypeInfos: make(map[string]*typeTreeItem),
		children:  nil,
	}
}


func (scope *TypeInfoScope) add(name string, item *typeTreeItem, line int) {
	scope.names = append(scope.names, name)
	scope.variableTypeInfos[name] = item
	scope.nameLines[name] = line
}

func (scope *TypeInfoScope) get(name string) (result *typeTreeItem, line int, ok bool) {
	result, ok = scope.variableTypeInfos[name]
	line, lineOk := scope.nameLines[name]
	if ok && lineOk {
		return
	}
	ok = false
	if scope.parent == nil {
		return
	}
	result, line, ok = scope.parent.get(name)
	return
}

type TypeChecker struct {
	currentProtoScope *TypeInfoScope // 当前parse的proto的类型信息作用域
	rootScope *TypeInfoScope // 根类型信息作用域
}

func NewTypeChecker() *TypeChecker {
	rootScope := NewTypeInfoScope()
	globalTypes := []string{"int", "string", "Array", "Map", "table", "function"}
	for _, t := range globalTypes {
		rootScope.add(t, &typeTreeItem{
			itemType: simpleInnerType,
			name: t,
		}, 0)
	}
	// TODO: 内置函数和内置模块的类型信息需要初始化时加入
	return &TypeChecker{
		rootScope: rootScope,
		currentProtoScope: rootScope,
	}
}

// parser进入一个新的词法作用域的时候需要调用enterLevel
func (checker *TypeChecker) enterLevel() {
	newScope := NewTypeInfoScope()
	newScope.parent = checker.currentProtoScope
	checker.currentProtoScope.children = append(checker.currentProtoScope.children, newScope)
	checker.currentProtoScope = newScope
}

// parser离开一个词法作用域的时候需要调用leaveLevel
func (checker *TypeChecker) leaveLevel() {
	if checker.currentProtoScope.parent == nil {
		log.Fatalln("invalid scope level when TypeChecker::leaveLevel")
		return
	}
	checker.currentProtoScope = checker.currentProtoScope.parent
}

func (checker *TypeChecker) AddGlobalType(name string, item *typeTreeItem, line int) {
	checker.rootScope.add(name, item, line)
}

func (checker *TypeChecker) AddVariable(name string, item *typeTreeItem, line int) {
	checker.currentProtoScope.add(name, item, line)
}

func (checker *TypeChecker) Contains(name string) bool {
	_, _, ok := checker.currentProtoScope.get(name)
	return ok
}

func (checker *TypeChecker) IsRecordType(name string) bool {
	info, _, ok := checker.currentProtoScope.get(name)
	if !ok {
		return false
	}
	return info.itemType == simpleNameType || info.itemType == simpleRecordType // TODO: 暂时因为没有向上resolve，所以可能是simpleName/record类型
}

// 把词法作用域的类型信息树dump成树形字符串用来显示
func (checker *TypeChecker) ToTreeString() (result string, err error) {
	// TODO
	return
}

// 验证整个类型信息树是否正确，包括其中有根据名字引用其他类型暂时还没resolve的也这时候resolve出来验证
func (checker *TypeChecker) Validate() (warnings []error, errs []error) {
	// TODO
	return
}
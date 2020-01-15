package parser

import (
	"encoding/json"
	"fmt"
	"log"
)

// glua compile-time static type system

// 类型信息的约束
type TypeInfoConstraint struct {
	Name            string        // 变量名称
	Line            int           // 使用地方在proto中的行数
	UsingAsTypeInfo *TypeTreeItem // name被当成什么类型来使用。要求name的实际类型能和这个类型兼容，也就是需要name的类型是usingAsTypeInfo的子类型或者本身
}

// 类型信息作用域
type TypeInfoScope struct {
	StartLine int
	EndLine int
	Names             []string
	NameLines         map[string]int // 变量申明时所在的proto的函数
	VariableTypeInfos map[string]*TypeTreeItem  `json:"VariableTypeInfos,omitempty"`
	Constraints       []*TypeInfoConstraint  `json:"Constraints,omitempty"` // 本词法作用域中的类型约束

	Children []*TypeInfoScope   `json:"Children,omitempty"` // 子作用域
	Parent   *TypeInfoScope   `json:"-"` // 上级作用域
}

func NewTypeInfoScope() *TypeInfoScope {
	return &TypeInfoScope{
		Names:             nil,
		NameLines:         make(map[string]int),
		VariableTypeInfos: make(map[string]*TypeTreeItem),
		Children:          nil,
	}
}


func (scope *TypeInfoScope) add(name string, item *TypeTreeItem, line int) {
	scope.Names = append(scope.Names, name)
	scope.VariableTypeInfos[name] = item
	scope.NameLines[name] = line
}

func (scope *TypeInfoScope) get(name string) (result *TypeTreeItem, line int, ok bool) {
	result, ok = scope.VariableTypeInfos[name]
	line, lineOk := scope.NameLines[name]
	if ok && lineOk {
		return
	}
	ok = false
	if scope.Parent == nil {
		return
	}
	result, line, ok = scope.Parent.get(name)
	return
}

// 如果类型信息还没展开(比如是名称，或者是typedef的类型)，则展开这种类型
func (scope *TypeInfoScope) resolve(typeInfo *TypeTreeItem) (result *TypeTreeItem) {
	result = typeInfo
	if typeInfo.ItemType == simpleNameType {
		resolved, _, ok := scope.get(typeInfo.Name)
		if !ok {
			return
		}
		result = resolved
		if result == typeInfo {
			return
		}
		result = scope.resolve(result)
		return
	}
	// typedef等类型的展开，比如P<T1, T2> 展开
	if typeInfo.ItemType == simpleAliasType {
		// TODO: 如果有泛型参数，给alias target type的泛型实例参数增加新项
		resolved, _, ok := scope.get(typeInfo.AliasTypeName)
		if !ok {
			return
		}
		result = resolved
		if result == typeInfo {
			return
		}
		result = scope.resolve(result)
		return
	}
	return
}

type TypeChecker struct {
	CurrentProtoScope *TypeInfoScope `json:"-"`         // 当前parse的proto的类型信息作用域
	RootScope         *TypeInfoScope `json:"RootScope"` // 根类型信息作用域
}

func NewTypeChecker() *TypeChecker {
	rootScope := NewTypeInfoScope()
	globalTypes := []string{"int", "string", "Array", "Map", "table", "function"}
	for _, t := range globalTypes {
		rootScope.add(t, &TypeTreeItem{
			ItemType: simpleInnerType,
			Name:     t,
		}, 0)
	}
	// TODO: 内置函数和内置模块的类型信息需要初始化时加入
	return &TypeChecker{
		RootScope:         rootScope,
		CurrentProtoScope: rootScope,
	}
}

// parser进入一个新的词法作用域的时候需要调用enterLevel
func (checker *TypeChecker) enterLevel(line int) {
	newScope := NewTypeInfoScope()
	newScope.StartLine = line
	newScope.Parent = checker.CurrentProtoScope
	checker.CurrentProtoScope.Children = append(checker.CurrentProtoScope.Children, newScope)
	checker.CurrentProtoScope = newScope
}

// parser离开一个词法作用域的时候需要调用leaveLevel
func (checker *TypeChecker) leaveLevel(line int) {
	if checker.CurrentProtoScope.Parent == nil {
		log.Fatalln("invalid scope level when TypeChecker::leaveLevel")
		return
	}
	checker.CurrentProtoScope.EndLine = line
	checker.CurrentProtoScope = checker.CurrentProtoScope.Parent
}

func (checker *TypeChecker) AddGlobalType(name string, item *TypeTreeItem, line int) {
	checker.RootScope.add(name, item, line)
}

func (checker *TypeChecker) AddVariable(name string, item *TypeTreeItem, line int) {
	checker.CurrentProtoScope.add(name, item, line)
}

func (checker *TypeChecker) AddConstraint(name string, usingAsTypeInfo *TypeTreeItem, line int) {
	checker.CurrentProtoScope.Constraints = append(checker.CurrentProtoScope.Constraints, &TypeInfoConstraint{
		Name: name,
		Line: line,
		UsingAsTypeInfo: usingAsTypeInfo,
	})
}

func (checker *TypeChecker) Contains(name string) bool {
	_, _, ok := checker.CurrentProtoScope.get(name)
	return ok
}

func (checker *TypeChecker) IsRecordType(name string) bool {
	info, _, ok := checker.CurrentProtoScope.get(name)
	if !ok {
		return false
	}
	return info.ItemType == simpleNameType || info.ItemType == simpleRecordType // TODO: 暂时因为没有向上resolve，所以可能是simpleName/record类型
}

// 把词法作用域的类型信息树dump成树形字符串用来显示
func (checker *TypeChecker) ToTreeString() (result string, err error) {
	bytes, err := json.Marshal(checker)
	if err != nil {
		return
	}
	result = string(bytes)
	return
}

// 验证整个类型信息树是否正确，包括其中有根据名字引用其他类型暂时还没resolve的也这时候resolve出来验证
func (scope *TypeInfoScope) Validate() (warnings []error, errs []error) {
	for _, constraint := range scope.Constraints {
		varName := constraint.Name
		varDeclareType, _, ok := scope.get(varName)
		usingAsTypeInfo := constraint.UsingAsTypeInfo
		if !ok {
			warnings = append(warnings, fmt.Errorf("can't find variable %s at line %d", varName, constraint.Line))
			continue
		}
		varDeclareType = scope.resolve(varDeclareType)
		usingAsTypeInfo = scope.resolve(usingAsTypeInfo)

		log.Println(varName)
		if !IsTypeAssignable(usingAsTypeInfo, varDeclareType) {
			warnings = append(warnings, fmt.Errorf("variable %s declared as %s but got %s at line %d",
				varName, varDeclareType.String(), usingAsTypeInfo.String(), constraint.Line))
			continue
		}
	}
	for _, child := range scope.Children {
		subWarnings, subErrors := child.Validate()
		warnings = append(warnings, subWarnings...)
		errs = append(errs, subErrors...)
	}
	return
}

func (checker *TypeChecker) Validate() (warnings []error, errs []error) {
	return checker.RootScope.Validate()
}
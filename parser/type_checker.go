package parser

import (
	"encoding/json"
	"fmt"
	"log"
)

// glua compile-time static type system

type TypeChecker struct {
	CurrentProtoScope *TypeInfoScope `json:"-"`         // 当前parse的proto的类型信息作用域
	RootScope         *TypeInfoScope `json:"RootScope"` // 根类型信息作用域
	Events            []string // emit出的eventName列表
}

func NewTypeChecker() *TypeChecker {
	rootScope := NewTypeInfoScope()
	globalTypes := []string{"int", "string", "Array", "Map", "table", "function"}
	for _, t := range globalTypes {
		rootScope.add(t, &TypeTreeItem{
			ItemType: simpleInnerType,
			Name:     t,
		}, 0, CONST_VARIABLE)
	}
	// TODO: 内置函数和内置模块的类型信息需要初始化时加入
	return &TypeChecker{
		RootScope:         rootScope,
		CurrentProtoScope: rootScope,
		Events: make([]string, 0),
	}
}

func (checker *TypeChecker) AddEventName(eventName string) {
	if !ContainsString(checker.Events, eventName) {
		checker.Events = append(checker.Events, eventName)
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
	checker.RootScope.add(name, item, line, VAR_VARIABLE) // 目前把全局变量当成可变变量
}

func (checker *TypeChecker) AddVariable(name string, item *TypeTreeItem, line int, varType VariableType) {
	checker.CurrentProtoScope.add(name, item, line, varType)
}

func (checker *TypeChecker) AddConstraint(name string, usingAsTypeInfo *TypeTreeItem, line int) {
	checker.CurrentProtoScope.Constraints = append(checker.CurrentProtoScope.Constraints, &TypeInfoConstraint{
		Name:            name,
		Line:            line,
		UsingAsTypeInfo: usingAsTypeInfo,
	})
}

func (checker *TypeChecker) AddAssignConstraint(name string, valueTypeInfo *TypeTreeItem, line int) {
	checker.CurrentProtoScope.AssignConstraints = append(checker.CurrentProtoScope.AssignConstraints, &AssignConstraint{
		Name:          name,
		Line:          line,
		ValueTypeInfo: valueTypeInfo,
	})
}

func (checker *TypeChecker) SetVariableType(name string, valueTypeInfo *TypeTreeItem) {
	checker.CurrentProtoScope.VariableTypeInfos[name] = valueTypeInfo
}

func (checker *TypeChecker) Contains(name string) bool {
	_, _, _, ok := checker.CurrentProtoScope.get(name)
	return ok
}

func (checker *TypeChecker) IsRecordType(name string) bool {
	info, _, _, ok := checker.CurrentProtoScope.get(name)
	if !ok {
		return false
	}
	info = checker.CurrentProtoScope.resolve(info)
	return info.ItemType == simpleRecordType
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
		varDeclareType, _, _, ok := scope.get(varName)
		usingAsTypeInfo := constraint.UsingAsTypeInfo
		if !ok {
			warnings = append(warnings, fmt.Errorf("can't find variable %s at line %d", varName, constraint.Line))
			continue
		}
		varDeclareType = scope.resolve(varDeclareType)
		usingAsTypeInfo = scope.resolve(usingAsTypeInfo)

		if !IsTypeAssignable(usingAsTypeInfo, varDeclareType) {
			warnings = append(warnings, fmt.Errorf("variable %s declared as %s but got %s at line %d",
				varName, varDeclareType.String(), usingAsTypeInfo.String(), constraint.Line))
			continue
		}
	}
	// 找出对变量重新赋值的语句，检查类型和是否const变量
	for _, constraint := range scope.AssignConstraints {
		varName := constraint.Name
		varDeclareType, _, _, ok := scope.get(varName)
		usingAsTypeInfo := constraint.ValueTypeInfo
		if !ok {
			warnings = append(warnings, fmt.Errorf("can't find variable %s at line %d", varName, constraint.Line))
			continue
		}
		varDeclareType = scope.resolve(varDeclareType)
		usingAsTypeInfo = scope.resolve(usingAsTypeInfo)

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

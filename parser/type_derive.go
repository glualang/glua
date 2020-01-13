package parser

import (
	"log"
)

// 类型推导

// 尝试推导表达式的类型
func (checker *TypeChecker) deriveExprType(e exprDesc) (result *TypeTreeItem) {
	switch e.kind {
	case kindTrue:
		return boolTypeTreeItem
	case kindNil:
		return nilTypeTreeItem
	case kindConstant:
		return stringTypeTreeItem // TODO: 多种常量类型。目前这里都是字符串类型
	case kindNumber:
		return numberTypeTreeItem // TODO: 整数或者浮点数
	case kindCall:
		return notDerivedTypeTreeItem // TODO
	case kindLocal:
		return notDerivedTypeTreeItem
	case kindUpValue:
		return notDerivedTypeTreeItem
	default:
		return notDerivedTypeTreeItem
	}
}

// 判断valueType类型是否可以当成declareType用或者赋值给declareType类型
func IsTypeAssignable(valueType *TypeTreeItem, declareType *TypeTreeItem) bool {
	log.Printf("value: %s-%d, declare: %s-%d\n", valueType.Name, valueType.ItemType, declareType.Name, declareType.ItemType)
	if declareType == nil || valueType == nil {
		return true
	}
	if valueType.ItemType == simpleNotDerivedType {
		return true
	}
	if declareType.ItemType == simpleInnerType && declareType.Name == "object" {
		return true
	}
	if valueType.ItemType == simpleInnerType && valueType.Name == "nil" {
		return true
	}

	// TODO: 提前准备类型的继承树，方便判断类型

	if valueType.ItemType == simpleInnerType && declareType.ItemType == simpleInnerType {
		if valueType.Name == declareType.Name {
			return true
		}
		if (declareType.Name == "int" && valueType.Name == "number") || (valueType.Name == "int" && declareType.Name == "number") {
			return true
		}
		if declareType.Name == "table" && (valueType.Name == "Array" || valueType.Name == "Map") {
			return true
		}
		return false
	}
	// TODO
	return true
}
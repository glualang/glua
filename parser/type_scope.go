package parser


// 类型信息的约束
type TypeInfoConstraint struct {
	Name            string        // 变量名称
	Line            int           // 使用地方所在的代码行
	UsingAsTypeInfo *TypeTreeItem // name被当成什么类型来使用。要求name的实际类型能和这个类型兼容，也就是需要name的类型是usingAsTypeInfo的子类型或者本身
}

type VariableType int

const (
	VAR_VARIABLE   VariableType = iota // 可变类型变量
	CONST_VARIABLE                     // 不可变类型变量
)

// 修改变量的语句的约束
type AssignConstraint struct {
	Name          string        // 变量名称
	Line          int           // 所在代码行
	ValueTypeInfo *TypeTreeItem // 新的值的类型
}

// 类型信息作用域
type TypeInfoScope struct {
	StartLine         int
	EndLine           int
	Names             []string
	NameLines         map[string]int           // 变量申明时所在的proto的函数
	NameDeclareTypes  map[string]VariableType  // 变量的变量类型，比如是可变变量还是不可变变量
	VariableTypeInfos map[string]*TypeTreeItem `json:"VariableTypeInfos,omitempty"` // 变量当前的值类型
	Constraints       []*TypeInfoConstraint    `json:"Constraints,omitempty"`       // 本词法作用域中的类型约束
	AssignConstraints []*AssignConstraint      `json:"AssignConstraints,omitempty"` // 本词法作用域中的变量赋值的约束
	ReturnTypes []*TypeTreeItem // 所有返回语句返回的表达式类型

	Children []*TypeInfoScope `json:"Children,omitempty"` // 子作用域
	Parent   *TypeInfoScope   `json:"-"`                  // 上级作用域
}

func NewTypeInfoScope() *TypeInfoScope {
	return &TypeInfoScope{
		Names:             nil,
		NameLines:         make(map[string]int),
		NameDeclareTypes:  make(map[string]VariableType),
		VariableTypeInfos: make(map[string]*TypeTreeItem),
		Children:          nil,
	}
}

func (scope *TypeInfoScope) add(name string, item *TypeTreeItem, line int, varType VariableType) {
	scope.Names = append(scope.Names, name)
	scope.VariableTypeInfos[name] = item
	scope.NameLines[name] = line
	scope.NameDeclareTypes[name] = varType
	scope.VariableTypeInfos[name] = nilTypeTreeItem
}

func (scope *TypeInfoScope) get(name string) (result *TypeTreeItem, line int, varType VariableType, ok bool) {
	result, ok = scope.VariableTypeInfos[name]
	line, lineOk := scope.NameLines[name]
	varType, varTypeOk := scope.NameDeclareTypes[name]
	if ok && lineOk && varTypeOk {
		result, _ = scope.VariableTypeInfos[name]
		return
	}
	ok = false
	if scope.Parent == nil {
		return
	}
	result, line, varType, ok = scope.Parent.get(name)
	return
}

// 如果类型信息还没展开(比如是名称，或者是typedef的类型)，则展开这种类型
func (scope *TypeInfoScope) resolve(typeInfo *TypeTreeItem) (result *TypeTreeItem) {
	result = typeInfo
	if typeInfo.ItemType == simpleNameType {
		resolved, _, _, ok := scope.get(typeInfo.Name)
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
		resolved, _, _, ok := scope.get(typeInfo.AliasTypeName)
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

func (scope *TypeInfoScope) addReturnType(returnType *TypeTreeItem) {
	scope.ReturnTypes = append(scope.ReturnTypes, returnType)
}

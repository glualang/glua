package packager

import "github.com/glualang/gluac/parser"

func DumpCodeInfoFromTypeChecker(checker *parser.TypeChecker) (info *CodeInfo, err error) {
	rootScope := checker.RootScope
	if rootScope == nil {
		return
	}
	if len(rootScope.ReturnTypes) < 1 {
		return
	}
	lastReturnType := rootScope.ReturnTypes[len(rootScope.ReturnTypes)-1]
	if !lastReturnType.IsRecordType() || lastReturnType.RecordType == nil {
		return
	}
	recordTypeInfo := lastReturnType.RecordType
	props := recordTypeInfo.Props
	apis := make([]string, 0)
	offlineApis := make([]string, 0)
	allApis := make([]string, 0)
	for _, prop := range props {
		if prop.PropType.IsFuncType() {
			propName := prop.PropName
			apis = append(apis, propName)
			allApis = append(allApis, propName)
			// 如果是offline的方法属性，则也要加入offlineApis
			if prop.Offline {
				offlineApis = append(offlineApis, propName)
			}
		}
	}
	events := checker.Events

	storagePropertiesTypes := make([][]interface{}, 0)
	// find Storage type
	storageType, ok := recordTypeInfo.FindProp("storage")
	if !ok {
		return
	}
	if !storageType.IsRecordType() {
		return
	}
	for _, p := range storageType.RecordType.Props {
		// 把propType转换成codeInfo中的typeInt
		var propTypeInt StorageValueType = SVT_NIL
		// TODO: p.propType要在binding中进行apply
		if !p.PropType.IsInnerType() && !p.PropType.IsSimpleNameType() && !p.PropType.IsSimpleNameWithGenericTypesType() {
			continue
		}
		innerTypeName := p.PropType.Name
		switch innerTypeName {
		case "bool":
			{
				propTypeInt = SVT_BOOL
			}
		case "int":
			{
				propTypeInt = SVT_INT
			}
		case "number":
			{
				propTypeInt = SVT_NUMBER
			}
		case "string":
			{
				propTypeInt = SVT_STRING
			}
		case "Map":
			{
				if len(p.PropType.GenericTypeParams) < 1 {
					return
				}
				mapParamType := p.PropType.GenericTypeParams[0]
				switch mapParamType.Name {
				case "bool":
					{
						propTypeInt = SVT_BOOL_TABLE
					}
				case "int":
					{
						propTypeInt = SVT_INT_TABLE
					}
				case "number":
					{
						propTypeInt = SVT_NUMBER_TABLE
					}
				case "string":
					{
						propTypeInt = SVT_STRING_TABLE
					}
				default:
					return
				}
			}
		case "Array":
			{
				if len(p.PropType.GenericTypeParams) < 1 {
					return
				}
				mapParamType := p.PropType.GenericTypeParams[0]
				switch mapParamType.Name {
				case "bool":
					{
						propTypeInt = SVT_BOOL_ARRAY
					}
				case "int":
					{
						propTypeInt = SVT_INT_ARRAY
					}
				case "number":
					{
						propTypeInt = SVT_NUMBER_ARRAY
					}
				case "string":
					{
						propTypeInt = SVT_STRING_ARRAY
					}
				default:
					return
				}
			}
		default:
			continue
		}
		if propTypeInt == SVT_NIL {
			continue
		}
		item := []interface{}{p.PropName, propTypeInt}
		storagePropertiesTypes = append(storagePropertiesTypes, item)
	}

	//  set arg type. 并且init，on_deposit, on_deposit_asset, on_upgrade, on_destroy等特殊方法的参数需要特殊处理，其他的参数是字符串的类型
	apiArgsTypes := make([][]interface{}, 0)
	for _, api := range allApis {
		if api == "init" || api == "on_destroy" || api == "on_upgrade" {
			apiArgsTypes = append(apiArgsTypes, []interface{}{api, []interface{}{}})
			continue
		}
		if api == "on_deposit" {
			apiArgsTypes = append(apiArgsTypes, []interface{}{api, []interface{}{LTI_INT}})
			continue
		}
		if api == "on_deposit_asset" {
			apiArgsTypes = append(apiArgsTypes, []interface{}{api, []interface{}{LTI_STRING, LTI_INT}})
			continue
		}
		apiArgsTypes = append(apiArgsTypes, []interface{}{api, []interface{}{LTI_STRING}})
	}

	info = &CodeInfo{
		Apis:                   apis,
		OfflineApis:            offlineApis,
		Events:                 events,
		StoragePropertiesTypes: storagePropertiesTypes,
		ApiArgsTypes:           apiArgsTypes,
	}
	return
}

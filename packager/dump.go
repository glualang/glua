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
			allApis = append(apis, propName)
			// TODO: 如果是offline的方法属性，则也要加入offlineApis
		}
	}
	events := checker.Events

	// TODO: find Storage type

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
	}

	info = &CodeInfo{
		Apis:                   apis,
		OfflineApis:            offlineApis,
		Events:                 events,
		StoragePropertiesTypes: make([][]interface{}, 0),
		ApiArgsTypes:           apiArgsTypes,
	}
	return
}

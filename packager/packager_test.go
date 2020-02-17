package packager

import (
	"log"
	"testing"
)

func TestPackageBytecodeWithCodeInfo(t *testing.T)  {
	bytecode := []byte {1,2,3}
	codeInfo := &CodeInfo{
		Apis:                   []string {"hello", "query"},
		OfflineApis:            []string {"query"},
		Events:                 []string{"Init", "Upgrade"},
		StoragePropertiesTypes: [][]interface{}{{"name", 3}},
		ApiArgsTypes:           [][]interface{} {{"query", []CodeValueType{3}}},
	}
	result, err := PackageBytecodeWithCodeInfo(bytecode, codeInfo)
	if err != nil {
		t.Error(err)
		return
	}
	log.Printf("package result %x", result)
}
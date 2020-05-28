package packager

import "github.com/glualang/gluac/parser"

type CodeValueType int

const (
	LTI_OBJECT        = 0
	LTI_NIL           = 1
	LTI_STRING        = 2
	LTI_INT           = 3
	LTI_NUMBER        = 4
	LTI_BOOL          = 5
	LTI_TABLE         = 6
	LTI_FUNCTION      = 7 // coroutine as function type
	LTI_UNION         = 8
	LTI_RECORD        = 9   // , type <RecordName> = { <name> : <type> , ... }
	LTI_GENERIC       = 10  // ，
	LTI_ARRAY         = 11  // ，
	LTI_MAP           = 12  // ，，key
	LTI_LITERIAL_TYPE = 13  // ，literal type //union，: "Male" | "Female"
	LTI_UNDEFINED     = 100 // ，undefined
)

type StorageValueType int

const (
	SVT_NIL    = 0
	SVT_INT    = 1
	SVT_NUMBER = 2
	SVT_BOOL   = 3
	SVT_STRING = 4
	SVT_STREAM = 5

	SVT_UNKNOWN_TABLE = 50
	SVT_INT_TABLE     = 51
	SVT_NUMBER_TABLE  = 52
	SVT_BOOL_TABLE    = 53
	SVT_STRING_TABLE  = 54
	SVT_STREAM_TABLE  = 55

	SVT_UNKNOWN_ARRAY = 100
	SVT_INT_ARRAY     = 101
	SVT_NUMBER_ARRAY  = 102
	SVT_BOOL_ARRAY    = 103
	SVT_STRING_ARRAY  = 104
	SVT_STREAM_ARRAY  = 105
)

type CodeInfo struct {
	Apis                   []string        `json:"api"`
	OfflineApis            []string        `json:"offline_api"`
	Events                 []string        `json:"event"`
	StoragePropertiesTypes [][]interface{} `json:"storage_properties_types"` // list of [storageName, storageTypeInt] pairs
	ApiArgsTypes           [][]interface{} `json:"api_args_types"`           // list of [apiName, [list of apiArgumentTypes]] pairs
}

func (info CodeInfo) IsOfflineApi(apiName string) bool {
	return parser.ContainsString(info.OfflineApis, apiName)
}

func (info CodeInfo) NonOfflineApis() []string {
	result := make([]string, 0)
	for _, item := range info.Apis {
		if !info.IsOfflineApi(item) {
			result = append(result, item)
		}
	}
	return result
}

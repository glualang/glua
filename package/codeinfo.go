package _package

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

type CodeInfo struct {
	Apis                   []string        `json:"api"`
	OfflineApis            []string        `json:"offline_api"`
	Events                 []string        `json:"event"`
	StoragePropertiesTypes [][]interface{} `json:"storage_properties_types"` // list of [storageName, storageTypeInt] pairs
	ApiArgsTypes [][]interface{} `json:"api_args_types"` // list of [apiName, [list of apiArgumentTypes]] pairs
}

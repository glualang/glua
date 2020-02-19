package packager

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"github.com/glualang/gluac/utils"
)

func writeIntToPackage(stream utils.ByteStream, value int) (err error) {
	// 高位在前,大端序
	uc4 := (byte)((value & 0xFF000000) >> 24)
	uc3 := (byte)((value & 0x00FF0000) >> 16)
	uc2 := (byte)((value & 0x0000FF00) >> 8)
	uc1 := (byte)(value & 0x000000FF)
	err = stream.WriteByte(uc4)
	if err != nil {
		return
	}
	err = stream.WriteByte(uc3)
	if err != nil {
		return
	}
	err = stream.WriteByte(uc2)
	if err != nil {
		return
	}
	err = stream.WriteByte(uc1)
	return
}

func writeStringToPackage(stream utils.ByteStream, str string) (err error) {
	err = writeIntToPackage(stream, len(str))
	if err != nil {
		return
	}
	err = stream.WriteString(str)
	return
}

func readMaybeInt(value interface{}) (result int, err error) {
	intVal, ok := value.(int)
	if ok {
		result = intVal
		return
	}
	int64Val, ok := value.(int64)
	if ok {
		result = int(int64Val)
		return
	}
	float64Val, ok := value.(float64)
	if ok {
		result = int(float64Val)
		return
	}
	float32Val, ok := value.(float32)
	if ok {
		result = int(float32Val)
		return
	}
	numVal, ok := value.(json.Number)
	if ok {
		num64, err2 := numVal.Int64()
		if err2 != nil {
			err = err2
			return
		}
		result = int(num64)
		return
	}
	err = errors.New("decode int error")
	return
}

// 把字节码和源码中的基本信息(各方法，各方法的参数，主类的各字段类型等)
func PackageBytecodeWithCodeInfo(bytecode []byte, codeInfo *CodeInfo) (result []byte, err error) {
	resultBuf := utils.NewSimpleByteStream()
	// bytecode digest
	bytecodeDigest := sha1.Sum(bytecode)
	_, err = resultBuf.Write(bytecodeDigest[:])
	if err != nil {
		return
	}
	// bytecode
	err = writeIntToPackage(resultBuf, len(bytecode))
	if err != nil {
		return
	}
	_, err = resultBuf.Write(bytecode)
	if err != nil {
		return
	}
	// apis
	// apis要排除offline apis
	nonOfflineApis := codeInfo.NonOfflineApis()
	err = writeIntToPackage(resultBuf, len(nonOfflineApis))
	if err != nil {
		return
	}
	for _, api := range nonOfflineApis {
		err = writeStringToPackage(resultBuf, api)
		if err != nil {
			return
		}
	}
	// offline_apis
	err = writeIntToPackage(resultBuf, len(codeInfo.OfflineApis))
	if err != nil {
		return
	}
	for _, api := range codeInfo.OfflineApis {
		err = writeStringToPackage(resultBuf, api)
		if err != nil {
			return
		}
	}
	// events
	err = writeIntToPackage(resultBuf, len(codeInfo.Events))
	if err != nil {
		return
	}
	for _, eventName := range codeInfo.Events {
		err = writeStringToPackage(resultBuf, eventName)
		if err != nil {
			return
		}
	}
	// contract_storage_properties
	err = writeIntToPackage(resultBuf, len(codeInfo.StoragePropertiesTypes))
	if err != nil {
		return
	}
	for _, storageInfoPair := range codeInfo.StoragePropertiesTypes {
		if len(storageInfoPair) != 2 {
			err = errors.New("invalid storage info pairs")
			return
		}
		storageName := storageInfoPair[0].(string)
		err = writeStringToPackage(resultBuf, storageName)
		if err != nil {
			return
		}
		storageType, err2 := readMaybeInt(storageInfoPair[1])
		if err2 != nil {
			err = err2
			return
		}
		err = writeIntToPackage(resultBuf, int(storageType))
		if err != nil {
			return
		}
	}
	// contract_api_arg_types
	err = writeIntToPackage(resultBuf, len(codeInfo.ApiArgsTypes))
	if err != nil {
		return
	}
	for _, apiArgsInfo := range codeInfo.ApiArgsTypes {
		apiName := apiArgsInfo[0].(string)
		err = writeStringToPackage(resultBuf, apiName)
		if err != nil {
			return
		}
		apiArgsJson, err2 := json.Marshal(apiArgsInfo[1])
		if err2 != nil {
			err = err2
			return
		}
		var apiArgs []interface{}
		err = json.Unmarshal(apiArgsJson, &apiArgs)
		if err != nil {
			return
		}
		err = writeIntToPackage(resultBuf, len(apiArgs))
		if err != nil {
			return
		}
		for _, apiArg := range apiArgs {
			argInt, err2 := readMaybeInt(apiArg)
			if err2 != nil {
				err = err2
				return
			}
			err = writeIntToPackage(resultBuf, argInt)
			if err != nil {
				return
			}
		}
	}
	result = resultBuf.ToBytes()
	return
}

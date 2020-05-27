package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/glualang/gluac/assembler"
	"github.com/glualang/gluac/packager"
	"github.com/glualang/gluac/parser"
	"github.com/glualang/gluac/utils"
	"io/ioutil"
	"log"
	"os"
)

var targetTypeFlag = flag.String("target", "asm", "target type(asm or binary or meta)")

var vmTypeFlag = flag.String("vm", "lua53", "target bytecode type(lua53 or glua)")

var packageFlag = flag.Bool("package", false, "package bytecode with code info to single file")

var metaInfoFlag = flag.String("meta", "", "meta info json file path if you want package")

var meterFlag = flag.Bool("meter", false, "add meter op")

type commandType int

const (
	SHOW_HELP_COMMAND commandType = iota
	COMPILE_TO_ASM_COMMAND
)

func isNoArgsCommand(cmdType commandType) bool {
	switch cmdType {
	case SHOW_HELP_COMMAND:
		return true
	default:
		return false
	}
}

func createFileIfNotExists(filepath string) (err error) {
	fileExist, checkFileErr := parser.CheckFileExists(filepath)
	if checkFileErr != nil {
		err = checkFileErr
		return
	}
	if !fileExist {
		f, createFileErr := os.Create(filepath)
		if createFileErr != nil {
			err = createFileErr
			return
		}
		defer f.Close()
	}
	return
}

func programMain() (err error) {
	var programCmdType = COMPILE_TO_ASM_COMMAND
	flag.Parse()

	targetType := *targetTypeFlag
	vmType := *vmTypeFlag
	packageToSingleFile := *packageFlag
	metaInfoFilePath := *metaInfoFlag
	isMeter := *meterFlag

	otherArgs := flag.Args()

	if vmType == "lua53" {
		assembler.CurrentLuaConfig = assembler.Lua53Config
	} else if vmType == "glua" {
		assembler.CurrentLuaConfig = assembler.GluaConfig
	} else {
		err = errors.New("invalid target bytecode type " + vmType)
		return
	}

	if isNoArgsCommand(programCmdType) {
		return
	}
	if len(otherArgs) < 1 {
		fmt.Println("please pass the filename as argument or -h to see help")
		os.Exit(1)
		return
	}
	filename := otherArgs[0]
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()
	r := bufio.NewReader(f)
	proto, typeChecker := parser.ParseToPrototype(r, filename)

	// dump AST tree to tree string
	typeTree, err := typeChecker.ToTreeString()
	if err != nil {
		return
	}
	log.Println("type tree: ", typeTree)

	readFileMode := os.O_RDONLY
	createReadWriteFileMode := os.O_CREATE | os.O_RDWR | os.O_TRUNC
	var writeFilePerMode os.FileMode = 0644

	if targetType == "asm" {
		// dump AST to lua-asm
		dumpAsmProtoFilename := filename + ".asm"
		err = createFileIfNotExists(dumpAsmProtoFilename)
		if err != nil {
			return
		}

		dumpProtoF, openFileErr := os.OpenFile(dumpAsmProtoFilename, createReadWriteFileMode, writeFilePerMode)
		if openFileErr != nil {
			err = openFileErr
			return
		}
		if(isMeter){
			err = proto.AddMeter(true)
			if err != nil {
				return
			}
		}
		defer dumpProtoF.Close()
		asmOutStream := utils.NewSimpleByteStream()
		proto.ToFuncAsm(asmOutStream, true)
		dumpProtoF.Write(asmOutStream.ToBytes())
		_ = proto
	} else if targetType == "binary" {
		// dump AST to binary
		dumpBinaryProtoFilename := filename + ".out"
		err = createFileIfNotExists(dumpBinaryProtoFilename)
		if err != nil {
			return
		}

		dumpProtoF, openFileErr := os.OpenFile(dumpBinaryProtoFilename, createReadWriteFileMode, writeFilePerMode)
		if openFileErr != nil {
			err = openFileErr
			return
		}
		defer dumpProtoF.Close()
		asmOutStream := utils.NewSimpleByteStream()
		proto.ToFuncAsm(asmOutStream, true)
		asmStr := string(asmOutStream.ToBytes())
		ass := assembler.NewAssembler()
		binaryBytes, parseAsmErr := ass.ParseAsmContent(asmStr)
		if parseAsmErr != nil {
			err = parseAsmErr
			return
		}
		dumpProtoF.Write(binaryBytes)
		_ = proto

		if packageToSingleFile {
			// 如果要把字节码和元信息json文件一起打包到单独一个文件的话
			if len(metaInfoFilePath) < 1 {
				err = errors.New("please specify meta info json path")
				return
			}
			metaInfoFile, openMetaInfoFileErr := os.OpenFile(metaInfoFilePath, readFileMode, writeFilePerMode)
			if openMetaInfoFileErr != nil {
				err = openMetaInfoFileErr
				return
			}
			defer metaInfoFile.Close()
			metaInfo, readMetaInfoErr := ioutil.ReadAll(metaInfoFile)
			if readMetaInfoErr != nil {
				err = readMetaInfoErr
				return
			}
			var codeInfo packager.CodeInfo
			err = json.Unmarshal(metaInfo, &codeInfo)
			if err != nil {
				return
			}
			packagedBytes, packageErr := packager.PackageBytecodeWithCodeInfo(binaryBytes, &codeInfo)
			if packageErr != nil {
				err = packageErr
				return
			}
			packagedFileName := filename + ".gpc"
			err = createFileIfNotExists(packagedFileName)
			if err != nil {
				return
			}
			packageFile, openPackageFileErr := os.OpenFile(packagedFileName, createReadWriteFileMode, writeFilePerMode)
			if openPackageFileErr != nil {
				err = openPackageFileErr
				return
			}
			defer packageFile.Close()
			_, err = packageFile.Write(packagedBytes)
			if err != nil {
				return
			}
		}

	} else if targetType == "meta" {
		// generate contract meta info file
		var dumpCodeInfo *packager.CodeInfo
		dumpCodeInfo, err = packager.DumpCodeInfoFromTypeChecker(typeChecker)
		if err != nil {
			return
		}
		var dumpCodeInfoBytes []byte
		dumpCodeInfoBytes, err = json.Marshal(dumpCodeInfo)
		if err != nil {
			return
		}
		codeInfoFilePath := filename + ".gen.meta.json"
		err = createFileIfNotExists(codeInfoFilePath)
		if err != nil {
			return
		}

		dumpCodeInfoF, openFileErr := os.OpenFile(codeInfoFilePath, createReadWriteFileMode, writeFilePerMode)
		if openFileErr != nil {
			err = openFileErr
			return
		}
		defer dumpCodeInfoF.Close()
		_, err = dumpCodeInfoF.Write(dumpCodeInfoBytes)
		if err != nil {
			return
		}
	} else {
		panic("not supported target type " + targetType)
	}

	warinings, compileErrs := typeChecker.Validate()
	if len(warinings) > 0 {
		fmt.Println("compile warnings:")
		for _, warning := range warinings {
			fmt.Println(warning.Error())
		}
	}
	if len(compileErrs) > 0 {
		fmt.Println("compile errors:")
		for _, compileErr := range compileErrs {
			fmt.Println(compileErr.Error())
		}
	}
	return
}

func main() {
	err := programMain()
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("loaded")
}

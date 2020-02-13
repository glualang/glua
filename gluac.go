package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/glualang/gluac/assembler"
	"github.com/glualang/gluac/parser"
	"log"
	"os"
)

var targetTypeFlag = flag.String("target", "asm", "target type(asm or binary)")

var vmTypeFlag = flag.String("vm", "lua53", "target bytecode type(lua53 or glua)")

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

func programMain() (err error) {
	var programCmdType = COMPILE_TO_ASM_COMMAND
	flag.Parse()

	targetType := *targetTypeFlag
	vmType := *vmTypeFlag

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


	if targetType == "asm" {
		// dump AST to lua-asm
		dumpAsmProtoFilename := filename + ".asm"
		dumpProtoFileExists, checkFileErr := parser.CheckFileExists(dumpAsmProtoFilename)
		if checkFileErr != nil {
			err = checkFileErr
			return
		}
		if !dumpProtoFileExists {
			f, createFileErr := os.Create(dumpAsmProtoFilename)
			if createFileErr != nil {
				err = createFileErr
				return
			}
			defer f.Close()
		}

		dumpProtoF, openFileErr := os.OpenFile(dumpAsmProtoFilename, os.O_WRONLY, os.ModeAppend)
		if openFileErr != nil {
			err = openFileErr
			return
		}
		defer dumpProtoF.Close()
		asmOutStream := parser.NewSimpleByteStream()
		proto.ToFuncAsm(asmOutStream, true)
		dumpProtoF.Write(asmOutStream.ToBytes())
		_ = proto
	} else if targetType == "binary" {
		// dump AST to binary
		dumpBinaryProtoFilename := filename + ".out"
		dumpProtoFileExists, checkFileErr := parser.CheckFileExists(dumpBinaryProtoFilename)
		if checkFileErr != nil {
			err = checkFileErr
			return
		}
		if !dumpProtoFileExists {
			f, createFileErr := os.Create(dumpBinaryProtoFilename)
			if createFileErr != nil {
				err = createFileErr
				return
			}
			defer f.Close()
		}

		dumpProtoF, openFileErr := os.OpenFile(dumpBinaryProtoFilename, os.O_WRONLY, os.ModeAppend)
		if openFileErr != nil {
			err = openFileErr
			return
		}
		defer dumpProtoF.Close()
		asmOutStream := parser.NewSimpleByteStream()
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

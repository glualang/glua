package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/zoowii/gluac/parser"
	"log"
	"os"
)

var targetTypeFlag = flag.String("target", "asm", "target type")

type commandType int

const (
	SHOW_HELP_COMMAND commandType = iota
	COMPILE_TO_ASM_COMMAND
)

func isNoArgsCommand(cmdType commandType) bool {
	switch cmdType {
	case SHOW_HELP_COMMAND: return true
	default:
		return false
	}
}

func programMain() (err error) {
	var programCmdType = COMPILE_TO_ASM_COMMAND
	flag.Parse()

	otherArgs := flag.Args()

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

	// dump AST to lua-asm
	dumpProtoFilename := filename + ".asm"
	dumpProtoFileExists, err := parser.CheckFileExists(dumpProtoFilename)
	if err != nil {
		return
	}
	if !dumpProtoFileExists {
		f, createFileErr := os.Create(dumpProtoFilename)
		if createFileErr != nil {
			err = createFileErr
			return
		}
		defer f.Close()
	}
	dumpProtoF, err := os.OpenFile(dumpProtoFilename, os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return
	}
	defer dumpProtoF.Close()
	proto.ToFuncAsm(dumpProtoF, true)
	_ = proto
	return
}

func main()  {
	err := programMain()
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("loaded")
}

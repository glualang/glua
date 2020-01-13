package main

import (
	"bufio"
	"fmt"
	"github.com/zoowii/gluac/parser"
	"log"
	"os"
)

func main()  {
	filename := "example/record.lua" // TODO: read from command line args
	f, err := os.Open(filename)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()
	r := bufio.NewReader(f)
	proto := parser.ParseToPrototype(r, filename)
	// TODO: add glua syntax to scanner and parser
	// TODO: add type checker
	// TODO: dump AST tree to tree string

	// dump AST to lua-asm
	dumpProtoFilename := filename + ".asm"
	os.Remove(dumpProtoFilename)
	os.Create(dumpProtoFilename)
	dumpProtoF, err := os.OpenFile(dumpProtoFilename, os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Println(err)
		return
	}
	defer dumpProtoF.Close()
	proto.ToFuncAsm(dumpProtoF, true)
	_ = proto
	fmt.Println("loaded")
}

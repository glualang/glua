package parser

import (
	"strings"
)

func chunkID(source string) string {
	switch source[0] {
	case '=': // "literal" source
		if len(source) <= idSize {
			return source[1:]
		}
		return source[1:idSize]
	case '@': // file Name
		if len(source) <= idSize {
			return source[1:]
		}
		return "..." + source[1:idSize-3]
	}
	source = strings.Split(source, "\n")[0]
	if l := len("[string \"...\"]"); len(source) > idSize-l {
		return "[string \"" + source + "...\"]"
	}
	return "[string \"" + source + "\"]"
}

package assembler

import (
	"strings"
	"unicode"
)

func Trim(str string) string {
	return strings.Trim(str, " \t\r\n")
}

func StringFirstIndexOf(str string, fn func(byte) bool) int {
	for i := 0; i < len(str); i++ {
		c := str[i]
		if fn(c) {
			return i
		}
	}
	return len(str)
}

func StringFirstIndexOfNot(str string, fn func(byte) bool) int {
	for i := 0; i < len(str); i++ {
		c := str[i]
		if !fn(c) {
			return i
		}
	}
	return len(str)
}

func IsEmptyChar(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n'
}

func IsSymbolStartChar(c rune) bool {
	return '_' == c || unicode.IsLetter(c)
}

func IsSameStringIgnoreCase(a, b string) bool {
	return strings.ToLower(a) == strings.ToLower(b)
}
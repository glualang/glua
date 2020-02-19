package parser

import (
	"log"
	"os"
)

func lua_assert(cond bool) {
	if !cond {
		log.Fatal("assertion failure")
	}
}

func lua_throw_error(err error, msg string) {
	log.Fatalf(err.Error()+": %s\n", msg)
}

func CheckFileExists(filepath string) (bool, error) {
	_, err := os.Stat(filepath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func ContainsString(array []string, term string) bool {
	for _, item := range array {
		if item == term {
			return true
		}
	}
	return false
}
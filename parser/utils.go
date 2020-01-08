package parser

import "log"

func lua_assert(cond bool) {
	if !cond {
		log.Fatal("assertion failure")
	}
}

func lua_throw_error(err error, msg string) {
	log.Fatalf(err.Error() + ": %s\n", msg)
}
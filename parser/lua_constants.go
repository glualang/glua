package parser

import "errors"

// MultipleReturns is the argument for argCount or resultCount in ProtectedCall and Call.
const MultipleReturns = -1

// An Operator is an op argument for Arith.
type Operator int

// Valid Operator values for Arith.
const (
	OpAdd        Operator = iota // Performs addition (+).
	OpSub                        // Performs subtraction (-).
	OpMul                        // Performs multiplication (*).
	OpDiv                        // Performs division (/).
	OpMod                        // Performs modulo (%).
	OpPow                        // Performs exponentiation (^).
	OpUnaryMinus                 // Performs mathematical negation (unary -).
)

// Errors introduced by the Lua VM.
var (
	SyntaxError = errors.New("syntax error")
	MemoryError = errors.New("memory error")
	ErrorError  = errors.New("error within the error handler")
	FileError   = errors.New("file error")
)

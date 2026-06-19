package interpreter

import (
	tunaparser "tunascript/src/parser"
	"tunascript/src/directives"
)

type ValueKind int
type SignalKind int

type RuntimeValue struct {
	Kind  ValueKind
	Value any
}

type ScriptCallback func(hdc uintptr)

type FunctionValue struct {
	Name       string
	Parameters []tunaparser.FunctionParameter
	ReturnType tunaparser.AstType
	Body       tunaparser.BlockStatement
	Env        *Environment
}

type NativeFunction struct {
	Name string
	Call func(args []RuntimeValue) RuntimeValue
}

type EvalResult struct {
	Value  RuntimeValue
	Signal SignalKind
}

type Environment struct {
	variables     map[string]RuntimeValue
	constants     map[string]bool
	declaredTypes map[string]tunaparser.AstType
	parent        *Environment
}

type ExecContext struct {
	inLoop      bool
	inFunction  bool
	filePath    string
	moduleCache map[string]map[string]RuntimeValue
	builtinNames map[string]bool
	rootDir     string
	Cfg          directives.Config
}
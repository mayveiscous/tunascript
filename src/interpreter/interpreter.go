package interpreter

import (
	tunaparser "tunascript/src/parser"

	"tunascript/src/lexer"
	"tunascript/src/directives"
	"path/filepath"
	"fmt"
	"os"
)

func returnResult(v RuntimeValue) EvalResult  {
	return EvalResult{Value: v, Signal: sigReturn} 
}

func breakResult() EvalResult {
	return EvalResult{Value: RuntimeValue{Kind: NullVal}, Signal: sigBreak} 
}

func continueResult() EvalResult {
	return EvalResult{Value: RuntimeValue{Kind: NullVal}, Signal: sigContinue} 
}

func valueResult(v RuntimeValue) EvalResult {
	return EvalResult{Value: v}
}

var currentFilePath string

func SetCurrentFile(path string) {
	currentFilePath = path
}

func TunaError(msg string) string {
	if currentFilePath != "" {
		return fmt.Sprintf("\033[31m[Tunascript Error]\033[0m %s: %s", currentFilePath, msg)
	}
	return fmt.Sprintf("\033[31m[Tunascript Error]\033[0m %s", msg)
}

// schoolRegistry holds all declared school types, keyed by name.
// Populated at runtime when a SchoolStatement is evaluated.
var schoolRegistry = map[string]tunaparser.SchoolStatement{}

func resolveTypeName(t tunaparser.AstType) string {
	switch v := t.(type) {
	case tunaparser.SymbolType:
		return v.Name
	case tunaparser.ArrayType:
		inner := resolveTypeName(v.Underlying)
		if inner != "" {
			return "[]" + inner
		}
		return "array"
	default:
		return ""
	}
}

func checkType(label string, val RuntimeValue, expected tunaparser.AstType) {
	if expected == nil {
		return
	}
	switch exp := expected.(type) {
	case tunaparser.ArrayType:
		if val.Kind == NullVal {
			panic(TunaError(fmt.Sprintf(
				"type mismatch for '%s': cannot assign null to '%s'",
				label, resolveTypeName(expected))))
		}
		if val.Kind != ArrayVal {
			panic(TunaError(fmt.Sprintf(
				"type mismatch for '%s': expected '%s' but got '%s'",
				label, resolveTypeName(expected), val.Kind.String())))
		}
		// Validate each element against the array's element type
		elemTypeName := resolveTypeName(exp.Underlying)
		if elemTypeName != "" {
			for i, elem := range val.Value.([]RuntimeValue) {
				checkType(fmt.Sprintf("%s[%d]", label, i), elem, exp.Underlying)
			}
		}
	case tunaparser.SymbolType:
		want := exp.Name
		if want == "" {
			return
		}
		// Check if this is a user-defined school type
		if school, isSchool := schoolRegistry[want]; isSchool {
			checkSchoolType(label, val, school)
			return
		}
		if val.Kind == NullVal {
			if want == "null" || want == "void" {
				return
			}
			panic(TunaError(fmt.Sprintf(
				"type mismatch for '%s': cannot assign null to '%s'",
				label, want)))
		}
		got := val.Kind.String()
		if got != want {
			panic(TunaError(fmt.Sprintf(
				"type mismatch for '%s': expected '%s' but got '%s'",
				label, want, got)))
		}
	}
}

func checkSchoolType(label string, val RuntimeValue, school tunaparser.SchoolStatement) {
	if val.Kind == NullVal {
		panic(TunaError(fmt.Sprintf(
			"type mismatch for '%s': cannot assign null to '%s'",
			label, school.Name)))
	}
	if val.Kind != ObjectVal {
		panic(TunaError(fmt.Sprintf(
			"type mismatch for '%s': expected '%s' (object) but got '%s'",
			label, school.Name, val.Kind.String())))
	}
	props := val.Value.(map[string]RuntimeValue)

	// Check for missing or mistyped fields
	for _, field := range school.Fields {
		fieldVal, ok := props[field.Name]
		if !ok {
			panic(TunaError(fmt.Sprintf(
				"type mismatch for '%s': missing field '%s' required by '%s'",
				label, field.Name, school.Name)))
		}
		checkType(fmt.Sprintf("%s.%s", label, field.Name), fieldVal, field.Type)
	}

	// Strict: no extra fields beyond what the school defines
	allowed := map[string]bool{}
	for _, field := range school.Fields {
		allowed[field.Name] = true
	}
	for key := range props {
		if !allowed[key] {
			panic(TunaError(fmt.Sprintf(
				"type mismatch for '%s': unexpected field '%s' not defined in '%s'",
				label, key, school.Name)))
		}
	}
}

func (ctx ExecContext) withLoop() ExecContext {
	ctx.inLoop = true
	return ctx
}

func (ctx ExecContext) withFunction() ExecContext {
	ctx.inLoop = false
	ctx.inFunction = true
	return ctx
}

func CallFunctionValue(f FunctionValue, args []RuntimeValue, env *Environment, ctx ExecContext) RuntimeValue {
	hasVariadic := len(f.Parameters) > 0 && f.Parameters[len(f.Parameters)-1].IsVariadic
	fixedCount := len(f.Parameters)
	if hasVariadic {
		fixedCount--
	}

	if !hasVariadic && len(args) != len(f.Parameters) {
		panic("argument mismatch")
	}

	declEnv := f.Env
	if declEnv == nil {
		declEnv = env
	}

	fnEnv := NewEnvironment(declEnv)

	for i := 0; i < fixedCount; i++ {
		param := f.Parameters[i]
		argVal := args[i]

		if param.Type != nil {
			checkType(param.Name, argVal, param.Type)
		}

		fnEnv.SetTyped(param.Name, argVal, param.Type)
	}

	if hasVariadic {
		varParam := f.Parameters[len(f.Parameters)-1]
		rest := args[fixedCount:]
		fnEnv.Set(varParam.Name, RuntimeValue{
			Kind:  ArrayVal,
			Value: rest,
		})
	}

	fnCtx := ctx.withFunction()
	result := EvaluateBlock(f.Body, fnEnv, fnCtx)

	if result.Signal == sigReturn {
		return result.Value
	}

	return RuntimeValue{Kind: NullVal}
}

func Interpret(block tunaparser.BlockStatement, filePath string, cfg directives.Config) RuntimeValue {
	if cfg.Mode == directives.ModeCompile {
		panic(TunaError("compile mode requested ('>?? !compile') but no compiler is available"))
	}

	absPath, _ := filepath.Abs(filePath)
	SetCurrentFile(absPath)

	rootDir := filepath.Dir(absPath)
	env := NewEnvironment(nil)

	builtinNames := map[string]bool{}
	ctx := ExecContext{
		filePath:     absPath,
		moduleCache:  map[string]map[string]RuntimeValue{},
		builtinNames: builtinNames,
		rootDir:      rootDir,
		Cfg:          cfg,
	}

	registerBuiltins(env, ctx)

	for _, stmt := range block.Body {
		result := EvaluateStatement(stmt, env, ctx)
		if result.Signal != sigNone {
			break
		}
	}

	if mainFn, ok := env.Get("main"); ok && mainFn.Kind == FunctionVal {
		if fn, ok := mainFn.Value.(FunctionValue); ok && len(fn.Parameters) == 0 {
			fnEnv := NewEnvironment(fn.Env)
			EvaluateBlock(fn.Body, fnEnv, ctx.withFunction())
		}
	}

	return RuntimeValue{Kind: NullVal}
}

func EvaluateBlock(block tunaparser.BlockStatement, env *Environment, ctx ExecContext) EvalResult {
	result := NullResult
	for _, stmt := range block.Body {
		result = EvaluateStatement(stmt, env, ctx)

		if result.Signal != sigNone {
			return result
		}
	}
	return result
}

func NewREPLEnvironment() *Environment {
	env := NewEnvironment(nil)
	cfg := directives.Config{}
	ctx := ExecContext{Cfg: cfg}
	registerBuiltins(env, ctx)
	return env
}

func NewRuntimeEnvironmentWithRoot(filePath string, rootDir string, ctx ExecContext) *Environment {
	env := NewEnvironment(nil)

	registerBuiltins(env, ctx)

	absPath, _ := filepath.Abs(filePath)
	scriptDir := filepath.Dir(absPath)

	osVal := env.MustGet("os")
	osProps := osVal.Value.(map[string]RuntimeValue)

	osProps["scriptDir"] = RuntimeValue{
		Kind:  StringVal,
		Value: scriptDir,
	}

	osProps["rootDir"] = RuntimeValue{
		Kind:  StringVal,
		Value: rootDir,
	}

	return env
}

func loadModule(importPath string, ctx ExecContext) map[string]RuntimeValue {
	if filepath.Ext(importPath) == "" {
		importPath += ".tuna"
	}

	var absPath string
	var err error

	if filepath.IsAbs(importPath) {
		absPath = importPath
	} else {
		baseDir := filepath.Dir(ctx.filePath)

		joined := filepath.Join(baseDir, importPath)
		absPath, err = filepath.Abs(joined)
		if err != nil {
			panic(TunaError(fmt.Sprintf(
				"cannot resolve import path '%s': %s",
				importPath, err)))
		}
	}

	absPath = filepath.Clean(absPath)

	if cached, ok := ctx.moduleCache[absPath]; ok {
		if cached == nil {
			panic(TunaError(fmt.Sprintf(
				"import cycle detected involving '%s'",
				absPath)))
		}
		return cached
	}

	ctx.moduleCache[absPath] = nil

	src, err := os.ReadFile(absPath)
	if err != nil {
		panic(TunaError(fmt.Sprintf(
			"cannot read module '%s': %s",
			absPath, err)))
	}

	tokens := lexer.Lex(string(src), absPath)
	block := tunaparser.Parse(tokens, absPath)

	exports := executeModule(block, absPath, ctx)

	ctx.moduleCache[absPath] = exports
	return exports
}

func executeModule(block tunaparser.BlockStatement, absPath string, parentCtx ExecContext) map[string]RuntimeValue {
	SetCurrentFile(absPath)
	ctx := ExecContext{
		filePath:     absPath,
		moduleCache:  parentCtx.moduleCache,
		builtinNames: parentCtx.builtinNames,
		rootDir:      parentCtx.rootDir,
	}

	env := NewEnvironment(nil)
	registerBuiltins(env, ctx)

	exports := map[string]RuntimeValue{}

	for _, stmt := range block.Body {
		if cast, ok := stmt.(tunaparser.CastStatement); ok {
			EvaluateStatement(cast, env, ctx)
			switch inner := cast.Inner.(type) {
			case tunaparser.FunctionDecStatement:
				exports[inner.Name] = env.MustGet(inner.Name)
			case tunaparser.VariableDecStatement:
				exports[inner.VariableName] = env.MustGet(inner.VariableName)
			}
		} else {
			EvaluateStatement(stmt, env, ctx)
		}
	}
	return exports
}

func IsTruthy(val RuntimeValue) bool {
	switch val.Kind {
	case BoolVal:
		return val.Value.(bool)
	case NumberVal:
		return val.Value.(float64) != 0
	case StringVal:
		return val.Value.(string) != ""
	case NullVal:
		return false
	default:
		return true
	}
}
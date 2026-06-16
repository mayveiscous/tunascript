package interpreter

import (
	tunaparser "tunascript/src/parser"

	"tunascript/src/lexer"
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

func TunaError(msg string) string {
	return fmt.Sprintf("\033[31m[Tunascript Error]\033[0m %s", msg)
}

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
		elemTypeName := resolveTypeName(exp.Underlying)
		if elemTypeName != "" {
			for _, elem := range val.Value.([]RuntimeValue) {
				got := elem.Kind.String()
				if got != elemTypeName {
					panic(TunaError(fmt.Sprintf(
						"type mismatch for '%s': expected '[]%s' contents but got '%s'",
						label, elemTypeName, got)))
				}
			}
		}
	case tunaparser.SymbolType:
		want := exp.Name
		if want == "" {
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

func (ctx ExecContext) withLoop() ExecContext {
	ctx.inLoop = true
	return ctx
}

func (ctx ExecContext) withFunction() ExecContext {
	ctx.inLoop = false
	ctx.inFunction = true
	return ctx
}


func Interpret(block tunaparser.BlockStatement, filePath string) RuntimeValue {
	env := NewEnvironment(nil)
	registerBuiltins(env)

	builtinNames := map[string]bool{}
	for k := range env.variables {
		 builtinNames[k] = true
	}

	absPath, _ := filepath.Abs(filePath)
	ctx := ExecContext{
		 filePath:     absPath,
		 moduleCache:  map[string]map[string]RuntimeValue{},
		 builtinNames: builtinNames,
	}

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

func loadModule(importPath string, ctx ExecContext) map[string]RuntimeValue {
	if filepath.Ext(importPath) == "" {
		importPath += ".tuna"
	}

	var absPath string
	var err error

	baseDir := filepath.Dir(ctx.filePath)
	if filepath.IsAbs(importPath) {
		absPath = importPath
	} else {
		absPath, err = filepath.Abs(filepath.Join(baseDir, importPath))
		if err != nil {
			panic(TunaError(fmt.Sprintf("cannot resolve import path '%s': %s", importPath, err)))
		}
	}

	if cached, ok := ctx.moduleCache[absPath]; ok {
		if cached == nil {
			panic(TunaError(fmt.Sprintf(
				"import cycle detected involving '%s'", absPath)))
		}
		return cached
	}

	ctx.moduleCache[absPath] = nil

	src, err := os.ReadFile(absPath)
	if err != nil {
		panic(TunaError(fmt.Sprintf("cannot read module '%s': %s", absPath, err)))
	}

	tokens := lexer.Lex(string(src))
	block := tunaparser.Parse(tokens)

	exports := executeModule(block, absPath, ctx)
	ctx.moduleCache[absPath] = exports
	return exports
}

func executeModule(block tunaparser.BlockStatement, absPath string, parentCtx ExecContext) map[string]RuntimeValue {
	env := NewEnvironment(nil)
	registerBuiltins(env)
	exports := map[string]RuntimeValue{}
	ctx := ExecContext{
		filePath:    absPath,
		moduleCache: parentCtx.moduleCache,
		builtinNames: parentCtx.builtinNames,
	}

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
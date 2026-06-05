package interpreter

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"test-go/src/lexer"
	"test-go/src/parser"
)

type ValueKind int

const (
	NumberVal ValueKind = iota
	StringVal
	BoolVal
	NullVal
	FunctionVal
	ArrayVal
	ObjectVal
)

func (k ValueKind) String() string {
	switch k {
	case NumberVal:
		return "number"
	case StringVal:
		return "string"
	case BoolVal:
		return "bool"
	case NullVal:
		return "null"
	case FunctionVal:
		return "function"
	case ArrayVal:
		return "array"
	case ObjectVal:
		return "object"
	default:
		return "unknown"
	}
}

type RuntimeValue struct {
	Kind  ValueKind
	Value any
}

type FunctionValue struct {
	Name       string
	Parameters []parser.FunctionParameter
	ReturnType parser.AstType
	Body       parser.BlockStatement
	Env        *Environment
}

type NativeFunction struct {
	Name string
	Call func(args []RuntimeValue) RuntimeValue
}

type signalKind int

const (
	sigNone signalKind = iota
	sigReturn
	sigBreak
	sigContinue
)

type EvalResult struct {
	Value  RuntimeValue
	Signal signalKind
}

var nullResult = EvalResult{Value: RuntimeValue{Kind: NullVal}}

func returnResult(v RuntimeValue) EvalResult  { return EvalResult{Value: v, Signal: sigReturn} }
func breakResult() EvalResult                 { return EvalResult{Value: RuntimeValue{Kind: NullVal}, Signal: sigBreak} }
func continueResult() EvalResult              { return EvalResult{Value: RuntimeValue{Kind: NullVal}, Signal: sigContinue} }
func valueResult(v RuntimeValue) EvalResult   { return EvalResult{Value: v} }

type Environment struct {
	variables     map[string]RuntimeValue
	constants     map[string]bool
	declaredTypes map[string]parser.AstType
	parent        *Environment
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		variables:     map[string]RuntimeValue{},
		constants:     map[string]bool{},
		declaredTypes: map[string]parser.AstType{},
		parent:        parent,
	}
}

func (e *Environment) Update(name string, val RuntimeValue) (RuntimeValue, error) {
	if _, ok := e.variables[name]; ok {
		if e.constants[name] {
			return RuntimeValue{}, fmt.Errorf("cannot reassign constant '%s'", name)
		}
		e.variables[name] = val
		return val, nil
	}
	if e.parent != nil {
		return e.parent.Update(name, val)
	}
	return RuntimeValue{}, fmt.Errorf("cannot assign to undefined variable '%s'", name)
}

func (e *Environment) MustUpdate(name string, val RuntimeValue) RuntimeValue {
	v, err := e.Update(name, val)
	if err != nil {
		panic(tunaError(err.Error()))
	}
	return v
}

func (e *Environment) Set(name string, val RuntimeValue) RuntimeValue {
	e.variables[name] = val
	return val
}

func (e *Environment) SetTyped(name string, val RuntimeValue, t parser.AstType) RuntimeValue {
	e.variables[name] = val
	if t != nil {
		e.declaredTypes[name] = t
	}
	return val
}

func (e *Environment) SetConst(name string, val RuntimeValue) RuntimeValue {
	e.variables[name] = val
	e.constants[name] = true
	return val
}

func (e *Environment) Get(name string) (RuntimeValue, bool) {
	if val, ok := e.variables[name]; ok {
		return val, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return RuntimeValue{}, false
}

func (e *Environment) MustGet(name string) RuntimeValue {
	v, ok := e.Get(name)
	if !ok {
		panic(tunaError(fmt.Sprintf("undefined variable '%s'", name)))
	}
	return v
}

func (e *Environment) GetDeclaredType(name string) parser.AstType {
	if t, ok := e.declaredTypes[name]; ok {
		return t
	}
	if e.parent != nil {
		return e.parent.GetDeclaredType(name)
	}
	return nil
}

func resolveTypeName(t parser.AstType) string {
	switch v := t.(type) {
	case parser.SymbolType:
		return v.Name
	case parser.ArrayType:
		inner := resolveTypeName(v.Underlying)
		if inner != "" {
			return "[]" + inner
		}
		return "array"
	default:
		return ""
	}
}

func checkType(label string, val RuntimeValue, expected parser.AstType) {
	if expected == nil {
		return
	}
	switch exp := expected.(type) {
	case parser.ArrayType:
		if val.Kind == NullVal {
			panic(tunaError(fmt.Sprintf(
				"type mismatch for '%s': cannot assign null to '%s'",
				label, resolveTypeName(expected))))
		}
		if val.Kind != ArrayVal {
			panic(tunaError(fmt.Sprintf(
				"type mismatch for '%s': expected '%s' but got '%s'",
				label, resolveTypeName(expected), val.Kind.String())))
		}
		elemTypeName := resolveTypeName(exp.Underlying)
		if elemTypeName != "" {
			for _, elem := range val.Value.([]RuntimeValue) {
				got := elem.Kind.String()
				if got != elemTypeName {
					panic(tunaError(fmt.Sprintf(
						"type mismatch for '%s': expected '[]%s' contents but got '%s'",
						label, elemTypeName, got)))
				}
			}
		}
	case parser.SymbolType:
		want := exp.Name
		if want == "" {
			return
		}
		if val.Kind == NullVal {
			if want == "null" || want == "void" {
				return
			}
			panic(tunaError(fmt.Sprintf(
				"type mismatch for '%s': cannot assign null to '%s'",
				label, want)))
		}
		got := val.Kind.String()
		if got != want {
			panic(tunaError(fmt.Sprintf(
				"type mismatch for '%s': expected '%s' but got '%s'",
				label, want, got)))
		}
	}
}

type execContext struct {
	inLoop     bool
	inFunction bool
	filePath   string
	moduleCache map[string]map[string]RuntimeValue
	builtinNames map[string]bool
}

func (ctx execContext) withLoop() execContext {
	ctx.inLoop = true
	return ctx
}

func (ctx execContext) withFunction() execContext {
	ctx.inLoop = false
	ctx.inFunction = true
	return ctx
}

func Interpret(block parser.BlockStatement, filePath string) RuntimeValue {
	env := NewEnvironment(nil)
	registerBuiltins(env)

	builtinNames := map[string]bool{}
	for k := range env.variables {
		builtinNames[k] = true
	}

	absPath, _ := filepath.Abs(filePath)
	ctx := execContext{
		filePath:    absPath,
		moduleCache: map[string]map[string]RuntimeValue{},
		builtinNames: builtinNames,
	}

	result := nullResult
	for _, stmt := range block.Body {
		result = EvaluateStatement(stmt, env, ctx)
		if result.Signal != sigNone {
			break
		}
	}
	return result.Value
}

func EvaluateStatement(stmt parser.Statement, env *Environment, ctx execContext) EvalResult {
	switch s := stmt.(type) {

	case parser.ExpressionStatement:
		val := EvaluateExpression(s.Expression, env, ctx)
		return valueResult(val)

	case parser.VariableDecStatement:
		val := RuntimeValue{Kind: NullVal}
		if s.AssignedValue != nil {
			val = EvaluateExpression(s.AssignedValue, env, ctx)
			checkType(s.VariableName, val, s.ExplicitType)
		}
		if s.IsConstant {
			env.SetConst(s.VariableName, val)
		} else {
			env.SetTyped(s.VariableName, val, s.ExplicitType)
		}
		return valueResult(val)

	case parser.BreakStatement:
		if !ctx.inLoop {
			if ctx.inFunction {
				panic(tunaError("`break` cannot cross a function boundary"))
			}
			panic(tunaError("`break` can only be used inside a loop"))
		}
		return breakResult()

	case parser.ContinueStatement:
		if !ctx.inLoop {
			if ctx.inFunction {
				panic(tunaError("`continue` cannot cross a function boundary"))
			}
			panic(tunaError("`continue` can only be used inside a loop"))
		}
		return continueResult()

	case parser.ReturnStatement:
		if !ctx.inFunction {
			panic(tunaError("`serve` can only be used inside a function"))
		}
		val := EvaluateExpression(s.Value, env, ctx)
		return returnResult(val)

	case parser.IfStatement:
		condition := EvaluateExpression(s.Condition, env, ctx)
		if isTruthy(condition) {
			return EvaluateBlock(s.Then, NewEnvironment(env), ctx)
		} else if s.Else != nil {
			return EvaluateBlock(*s.Else, NewEnvironment(env), ctx)
		}
		return nullResult

	case parser.WhileStatement:
		result := nullResult
		loopCtx := ctx.withLoop()
		for {
			if !isTruthy(EvaluateExpression(s.Condition, env, ctx)) {
				break
			}
			r := EvaluateBlock(s.Body, NewEnvironment(env), loopCtx)
			switch r.Signal {
			case sigBreak:
				return nullResult
			case sigContinue:
				continue
			case sigReturn:
				return r
			}
			result = r
		}
		return result

	case parser.ForInStatement:
		iterable := EvaluateExpression(s.Iterable, env, ctx)
		if iterable.Kind != ArrayVal {
			panic(tunaError(fmt.Sprintf(
				"cannot iterate over non-array value of type '%s' in for/in", iterable.Kind)))
		}
		result := nullResult
		loopCtx := ctx.withLoop()
		for _, element := range iterable.Value.([]RuntimeValue) {
			loopEnv := NewEnvironment(env)
			loopEnv.Set(s.Iterator, element)
			r := EvaluateBlock(s.Body, loopEnv, loopCtx)
			switch r.Signal {
			case sigBreak:
				return nullResult
			case sigContinue:
				continue
			case sigReturn:
				return r
			}
			result = r
		}
		return result

	case parser.FunctionDecStatement:
		closureEnv := NewEnvironment(env)
		fn := RuntimeValue{
			Kind: FunctionVal,
			Value: FunctionValue{
				Name:       s.Name,
				Parameters: s.Parameters,
				ReturnType: s.ReturnType,
				Body:       s.Body,
				Env:        closureEnv,
			},
		}
		closureEnv.Set(s.Name, fn)

		if ctx.builtinNames[s.Name] {
			fmt.Printf("\033[33m[TunaScript Warning]\033[0m overwriting builtin '%s'\n", s.Name)
		}
		env.Set(s.Name, fn)
		return valueResult(fn)

	case parser.BlockStatement:
		return EvaluateBlock(s, NewEnvironment(env), ctx)

	case parser.CastStatement:
		return EvaluateStatement(s.Inner, env, ctx)

	case parser.ImportStatement:
		exports := loadModule(s.Path, ctx)
		for _, item := range s.Items {
			val, ok := exports[item.Name]
			if !ok {
				panic(tunaError(fmt.Sprintf(
					"module '%s' does not export '%s'", s.Path, item.Name)))
			}
			env.Set(item.Alias, val)
		}
		return nullResult

	default:
		panic(tunaError(fmt.Sprintf("unknown statement type: %T", stmt)))
	}
}

func EvaluateBlock(block parser.BlockStatement, env *Environment, ctx execContext) EvalResult {
	result := nullResult
	for _, stmt := range block.Body {
		result = EvaluateStatement(stmt, env, ctx)

		if result.Signal != sigNone {
			return result
		}
	}
	return result
}

func loadModule(importPath string, ctx execContext) map[string]RuntimeValue {
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
			panic(tunaError(fmt.Sprintf("cannot resolve import path '%s': %s", importPath, err)))
		}
	}

	if cached, ok := ctx.moduleCache[absPath]; ok {
		if cached == nil {
			panic(tunaError(fmt.Sprintf(
				"import cycle detected involving '%s'", absPath)))
		}
		return cached
	}

	ctx.moduleCache[absPath] = nil

	src, err := os.ReadFile(absPath)
	if err != nil {
		panic(tunaError(fmt.Sprintf("cannot read module '%s': %s", absPath, err)))
	}

	tokens := lexer.Lex(string(src))
	block := parser.Parse(tokens)

	exports := executeModule(block, absPath, ctx)
	ctx.moduleCache[absPath] = exports
	return exports
}

func executeModule(block parser.BlockStatement, absPath string, parentCtx execContext) map[string]RuntimeValue {
	env := NewEnvironment(nil)
	registerBuiltins(env)
	exports := map[string]RuntimeValue{}
	ctx := execContext{
		filePath:    absPath,
		moduleCache: parentCtx.moduleCache,
		builtinNames: parentCtx.builtinNames,
	}

	for _, stmt := range block.Body {
		if cast, ok := stmt.(parser.CastStatement); ok {
			EvaluateStatement(cast, env, ctx)
			switch inner := cast.Inner.(type) {
			case parser.FunctionDecStatement:
				exports[inner.Name] = env.MustGet(inner.Name)
			case parser.VariableDecStatement:
				exports[inner.VariableName] = env.MustGet(inner.VariableName)
			}
		} else {
			EvaluateStatement(stmt, env, ctx)
		}
	}
	return exports
}

func EvaluateExpression(expr parser.Expression, env *Environment, ctx execContext) RuntimeValue {
	switch e := expr.(type) {

	case parser.NumberExpression:
		return RuntimeValue{Kind: NumberVal, Value: e.Value}

	case parser.StringExpression:
		return RuntimeValue{Kind: StringVal, Value: e.Value}

	case parser.SymbolExpression:
		if e.Value == "nil" {
			return RuntimeValue{Kind: NullVal}
		}
		return env.MustGet(e.Value)

	case parser.BoolExpression:
		return RuntimeValue{Kind: BoolVal, Value: e.Value}

	case parser.IndexExpression:
		left := EvaluateExpression(e.Left, env, ctx)
		index := EvaluateExpression(e.Index, env, ctx)
		if index.Kind != NumberVal {
			panic(tunaError("index must be a number"))
		}
		i := int(index.Value.(float64))
		if left.Kind == ArrayVal {
			arr := left.Value.([]RuntimeValue)
			if i < 0 {
				panic(tunaError(fmt.Sprintf("negative indices are not supported (got %d)", i)))
			}
			if i >= len(arr) {
				panic(tunaError(fmt.Sprintf("index %d out of bounds (length %d)", i, len(arr))))
			}
			return arr[i]
		}
		if left.Kind == StringVal {
			runes := []rune(left.Value.(string))
			if i < 0 {
				panic(tunaError(fmt.Sprintf("negative indices are not supported (got %d)", i)))
			}
			if i >= len(runes) {
				panic(tunaError(fmt.Sprintf("index %d out of bounds (length %d)", i, len(runes))))
			}
			return RuntimeValue{Kind: StringVal, Value: string(runes[i])}
		}
		panic(tunaError(fmt.Sprintf("cannot index into type '%s'", left.Kind)))

	case parser.ArrayLiteral:
		elements := make([]RuntimeValue, len(e.Elements))
		for i, el := range e.Elements {
			elements[i] = EvaluateExpression(el, env, ctx)
		}
		return RuntimeValue{Kind: ArrayVal, Value: elements}

	case parser.ObjectLiteral:
		props := map[string]RuntimeValue{}
		for _, prop := range e.Properties {
			props[prop.Key] = EvaluateExpression(prop.Value, env, ctx)
		}
		return RuntimeValue{Kind: ObjectVal, Value: props}

	case parser.MemberExpression:
		obj := EvaluateExpression(e.Object, env, ctx)
		if obj.Kind != ObjectVal {
			panic(tunaError(fmt.Sprintf(
				"cannot access property '%s' on type '%s'", e.Property, obj.Kind)))
		}
		props := obj.Value.(map[string]RuntimeValue)
		val, ok := props[e.Property]
		if !ok {
			panic(tunaError(fmt.Sprintf("property '%s' does not exist on object", e.Property)))
		}
		return val

	case parser.PrefixExpression:
		return EvaluatePrefixExpression(e, env, ctx)

	case parser.BinaryExpression:
		return EvaluateBinaryExpression(e, env, ctx)

	case parser.AssignmentExpression:
		return EvaluateAssignmentExpression(e, env, ctx)

	case parser.CallExpression:
		return EvaluateCallExpression(e, env, ctx)

	case parser.PostfixExpression:
		return EvaluatePostfixExpression(e, env, ctx)

	case parser.TypeofExpression:
		val := EvaluateExpression(e.Expr, env, ctx)
		return RuntimeValue{Kind: StringVal, Value: val.Kind.String()}

	default:
		panic(tunaError(fmt.Sprintf("unknown expression type: %T", expr)))
	}
}

func EvaluatePrefixExpression(e parser.PrefixExpression, env *Environment, ctx execContext) RuntimeValue {
	right := EvaluateExpression(e.RightExpression, env, ctx)
	switch e.Operator.Kind {
	case lexer.DASH:
		if right.Kind != NumberVal {
			panic(tunaError(fmt.Sprintf("unary '-' cannot be applied to type '%s'", right.Kind)))
		}
		return RuntimeValue{Kind: NumberVal, Value: -right.Value.(float64)}
	case lexer.NOT:
		return RuntimeValue{Kind: BoolVal, Value: !isTruthy(right)}
	default:
		panic(tunaError(fmt.Sprintf("unknown prefix operator '%s'", e.Operator.Value)))
	}
}

func EvaluateBinaryExpression(e parser.BinaryExpression, env *Environment, ctx execContext) RuntimeValue {
	if e.Operator.Kind == lexer.AND {
		left := EvaluateExpression(e.Left, env, ctx)
		if !isTruthy(left) {
			return RuntimeValue{Kind: BoolVal, Value: false}
		}
		right := EvaluateExpression(e.Right, env, ctx)
		return RuntimeValue{Kind: BoolVal, Value: isTruthy(right)}
	}
	if e.Operator.Kind == lexer.OR {
		left := EvaluateExpression(e.Left, env, ctx)
		if isTruthy(left) {
			return RuntimeValue{Kind: BoolVal, Value: true}
		}
		right := EvaluateExpression(e.Right, env, ctx)
		return RuntimeValue{Kind: BoolVal, Value: isTruthy(right)}
	}

	left := EvaluateExpression(e.Left, env, ctx)
	right := EvaluateExpression(e.Right, env, ctx)

	if left.Kind == NumberVal && right.Kind == NumberVal {
		l := left.Value.(float64)
		r := right.Value.(float64)
		switch e.Operator.Kind {
		case lexer.PLUS:
			return RuntimeValue{Kind: NumberVal, Value: l + r}
		case lexer.DASH:
			return RuntimeValue{Kind: NumberVal, Value: l - r}
		case lexer.STAR:
			return RuntimeValue{Kind: NumberVal, Value: l * r}
		case lexer.SLASH:
			if r == 0 {
				panic(tunaError("division by zero"))
			}
			return RuntimeValue{Kind: NumberVal, Value: l / r}
		case lexer.PERCENT:
			if r == 0 {
				panic(tunaError("modulo by zero"))
			}
			return RuntimeValue{Kind: NumberVal, Value: math.Mod(l, r)}
		case lexer.LESS:
			return RuntimeValue{Kind: BoolVal, Value: l < r}
		case lexer.LESS_EQUALS:
			return RuntimeValue{Kind: BoolVal, Value: l <= r}
		case lexer.GREATER:
			return RuntimeValue{Kind: BoolVal, Value: l > r}
		case lexer.GREATER_EQUALS:
			return RuntimeValue{Kind: BoolVal, Value: l >= r}
		case lexer.EQUALS:
			return RuntimeValue{Kind: BoolVal, Value: l == r}
		case lexer.NOT_EQUALS:
			return RuntimeValue{Kind: BoolVal, Value: l != r}
		}
	}

	if left.Kind == StringVal && right.Kind == StringVal {
		l := left.Value.(string)
		r := right.Value.(string)
		switch e.Operator.Kind {
		case lexer.PLUS:
			return RuntimeValue{Kind: StringVal, Value: l + r}
		case lexer.EQUALS:
			return RuntimeValue{Kind: BoolVal, Value: l == r}
		case lexer.NOT_EQUALS:
			return RuntimeValue{Kind: BoolVal, Value: l != r}
		}
	}

	if left.Kind == StringVal && right.Kind == NumberVal {
		if e.Operator.Kind == lexer.PLUS {
			return RuntimeValue{Kind: StringVal, Value: left.Value.(string) + nativeToString(right)}
		}
	}
	if left.Kind == NumberVal && right.Kind == StringVal {
		if e.Operator.Kind == lexer.PLUS {
			return RuntimeValue{Kind: StringVal, Value: nativeToString(left) + right.Value.(string)}
		}
	}

	if left.Kind == BoolVal && right.Kind == BoolVal {
		l := left.Value.(bool)
		r := right.Value.(bool)
		switch e.Operator.Kind {
		case lexer.EQUALS:
			return RuntimeValue{Kind: BoolVal, Value: l == r}
		case lexer.NOT_EQUALS:
			return RuntimeValue{Kind: BoolVal, Value: l != r}
		}
	}

	panic(tunaError(fmt.Sprintf("operator '%s' cannot be applied to types '%s' and '%s'",
		e.Operator.Value, left.Kind, right.Kind)))
}

func applyCompoundOp(op lexer.Token, current, rhs RuntimeValue) RuntimeValue {
	switch op.Kind {
	case lexer.PLUS_EQUALS:
		if current.Kind == NumberVal && rhs.Kind == NumberVal {
			return RuntimeValue{Kind: NumberVal, Value: current.Value.(float64) + rhs.Value.(float64)}
		}
		if current.Kind == StringVal && rhs.Kind == StringVal {
			return RuntimeValue{Kind: StringVal, Value: current.Value.(string) + rhs.Value.(string)}
		}
		panic(tunaError(fmt.Sprintf("'+=' cannot be applied to types '%s' and '%s'", current.Kind, rhs.Kind)))
	case lexer.MINUS_EQUALS:
		if current.Kind != NumberVal || rhs.Kind != NumberVal {
			panic(tunaError(fmt.Sprintf("'-=' requires number operands, got '%s' and '%s'", current.Kind, rhs.Kind)))
		}
		return RuntimeValue{Kind: NumberVal, Value: current.Value.(float64) - rhs.Value.(float64)}
	case lexer.STAR_EQUALS:
		if current.Kind != NumberVal || rhs.Kind != NumberVal {
			panic(tunaError(fmt.Sprintf("'*=' requires number operands, got '%s' and '%s'", current.Kind, rhs.Kind)))
		}
		return RuntimeValue{Kind: NumberVal, Value: current.Value.(float64) * rhs.Value.(float64)}
	case lexer.SLASH_EQUALS:
		if current.Kind != NumberVal || rhs.Kind != NumberVal {
			panic(tunaError(fmt.Sprintf("'/=' requires number operands, got '%s' and '%s'", current.Kind, rhs.Kind)))
		}
		if rhs.Value.(float64) == 0 {
			panic(tunaError("division by zero"))
		}
		return RuntimeValue{Kind: NumberVal, Value: current.Value.(float64) / rhs.Value.(float64)}
	default:
		panic(tunaError(fmt.Sprintf("unknown compound operator '%s'", op.Value)))
	}
}

func EvaluateAssignmentExpression(e parser.AssignmentExpression, env *Environment, ctx execContext) RuntimeValue {
	if mem, ok := e.Assigne.(parser.MemberExpression); ok {
		obj := EvaluateExpression(mem.Object, env, ctx)
		if obj.Kind != ObjectVal {
			panic(tunaError(fmt.Sprintf(
				"cannot assign property '%s' on type '%s'", mem.Property, obj.Kind)))
		}
		props := obj.Value.(map[string]RuntimeValue)
		rhs := EvaluateExpression(e.Value, env, ctx)
		var newVal RuntimeValue
		if e.Operator.Kind == lexer.ASSIGNMENT {
			newVal = rhs
		} else {
			current, exists := props[mem.Property]
			if !exists {
				panic(tunaError(fmt.Sprintf("property '%s' does not exist on object", mem.Property)))
			}
			newVal = applyCompoundOp(e.Operator, current, rhs)
		}
		props[mem.Property] = newVal
		return newVal
	}

	if idx, ok := e.Assigne.(parser.IndexExpression); ok {
		arrVal := EvaluateExpression(idx.Left, env, ctx)
		if arrVal.Kind != ArrayVal {
			panic(tunaError(fmt.Sprintf("cannot index-assign into type '%s'", arrVal.Kind)))
		}
		indexVal := EvaluateExpression(idx.Index, env, ctx)
		if indexVal.Kind != NumberVal {
			panic(tunaError("index must be a number"))
		}
		i := int(indexVal.Value.(float64))
		arr := arrVal.Value.([]RuntimeValue)
		if i < 0 {
			panic(tunaError(fmt.Sprintf("negative indices are not supported (got %d)", i)))
		}
		if i >= len(arr) {
			panic(tunaError(fmt.Sprintf("index %d out of bounds (length %d)", i, len(arr))))
		}

		rhs := EvaluateExpression(e.Value, env, ctx)
		var newVal RuntimeValue
		if e.Operator.Kind == lexer.ASSIGNMENT {
			newVal = rhs
		} else {
			newVal = applyCompoundOp(e.Operator, arr[i], rhs)
		}

		if sym, ok2 := idx.Left.(parser.SymbolExpression); ok2 {
			if declType := env.GetDeclaredType(sym.Value); declType != nil {
				if arrType, ok3 := declType.(parser.ArrayType); ok3 {
					checkType(sym.Value, newVal, arrType.Underlying)
				}
			}
		}
		arr[i] = newVal
		if sym, ok2 := idx.Left.(parser.SymbolExpression); ok2 {
			env.MustUpdate(sym.Value, RuntimeValue{Kind: ArrayVal, Value: arr})
		}
		return newVal
	}

	symbol, ok := e.Assigne.(parser.SymbolExpression)
	if !ok {
		panic(tunaError("invalid assignment target"))
	}

	current := env.MustGet(symbol.Value)
	var newVal RuntimeValue
	if e.Operator.Kind == lexer.ASSIGNMENT {
		newVal = EvaluateExpression(e.Value, env, ctx)
	} else {
		rhs := EvaluateExpression(e.Value, env, ctx)
		newVal = applyCompoundOp(e.Operator, current, rhs)
	}

	declaredType := env.GetDeclaredType(symbol.Value)
	if declaredType != nil {
		checkType(symbol.Value, newVal, declaredType)
	}

	if ctx.builtinNames[symbol.Value] {
		fmt.Printf("\033[33m[TunaScript Warning]\033[0m overwriting builtin '%s'\n", symbol.Value)
	}

	env.MustUpdate(symbol.Value, newVal)
	return newVal
}

func EvaluateCallExpression(e parser.CallExpression, env *Environment, ctx execContext) RuntimeValue {
	callee := EvaluateExpression(e.Callee, env, ctx)
	if callee.Kind != FunctionVal {
		panic(tunaError(fmt.Sprintf("cannot call a value of type '%s'", callee.Kind)))
	}

	switch f := callee.Value.(type) {
	case NativeFunction:
		args := make([]RuntimeValue, len(e.Arguments))
		for i, arg := range e.Arguments {
			args[i] = EvaluateExpression(arg, env, ctx)
		}
		return f.Call(args)

	case FunctionValue:
		if len(e.Arguments) != len(f.Parameters) {
			panic(tunaError(fmt.Sprintf(
				"function '%s' expects %d argument(s) but got %d",
				f.Name, len(f.Parameters), len(e.Arguments))))
		}

		declEnv := f.Env
		if declEnv == nil {
			declEnv = env
		}
		fnEnv := NewEnvironment(declEnv)
		for i, param := range f.Parameters {
			argVal := EvaluateExpression(e.Arguments[i], env, ctx)
			if param.Type != nil {
				checkType(param.Name, argVal, param.Type)
			}
			fnEnv.SetTyped(param.Name, argVal, param.Type)
		}

		fnCtx := ctx.withFunction()
		result := EvaluateBlock(f.Body, fnEnv, fnCtx)

		if result.Signal == sigReturn {
			ret := result.Value
			checkType(f.Name+"() return", ret, f.ReturnType)
			return ret
		}

		if f.ReturnType != nil {
			want := resolveTypeName(f.ReturnType)
			if want != "" && want != "null" && want != "void" {
				panic(tunaError(fmt.Sprintf(
					"function '%s' must return a '%s' but reached end without a 'serve'",
					f.Name, want)))
			}
		}
		return RuntimeValue{Kind: NullVal}

	default:
		panic(tunaError("cannot call non-function value"))
	}
}

func EvaluatePostfixExpression(e parser.PostfixExpression, env *Environment, ctx execContext) RuntimeValue {
	symbol, ok := e.Left.(parser.SymbolExpression)
	if !ok {
		panic(tunaError("'++' and '--' can only be applied to a variable"))
	}
	current := env.MustGet(symbol.Value)
	if current.Kind != NumberVal {
		panic(tunaError(fmt.Sprintf("'%s' cannot be applied to type '%s'", e.Operator.Value, current.Kind)))
	}
	val := current.Value.(float64)
	switch e.Operator.Kind {
	case lexer.PLUS_PLUS:
		env.MustUpdate(symbol.Value, RuntimeValue{Kind: NumberVal, Value: val + 1})
	case lexer.MINUS_MINUS:
		env.MustUpdate(symbol.Value, RuntimeValue{Kind: NumberVal, Value: val - 1})
	}

	return current
}

func isTruthy(val RuntimeValue) bool {
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

func tunaError(msg string) string {
	return fmt.Sprintf("\033[31m[TunaScript Error]\033[0m %s", msg)
}
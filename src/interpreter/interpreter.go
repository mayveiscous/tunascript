package interpreter

import (
	"fmt"
	"math"
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

type ReturnSignal struct {
	Value RuntimeValue
}

type FunctionValue struct {
	Name       string
	Parameters []parser.FunctionParameter
	ReturnType parser.AstType
	Body       parser.BlockStatement
	Env        *Environment
}

type BreakSignal struct{}
type ContinueSignal struct{}

type NativeFunction struct {
	Name string
	Call func(args []RuntimeValue) RuntimeValue
}

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

func (e *Environment) Update(name string, val RuntimeValue) RuntimeValue {
	if _, ok := e.variables[name]; ok {
		if e.constants[name] {
			panic(fmt.Sprintf("cannot reassign constant '%s'", name))
		}
		e.variables[name] = val
		return val
	}
	if e.parent != nil {
		return e.parent.Update(name, val)
	}
	panic(fmt.Sprintf("cannot assign to undefined variable '%s'", name))
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

func (e *Environment) Get(name string) RuntimeValue {
	if val, ok := e.variables[name]; ok {
		return val
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	panic(fmt.Sprintf("undefined variable '%s'", name))
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
			return
		}
		if val.Kind != ArrayVal {
			want := resolveTypeName(expected)
			panic(fmt.Sprintf("type mismatch for '%s': expected '%s' but got '%s'", label, want, val.Kind.String()))
		}
		elemTypeName := resolveTypeName(exp.Underlying)
		if elemTypeName != "" {
			for _, elem := range val.Value.([]RuntimeValue) {
				got := elem.Kind.String()
				if got != elemTypeName {
					panic(fmt.Sprintf("type mismatch for '%s': expected '[]%s' contents but got '%s'", label, elemTypeName, got))
				}
			}
		}
	case parser.SymbolType:
		want := exp.Name
		if want == "" {
			 return
		}
		validTypes := map[string]bool{
			 "number": true, "string": true, "bool": true,
			 "function": true, "void": true, "null": true, "array": true, "object": true,
		}
		if !validTypes[want] {
			 panic(fmt.Sprintf("unknown type '%s': valid types are number, string, bool, function, void, null, array", want))
		}
  
		if val.Kind == NullVal {
			 return
		}
		got := val.Kind.String()
		if got != want {
			 panic(fmt.Sprintf("type mismatch for '%s': expected '%s' but got '%s'", label, want, got))
		}
	}
}

type execContext struct {
	inLoop     bool
	inFunction bool
}

func Interpret(block parser.BlockStatement) RuntimeValue {
	env := NewEnvironment(nil)
	registerBuiltins(env)
	builtinNames = map[string]bool{}
	for k := range env.variables {
		builtinNames[k] = true
	}
	var result RuntimeValue
	ctx := execContext{}
	for _, stmt := range block.Body {
		result = EvaluateStatement(stmt, env, ctx)
	}
	return result
}

var builtinNames map[string]bool

func EvaluateStatement(stmt parser.Statement, env *Environment, ctx execContext) RuntimeValue {
	switch s := stmt.(type) {

	case parser.ExpressionStatement:
		return EvaluateExpression(s.Expression, env, ctx)
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
		return val
	case parser.BreakStatement:
		if !ctx.inLoop {
			if ctx.inFunction {
				panic(tunaError("`break` cannot cross a function boundary"))
			}
			panic(tunaError("`break` can only be used inside a loop"))
		}
		panic(BreakSignal{})
	case parser.ContinueStatement:
		if !ctx.inLoop {
			if ctx.inFunction {
				panic(tunaError("`continue` cannot cross a function boundary"))
			}
			panic(tunaError("`continue` can only be used inside a loop"))
		}
		panic(ContinueSignal{})
	case parser.ReturnStatement:
		if !ctx.inFunction {
			panic(tunaError("`serve` can only be used inside a function"))
		}
		val := EvaluateExpression(s.Value, env, ctx)
		panic(ReturnSignal{Value: val})
	case parser.IfStatement:
		condition := EvaluateExpression(s.Condition, env, ctx)
		if isTruthy(condition) {
			return EvaluateBlock(s.Then, NewEnvironment(env), ctx)
		} else if s.Else != nil {
			return EvaluateBlock(*s.Else, NewEnvironment(env), ctx)
		}
		return RuntimeValue{Kind: NullVal}
	case parser.WhileStatement:
		result := RuntimeValue{Kind: NullVal}
		shouldBreak := false
		loopCtx := execContext{inLoop: true, inFunction: ctx.inFunction}
		for !shouldBreak {
			if !isTruthy(EvaluateExpression(s.Condition, env, ctx)) {
				break
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						switch r.(type) {
						case BreakSignal:
							shouldBreak = true
						case ContinueSignal:
						default:
							panic(r)
						}
					}
				}()
				result = EvaluateBlock(s.Body, NewEnvironment(env), loopCtx)
			}()
		}
		return result
	case parser.ForInStatement:
		iterable := EvaluateExpression(s.Iterable, env, ctx)
		if iterable.Kind != ArrayVal {
			panic(fmt.Sprintf("cannot iterate over non-array value of type '%s' in for/in", iterable.Kind))
		}
		result := RuntimeValue{Kind: NullVal}
		shouldBreak := false
		loopCtx := execContext{inLoop: true, inFunction: ctx.inFunction}
		for _, element := range iterable.Value.([]RuntimeValue) {
			if shouldBreak {
				break
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						switch r.(type) {
						case BreakSignal:
							shouldBreak = true
						case ContinueSignal:
						default:
							panic(r)
						}
					}
				}()
				loopEnv := NewEnvironment(env)
				loopEnv.Set(s.Iterator, element)
				result = EvaluateBlock(s.Body, loopEnv, loopCtx)
			}()
		}
		return result
	case parser.FunctionDecStatement:
		fn := RuntimeValue{
			Kind: FunctionVal,
			Value: FunctionValue{
				Name:       s.Name,
				Parameters: s.Parameters,
				ReturnType: s.ReturnType,
				Body:       s.Body,
				Env:        env,
			},
		}
		if builtinNames[s.Name] {
			fmt.Printf("\033[33m[TunaScript Warning]\033[0m overwriting builtin '%s'\n", s.Name)
		}
		env.Set(s.Name, fn)
		return fn
	case parser.BlockStatement:
		return EvaluateBlock(s, NewEnvironment(env), ctx)
	default:
		panic(fmt.Sprintf("unknown statement type: %T", stmt))
	}
}

func EvaluateBlock(block parser.BlockStatement, env *Environment, ctx execContext) RuntimeValue {
	result := RuntimeValue{Kind: NullVal}
	for _, stmt := range block.Body {
		result = EvaluateStatement(stmt, env, ctx)
	}
	return result
}

func tunaError(msg string) string {
	return fmt.Sprintf("\033[31m[TunaScript Error]\033[0m %s", msg)
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
		return env.Get(e.Value)

	case parser.IndexExpression:
		left := EvaluateExpression(e.Left, env, ctx)
		index := EvaluateExpression(e.Index, env, ctx)
		if index.Kind != NumberVal {
			panic("index must be a number")
		}
		i := int(index.Value.(float64))
		if left.Kind == ArrayVal {
			arr := left.Value.([]RuntimeValue)
			if i < 0 || i >= len(arr) {
				panic(fmt.Sprintf("index %d out of bounds (length %d)", i, len(arr)))
			}
			return arr[i]
		}
		if left.Kind == StringVal {
			runes := []rune(left.Value.(string))
			if i < 0 || i >= len(runes) {
				panic(fmt.Sprintf("index %d out of bounds (length %d)", i, len(runes)))
			}
			return RuntimeValue{Kind: StringVal, Value: string(runes[i])}
		}
		panic(fmt.Sprintf("cannot index into type '%s'", left.Kind))

	case parser.ArrayLiteral:
		elements := []RuntimeValue{}
		for _, el := range e.Elements {
			elements = append(elements, EvaluateExpression(el, env, ctx))
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
			panic(fmt.Sprintf("cannot access property '%s' on type '%s'", e.Property, obj.Kind))
		}
		props := obj.Value.(map[string]RuntimeValue)
		val, ok := props[e.Property]
		if !ok {
			panic(fmt.Sprintf("property '%s' does not exist on object", e.Property))
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
	case parser.BoolExpression:
		return RuntimeValue{Kind: BoolVal, Value: e.Value}

	case parser.PostfixExpression:
		return EvaluatePostfixExpression(e, env, ctx)

	case parser.TypeofExpression:
		val := EvaluateExpression(e.Expr, env, ctx)
		return RuntimeValue{Kind: StringVal, Value: val.Kind.String()}

	default:
		panic(fmt.Sprintf("unknown expression type: %T", expr))
	}
}
func EvaluatePrefixExpression(e parser.PrefixExpression, env *Environment, ctx execContext) RuntimeValue {
	right := EvaluateExpression(e.RightExpression, env, ctx)
	switch e.Operator.Kind {
	case lexer.DASH:
		if right.Kind != NumberVal {
			panic(fmt.Sprintf("unary '-' cannot be applied to type '%s'", right.Kind))
		}
		return RuntimeValue{Kind: NumberVal, Value: -right.Value.(float64)}
	case lexer.NOT:
		return RuntimeValue{Kind: BoolVal, Value: !isTruthy(right)}
	default:
		panic(fmt.Sprintf("unknown prefix operator '%s'", e.Operator.Value))
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
				panic("division by zero")
			}
			return RuntimeValue{Kind: NumberVal, Value: l / r}
		case lexer.PERCENT:
			if r == 0 {
				panic("division by zero")
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

	panic(fmt.Sprintf("operator '%s' cannot be applied to types '%s' and '%s'",
		e.Operator.Value, left.Kind, right.Kind))
}

func EvaluateAssignmentExpression(e parser.AssignmentExpression, env *Environment, ctx execContext) RuntimeValue {
	// obj.key = val
	if mem, ok := e.Assigne.(parser.MemberExpression); ok {
		obj := EvaluateExpression(mem.Object, env, ctx)
		if obj.Kind != ObjectVal {
			panic(fmt.Sprintf("cannot assign property '%s' on type '%s'", mem.Property, obj.Kind))
		}
		rhs := EvaluateExpression(e.Value, env, ctx)
		obj.Value.(map[string]RuntimeValue)[mem.Property] = rhs
		return rhs
	}

	// arr[i] = val
	if idx, ok := e.Assigne.(parser.IndexExpression); ok {
		arrVal := EvaluateExpression(idx.Left, env, ctx)
		if arrVal.Kind != ArrayVal {
			panic(fmt.Sprintf("cannot index-assign into type '%s'", arrVal.Kind))
		}
		indexVal := EvaluateExpression(idx.Index, env, ctx)
		if indexVal.Kind != NumberVal {
			panic("index must be a number")
		}
		i := int(indexVal.Value.(float64))
		arr := arrVal.Value.([]RuntimeValue)
		if i < 0 || i >= len(arr) {
			panic(fmt.Sprintf("index %d out of bounds (length %d)", i, len(arr)))
		}
		rhs := EvaluateExpression(e.Value, env, ctx)
		arr[i] = rhs
		return rhs
	}

	symbol, ok := e.Assigne.(parser.SymbolExpression)
	if !ok {
		panic(tunaError("invalid assignment target"))
	}

	current := env.Get(symbol.Value)
	var newVal RuntimeValue

	switch e.Operator.Kind {
	case lexer.ASSIGNMENT:
		newVal = EvaluateExpression(e.Value, env, ctx)
	case lexer.PLUS_EQUALS:
		rhs := EvaluateExpression(e.Value, env, ctx)
		if current.Kind == NumberVal && rhs.Kind == NumberVal {
			newVal = RuntimeValue{Kind: NumberVal, Value: current.Value.(float64) + rhs.Value.(float64)}
		} else if current.Kind == StringVal && rhs.Kind == StringVal {
			newVal = RuntimeValue{Kind: StringVal, Value: current.Value.(string) + rhs.Value.(string)}
		} else {
			panic(fmt.Sprintf("'+=' cannot be applied to types '%s' and '%s'", current.Kind, rhs.Kind))
		}
	case lexer.MINUS_EQUALS:
		rhs := EvaluateExpression(e.Value, env, ctx)
		if current.Kind != NumberVal || rhs.Kind != NumberVal {
			panic(fmt.Sprintf("'-=' requires number operands, got '%s' and '%s'", current.Kind, rhs.Kind))
		}
		newVal = RuntimeValue{Kind: NumberVal, Value: current.Value.(float64) - rhs.Value.(float64)}
	case lexer.STAR_EQUALS:
		rhs := EvaluateExpression(e.Value, env, ctx)
		if current.Kind != NumberVal || rhs.Kind != NumberVal {
			panic(fmt.Sprintf("'*=' requires number operands, got '%s' and '%s'", current.Kind, rhs.Kind))
		}
		newVal = RuntimeValue{Kind: NumberVal, Value: current.Value.(float64) * rhs.Value.(float64)}
	case lexer.SLASH_EQUALS:
		rhs := EvaluateExpression(e.Value, env, ctx)
		if current.Kind != NumberVal || rhs.Kind != NumberVal {
			panic(fmt.Sprintf("'/=' requires number operands, got '%s' and '%s'", current.Kind, rhs.Kind))
		}
		if rhs.Value.(float64) == 0 {
			panic("division by zero")
		}
		newVal = RuntimeValue{Kind: NumberVal, Value: current.Value.(float64) / rhs.Value.(float64)}
	default:
		panic(fmt.Sprintf("unknown assignment operator '%s'", e.Operator.Value))
	}

	declaredType := env.GetDeclaredType(symbol.Value)
	if declaredType != nil {
		checkType(symbol.Value, newVal, declaredType)
	}

	if builtinNames[symbol.Value] {
		fmt.Printf("\033[33m[TunaScript Warning]\033[0m overwriting builtin '%s'\n", symbol.Value)
	}

	env.Update(symbol.Value, newVal)
	return newVal
}

func EvaluateCallExpression(e parser.CallExpression, env *Environment, ctx execContext) RuntimeValue {
	callee := EvaluateExpression(e.Callee, env, ctx)
	if callee.Kind != FunctionVal {
		panic(fmt.Sprintf("cannot call a value of type '%s'", callee.Kind))
	}

	switch f := callee.Value.(type) {
	case NativeFunction:
		args := []RuntimeValue{}
		for _, arg := range e.Arguments {
			args = append(args, EvaluateExpression(arg, env, ctx))
		}
		return f.Call(args)

	case FunctionValue:
		if len(e.Arguments) != len(f.Parameters) {
			panic(fmt.Sprintf("function '%s' expects %d argument(s) but got %d",
				f.Name, len(f.Parameters), len(e.Arguments)))
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
			fnEnv.Set(param.Name, argVal)
		}

		result := RuntimeValue{Kind: NullVal}
		fnCtx := execContext{inLoop: false, inFunction: true}
		func() {
			defer func() {
				if r := recover(); r != nil {
					if ret, ok := r.(ReturnSignal); ok {
						checkType(f.Name+"() return", ret.Value, f.ReturnType)
						result = ret.Value
					} else {
						panic(r)
					}
				}
			}()
			EvaluateBlock(f.Body, fnEnv, fnCtx)
		}()
		if f.ReturnType != nil && result.Kind == NullVal {
			want := resolveTypeName(f.ReturnType)
			if want != "" && want != "null" {
				panic(fmt.Sprintf("type mismatch for '%s() return': expected '%s' but got 'null'", f.Name, want))
			}
		}
		return result

	default:
		panic("cannot call non-function value")
	}
}

func EvaluatePostfixExpression(e parser.PostfixExpression, env *Environment, ctx execContext) RuntimeValue {
	symbol, ok := e.Left.(parser.SymbolExpression)
	if !ok {
		panic("'++' and '--' can only be applied to a variable")
	}
	current := env.Get(symbol.Value)
	if current.Kind != NumberVal {
		panic(fmt.Sprintf("'%s' cannot be applied to type '%s'", e.Operator.Value, current.Kind))
	}
	val := current.Value.(float64)
	switch e.Operator.Kind {
	case lexer.PLUS_PLUS:
		env.Update(symbol.Value, RuntimeValue{Kind: NumberVal, Value: val + 1})
	case lexer.MINUS_MINUS:
		env.Update(symbol.Value, RuntimeValue{Kind: NumberVal, Value: val - 1})
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
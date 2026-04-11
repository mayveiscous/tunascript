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
	Env        *Environment // declaration-scope environment (for closures)
}

type BreakSignal struct{}
type ContinueSignal struct{}

type NativeFunction struct {
	Name string
	Call func(args []RuntimeValue) RuntimeValue
}

type Environment struct {
	variables map[string]RuntimeValue
	constants map[string]bool
	parent    *Environment
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		variables: map[string]RuntimeValue{},
		constants: map[string]bool{},
		parent:    parent,
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

// resolveTypeName maps an AstType to a ValueKind string for type checking.
func resolveTypeName(t parser.AstType) string {
	switch v := t.(type) {
	case parser.SymbolType:
		return v.Name
	case parser.ArrayType:
		return "array"
	default:
		return ""
	}
}

func checkType(label string, val RuntimeValue, expected parser.AstType) {
	if expected == nil {
		return
	}
	want := resolveTypeName(expected)
	if want == "" {
		return
	}
	got := val.Kind.String()
	if got != want {
		panic(fmt.Sprintf("type mismatch for '%s': expected '%s' but got '%s'", label, want, got))
	}
}

func Interpret(block parser.BlockStatement) RuntimeValue {
	env := NewEnvironment(nil)
	registerBuiltins(env)
	var result RuntimeValue
	for _, stmt := range block.Body {
		result = EvaluateStatement(stmt, env)
	}
	return result
}

func EvaluateStatement(stmt parser.Statement, env *Environment) RuntimeValue {
	switch s := stmt.(type) {

	case parser.ExpressionStatement:
		return EvaluateExpression(s.Expression, env)
	case parser.VariableDecStatement:
		val := RuntimeValue{Kind: NullVal}
		if s.AssignedValue != nil {
			val = EvaluateExpression(s.AssignedValue, env)
		}
		checkType(s.VariableName, val, s.ExplicitType)
		if s.IsConstant {
			env.SetConst(s.VariableName, val)
		} else {
			env.Set(s.VariableName, val)
		}
		return val
	case parser.BreakStatement:
		panic(BreakSignal{})
  case parser.ContinueStatement:
		panic(ContinueSignal{})
	case parser.ReturnStatement:
		val := EvaluateExpression(s.Value, env)
		panic(ReturnSignal{Value: val})
	case parser.IfStatement:
		condition := EvaluateExpression(s.Condition, env)
		if isTruthy(condition) {
			return EvaluateBlock(s.Then, NewEnvironment(env))
		} else if s.Else != nil {
			return EvaluateBlock(*s.Else, NewEnvironment(env))
		}
		return RuntimeValue{Kind: NullVal}
	case parser.WhileStatement:
		result := RuntimeValue{Kind: NullVal}
		shouldBreak := false
		for !shouldBreak {
			if !isTruthy(EvaluateExpression(s.Condition, env)) {
				break
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						switch r.(type) {
						case BreakSignal:
							shouldBreak = true
						case ContinueSignal:
							// just stop this iteration, condition re-evaluates next
						default:
							panic(r)
						}
					}
				}()
				result = EvaluateBlock(s.Body, NewEnvironment(env))
			}()
		}
		return result
	case parser.ForInStatement:
		iterable := EvaluateExpression(s.Iterable, env)
		if iterable.Kind != ArrayVal {
			panic(fmt.Sprintf("cannot iterate over non-array value of type '%s' in for/in", iterable.Kind))
		}
		result := RuntimeValue{Kind: NullVal}
		shouldBreak := false
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
							// just stop this iteration
						default:
							panic(r)
						}
					}
				}()
				loopEnv := NewEnvironment(env)
				loopEnv.Set(s.Iterator, element)
				result = EvaluateBlock(s.Body, loopEnv)
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
				Env:        env, // capture declaration scope
			},
		}
		env.Set(s.Name, fn)
		return fn
	case parser.BlockStatement:
		return EvaluateBlock(s, NewEnvironment(env))
	default:
		panic(fmt.Sprintf("unknown statement type: %T", stmt))
	}
}

func EvaluateExpression(expr parser.Expression, env *Environment) RuntimeValue {
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
		left := EvaluateExpression(e.Left, env)
		index := EvaluateExpression(e.Index, env)
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
			 s := left.Value.(string)
			 if i < 0 || i >= len(s) {
				  panic(fmt.Sprintf("index %d out of bounds (length %d)", i, len(s)))
			 }
			 return RuntimeValue{Kind: StringVal, Value: string(s[i])}
		}
		panic(fmt.Sprintf("cannot index into type '%s'", left.Kind))
	case parser.ArrayLiteral:
		elements := []RuntimeValue{}
		for _, el := range e.Elements {
			elements = append(elements, EvaluateExpression(el, env))
		}
		return RuntimeValue{Kind: ArrayVal, Value: elements}

	case parser.PrefixExpression:
		return EvaluatePrefixExpression(e, env)

	case parser.BinaryExpression:
		return EvaluateBinaryExpression(e, env)

	case parser.AssignmentExpression:
		return EvaluateAssignmentExpression(e, env)

	case parser.CallExpression:
		return EvaluateCallExpression(e, env)
	case parser.BoolExpression:
		return RuntimeValue{Kind: BoolVal, Value: e.Value}

	case parser.PostfixExpression:
		return EvaluatePostfixExpression(e, env)

	case parser.TypeofExpression:
		val := EvaluateExpression(e.Expr, env)
		return RuntimeValue{Kind: StringVal, Value: val.Kind.String()}

	default:
		panic(fmt.Sprintf("unknown expression type: %T", expr))
	}
}

func EvaluateBlock(block parser.BlockStatement, env *Environment) RuntimeValue {
	result := RuntimeValue{Kind: NullVal}
	for _, stmt := range block.Body {
		result = EvaluateStatement(stmt, env)
	}
	return result
}

func EvaluatePrefixExpression(e parser.PrefixExpression, env *Environment) RuntimeValue {
	right := EvaluateExpression(e.RightExpression, env)
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

func EvaluateBinaryExpression(e parser.BinaryExpression, env *Environment) RuntimeValue {
	// Short-circuit logical operators before evaluating right side
	if e.Operator.Kind == lexer.AND {
		left := EvaluateExpression(e.Left, env)
		if !isTruthy(left) {
			return RuntimeValue{Kind: BoolVal, Value: false}
		}
		right := EvaluateExpression(e.Right, env)
		return RuntimeValue{Kind: BoolVal, Value: isTruthy(right)}
	}
	if e.Operator.Kind == lexer.OR {
		left := EvaluateExpression(e.Left, env)
		if isTruthy(left) {
			return RuntimeValue{Kind: BoolVal, Value: true}
		}
		right := EvaluateExpression(e.Right, env)
		return RuntimeValue{Kind: BoolVal, Value: isTruthy(right)}
	}

	left := EvaluateExpression(e.Left, env)
	right := EvaluateExpression(e.Right, env)

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

func EvaluateAssignmentExpression(e parser.AssignmentExpression, env *Environment) RuntimeValue {
	symbol, ok := e.Assigne.(parser.SymbolExpression)
	if !ok {
		panic("left side of assignment must be a variable name")
	}

	current := env.Get(symbol.Value)
	var newVal RuntimeValue

	switch e.Operator.Kind {
	case lexer.ASSIGNMENT:
		newVal = EvaluateExpression(e.Value, env)
	case lexer.PLUS_EQUALS:
		rhs := EvaluateExpression(e.Value, env)
		if current.Kind == NumberVal && rhs.Kind == NumberVal {
			newVal = RuntimeValue{Kind: NumberVal, Value: current.Value.(float64) + rhs.Value.(float64)}
		} else if current.Kind == StringVal && rhs.Kind == StringVal {
			newVal = RuntimeValue{Kind: StringVal, Value: current.Value.(string) + rhs.Value.(string)}
		} else {
			panic(fmt.Sprintf("'+=' cannot be applied to types '%s' and '%s'", current.Kind, rhs.Kind))
		}
	case lexer.MINUS_EQUALS:
		rhs := EvaluateExpression(e.Value, env)
		if current.Kind != NumberVal || rhs.Kind != NumberVal {
			panic(fmt.Sprintf("'-=' requires number operands, got '%s' and '%s'", current.Kind, rhs.Kind))
		}
		newVal = RuntimeValue{Kind: NumberVal, Value: current.Value.(float64) - rhs.Value.(float64)}
	case lexer.STAR_EQUALS:
		rhs := EvaluateExpression(e.Value, env)
		if current.Kind != NumberVal || rhs.Kind != NumberVal {
			panic(fmt.Sprintf("'*=' requires number operands, got '%s' and '%s'", current.Kind, rhs.Kind))
		}
		newVal = RuntimeValue{Kind: NumberVal, Value: current.Value.(float64) * rhs.Value.(float64)}
	case lexer.SLASH_EQUALS:
		rhs := EvaluateExpression(e.Value, env)
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

	env.Update(symbol.Value, newVal)
	return newVal
}

func EvaluateCallExpression(e parser.CallExpression, env *Environment) RuntimeValue {
	callee := EvaluateExpression(e.Callee, env)
	if callee.Kind != FunctionVal {
		panic(fmt.Sprintf("cannot call a value of type '%s'", callee.Kind))
	}

	switch f := callee.Value.(type) {
	case NativeFunction:
		args := []RuntimeValue{}
		for _, arg := range e.Arguments {
			args = append(args, EvaluateExpression(arg, env))
		}
		return f.Call(args)

	case FunctionValue:
		if len(e.Arguments) != len(f.Parameters) {
			panic(fmt.Sprintf("function '%s' expects %d argument(s) but got %d",
				f.Name, len(f.Parameters), len(e.Arguments)))
		}

		// Use the declaration-scope environment as parent (lexical scoping / closures)
		declEnv := f.Env
		if declEnv == nil {
			declEnv = env
		}
		fnEnv := NewEnvironment(declEnv)
		for i, param := range f.Parameters {
			fnEnv.Set(param.Name, EvaluateExpression(e.Arguments[i], env))
		}

		result := RuntimeValue{Kind: NullVal}
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
			EvaluateBlock(f.Body, fnEnv)
		}()
		return result

	default:
		panic("cannot call non-function value")
	}
}

func EvaluatePostfixExpression(e parser.PostfixExpression, env *Environment) RuntimeValue {
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
	return current // return value before increment (postfix behaviour)
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
package interpreter

import (
	"fmt"
	"math"

	tunaparser "tunascript/src/parser"
	"tunascript/src/lexer"
)

func EvaluateExpression(expr tunaparser.Expression, env *Environment, ctx ExecContext) RuntimeValue {
	switch e := expr.(type) {

	case tunaparser.NumberExpression:
		return RuntimeValue{Kind: NumberVal, Value: e.Value}

	case tunaparser.StringExpression:
		return RuntimeValue{Kind: StringVal, Value: e.Value}

	case tunaparser.SymbolExpression:
		if e.Value == "nil" {
			return RuntimeValue{Kind: NullVal}
		}
		return env.MustGet(e.Value)

	case tunaparser.BoolExpression:
		return RuntimeValue{Kind: BoolVal, Value: e.Value}

	case tunaparser.IndexExpression:
		left  := EvaluateExpression(e.Left, env, ctx)
		index := EvaluateExpression(e.Index, env, ctx)
  
		if left.Kind == ArrayVal {
			 if index.Kind != NumberVal {
				  panic(TunaError("array index must be a number"))
			 }
			 i := int(index.Value.(float64))
			 arr := left.Value.([]RuntimeValue)
			 if i < 0 {
				  panic(TunaError(fmt.Sprintf("negative indices are not supported (got %d)", i)))
			 }
			 if i >= len(arr) {
				  panic(TunaError(fmt.Sprintf("index %d out of bounds (length %d)", i, len(arr))))
			 }
			 return arr[i]
		}
  
		if left.Kind == StringVal {
			 if index.Kind != NumberVal {
				  panic(TunaError("string index must be a number"))
			 }
			 i := int(index.Value.(float64))
			 runes := []rune(left.Value.(string))
			 if i < 0 {
				  panic(TunaError(fmt.Sprintf("negative indices are not supported (got %d)", i)))
			 }
			 if i >= len(runes) {
				  panic(TunaError(fmt.Sprintf("index %d out of bounds (length %d)", i, len(runes))))
			 }
			 return RuntimeValue{Kind: StringVal, Value: string(runes[i])}
		}
  
		if left.Kind == ObjectVal {
			 if index.Kind != StringVal {
				  // A non-string key (including nil) can never match a property,
				  // so treat it as a miss rather than a hard error. This lets
				  // checks like `if obj[someValue] == nil` or `!obj[someValue]`
				  // work even when someValue turns out to be nil.
				  return RuntimeValue{Kind: NullVal}
			 }
			 props := left.Value.(map[string]RuntimeValue)
			 key := index.Value.(string)
			 val, ok := props[key]
			 if !ok {
				  return RuntimeValue{Kind: NullVal}
			 }
			 return val
		}
  
		panic(TunaError(fmt.Sprintf("cannot index into type '%s'", left.Kind)))
	case tunaparser.ArrayLiteral:
		elements := make([]RuntimeValue, len(e.Elements))
		for i, el := range e.Elements {
			elements[i] = EvaluateExpression(el, env, ctx)
		}
		return RuntimeValue{Kind: ArrayVal, Value: elements}

	case tunaparser.ObjectLiteral:
		props := map[string]RuntimeValue{}
		for _, prop := range e.Properties {
			props[prop.Key] = EvaluateExpression(prop.Value, env, ctx)
		}
		return RuntimeValue{Kind: ObjectVal, Value: props}

	case tunaparser.MemberExpression:
		obj := EvaluateExpression(e.Object, env, ctx)
		if obj.Kind != ObjectVal {
			panic(TunaError(fmt.Sprintf(
				"cannot access property '%s' on type '%s'", e.Property, obj.Kind)))
		}
		props := obj.Value.(map[string]RuntimeValue)
		val, ok := props[e.Property]
		if !ok {
			panic(TunaError(fmt.Sprintf("property '%s' does not exist on object", e.Property)))
		}
		return val

	case tunaparser.PrefixExpression:
		return EvaluatePrefixExpression(e, env, ctx)

	case tunaparser.BinaryExpression:
		return EvaluateBinaryExpression(e, env, ctx)

	case tunaparser.AssignmentExpression:
		return EvaluateAssignmentExpression(e, env, ctx)

	case tunaparser.CallExpression:
		return EvaluateCallExpression(e, env, ctx)

	case tunaparser.PostfixExpression:
		return EvaluatePostfixExpression(e, env, ctx)

	case tunaparser.TypeofExpression:
		val := EvaluateExpression(e.Expr, env, ctx)
		return RuntimeValue{Kind: StringVal, Value: val.Kind.String()}

	default:
		panic(TunaError(fmt.Sprintf("unknown expression type: %T", expr)))
	}
}

func EvaluatePrefixExpression(e tunaparser.PrefixExpression, env *Environment, ctx ExecContext) RuntimeValue {
	right := EvaluateExpression(e.RightExpression, env, ctx)
	switch e.Operator.Kind {
	case lexer.DASH:
		if right.Kind != NumberVal {
			panic(TunaError(fmt.Sprintf("unary '-' cannot be applied to type '%s'", right.Kind)))
		}
		return RuntimeValue{Kind: NumberVal, Value: -right.Value.(float64)}
	case lexer.NOT:
		return RuntimeValue{Kind: BoolVal, Value: !IsTruthy(right)}
	default:
		panic(TunaError(fmt.Sprintf("unknown prefix operator '%s'", e.Operator.Value)))
	}
}

func EvaluateBinaryExpression(e tunaparser.BinaryExpression, env *Environment, ctx ExecContext) RuntimeValue {
	if e.Operator.Kind == lexer.AND {
		left := EvaluateExpression(e.Left, env, ctx)
		if !IsTruthy(left) {
			return RuntimeValue{Kind: BoolVal, Value: false}
		}
		right := EvaluateExpression(e.Right, env, ctx)
		return RuntimeValue{Kind: BoolVal, Value: IsTruthy(right)}
	}
	if e.Operator.Kind == lexer.OR {
		left := EvaluateExpression(e.Left, env, ctx)
		if IsTruthy(left) {
			return RuntimeValue{Kind: BoolVal, Value: true}
		}
		right := EvaluateExpression(e.Right, env, ctx)
		return RuntimeValue{Kind: BoolVal, Value: IsTruthy(right)}
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
				panic(TunaError("division by zero"))
			}
			return RuntimeValue{Kind: NumberVal, Value: l / r}
		case lexer.PERCENT:
			if r == 0 {
				panic(TunaError("modulo by zero"))
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

	if e.Operator.Kind == lexer.EQUALS {
		return RuntimeValue{Kind: BoolVal, Value: left.Kind == right.Kind}
  }
  if e.Operator.Kind == lexer.NOT_EQUALS {
		return RuntimeValue{Kind: BoolVal, Value: left.Kind != right.Kind}
  }

	panic(TunaError(fmt.Sprintf("operator '%s' cannot be applied to types '%s' and '%s'",
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
		panic(TunaError(fmt.Sprintf("'+=' cannot be applied to types '%s' and '%s'", current.Kind, rhs.Kind)))
	case lexer.MINUS_EQUALS:
		if current.Kind != NumberVal || rhs.Kind != NumberVal {
			panic(TunaError(fmt.Sprintf("'-=' requires number operands, got '%s' and '%s'", current.Kind, rhs.Kind)))
		}
		return RuntimeValue{Kind: NumberVal, Value: current.Value.(float64) - rhs.Value.(float64)}
	case lexer.STAR_EQUALS:
		if current.Kind != NumberVal || rhs.Kind != NumberVal {
			panic(TunaError(fmt.Sprintf("'*=' requires number operands, got '%s' and '%s'", current.Kind, rhs.Kind)))
		}
		return RuntimeValue{Kind: NumberVal, Value: current.Value.(float64) * rhs.Value.(float64)}
	case lexer.SLASH_EQUALS:
		if current.Kind != NumberVal || rhs.Kind != NumberVal {
			panic(TunaError(fmt.Sprintf("'/=' requires number operands, got '%s' and '%s'", current.Kind, rhs.Kind)))
		}
		if rhs.Value.(float64) == 0 {
			panic(TunaError("division by zero"))
		}
		return RuntimeValue{Kind: NumberVal, Value: current.Value.(float64) / rhs.Value.(float64)}
	default:
		panic(TunaError(fmt.Sprintf("unknown compound operator '%s'", op.Value)))
	}
}

func EvaluateAssignmentExpression(e tunaparser.AssignmentExpression, env *Environment, ctx ExecContext) RuntimeValue {
	if mem, ok := e.Assignee.(tunaparser.MemberExpression); ok {
		obj := EvaluateExpression(mem.Object, env, ctx)
		if obj.Kind != ObjectVal {
			panic(TunaError(fmt.Sprintf(
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
				panic(TunaError(fmt.Sprintf("property '%s' does not exist on object", mem.Property)))
			}
			newVal = applyCompoundOp(e.Operator, current, rhs)
		}
		props[mem.Property] = newVal
		return newVal
	}

	if idx, ok := e.Assignee.(tunaparser.IndexExpression); ok {
		left := EvaluateExpression(idx.Left, env, ctx)
		indexVal := EvaluateExpression(idx.Index, env, ctx)

		if left.Kind == ObjectVal {
			if indexVal.Kind != StringVal {
				panic(TunaError("object key must be a string"))
			}
			props := left.Value.(map[string]RuntimeValue)
			key := indexVal.Value.(string)
			rhs := EvaluateExpression(e.Value, env, ctx)
			var newVal RuntimeValue
			if e.Operator.Kind == lexer.ASSIGNMENT {
				newVal = rhs
			} else {
				current, exists := props[key]
				if !exists {
					panic(TunaError(fmt.Sprintf("property '%s' does not exist on object", key)))
				}
				newVal = applyCompoundOp(e.Operator, current, rhs)
			}
			props[key] = newVal
			if sym, ok2 := idx.Left.(tunaparser.SymbolExpression); ok2 {
				env.MustUpdate(sym.Value, RuntimeValue{Kind: ObjectVal, Value: props})
			}
			return newVal
		}

		if left.Kind != ArrayVal {
			panic(TunaError(fmt.Sprintf("cannot index-assign into type '%s'", left.Kind)))
		}
		if indexVal.Kind != NumberVal {
			panic(TunaError("index must be a number"))
		}
		i := int(indexVal.Value.(float64))
		arr := left.Value.([]RuntimeValue)
		if i < 0 {
			panic(TunaError(fmt.Sprintf("negative indices are not supported (got %d)", i)))
		}
		if i >= len(arr) {
			panic(TunaError(fmt.Sprintf("index %d out of bounds (length %d)", i, len(arr))))
		}
		rhs := EvaluateExpression(e.Value, env, ctx)
		var newVal RuntimeValue
		if e.Operator.Kind == lexer.ASSIGNMENT {
			newVal = rhs
		} else {
			newVal = applyCompoundOp(e.Operator, arr[i], rhs)
		}
		if sym, ok2 := idx.Left.(tunaparser.SymbolExpression); ok2 {
			if declType := env.GetDeclaredType(sym.Value); declType != nil {
				if arrType, ok3 := declType.(tunaparser.ArrayType); ok3 {
					checkType(sym.Value, newVal, arrType.Underlying)
				}
			}
		}
		arr[i] = newVal
		if sym, ok2 := idx.Left.(tunaparser.SymbolExpression); ok2 {
			env.MustUpdate(sym.Value, RuntimeValue{Kind: ArrayVal, Value: arr})
		}
		return newVal
	}

	symbol, ok := e.Assignee.(tunaparser.SymbolExpression)
	if !ok {
		panic(TunaError("invalid assignment target"))
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
		fmt.Printf("\033[33m[Tunascript Warning]\033[0m overwriting builtin '%s'\n", symbol.Value)
	}

	env.MustUpdate(symbol.Value, newVal)
	return newVal
}

func EvaluateCallExpression(e tunaparser.CallExpression, env *Environment, ctx ExecContext) RuntimeValue {
	callee := EvaluateExpression(e.Callee, env, ctx)
	if callee.Kind != FunctionVal {
		panic(TunaError(fmt.Sprintf("cannot call a value of type '%s'", callee.Kind)))
	}

	switch f := callee.Value.(type) {
	case NativeFunction:
		args := make([]RuntimeValue, len(e.Arguments))
		for i, arg := range e.Arguments {
			args[i] = EvaluateExpression(arg, env, ctx)
		}
		return f.Call(args)

	case FunctionValue:
		args := make([]RuntimeValue, len(e.Arguments))
		for i, arg := range e.Arguments {
			args[i] = EvaluateExpression(arg, env, ctx)
		}
	
		return CallFunctionValue(f, args, env, ctx)
	default:
		panic(TunaError("cannot call non-function value"))
	}
}

func EvaluatePostfixExpression(e tunaparser.PostfixExpression, env *Environment, ctx ExecContext) RuntimeValue {
	symbol, ok := e.Left.(tunaparser.SymbolExpression)
	if !ok {
		panic(TunaError("'++' and '--' can only be applied to a variable"))
	}
	current := env.MustGet(symbol.Value)
	if current.Kind != NumberVal {
		panic(TunaError(fmt.Sprintf("'%s' cannot be applied to type '%s'", e.Operator.Value, current.Kind)))
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
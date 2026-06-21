package interpreter

import (
	tunaparser "tunascript/src/parser"
	"fmt"
	"sort"
)

func EvaluateStatement(stmt tunaparser.Statement, env *Environment, ctx ExecContext) EvalResult {
	switch s := stmt.(type) {

	case tunaparser.ExpressionStatement:
		val := EvaluateExpression(s.Expression, env, ctx)
		return valueResult(val)

	case tunaparser.VariableDecStatement:
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

	case tunaparser.BreakStatement:
		if !ctx.inLoop && !ctx.Cfg.NonStrict {
			if ctx.inFunction {
				panic(TunaError("`break` cannot cross a function boundary"))
			}
			panic(TunaError("`break` can only be used inside a loop"))
		}
		return breakResult()

	case tunaparser.ContinueStatement:
		if !ctx.inLoop && !ctx.Cfg.NonStrict {
			if ctx.inFunction {
				panic(TunaError("`continue` cannot cross a function boundary"))
			}
			panic(TunaError("`continue` can only be used inside a loop"))
		}
		return continueResult()

	case tunaparser.ReturnStatement:
		if !ctx.inFunction {
			if !ctx.Cfg.NonStrict {
				panic(TunaError("`serve` can only be used inside a function"))
			}
			val := EvaluateExpression(s.Value, env, ctx)
			return valueResult(val)
		}
		val := EvaluateExpression(s.Value, env, ctx)
		return returnResult(val)

	case tunaparser.IfStatement:
		condition := EvaluateExpression(s.Condition, env, ctx)
		if IsTruthy(condition) {
			return EvaluateBlock(s.Then, NewEnvironment(env), ctx)
		} else if s.Else != nil {
			return EvaluateBlock(*s.Else, NewEnvironment(env), ctx)
		}
		return NullResult

	case tunaparser.WhileStatement:
		result := NullResult
		loopCtx := ctx.withLoop()
		for {
			if !IsTruthy(EvaluateExpression(s.Condition, env, ctx)) {
				break
			}
			r := EvaluateBlock(s.Body, NewEnvironment(env), loopCtx)
			switch r.Signal {
			case sigBreak:
				return NullResult
			case sigContinue:
				continue
			case sigReturn:
				return r
			}
			result = r
		}
		return result

	case tunaparser.ForInStatement:
		iterable := EvaluateExpression(s.Iterable, env, ctx)
		result := NullResult
		loopCtx := ctx.withLoop()

		switch iterable.Kind {
		case ArrayVal:
			for i, element := range iterable.Value.([]RuntimeValue) {
				loopEnv := NewEnvironment(env)
				if s.KeyVar != "" {
					loopEnv.Set(s.KeyVar, RuntimeValue{Kind: NumberVal, Value: float64(i)})
				}
				loopEnv.Set(s.Iterator, element)
				r := EvaluateBlock(s.Body, loopEnv, loopCtx)
				switch r.Signal {
				case sigBreak:
					return NullResult
				case sigContinue:
					continue
				case sigReturn:
					return r
				}
				result = r
			}
			return result

		case ObjectVal:
			props := iterable.Value.(map[string]RuntimeValue)
			keys := make([]string, 0, len(props))
			for k := range props {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				loopEnv := NewEnvironment(env)
				if s.KeyVar != "" {
					loopEnv.Set(s.KeyVar, RuntimeValue{Kind: StringVal, Value: k})
					loopEnv.Set(s.Iterator, props[k])
				} else {

					loopEnv.Set(s.Iterator, props[k])
				}
				r := EvaluateBlock(s.Body, loopEnv, loopCtx)
				switch r.Signal {
				case sigBreak:
					return NullResult
				case sigContinue:
					continue
				case sigReturn:
					return r
				}
				result = r
			}
			return result

		default:
			panic(TunaError(fmt.Sprintf(
				"cannot iterate over non-array, non-object value of type '%s' in for/in", iterable.Kind)))
		}

	case tunaparser.FunctionDecStatement:
		closureEnv := NewEnvironment(env)
		fn := RuntimeValue{
			Kind:	FunctionVal,
			Value: FunctionValue{
				Name:		s.Name,
				Parameters:	s.Parameters,
				ReturnType:	s.ReturnType,
				Body:		s.Body,
				Env:		closureEnv,
			},
		}
		closureEnv.Set(s.Name, fn)

		if ctx.builtinNames[s.Name] {
			fmt.Printf("\033[33m[Tunascript Warning]\033[0m overwriting builtin '%s'\n", s.Name)
		}
		env.Set(s.Name, fn)
		return valueResult(fn)

	case tunaparser.BlockStatement:
		return EvaluateBlock(s, NewEnvironment(env), ctx)

	case tunaparser.CastStatement:
		return EvaluateStatement(s.Inner, env, ctx)

	case tunaparser.SchoolStatement:
		if _, exists := schoolRegistry[s.Name]; exists {
			fmt.Printf("\033[33m[Tunascript Warning]\033[0m redefining school '%s'\n", s.Name)
		}
		schoolRegistry[s.Name] = s
		return NullResult

	case tunaparser.ImportStatement:
		exports := loadModule(s.Path, ctx)
		for _, item := range s.Items {
			val, ok := exports[item.Name]
			if !ok {
				panic(TunaError(fmt.Sprintf(
					"module '%s' does not export '%s'", s.Path, item.Name)))
			}
			env.Set(item.Alias, val)
		}
		return NullResult
	case tunaparser.SwapStatement:
		vals := make([]RuntimeValue, len(s.Values))
		for i, v := range s.Values {
			vals[i] = EvaluateExpression(v, env, ctx)
		}
		for i, target := range s.Targets {
			sym, ok := target.(tunaparser.SymbolExpression)
			if !ok {
				panic(TunaError("swap targets must be simple variables"))
			}
			env.MustUpdate(sym.Value, vals[i])
		}
		return NullResult
	case tunaparser.TryStatement:
		var result EvalResult
		func() {
			defer func() {
				if r := recover(); r != nil {
					var msg string
					switch v := r.(type) {
					case string:
						msg = v
					case error:
						msg = v.Error()
					default:
						msg = fmt.Sprintf("%v", v)
					}
					hookEnv := NewEnvironment(env)
					hookEnv.Set(s.ErrName, RuntimeValue{Kind: StringVal, Value: msg})
					result = EvaluateBlock(s.Hook, hookEnv, ctx)
				}
			}()
			result = EvaluateBlock(s.Body, NewEnvironment(env), ctx)
		}()
		return result
	default:
		panic(TunaError(fmt.Sprintf("unknown statement type: %T", stmt)))
	}
}

package interpreter

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
)

func registerBuiltins(env *Environment) {
	env.Set("bubble", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "bubble",
		Call: func(args []RuntimeValue) RuntimeValue {
			parts := make([]string, len(args))
			for i, arg := range args {
				parts[i] = nativeToString(arg)
			}
			fmt.Println(strings.Join(parts, " "))
			return RuntimeValue{Kind: NullVal}
		},
	}})

	env.Set("typeOf", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "typeOf",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("typeof", args, 1)
			return RuntimeValue{Kind: StringVal, Value: args[0].Kind.String()}
		},
	}})

	env.Set("toNumber", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "toNumber",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("toNumber", args, 1)
			switch args[0].Kind {
			case NumberVal:
				return args[0]
			case StringVal:
				n, err := strconv.ParseFloat(args[0].Value.(string), 64)
				if err != nil {
					panic(fmt.Sprintf("toNumber() cannot convert \"%s\" to a number", args[0].Value.(string)))
				}
				return RuntimeValue{Kind: NumberVal, Value: n}
			case BoolVal:
				if args[0].Value.(bool) {
					return RuntimeValue{Kind: NumberVal, Value: float64(1)}
				}
				return RuntimeValue{Kind: NumberVal, Value: float64(0)}
			default:
				panic(fmt.Sprintf("toNumber() cannot convert type '%s'", args[0].Kind))
			}
		},
	}})

	env.Set("toString", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "toString",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("toString", args, 1)
			return RuntimeValue{Kind: StringVal, Value: nativeToString(args[0])}
		},
	}})

	env.Set("len", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "len",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("len", args, 1)
			switch args[0].Kind {
			case ArrayVal:
				return RuntimeValue{Kind: NumberVal, Value: float64(len(args[0].Value.([]RuntimeValue)))}
			case StringVal:
				return RuntimeValue{Kind: NumberVal, Value: float64(len([]rune(args[0].Value.(string))))}
			default:
				panic(fmt.Sprintf("len() expects a string or array, got '%s'", args[0].Kind))
			}
		},
	}})

	env.Set("split", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "split",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("split", args, 2)
			assertKind("split", args[0], StringVal)
			assertKind("split", args[1], StringVal)
			parts := strings.Split(args[0].Value.(string), args[1].Value.(string))
			elements := make([]RuntimeValue, len(parts))
			for i, p := range parts {
				elements[i] = RuntimeValue{Kind: StringVal, Value: p}
			}
			return RuntimeValue{Kind: ArrayVal, Value: elements}
		},
	}})

	env.Set("join", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "join",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("join", args, 2)
			assertKind("join", args[0], ArrayVal)
			assertKind("join", args[1], StringVal)
			elements := args[0].Value.([]RuntimeValue)
			parts := make([]string, len(elements))
			for i, e := range elements {
				parts[i] = nativeToString(e)
			}
			return RuntimeValue{Kind: StringVal, Value: strings.Join(parts, args[1].Value.(string))}
		},
	}})

	env.Set("contains", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "contains",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("contains", args, 2)
			switch args[0].Kind {
			case StringVal:
				assertKind("contains", args[1], StringVal)
				return RuntimeValue{Kind: BoolVal, Value: strings.Contains(args[0].Value.(string), args[1].Value.(string))}
			case ArrayVal:
				needle := args[1]
				for _, el := range args[0].Value.([]RuntimeValue) {
					if runtimeEqual(el, needle) {
						return RuntimeValue{Kind: BoolVal, Value: true}
					}
				}
				return RuntimeValue{Kind: BoolVal, Value: false}
			default:
				panic(fmt.Sprintf("contains() expects a string or array, got '%s'", args[0].Kind))
			}
		},
	}})

	env.Set("upper", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "upper",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("upper", args, 1)
			assertKind("upper", args[0], StringVal)
			return RuntimeValue{Kind: StringVal, Value: strings.ToUpper(args[0].Value.(string))}
		},
	}})

	env.Set("lower", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "lower",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("lower", args, 1)
			assertKind("lower", args[0], StringVal)
			return RuntimeValue{Kind: StringVal, Value: strings.ToLower(args[0].Value.(string))}
		},
	}})

	env.Set("trim", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "trim",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("trim", args, 1)
			assertKind("trim", args[0], StringVal)
			return RuntimeValue{Kind: StringVal, Value: strings.TrimSpace(args[0].Value.(string))}
		},
	}})

	env.Set("push", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "push",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("push", args, 2)
			assertKind("push", args[0], ArrayVal)
			existing := args[0].Value.([]RuntimeValue)
			newArr := make([]RuntimeValue, len(existing)+1)
			copy(newArr, existing)
			newArr[len(existing)] = args[1]
			return RuntimeValue{Kind: ArrayVal, Value: newArr}
		},
	}})

	env.Set("pop", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "pop",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("pop", args, 1)
			assertKind("pop", args[0], ArrayVal)
			existing := args[0].Value.([]RuntimeValue)
			if len(existing) == 0 {
				panic("pop() called on empty array")
			}
			newArr := make([]RuntimeValue, len(existing)-1)
			copy(newArr, existing[:len(existing)-1])
			return RuntimeValue{Kind: ArrayVal, Value: newArr}
		},
	}})

	env.Set("first", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "first",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("first", args, 1)
			assertKind("first", args[0], ArrayVal)
			arr := args[0].Value.([]RuntimeValue)
			if len(arr) == 0 {
				panic("first() called on empty array")
			}
			return arr[0]
		},
	}})

	env.Set("last", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "last",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("last", args, 1)
			assertKind("last", args[0], ArrayVal)
			arr := args[0].Value.([]RuntimeValue)
			if len(arr) == 0 {
				panic("last() called on empty array")
			}
			return arr[len(arr)-1]
		},
	}})

	env.Set("slice", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "slice",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("slice", args, 3)
			assertKind("slice", args[0], ArrayVal)
			assertKind("slice", args[1], NumberVal)
			assertKind("slice", args[2], NumberVal)
			arr := args[0].Value.([]RuntimeValue)
			start := int(args[1].Value.(float64))
			end := int(args[2].Value.(float64))
			if start < 0 || end > len(arr) || start > end {
				panic(fmt.Sprintf("slice() index out of bounds: [%d:%d] on array of length %d", start, end, len(arr)))
			}
			return RuntimeValue{Kind: ArrayVal, Value: arr[start:end]}
		},
	}})

	env.Set("sort", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "sort",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("sort", args, 1)
			assertKind("sort", args[0], ArrayVal)
			arr := args[0].Value.([]RuntimeValue)
			newArr := make([]RuntimeValue, len(arr))
			copy(newArr, arr)
			sort.Slice(newArr, func(i, j int) bool {
				a, b := newArr[i], newArr[j]
				if a.Kind == NumberVal && b.Kind == NumberVal {
					return a.Value.(float64) < b.Value.(float64)
				}
				if a.Kind == StringVal && b.Kind == StringVal {
					return a.Value.(string) < b.Value.(string)
				}
				panic("sort() requires an array of all numbers or all strings")
			})
			return RuntimeValue{Kind: ArrayVal, Value: newArr}
		},
	}})

	env.Set("floor", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "floor",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("floor", args, 1)
			assertKind("floor", args[0], NumberVal)
			return RuntimeValue{Kind: NumberVal, Value: math.Floor(args[0].Value.(float64))}
		},
	}})

	env.Set("ceil", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "ceil",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("ceil", args, 1)
			assertKind("ceil", args[0], NumberVal)
			return RuntimeValue{Kind: NumberVal, Value: math.Ceil(args[0].Value.(float64))}
		},
	}})

	env.Set("round", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "round",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("round", args, 1)
			assertKind("round", args[0], NumberVal)
			return RuntimeValue{Kind: NumberVal, Value: math.Round(args[0].Value.(float64))}
		},
	}})

	env.Set("abs", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "abs",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("abs", args, 1)
			assertKind("abs", args[0], NumberVal)
			return RuntimeValue{Kind: NumberVal, Value: math.Abs(args[0].Value.(float64))}
		},
	}})

	env.Set("min", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "min",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("min", args, 2)
			assertKind("min", args[0], NumberVal)
			assertKind("min", args[1], NumberVal)
			return RuntimeValue{Kind: NumberVal, Value: math.Min(args[0].Value.(float64), args[1].Value.(float64))}
		},
	}})

	env.Set("max", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "max",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("max", args, 2)
			assertKind("max", args[0], NumberVal)
			assertKind("max", args[1], NumberVal)
			return RuntimeValue{Kind: NumberVal, Value: math.Max(args[0].Value.(float64), args[1].Value.(float64))}
		},
	}})

	env.Set("pow", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "pow",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("pow", args, 2)
			assertKind("pow", args[0], NumberVal)
			assertKind("pow", args[1], NumberVal)
			return RuntimeValue{Kind: NumberVal, Value: math.Pow(args[0].Value.(float64), args[1].Value.(float64))}
		},
	}})

	env.Set("sqrt", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "sqrt",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("sqrt", args, 1)
			assertKind("sqrt", args[0], NumberVal)
			return RuntimeValue{Kind: NumberVal, Value: math.Sqrt(args[0].Value.(float64))}
		},
	}})

	env.Set("rand", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "rand",
		Call: func(args []RuntimeValue) RuntimeValue {
			return RuntimeValue{Kind: NumberVal, Value: rand.Float64()}
		},
	}})

	env.Set("randInt", RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
		Name: "randInt",
		Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("randInt", args, 2)
			assertKind("randInt", args[0], NumberVal)
			assertKind("randInt", args[1], NumberVal)
			lo := int(args[0].Value.(float64))
			hi := int(args[1].Value.(float64))
			if hi < lo {
				panic("randInt() max must be >= min")
			}
			return RuntimeValue{Kind: NumberVal, Value: float64(lo + rand.Intn(hi-lo+1))}
		},
	}})
}

func assertArgCount(name string, args []RuntimeValue, n int) {
	if len(args) != n {
		panic(fmt.Sprintf("%s() expects %d argument(s) but got %d", name, n, len(args)))
	}
}

func assertKind(name string, val RuntimeValue, expected ValueKind) {
	if val.Kind != expected {
		panic(fmt.Sprintf("%s() expects a %s but got '%s'", name, expected, val.Kind))
	}
}

func runtimeEqual(a, b RuntimeValue) bool {
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case NumberVal:
		return a.Value.(float64) == b.Value.(float64)
	case StringVal:
		return a.Value.(string) == b.Value.(string)
	case BoolVal:
		return a.Value.(bool) == b.Value.(bool)
	case NullVal:
		return true
	default:
		return false
	}
}

func nativeToString(val RuntimeValue) string {
	switch val.Kind {
	case NumberVal:
		return fmt.Sprintf("%g", val.Value.(float64))
	case StringVal:
		return val.Value.(string)
	case BoolVal:
		if val.Value.(bool) {
			return "true"
		}
		return "false"
	case NullVal:
		return "null"
	case ArrayVal:
		elements := val.Value.([]RuntimeValue)
		parts := make([]string, len(elements))
		for i, el := range elements {
			parts[i] = nativeToString(el)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case FunctionVal:
		return "function"
	default:
		return "unknown"
	}
}
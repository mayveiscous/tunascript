package interpreter

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
)

type NativeFunctionWithEnv struct {
	Name string
	Call func(args []RuntimeValue, rawArgs []interface{}, env *Environment) RuntimeValue
}

func registerBuiltins(env *Environment) {
	// ── globals ──────────────────────────────────────────────────────────────

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
			assertArgCount("typeOf", args, 1)
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

	// ── math namespace ───────────────────────────────────────────────────────

	mathNS := map[string]RuntimeValue{
		"floor": {Kind: FunctionVal, Value: NativeFunction{Name: "floor", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("math.floor", args, 1)
			assertKind("math.floor", args[0], NumberVal)
			return RuntimeValue{Kind: NumberVal, Value: math.Floor(args[0].Value.(float64))}
		}}},
		"ceil": {Kind: FunctionVal, Value: NativeFunction{Name: "ceil", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("math.ceil", args, 1)
			assertKind("math.ceil", args[0], NumberVal)
			return RuntimeValue{Kind: NumberVal, Value: math.Ceil(args[0].Value.(float64))}
		}}},
		"round": {Kind: FunctionVal, Value: NativeFunction{Name: "round", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("math.round", args, 1)
			assertKind("math.round", args[0], NumberVal)
			return RuntimeValue{Kind: NumberVal, Value: math.Round(args[0].Value.(float64))}
		}}},
		"abs": {Kind: FunctionVal, Value: NativeFunction{Name: "abs", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("math.abs", args, 1)
			assertKind("math.abs", args[0], NumberVal)
			return RuntimeValue{Kind: NumberVal, Value: math.Abs(args[0].Value.(float64))}
		}}},
		"min": {Kind: FunctionVal, Value: NativeFunction{Name: "min", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("math.min", args, 2)
			assertKind("math.min", args[0], NumberVal)
			assertKind("math.min", args[1], NumberVal)
			return RuntimeValue{Kind: NumberVal, Value: math.Min(args[0].Value.(float64), args[1].Value.(float64))}
		}}},
		"max": {Kind: FunctionVal, Value: NativeFunction{Name: "max", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("math.max", args, 2)
			assertKind("math.max", args[0], NumberVal)
			assertKind("math.max", args[1], NumberVal)
			return RuntimeValue{Kind: NumberVal, Value: math.Max(args[0].Value.(float64), args[1].Value.(float64))}
		}}},
		"pow": {Kind: FunctionVal, Value: NativeFunction{Name: "pow", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("math.pow", args, 2)
			assertKind("math.pow", args[0], NumberVal)
			assertKind("math.pow", args[1], NumberVal)
			return RuntimeValue{Kind: NumberVal, Value: math.Pow(args[0].Value.(float64), args[1].Value.(float64))}
		}}},
		"sqrt": {Kind: FunctionVal, Value: NativeFunction{Name: "sqrt", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("math.sqrt", args, 1)
			assertKind("math.sqrt", args[0], NumberVal)
			return RuntimeValue{Kind: NumberVal, Value: math.Sqrt(args[0].Value.(float64))}
		}}},
		"rand": {Kind: FunctionVal, Value: NativeFunction{Name: "rand", Call: func(args []RuntimeValue) RuntimeValue {
			return RuntimeValue{Kind: NumberVal, Value: rand.Float64()}
		}}},
		"randInt": {Kind: FunctionVal, Value: NativeFunction{Name: "randInt", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("math.randInt", args, 2)
			assertKind("math.randInt", args[0], NumberVal)
			assertKind("math.randInt", args[1], NumberVal)
			lo := int(args[0].Value.(float64))
			hi := int(args[1].Value.(float64))
			if hi < lo {
				panic("math.randInt() max must be >= min")
			}
			return RuntimeValue{Kind: NumberVal, Value: float64(lo + rand.Intn(hi-lo+1))}
		}}},
		"pi": {Kind: NumberVal, Value: math.Pi},
		"e":  {Kind: NumberVal, Value: math.E},
	}
	env.Set("math", RuntimeValue{Kind: ObjectVal, Value: mathNS})

	// ── string namespace ─────────────────────────────────────────────────────

	stringNS := map[string]RuntimeValue{
		"upper": {Kind: FunctionVal, Value: NativeFunction{Name: "upper", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("string.upper", args, 1)
			assertKind("string.upper", args[0], StringVal)
			return RuntimeValue{Kind: StringVal, Value: strings.ToUpper(args[0].Value.(string))}
		}}},
		"lower": {Kind: FunctionVal, Value: NativeFunction{Name: "lower", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("string.lower", args, 1)
			assertKind("string.lower", args[0], StringVal)
			return RuntimeValue{Kind: StringVal, Value: strings.ToLower(args[0].Value.(string))}
		}}},
		"trim": {Kind: FunctionVal, Value: NativeFunction{Name: "trim", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("string.trim", args, 1)
			assertKind("string.trim", args[0], StringVal)
			return RuntimeValue{Kind: StringVal, Value: strings.TrimSpace(args[0].Value.(string))}
		}}},
		"split": {Kind: FunctionVal, Value: NativeFunction{Name: "split", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("string.split", args, 2)
			assertKind("string.split", args[0], StringVal)
			assertKind("string.split", args[1], StringVal)
			parts := strings.Split(args[0].Value.(string), args[1].Value.(string))
			elements := make([]RuntimeValue, len(parts))
			for i, p := range parts {
				elements[i] = RuntimeValue{Kind: StringVal, Value: p}
			}
			return RuntimeValue{Kind: ArrayVal, Value: elements}
		}}},
		"contains": {Kind: FunctionVal, Value: NativeFunction{Name: "contains", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("string.contains", args, 2)
			assertKind("string.contains", args[0], StringVal)
			assertKind("string.contains", args[1], StringVal)
			return RuntimeValue{Kind: BoolVal, Value: strings.Contains(args[0].Value.(string), args[1].Value.(string))}
		}}},
		"replace": {Kind: FunctionVal, Value: NativeFunction{Name: "replace", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("string.replace", args, 3)
			assertKind("string.replace", args[0], StringVal)
			assertKind("string.replace", args[1], StringVal)
			assertKind("string.replace", args[2], StringVal)
			return RuntimeValue{Kind: StringVal, Value: strings.ReplaceAll(args[0].Value.(string), args[1].Value.(string), args[2].Value.(string))}
		}}},
		"startsWith": {Kind: FunctionVal, Value: NativeFunction{Name: "startsWith", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("string.startsWith", args, 2)
			assertKind("string.startsWith", args[0], StringVal)
			assertKind("string.startsWith", args[1], StringVal)
			return RuntimeValue{Kind: BoolVal, Value: strings.HasPrefix(args[0].Value.(string), args[1].Value.(string))}
		}}},
		"endsWith": {Kind: FunctionVal, Value: NativeFunction{Name: "endsWith", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("string.endsWith", args, 2)
			assertKind("string.endsWith", args[0], StringVal)
			assertKind("string.endsWith", args[1], StringVal)
			return RuntimeValue{Kind: BoolVal, Value: strings.HasSuffix(args[0].Value.(string), args[1].Value.(string))}
		}}},
		"repeat": {Kind: FunctionVal, Value: NativeFunction{Name: "repeat", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("string.repeat", args, 2)
			assertKind("string.repeat", args[0], StringVal)
			assertKind("string.repeat", args[1], NumberVal)
			n := int(args[1].Value.(float64))
			if n < 0 {
				panic("string.repeat() count must be >= 0")
			}
			return RuntimeValue{Kind: StringVal, Value: strings.Repeat(args[0].Value.(string), n)}
		}}},
	}
	env.Set("string", RuntimeValue{Kind: ObjectVal, Value: stringNS})

	// ── array namespace ──────────────────────────────────────────────────────
	arrayNS := map[string]RuntimeValue{
		"push": {Kind: FunctionVal, Value: MutatingNativeFunction{Name: "push",
			Call: func(args []RuntimeValue) RuntimeValue {
				assertArgCount("array.push", args, 2)
				assertKind("array.push", args[0], ArrayVal)
				existing := args[0].Value.([]RuntimeValue)
				newArr := append(existing, args[1])
				return RuntimeValue{Kind: ArrayVal, Value: newArr}
			},
		}},
		"pop": {Kind: FunctionVal, Value: MutatingNativeFunction{Name: "pop",
			Call: func(args []RuntimeValue) RuntimeValue {
				assertArgCount("array.pop", args, 1)
				assertKind("array.pop", args[0], ArrayVal)
				existing := args[0].Value.([]RuntimeValue)
				if len(existing) == 0 {
					panic("array.pop() called on empty array")
				}
				return RuntimeValue{Kind: ArrayVal, Value: existing[:len(existing)-1]}
			},
		}},
		"sort": {Kind: FunctionVal, Value: MutatingNativeFunction{Name: "sort",
			Call: func(args []RuntimeValue) RuntimeValue {
				assertArgCount("array.sort", args, 1)
				assertKind("array.sort", args[0], ArrayVal)
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
					panic("array.sort() requires an array of all numbers or all strings")
				})
				return RuntimeValue{Kind: ArrayVal, Value: newArr}
			},
		}},
		"reverse": {Kind: FunctionVal, Value: MutatingNativeFunction{Name: "reverse",
			Call: func(args []RuntimeValue) RuntimeValue {
				assertArgCount("array.reverse", args, 1)
				assertKind("array.reverse", args[0], ArrayVal)
				arr := args[0].Value.([]RuntimeValue)
				newArr := make([]RuntimeValue, len(arr))
				for i, v := range arr {
					newArr[len(arr)-1-i] = v
				}
				return RuntimeValue{Kind: ArrayVal, Value: newArr}
			},
		}},

		"first": {Kind: FunctionVal, Value: NativeFunction{Name: "first", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("array.first", args, 1)
			assertKind("array.first", args[0], ArrayVal)
			arr := args[0].Value.([]RuntimeValue)
			if len(arr) == 0 {
				panic("array.first() called on empty array")
			}
			return arr[0]
		}}},
		"last": {Kind: FunctionVal, Value: NativeFunction{Name: "last", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("array.last", args, 1)
			assertKind("array.last", args[0], ArrayVal)
			arr := args[0].Value.([]RuntimeValue)
			if len(arr) == 0 {
				panic("array.last() called on empty array")
			}
			return arr[len(arr)-1]
		}}},
		"slice": {Kind: FunctionVal, Value: MutatingNativeFunction{Name: "slice",
		Call: func(args []RuntimeValue) RuntimeValue {
			 assertArgCount("array.slice", args, 3)
			 assertKind("array.slice", args[0], ArrayVal)
			 assertKind("array.slice", args[1], NumberVal)
			 assertKind("array.slice", args[2], NumberVal)
			 arr := args[0].Value.([]RuntimeValue)
			 start := int(args[1].Value.(float64))
			 end := int(args[2].Value.(float64))
			 if start < 0 || end > len(arr) || start > end {
				  panic(fmt.Sprintf("array.slice() index out of bounds: [%d:%d] on array of length %d", start, end, len(arr)))
			 }
			 // Mutate the array in place to the sliced range
			 return RuntimeValue{Kind: ArrayVal, Value: arr[start:end]}
		},
  }},
		"contains": {Kind: FunctionVal, Value: NativeFunction{Name: "contains", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("array.contains", args, 2)
			assertKind("array.contains", args[0], ArrayVal)
			needle := args[1]
			for _, el := range args[0].Value.([]RuntimeValue) {
				if runtimeEqual(el, needle) {
					return RuntimeValue{Kind: BoolVal, Value: true}
				}
			}
			return RuntimeValue{Kind: BoolVal, Value: false}
		}}},
		"join": {Kind: FunctionVal, Value: NativeFunction{Name: "join", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("array.join", args, 2)
			assertKind("array.join", args[0], ArrayVal)
			assertKind("array.join", args[1], StringVal)
			elements := args[0].Value.([]RuntimeValue)
			parts := make([]string, len(elements))
			for i, e := range elements {
				parts[i] = nativeToString(e)
			}
			return RuntimeValue{Kind: StringVal, Value: strings.Join(parts, args[1].Value.(string))}
		}}},
	}
	env.Set("array", RuntimeValue{Kind: ObjectVal, Value: arrayNS})
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
	case ObjectVal:
		props := val.Value.(map[string]RuntimeValue)
		parts := make([]string, 0, len(props))
		for k, v := range props {
			parts = append(parts, k+": "+nativeToString(v))
		}
		sort.Strings(parts)
		return "{" + strings.Join(parts, ", ") + "}"
	default:
		return "unknown"
	}
}
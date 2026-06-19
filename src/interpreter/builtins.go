package interpreter

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"os"
	"time"
	"golang.org/x/term"
	"encoding/json"
	"tunascript/src/imui"
)

func registerBuiltins(env *Environment, ctx ExecContext) {
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
					panic(TunaError(fmt.Sprintf(
						"toNumber() cannot convert \"%s\" to a number", args[0].Value.(string))))
				}
				return RuntimeValue{Kind: NumberVal, Value: n}
			case BoolVal:
				if args[0].Value.(bool) {
					return RuntimeValue{Kind: NumberVal, Value: float64(1)}
				}
				return RuntimeValue{Kind: NumberVal, Value: float64(0)}
			default:
				panic(TunaError(fmt.Sprintf("toNumber() cannot convert type '%s'", args[0].Kind)))
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
				panic(TunaError(fmt.Sprintf("len() expects a string or array, got '%s'", args[0].Kind)))
			}
		},
	}})

	jsonNS := map[string]RuntimeValue{
		"encode": {Kind: FunctionVal, Value: NativeFunction{Name: "encode", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("json.encode", args, 1)
			native := runtimeToNative(args[0])
			b, err := json.MarshalIndent(native, "", "  ")
			if err != nil {
				panic(TunaError(fmt.Sprintf("json.encode() failed: %s", err)))
			}
			return RuntimeValue{Kind: StringVal, Value: string(b)}
		}}},
		"decode": {Kind: FunctionVal, Value: NativeFunction{Name: "decode", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("json.decode", args, 1)
			assertKind("json.decode", args[0], StringVal)
			var raw any
			if err := json.Unmarshal([]byte(args[0].Value.(string)), &raw); err != nil {
				panic(TunaError(fmt.Sprintf("json.decode() failed: %s", err)))
			}
			return nativeToRuntime(raw)
		}}},
	}
	env.Set("json", RuntimeValue{Kind: ObjectVal, Value: jsonNS})

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
				panic(TunaError("math.randInt() max must be >= min"))
			}
			return RuntimeValue{Kind: NumberVal, Value: float64(lo + rand.Intn(hi-lo+1))}
		}}},
		"inf": {Kind: NumberVal, Value: math.Inf(1)},
		"pi": {Kind: NumberVal, Value: math.Pi},
		"e":  {Kind: NumberVal, Value: math.E},
	}
	env.Set("math", RuntimeValue{Kind: ObjectVal, Value: mathNS})

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
			return RuntimeValue{Kind: StringVal, Value: strings.ReplaceAll(
				args[0].Value.(string), args[1].Value.(string), args[2].Value.(string))}
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
				panic(TunaError("string.repeat() count must be >= 0"))
			}
			return RuntimeValue{Kind: StringVal, Value: strings.Repeat(args[0].Value.(string), n)}
		}}},
		"slice": {Kind: FunctionVal, Value: NativeFunction{Name: "slice", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("string.slice", args, 3)
			assertKind("string.slice", args[0], StringVal)
			assertKind("string.slice", args[1], NumberVal)
			assertKind("string.slice", args[2], NumberVal)
			runes := []rune(args[0].Value.(string))
			start := int(args[1].Value.(float64))
			end   := int(args[2].Value.(float64))
			if start < 0 || end > len(runes) || start > end {
				 panic(TunaError(fmt.Sprintf(
					  "string.slice() index out of bounds: [%d:%d] on string of length %d",
					  start, end, len(runes))))
			}
			return RuntimeValue{Kind: StringVal, Value: string(runes[start:end])}
	  }}},
	  
	  "indexOf": {Kind: FunctionVal, Value: NativeFunction{Name: "indexOf", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("string.indexOf", args, 2)
			assertKind("string.indexOf", args[0], StringVal)
			assertKind("string.indexOf", args[1], StringVal)
			idx := strings.Index(args[0].Value.(string), args[1].Value.(string))
			return RuntimeValue{Kind: NumberVal, Value: float64(idx)}
	  }}},
	  
	  "charCode": {Kind: FunctionVal, Value: NativeFunction{Name: "charCode", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("string.charCode", args, 1)
			assertKind("string.charCode", args[0], StringVal)
			runes := []rune(args[0].Value.(string))
			if len(runes) == 0 {
				 panic(TunaError("string.charCode() cannot get char code of empty string"))
			}
			return RuntimeValue{Kind: NumberVal, Value: float64(runes[0])}
	  }}},
	  
	  "fromCharCode": {Kind: FunctionVal, Value: NativeFunction{Name: "fromCharCode", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("string.fromCharCode", args, 1)
			assertKind("string.fromCharCode", args[0], NumberVal)
			return RuntimeValue{Kind: StringVal, Value: string(rune(int(args[0].Value.(float64))))}
	  }}},
	}
	env.Set("string", RuntimeValue{Kind: ObjectVal, Value: stringNS})

	arrayNS := map[string]RuntimeValue{
		"push": {Kind: FunctionVal, Value: NativeFunction{Name: "push",
			Call: func(args []RuntimeValue) RuntimeValue {
				assertArgCount("array.push", args, 2)
				assertKind("array.push", args[0], ArrayVal)
				existing := args[0].Value.([]RuntimeValue)
				newArr := make([]RuntimeValue, len(existing)+1)
				copy(newArr, existing)
				newArr[len(existing)] = args[1]
				return RuntimeValue{Kind: ArrayVal, Value: newArr}
			},
		}},

		"pop": {Kind: FunctionVal, Value: NativeFunction{Name: "pop",
			Call: func(args []RuntimeValue) RuntimeValue {
				assertArgCount("array.pop", args, 1)
				assertKind("array.pop", args[0], ArrayVal)
				existing := args[0].Value.([]RuntimeValue)
				if len(existing) == 0 {
					panic(TunaError("array.pop() called on empty array"))
				}
				return existing[len(existing)-1]
			},
		}},

		"dropLast": {Kind: FunctionVal, Value: NativeFunction{Name: "dropLast",
			Call: func(args []RuntimeValue) RuntimeValue {
				assertArgCount("array.dropLast", args, 1)
				assertKind("array.dropLast", args[0], ArrayVal)
				existing := args[0].Value.([]RuntimeValue)
				if len(existing) == 0 {
					panic(TunaError("array.dropLast() called on empty array"))
				}
				newArr := make([]RuntimeValue, len(existing)-1)
				copy(newArr, existing[:len(existing)-1])
				return RuntimeValue{Kind: ArrayVal, Value: newArr}
			},
		}},

		"sort": {Kind: FunctionVal, Value: NativeFunction{Name: "sort",
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
					panic(TunaError("array.sort() requires an array of all numbers or all strings"))
				})
				return RuntimeValue{Kind: ArrayVal, Value: newArr}
			},
		}},

		"reverse": {Kind: FunctionVal, Value: NativeFunction{Name: "reverse",
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
				panic(TunaError("array.first() called on empty array"))
			}
			return arr[0]
		}}},

		"last": {Kind: FunctionVal, Value: NativeFunction{Name: "last", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("array.last", args, 1)
			assertKind("array.last", args[0], ArrayVal)
			arr := args[0].Value.([]RuntimeValue)
			if len(arr) == 0 {
				panic(TunaError("array.last() called on empty array"))
			}
			return arr[len(arr)-1]
		}}},

		"slice": {Kind: FunctionVal, Value: NativeFunction{Name: "slice",
			Call: func(args []RuntimeValue) RuntimeValue {
				assertArgCount("array.slice", args, 3)
				assertKind("array.slice", args[0], ArrayVal)
				assertKind("array.slice", args[1], NumberVal)
				assertKind("array.slice", args[2], NumberVal)
				arr := args[0].Value.([]RuntimeValue)
				start := int(args[1].Value.(float64))
				end := int(args[2].Value.(float64))
				if start < 0 || end > len(arr) || start > end {
					panic(TunaError(fmt.Sprintf(
						"array.slice() index out of bounds: [%d:%d] on array of length %d",
						start, end, len(arr))))
				}
				newArr := make([]RuntimeValue, end-start)
				copy(newArr, arr[start:end])
				return RuntimeValue{Kind: ArrayVal, Value: newArr}
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

	tuiNS := map[string]RuntimeValue{
		"clear": {Kind: FunctionVal, Value: NativeFunction{Name: "clear", Call: func(args []RuntimeValue) RuntimeValue {
			fmt.Print("\033[H\033[2J")
			return RuntimeValue{Kind: NullVal}
		}}},
		"move": {Kind: FunctionVal, Value: NativeFunction{Name: "move", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("tui.move", args, 2)
			assertKind("tui.move", args[0], NumberVal)
			assertKind("tui.move", args[1], NumberVal)
			row := int(args[0].Value.(float64))
			col := int(args[1].Value.(float64))
			fmt.Printf("\033[%d;%dH", row, col)
			return RuntimeValue{Kind: NullVal}
		}}},
		"color": {Kind: FunctionVal, Value: NativeFunction{Name: "color", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("tui.color", args, 2)
			assertKind("tui.color", args[0], StringVal)
			assertKind("tui.color", args[1], StringVal)
			colors := map[string]string{
				"black":   "\033[30m",
				"red":     "\033[31m",
				"green":   "\033[32m",
				"yellow":  "\033[33m",
				"blue":    "\033[34m",
				"magenta": "\033[35m",
				"cyan":    "\033[36m",
				"white":   "\033[37m",
				"bold":    "\033[1m",
				"dim":     "\033[2m",
			}
			code, ok := colors[args[0].Value.(string)]
			if !ok {
				panic(TunaError(fmt.Sprintf("tui.color() unknown color '%s'", args[0].Value.(string))))
			}
			return RuntimeValue{Kind: StringVal, Value: code + args[1].Value.(string) + "\033[0m"}
		}}},
		"bar": {Kind: FunctionVal, Value: NativeFunction{Name: "bar", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("tui.bar", args, 3)
			assertKind("tui.bar", args[0], NumberVal)
			assertKind("tui.bar", args[1], NumberVal)
			assertKind("tui.bar", args[2], NumberVal)
			current := args[0].Value.(float64)
			max     := args[1].Value.(float64)
			width   := int(args[2].Value.(float64))
			if max <= 0 || math.IsInf(max, 1) {
				return RuntimeValue{Kind: StringVal, Value: "\033[36m" + strings.Repeat("█", width) + "\033[0m"}
			}
			filled := int(math.Round((current / max) * float64(width)))
			if filled < 0 { filled = 0 }
			if filled > width { filled = width }
			empty := width - filled
			color := "\033[32m"
			if current/max < 0.5 { color = "\033[33m" }
			if current/max < 0.25 { color = "\033[31m" }
			return RuntimeValue{Kind: StringVal,
				Value: color + strings.Repeat("█", filled) + "\033[2m" + strings.Repeat("░", empty) + "\033[0m"}
		}}},
		"print": {Kind: FunctionVal, Value: NativeFunction{Name: "print", Call: func(args []RuntimeValue) RuntimeValue {
			parts := make([]string, len(args))
			for i, arg := range args {
				parts[i] = nativeToString(arg)
			}
			fmt.Print(strings.Join(parts, " "))
			return RuntimeValue{Kind: NullVal}
		}}},
		"println": {Kind: FunctionVal, Value: NativeFunction{Name: "println", Call: func(args []RuntimeValue) RuntimeValue {
			parts := make([]string, len(args))
			for i, arg := range args {
				parts[i] = nativeToString(arg)
			}
			fmt.Println(strings.Join(parts, " "))
			return RuntimeValue{Kind: NullVal}
		}}},
		"input": {Kind: FunctionVal, Value: NativeFunction{Name: "input", Call: func(args []RuntimeValue) RuntimeValue {
			if len(args) == 1 {
				assertKind("tui.input", args[0], StringVal)
				fmt.Print(args[0].Value.(string))
			}
			var line string
			fmt.Scanln(&line)
			return RuntimeValue{Kind: StringVal, Value: line}
		}}},
		"sleep": {Kind: FunctionVal, Value: NativeFunction{Name: "sleep", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("tui.sleep", args, 1)
			assertKind("tui.sleep", args[0], NumberVal)
			ms := args[0].Value.(float64)
			time.Sleep(time.Duration(ms) * time.Millisecond)
			return RuntimeValue{Kind: NullVal}
		}}},
		"width": {Kind: FunctionVal, Value: NativeFunction{Name: "width", Call: func(args []RuntimeValue) RuntimeValue {
			width, _, err := term.GetSize(int(os.Stdout.Fd()))
			if err != nil { width = 80 }
			return RuntimeValue{Kind: NumberVal, Value: float64(width)}
		}}},
		"height": {Kind: FunctionVal, Value: NativeFunction{Name: "height", Call: func(args []RuntimeValue) RuntimeValue {
			_, height, err := term.GetSize(int(os.Stdout.Fd()))
			if err != nil { height = 24 }
			return RuntimeValue{Kind: NumberVal, Value: float64(height)}
		}}},
	}
	env.Set("tui", RuntimeValue{Kind: ObjectVal, Value: tuiNS})

	osNS := map[string]RuntimeValue{
		"read": {Kind: FunctionVal, Value: NativeFunction{Name: "read", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("os.read", args, 1)
			assertKind("os.read", args[0], StringVal)
			data, err := os.ReadFile(args[0].Value.(string))
			if err != nil {
				panic(TunaError(fmt.Sprintf("os.read() could not read '%s': %s", args[0].Value.(string), err)))
			}
			return RuntimeValue{Kind: StringVal, Value: string(data)}
		}}},
		"write": {Kind: FunctionVal, Value: NativeFunction{Name: "write", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("os.write", args, 2)
			assertKind("os.write", args[0], StringVal)
			assertKind("os.write", args[1], StringVal)
			err := os.WriteFile(args[0].Value.(string), []byte(args[1].Value.(string)), 0644)
			if err != nil {
				panic(TunaError(fmt.Sprintf("os.write() could not write '%s': %s", args[0].Value.(string), err)))
			}
			return RuntimeValue{Kind: NullVal}
		}}},
		"open": {Kind: FunctionVal, Value: NativeFunction{Name: "open", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("os.open", args, 1)
			assertKind("os.open", args[0], StringVal)
			f, err := os.OpenFile(args[0].Value.(string), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
			if err != nil {
				panic(TunaError(fmt.Sprintf("os.open() could not open '%s': %s", args[0].Value.(string), err)))
			}
			fileObj := map[string]RuntimeValue{
				"path": {Kind: StringVal, Value: args[0].Value.(string)},
			}
			fileObj["__fd"] = RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
				Name: "__fd",
				Call: func(writeArgs []RuntimeValue) RuntimeValue {
					if len(writeArgs) == 0 {
						return RuntimeValue{Kind: NullVal}
					}
					assertKind("file.write", writeArgs[0], StringVal)
					_, werr := f.WriteString(writeArgs[0].Value.(string))
					if werr != nil {
						panic(TunaError(fmt.Sprintf("file write error: %s", werr)))
					}
					return RuntimeValue{Kind: NullVal}
				},
			}}
			fileObj["__close"] = RuntimeValue{Kind: FunctionVal, Value: NativeFunction{
				Name: "__close",
				Call: func(_ []RuntimeValue) RuntimeValue {
					f.Close()
					return RuntimeValue{Kind: NullVal}
				},
			}}
			return RuntimeValue{Kind: ObjectVal, Value: fileObj}
		}}},
		"append": {Kind: FunctionVal, Value: NativeFunction{Name: "append", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("os.append", args, 2)
			assertKind("os.append", args[0], StringVal)
			assertKind("os.append", args[1], StringVal)
			f, err := os.OpenFile(args[0].Value.(string), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				panic(TunaError(fmt.Sprintf("os.append() could not open '%s': %s", args[0].Value.(string), err)))
			}
			defer f.Close()
			_, err = f.WriteString(args[1].Value.(string))
			if err != nil {
				panic(TunaError(fmt.Sprintf("os.append() could not write to '%s': %s", args[0].Value.(string), err)))
			}
			return RuntimeValue{Kind: NullVal}
		}}},
		"close": {Kind: FunctionVal, Value: NativeFunction{Name: "close", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("os.close", args, 1)
			assertKind("os.close", args[0], ObjectVal)
			props := args[0].Value.(map[string]RuntimeValue)
			closer, ok := props["__close"]
			if !ok {
				panic(TunaError("os.close() argument is not a file handle"))
			}
			closer.Value.(NativeFunction).Call([]RuntimeValue{})
			return RuntimeValue{Kind: NullVal}
		}}},
		"clock": {Kind: FunctionVal, Value: NativeFunction{Name: "clock", Call: func(args []RuntimeValue) RuntimeValue {
			return RuntimeValue{Kind: NumberVal, Value: float64(time.Now().UnixMilli())}
		}}},
		"args": {Kind: FunctionVal, Value: NativeFunction{Name: "args", Call: func(args []RuntimeValue) RuntimeValue {
			osArgs := os.Args
			result := make([]RuntimeValue, len(osArgs))
			for i, a := range osArgs {
				result[i] = RuntimeValue{Kind: StringVal, Value: a}
			}
			return RuntimeValue{Kind: ArrayVal, Value: result}
		}}},
	}
	env.Set("os", RuntimeValue{Kind: ObjectVal, Value: osNS})

	// ------------------------------------------------------------------
	// Persistent widget object cache — allows value-based properties
	// (e.g. btn.textColor = "#FF0000") to survive across frames.
	// ------------------------------------------------------------------
	widgetObjCache := map[string]map[string]RuntimeValue{}

	// Apply a cached color string to widget struct
	applyColorProp := func(widget *imui.Widget, obj map[string]RuntimeValue, field string) {
		if v, ok := obj[field]; ok && v.Kind == StringVal && v.Value.(string) != "" {
			widget.SetColor(field, v.Value.(string))
		}
	}

	// Apply common design props shared by most widgets
	applyCommonDesignProps := func(widget *imui.Widget, obj map[string]RuntimeValue) {
		if v, ok := obj["borderColor"]; ok && v.Kind == StringVal && v.Value.(string) != "" {
			widget.SetColor("borderColor", v.Value.(string))
		}
		if v, ok := obj["borderThickness"]; ok && v.Kind == NumberVal {
			widget.BorderThickness = int(v.Value.(float64))
			widget.HasBorderThickness = true
		}
		if v, ok := obj["cornerRadius"]; ok && v.Kind == NumberVal {
			widget.CornerRadius = int(v.Value.(float64))
			widget.HasCornerRadius = true
		}
	}

	// Seed a new widget object with common entries and move() method
	initWidgetObj := func(obj map[string]RuntimeValue, widget *imui.Widget) {
		obj["px"] = RuntimeValue{Kind: NumberVal, Value: 0}
		obj["py"] = RuntimeValue{Kind: NumberVal, Value: 0}
		w := widget
		obj["move"] = RuntimeValue{Kind: FunctionVal, Value: NativeFunction{Name: "move",
			Call: func(cbArgs []RuntimeValue) RuntimeValue {
				if len(cbArgs) == 2 &&
					cbArgs[0].Kind == NumberVal && cbArgs[1].Kind == NumberVal {
					w.Move(int(cbArgs[0].Value.(float64)), int(cbArgs[1].Value.(float64)))
				}
				return RuntimeValue{Kind: NullVal}
			},
		}}
		obj["setSize"] = RuntimeValue{Kind: FunctionVal, Value: NativeFunction{Name: "setSize",
			Call: func(cbArgs []RuntimeValue) RuntimeValue {
				if len(cbArgs) == 2 &&
					cbArgs[0].Kind == NumberVal && cbArgs[1].Kind == NumberVal {
					w.SetSize(int(cbArgs[0].Value.(float64)), int(cbArgs[1].Value.(float64)))
				}
				return RuntimeValue{Kind: NullVal}
			},
		}}
		obj["setAnchor"] = RuntimeValue{Kind: FunctionVal, Value: NativeFunction{Name: "setAnchor",
			Call: func(cbArgs []RuntimeValue) RuntimeValue {
				if len(cbArgs) == 2 &&
					cbArgs[0].Kind == NumberVal && cbArgs[1].Kind == NumberVal {
					w.SetAnchor(cbArgs[0].Value.(float64), cbArgs[1].Value.(float64))
				}
				return RuntimeValue{Kind: NullVal}
			},
		}}
	}

	// Helper: wrap a script function value as a Go func()
	wrapVoidFn := func(fn RuntimeValue) func() {
		cEnv := env; cCtx := ctx
		return func() {
			switch f := fn.Value.(type) {
			case NativeFunction:
				f.Call([]RuntimeValue{})
			case FunctionValue:
				CallFunctionValue(f, []RuntimeValue{}, cEnv, cCtx)
			}
		}
	}

	// Helper: wrap a script function value as a Go func(bool)
	wrapBoolFn := func(fn RuntimeValue) func(bool) {
		cEnv := env; cCtx := ctx
		return func(b bool) {
			arg := []RuntimeValue{{Kind: BoolVal, Value: b}}
			switch f := fn.Value.(type) {
			case NativeFunction:
				f.Call(arg)
			case FunctionValue:
				CallFunctionValue(f, arg, cEnv, cCtx)
			}
		}
	}

	// Helper: wrap a script function value as a Go func(float64)
	wrapNumFn := func(fn RuntimeValue) func(float64) {
		cEnv := env; cCtx := ctx
		return func(n float64) {
			arg := []RuntimeValue{{Kind: NumberVal, Value: n}}
			switch f := fn.Value.(type) {
			case NativeFunction:
				f.Call(arg)
			case FunctionValue:
				CallFunctionValue(f, arg, cEnv, cCtx)
			}
		}
	}

	imuiNS := map[string]RuntimeValue{
		"createWindow": {Kind: FunctionVal, Value: NativeFunction{
			Name: "createWindow",
			Call: func(args []RuntimeValue) RuntimeValue {

				assertArgCount("imui.createWindow", args, 4)

				title := args[0].Value.(string)
				width := int(args[1].Value.(float64))
				height := int(args[2].Value.(float64))
				fn := args[3]

				capturedEnv := env
				capturedCtx := ctx
		
				callback := func(hdc uintptr) {
		
					switch f := fn.Value.(type) {
		
					case NativeFunction:
						f.Call([]RuntimeValue{
							{Kind: NumberVal, Value: float64(hdc)},
						})
		
					case FunctionValue:
						CallFunctionValue(
							f,
							[]RuntimeValue{
								{Kind: NumberVal, Value: float64(hdc)},
							},
							capturedEnv,
							capturedCtx,
						)
					}
				}
		
				imui.CreateWindow(title, width, height, callback)
		
				return RuntimeValue{Kind: NullVal}
			},
		}},

		"setElement": {Kind: FunctionVal, Value: NativeFunction{
			Name: "setElement",
			Call: func(args []RuntimeValue) RuntimeValue {
		
				assertArgCount("imui.setElement", args, 3)
				assertKind("imui.setElement", args[0], StringVal)
				assertKind("imui.setElement", args[1], StringVal)
		
				id := args[0].Value.(string)
				field := args[1].Value.(string)
		
				imui.SetElement(id, field, args[2].Value)
		
				return RuntimeValue{Kind: NullVal}
			},
		}},

		"resetFrame": {Kind: FunctionVal, Value: NativeFunction{Name: "resetFrame", Call: func(args []RuntimeValue) RuntimeValue {
			assertArgCount("imui.resetFrame", args, 1)
			assertKind("imui.resetFrame", args[0], NumberVal)
			imui.ResetFrame(uintptr(args[0].Value.(float64)))
			return RuntimeValue{Kind: NullVal}
		}}},

		// -------------------------------------------------------------------
		// button(id, text) → {
		//   clicked bool, px, py,
		//   onClick(fn), onHover(fn), move(x,y),
		//   idleColor, hoverColor, pressColor, textColor, borderColor,
		//   borderThickness, cornerRadius, width, height  (assignable)
		// }
		// -------------------------------------------------------------------
		"button": {Kind: FunctionVal, Value: NativeFunction{
			Name: "button",
			Call: func(args []RuntimeValue) RuntimeValue {
				assertArgCount("imui.button", args, 2)
				assertKind("imui.button", args[0], StringVal)
				assertKind("imui.button", args[1], StringVal)

				id := args[0].Value.(string)
				widget := imui.GetOrCreateWidget(id, "button")

				obj, exists := widgetObjCache[id]
				if !exists {
					obj = map[string]RuntimeValue{}
					initWidgetObj(obj, widget)
					w := widget
					obj["clicked"] = RuntimeValue{Kind: BoolVal, Value: false}
					obj["onClick"] = RuntimeValue{Kind: FunctionVal, Value: NativeFunction{Name: "onClick",
						Call: func(cbArgs []RuntimeValue) RuntimeValue {
							if len(cbArgs) == 1 { w.OnClick = wrapVoidFn(cbArgs[0]) }
							return RuntimeValue{Kind: NullVal}
						},
					}}
					obj["onHover"] = RuntimeValue{Kind: FunctionVal, Value: NativeFunction{Name: "onHover",
						Call: func(cbArgs []RuntimeValue) RuntimeValue {
							if len(cbArgs) == 1 { w.OnHover = wrapVoidFn(cbArgs[0]) }
							return RuntimeValue{Kind: NullVal}
						},
					}}
					widgetObjCache[id] = obj
				}

				// Apply cached properties before render
				applyColorProp(widget, obj, "idleColor")
				applyColorProp(widget, obj, "hoverColor")
				applyColorProp(widget, obj, "pressColor")
				applyColorProp(widget, obj, "textColor")
				applyCommonDesignProps(widget, obj)
				if v, ok := obj["width"]; ok && v.Kind == NumberVal {
					widget.Width = int(v.Value.(float64))
					widget.HasWidth = true
				}
				if v, ok := obj["height"]; ok && v.Kind == NumberVal {
					widget.Height = int(v.Value.(float64))
					widget.HasHeight = true
				}

				widget, clicked := imui.Button(args[0].Value.(string), args[1].Value.(string))

				obj["clicked"] = RuntimeValue{Kind: BoolVal, Value: clicked}
				obj["px"] = RuntimeValue{Kind: NumberVal, Value: float64(widget.X)}
				obj["py"] = RuntimeValue{Kind: NumberVal, Value: float64(widget.Y)}

				return RuntimeValue{Kind: ObjectVal, Value: obj}
			},
		}},

		// -------------------------------------------------------------------
		// text(id, content) → {
		//   px, py, textColor, text, borderColor,
		//   borderThickness, cornerRadius  (assignable),
		//   move(x,y)
		// }
		// -------------------------------------------------------------------
		"text": {Kind: FunctionVal, Value: NativeFunction{
			Name: "text",
			Call: func(args []RuntimeValue) RuntimeValue {
				assertArgCount("imui.text", args, 2)
				assertKind("imui.text", args[0], StringVal)
				assertKind("imui.text", args[1], StringVal)

				id := args[0].Value.(string)
				widget := imui.GetOrCreateWidget(id, "text")

				obj, exists := widgetObjCache[id]
				if !exists {
					obj = map[string]RuntimeValue{}
					initWidgetObj(obj, widget)
					widgetObjCache[id] = obj
				}

				applyColorProp(widget, obj, "textColor")
				applyCommonDesignProps(widget, obj)
				if v, ok := obj["text"]; ok && v.Kind == StringVal {
					widget.OverrideText = v.Value.(string)
					widget.HasOverrideText = true
				}

				widget = imui.Text(args[0].Value.(string), args[1].Value.(string))

				obj["px"] = RuntimeValue{Kind: NumberVal, Value: float64(widget.X)}
				obj["py"] = RuntimeValue{Kind: NumberVal, Value: float64(widget.Y)}

				return RuntimeValue{Kind: ObjectVal, Value: obj}
			},
		}},

		// -------------------------------------------------------------------
		// checkbox(id, label) → {
		//   checked bool, px, py,
		//   onChange(fn(bool)), move(x,y),
		//   checkColor, borderColor, textColor, borderThickness,
		//   cornerRadius, size, label  (assignable)
		// }
		// -------------------------------------------------------------------
		"checkbox": {Kind: FunctionVal, Value: NativeFunction{
			Name: "checkbox",
			Call: func(args []RuntimeValue) RuntimeValue {
				assertArgCount("imui.checkbox", args, 2)
				assertKind("imui.checkbox", args[0], StringVal)
				assertKind("imui.checkbox", args[1], StringVal)

				id := args[0].Value.(string)
				widget := imui.GetOrCreateWidget(id, "checkbox")

				obj, exists := widgetObjCache[id]
				if !exists {
					obj = map[string]RuntimeValue{}
					initWidgetObj(obj, widget)
					w := widget
					obj["checked"] = RuntimeValue{Kind: BoolVal, Value: false}
					obj["onChange"] = RuntimeValue{Kind: FunctionVal, Value: NativeFunction{Name: "onChange",
						Call: func(cbArgs []RuntimeValue) RuntimeValue {
							if len(cbArgs) == 1 { w.OnChange = wrapBoolFn(cbArgs[0]) }
							return RuntimeValue{Kind: NullVal}
						},
					}}
					widgetObjCache[id] = obj
				}

				applyColorProp(widget, obj, "checkColor")
				applyColorProp(widget, obj, "borderColor")
				applyColorProp(widget, obj, "textColor")
				applyCommonDesignProps(widget, obj)
				if v, ok := obj["label"]; ok && v.Kind == StringVal {
					widget.Label = v.Value.(string)
					widget.HasLabel = true
				}
				if v, ok := obj["size"]; ok && v.Kind == NumberVal {
					widget.Width = int(v.Value.(float64))
					widget.HasWidth = true
				}

				widget, checked := imui.Checkbox(args[0].Value.(string), args[1].Value.(string))

				obj["checked"] = RuntimeValue{Kind: BoolVal, Value: checked}
				obj["px"] = RuntimeValue{Kind: NumberVal, Value: float64(widget.X)}
				obj["py"] = RuntimeValue{Kind: NumberVal, Value: float64(widget.Y)}

				return RuntimeValue{Kind: ObjectVal, Value: obj}
			},
		}},

		// -------------------------------------------------------------------
		// toggle(id, label) → {
		//   on bool, px, py,
		//   onChange(fn(bool)), move(x,y),
		//   onColor, offColor, knobColor, textColor,
		//   trackWidth, trackHeight, label  (assignable)
		// }
		// -------------------------------------------------------------------
		"toggle": {Kind: FunctionVal, Value: NativeFunction{
			Name: "toggle",
			Call: func(args []RuntimeValue) RuntimeValue {
				assertArgCount("imui.toggle", args, 2)
				assertKind("imui.toggle", args[0], StringVal)
				assertKind("imui.toggle", args[1], StringVal)

				id := args[0].Value.(string)
				widget := imui.GetOrCreateWidget(id, "toggle")

				obj, exists := widgetObjCache[id]
				if !exists {
					obj = map[string]RuntimeValue{}
					initWidgetObj(obj, widget)
					w := widget
					obj["on"] = RuntimeValue{Kind: BoolVal, Value: false}
					obj["onChange"] = RuntimeValue{Kind: FunctionVal, Value: NativeFunction{Name: "onChange",
						Call: func(cbArgs []RuntimeValue) RuntimeValue {
							if len(cbArgs) == 1 { w.OnChange = wrapBoolFn(cbArgs[0]) }
							return RuntimeValue{Kind: NullVal}
						},
					}}
					widgetObjCache[id] = obj
				}

				applyColorProp(widget, obj, "onColor")
				applyColorProp(widget, obj, "offColor")
				applyColorProp(widget, obj, "knobColor")
				applyColorProp(widget, obj, "textColor")
				if v, ok := obj["label"]; ok && v.Kind == StringVal {
					widget.Label = v.Value.(string)
					widget.HasLabel = true
				}
				if v, ok := obj["trackWidth"]; ok && v.Kind == NumberVal {
					widget.Width = int(v.Value.(float64))
					widget.HasWidth = true
				}
				if v, ok := obj["trackHeight"]; ok && v.Kind == NumberVal {
					widget.Height = int(v.Value.(float64))
					widget.HasHeight = true
				}

				widget, on := imui.Toggle(args[0].Value.(string), args[1].Value.(string))

				obj["on"] = RuntimeValue{Kind: BoolVal, Value: on}
				obj["px"] = RuntimeValue{Kind: NumberVal, Value: float64(widget.X)}
				obj["py"] = RuntimeValue{Kind: NumberVal, Value: float64(widget.Y)}

				return RuntimeValue{Kind: ObjectVal, Value: obj}
			},
		}},

		// -------------------------------------------------------------------
		// slider(id, min, max, value) → {
		//   value float, px, py,
		//   onChange(fn(float)), move(x,y),
		//   trackColor, handleColor,
		//   min, max, trackWidth, trackHeight, handleRadius  (assignable)
		// }
		// -------------------------------------------------------------------
		"slider": {Kind: FunctionVal, Value: NativeFunction{
			Name: "slider",
			Call: func(args []RuntimeValue) RuntimeValue {
				assertArgCount("imui.slider", args, 4)
				assertKind("imui.slider", args[0], StringVal)
				assertKind("imui.slider", args[1], NumberVal)
				assertKind("imui.slider", args[2], NumberVal)
				assertKind("imui.slider", args[3], NumberVal)

				id := args[0].Value.(string)
				widget := imui.GetOrCreateWidget(id, "slider")

				obj, exists := widgetObjCache[id]
				if !exists {
					obj = map[string]RuntimeValue{}
					initWidgetObj(obj, widget)
					w := widget
					obj["value"] = RuntimeValue{Kind: NumberVal, Value: args[3].Value.(float64)}
					obj["onChange"] = RuntimeValue{Kind: FunctionVal, Value: NativeFunction{Name: "onChange",
						Call: func(cbArgs []RuntimeValue) RuntimeValue {
							if len(cbArgs) == 1 { w.OnSlide = wrapNumFn(cbArgs[0]) }
							return RuntimeValue{Kind: NullVal}
						},
					}}
					widgetObjCache[id] = obj
				}

				applyColorProp(widget, obj, "trackColor")
				applyColorProp(widget, obj, "handleColor")
				if v, ok := obj["min"]; ok && v.Kind == NumberVal {
					widget.OverrideMin = v.Value.(float64)
					widget.HasOverrideMin = true
				}
				if v, ok := obj["max"]; ok && v.Kind == NumberVal {
					widget.OverrideMax = v.Value.(float64)
					widget.HasOverrideMax = true
				}
				if v, ok := obj["trackWidth"]; ok && v.Kind == NumberVal {
					widget.Width = int(v.Value.(float64))
					widget.HasWidth = true
				}
				if v, ok := obj["trackHeight"]; ok && v.Kind == NumberVal {
					widget.Height = int(v.Value.(float64))
					widget.HasHeight = true
				}

				widget, val := imui.Slider(
					args[0].Value.(string),
					args[1].Value.(float64),
					args[2].Value.(float64),
					args[3].Value.(float64),
				)

				obj["value"] = RuntimeValue{Kind: NumberVal, Value: val}
				obj["px"] = RuntimeValue{Kind: NumberVal, Value: float64(widget.X)}
				obj["py"] = RuntimeValue{Kind: NumberVal, Value: float64(widget.Y)}

				return RuntimeValue{Kind: ObjectVal, Value: obj}
			},
		}},

		// -------------------------------------------------------------------
		// frame(id, x, y, w, h) → {
		//   px, py,
		//   move(x,y),
		//   bgColor, borderColor, borderThickness, cornerRadius,
		//   padding  (assignable)
		// }
		// Must be paired with imui.endFrame(id).
		// -------------------------------------------------------------------
		"frame": {Kind: FunctionVal, Value: NativeFunction{
			Name: "frame",
			Call: func(args []RuntimeValue) RuntimeValue {
				assertArgCount("imui.frame", args, 5)
				assertKind("imui.frame", args[0], StringVal)
				assertKind("imui.frame", args[1], NumberVal)
				assertKind("imui.frame", args[2], NumberVal)
				assertKind("imui.frame", args[3], NumberVal)
				assertKind("imui.frame", args[4], NumberVal)

				id := args[0].Value.(string)
				widget := imui.GetOrCreateWidget(id, "frame")

				obj, exists := widgetObjCache[id]
				if !exists {
					obj = map[string]RuntimeValue{}
					initWidgetObj(obj, widget)
					widgetObjCache[id] = obj
				}

				applyColorProp(widget, obj, "bgColor")
				applyColorProp(widget, obj, "borderColor")
				applyColorProp(widget, obj, "frameBorderColor")
				applyCommonDesignProps(widget, obj)
				if v, ok := obj["padding"]; ok && v.Kind == NumberVal {
					widget.Padding = int(v.Value.(float64))
					widget.HasPadding = true
				}

				widget = imui.Frame(
					args[0].Value.(string),
					int(args[1].Value.(float64)),
					int(args[2].Value.(float64)),
					int(args[3].Value.(float64)),
					int(args[4].Value.(float64)),
				)

				obj["px"] = RuntimeValue{Kind: NumberVal, Value: float64(widget.X)}
				obj["py"] = RuntimeValue{Kind: NumberVal, Value: float64(widget.Y)}

				return RuntimeValue{Kind: ObjectVal, Value: obj}
			},
		}},

		"endFrame": {Kind: FunctionVal, Value: NativeFunction{
			Name: "endFrame",
			Call: func(args []RuntimeValue) RuntimeValue {
				assertArgCount("imui.endFrame", args, 1)
				assertKind("imui.endFrame", args[0], StringVal)
				imui.EndFrame(args[0].Value.(string))
				return RuntimeValue{Kind: NullVal}
			},
		}},
	}
	env.Set("imui", RuntimeValue{Kind: ObjectVal, Value: imuiNS})
}

func assertArgCount(name string, args []RuntimeValue, n int) {
	if len(args) != n {
		panic(TunaError(fmt.Sprintf("%s() expects %d argument(s) but got %d", name, n, len(args))))
	}
}

func assertKind(name string, val RuntimeValue, expected ValueKind) {
	if val.Kind != expected {
		panic(TunaError(fmt.Sprintf("%s() expects a %s but got '%s'", name, expected, val.Kind)))
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
	case ArrayVal, ObjectVal:
		return runtimeDeepEqual(a, b)
	default:
		return false
	}
}

func runtimeToNative(val RuntimeValue) any {
	switch val.Kind {
	case NumberVal:
		 return val.Value.(float64)
	case StringVal:
		 return val.Value.(string)
	case BoolVal:
		 return val.Value.(bool)
	case NullVal:
		 return nil
	case ArrayVal:
		 arr := val.Value.([]RuntimeValue)
		 result := make([]any, len(arr))
		 for i, v := range arr {
			  result[i] = runtimeToNative(v)
		 }
		 return result
	case ObjectVal:
		 props := val.Value.(map[string]RuntimeValue)
		 result := map[string]any{}
		 for k, v := range props {
			  result[k] = runtimeToNative(v)
		 }
		 return result
	default:
		 return nil
	}
}

func nativeToRuntime(val any) RuntimeValue {
	if val == nil {
		 return RuntimeValue{Kind: NullVal}
	}
	switch v := val.(type) {
	case float64:
		 return RuntimeValue{Kind: NumberVal, Value: v}
	case string:
		 return RuntimeValue{Kind: StringVal, Value: v}
	case bool:
		 return RuntimeValue{Kind: BoolVal, Value: v}
	case []any:
		 arr := make([]RuntimeValue, len(v))
		 for i, el := range v {
			  arr[i] = nativeToRuntime(el)
		 }
		 return RuntimeValue{Kind: ArrayVal, Value: arr}
	case map[string]any:
		 props := map[string]RuntimeValue{}
		 for k, el := range v {
			  props[k] = nativeToRuntime(el)
		 }
		 return RuntimeValue{Kind: ObjectVal, Value: props}
	default:
		 return RuntimeValue{Kind: NullVal}
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
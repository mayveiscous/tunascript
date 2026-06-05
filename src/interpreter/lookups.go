package interpreter

const (
	NumberVal ValueKind = iota
	StringVal
	BoolVal
	NullVal
	FunctionVal
	ArrayVal
	ObjectVal
)

const (
	sigNone SignalKind = iota
	sigReturn
	sigBreak
	sigContinue
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


var NullResult = EvalResult{Value: RuntimeValue{Kind: NullVal}}
package interpreter

const (
	ValNumber	ValueKind	= iota
	ValString
	ValBool
	ValNull
	ValFunction
	ValArray
	ValObject
)

const (
	NumberVal	= ValNumber
	StringVal	= ValString
	BoolVal		= ValBool
	NullVal		= ValNull
	FunctionVal	= ValFunction
	ArrayVal	= ValArray
	ObjectVal	= ValObject
)

const (
	SigNone	SignalKind	= iota
	SigReturn
	SigBreak
	SigContinue
)

const (
	sigNone		= SigNone
	sigReturn	= SigReturn
	sigBreak	= SigBreak
	sigContinue	= SigContinue
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

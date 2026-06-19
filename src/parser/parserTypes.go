package parser

import "tunascript/src/lexer"

type parser struct {
	tokens   []lexer.Token
	pos      int
	filePath string
}

type Statement interface{ statement() }
type Expression interface{ expression() }
type AstType interface{ astType() }

type SwapStatement struct {
	Targets []Expression
	Values  []Expression
}

type BindingPower int

// Handler types
type statementHandler func(p *parser) Statement
type nudHandler        func(p *parser) Expression
type ledHandler        func(p *parser, left Expression, bp BindingPower) Expression
type typeNudHandler    func(p *parser) AstType
type typeLedHandler    func(p *parser, left AstType, bp BindingPower) AstType

// Literals
type NumberExpression struct{ Value float64 }
type StringExpression struct{ Value string }
type SymbolExpression struct {
	Token lexer.Token
	Value string
}
type BoolExpression    struct{ Value bool }

type ArrayLiteral struct{ Elements []Expression }
type ObjectLiteral struct{ Properties []ObjectProperty }

// Expressions
type ObjectProperty struct {
	Key   string
	Value Expression
}

type MemberExpression struct {
	Object   Expression
	Property string
}

type IndexExpression struct {
	Left  Expression
	Index Expression
}

type BinaryExpression struct {
	Left     Expression
	Operator lexer.Token
	Right    Expression
}

type PrefixExpression struct {
	Operator        lexer.Token
	RightExpression Expression
}

type PostfixExpression struct {
	Operator lexer.Token
	Left     Expression
}

type AssignmentExpression struct {
	Assignee Expression
	Operator lexer.Token
	Value    Expression
}

type CallExpression struct {
	Callee    Expression
	Arguments []Expression
}

type TypeofExpression struct {
	Expr Expression
}

// Statements
type BlockStatement struct{ Body []Statement }
type ExpressionStatement struct{ Expression Expression }
type ReturnStatement struct {
	Token lexer.Token
	Value Expression
}
type BreakStatement struct{ Token lexer.Token }
type ContinueStatement struct{ Token lexer.Token }

type VariableDecStatement struct {
	Token         lexer.Token
	VariableName  string
	IsConstant    bool
	AssignedValue Expression
	ExplicitType  AstType
}

type FunctionParameter struct {
	Token      lexer.Token
	Name       string
	Type       AstType
	IsVariadic bool
}

type FunctionDecStatement struct {
	Token      lexer.Token
	Name       string
	Parameters []FunctionParameter
	ReturnType AstType
	Body       BlockStatement
}

type FunctionExpression struct {
	Parameters []FunctionParameter
	ReturnType AstType
	Body       BlockStatement
}

type IfStatement struct {
	Condition Expression
	Then      BlockStatement
	Else      *BlockStatement
}

type WhileStatement struct {
	Condition Expression
	Body      BlockStatement
}

type ForInStatement struct {
	KeyVar   string
	Iterator string
	Iterable Expression
	Body     BlockStatement
}

type ImportItem struct {
	Token lexer.Token
	Name  string
	Alias string
}

type TryStatement struct {
	Body    BlockStatement
	ErrName string
	Hook    BlockStatement
}

type ImportStatement struct {
	Token lexer.Token
	Path  string
	Items []ImportItem
}

type CastStatement struct {
	Inner Statement
}

type SymbolType struct{ Name string }
type ArrayType  struct{ Underlying AstType }

type SchoolFieldDef struct {
	Name string
	Type AstType
}

type SchoolStatement struct {
	Name   string
	Fields []SchoolFieldDef
}
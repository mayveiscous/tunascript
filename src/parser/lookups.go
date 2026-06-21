package parser

import "tunascript/src/lexer"

const (
	defaultBp	BindingPower	= iota
	commaBp
	assignmentBp
	logicalBp
	relationalBp
	additiveBp
	multiplicativeBp
	unaryBp
	callBp
	memberBp
	primaryBp
)

var (
	statementLu	= map[lexer.TokenKind]statementHandler{}
	bpLu		= map[lexer.TokenKind]BindingPower{}
	nudLu		= map[lexer.TokenKind]nudHandler{}
	ledLu		= map[lexer.TokenKind]ledHandler{}
	typeBpLu	= map[lexer.TokenKind]BindingPower{}
	typeNudLu	= map[lexer.TokenKind]typeNudHandler{}
	typeLedLu	= map[lexer.TokenKind]typeLedHandler{}
)

func (e NumberExpression) expression()		{}
func (e StringExpression) expression()		{}
func (e SymbolExpression) expression()		{}
func (e BoolExpression) expression()		{}
func (e ArrayLiteral) expression()		{}
func (e ObjectLiteral) expression()		{}
func (e MemberExpression) expression()		{}
func (e IndexExpression) expression()		{}
func (e BinaryExpression) expression()		{}
func (e PrefixExpression) expression()		{}
func (e PostfixExpression) expression()		{}
func (e AssignmentExpression) expression()	{}
func (e CallExpression) expression()		{}
func (e TypeofExpression) expression()		{}
func (s SwapStatement) statement()		{}
func (s TryStatement) statement()		{}
func (s SchoolStatement) statement()		{}

func (e FunctionExpression) expression()	{}
func (s ExpressionStatement) statement()	{}
func (s BlockStatement) statement()		{}
func (s ReturnStatement) statement()		{}
func (s BreakStatement) statement()		{}
func (s ContinueStatement) statement()		{}
func (s VariableDecStatement) statement()	{}
func (s FunctionDecStatement) statement()	{}
func (s IfStatement) statement()		{}
func (s WhileStatement) statement()		{}
func (s ForInStatement) statement()		{}
func (s ImportStatement) statement()		{}
func (s CastStatement) statement()		{}

func (t SymbolType) astType()	{}
func (t ArrayType) astType()	{}

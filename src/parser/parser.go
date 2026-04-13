package parser

import (
	"fmt"
	"strconv"
	"test-go/src/lexer"
)

type Statement interface{ statement() }
type Expression interface{ expression() }
type AstType interface{ _type() }

type NumberExpression struct{ Value float64 }
type StringExpression struct{ Value string }
type SymbolExpression struct{ Value string }

type BinaryExpression struct {
	Left     Expression
	Operator lexer.Token
	Right    Expression
}

type IndexExpression struct {
	Left  Expression
	Index Expression
}

type PrefixExpression struct {
	Operator        lexer.Token
	RightExpression Expression
}

type AssignmentExpression struct {
	Assigne  Expression
	Operator lexer.Token
	Value    Expression
}

type CallExpression struct {
	Callee    Expression
	Arguments []Expression
}

type ArrayLiteral struct{ Elements []Expression }

type ObjectProperty struct {
	Key   string
	Value Expression
}

type ObjectLiteral struct{ Properties []ObjectProperty }

type MemberExpression struct {
	Object   Expression
	Property string
}

type PostfixExpression struct {
	Operator lexer.Token
	Left     Expression
}

type TypeofExpression struct {
	Expr Expression
}

func (n ArrayLiteral) expression()         {}
func (n ObjectLiteral) expression()        {}
func (n MemberExpression) expression()     {}
func (n CallExpression) expression()       {}
func (n NumberExpression) expression()     {}
func (n StringExpression) expression()     {}
func (n SymbolExpression) expression()     {}
func (n BinaryExpression) expression()     {}
func (n PrefixExpression) expression()     {}
func (n AssignmentExpression) expression() {}
func (n IndexExpression) expression()      {}
func (n PostfixExpression) expression()    {}
func (n TypeofExpression) expression()     {}

type BlockStatement struct{ Body []Statement }
type BoolExpression struct{ Value bool }
type ReturnStatement struct{ Value Expression }
type ExpressionStatement struct{ Expression Expression }

type VariableDecStatement struct {
	VariableName  string
	IsConstant    bool
	AssignedValue Expression
	ExplicitType  AstType
}

type FunctionParameter struct {
	Name string
	Type AstType
}

type FunctionDecStatement struct {
	Name       string
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
	Iterator string
	Iterable Expression
	Body     BlockStatement
}

type BreakStatement struct{}
type ContinueStatement struct{}

type ImportItem struct {
	Name  string
	Alias string
}

type ImportStatement struct {
	Path  string
	Items []ImportItem
}

type CastStatement struct {
	Inner Statement
}

func (n BreakStatement) statement()    {}
func (n ContinueStatement) statement() {}
func (n ImportStatement) statement()   {}
func (n CastStatement) statement()     {}
func (n BoolExpression) expression()    {}
func (n ReturnStatement) statement()    {}
func (n BlockStatement) statement()     {}
func (n ExpressionStatement) statement()  {}
func (n VariableDecStatement) statement() {}
func (n FunctionDecStatement) statement() {}
func (n IfStatement) statement()          {}
func (n WhileStatement) statement()       {}
func (n ForInStatement) statement()       {}

type SymbolType struct{ Name string }
type ArrayType struct{ Underlying AstType }

func (t SymbolType) _type() {}
func (t ArrayType) _type()  {}

type BindingPower int

const (
	default_bp BindingPower = iota
	comma_bp
	assignment_bp
	logical_bp
	relational_bp
	additive_bp
	multiplicative_bp
	unary_bp
	call_bp
	member_bp
	primary_bp
)

type statement_handler func(p *parser) Statement
type nud_handler func(p *parser) Expression
type led_handler func(p *parser, left Expression, bp BindingPower) Expression
type type_nud_handler func(p *parser) AstType
type type_led_handler func(p *parser, left AstType, bp BindingPower) AstType

var statement_lu = map[lexer.TokenKind]statement_handler{}
var bp_lu = map[lexer.TokenKind]BindingPower{}
var nud_lu = map[lexer.TokenKind]nud_handler{}
var led_lu = map[lexer.TokenKind]led_handler{}
var type_bp_lu = map[lexer.TokenKind]BindingPower{}
var type_nud_lu = map[lexer.TokenKind]type_nud_handler{}
var type_led_lu = map[lexer.TokenKind]type_led_handler{}

func led_reg(kind lexer.TokenKind, bp BindingPower, fn led_handler) {
	bp_lu[kind] = bp
	led_lu[kind] = fn
}

func nud_reg(kind lexer.TokenKind, fn nud_handler) {
	nud_lu[kind] = fn
}

func statement_reg(kind lexer.TokenKind, fn statement_handler) {
	bp_lu[kind] = default_bp
	statement_lu[kind] = fn
}

func type_nud_reg(kind lexer.TokenKind, fn type_nud_handler) {
	type_nud_lu[kind] = fn
}

func createTokenLookups() {
	led_reg(lexer.ASSIGNMENT, assignment_bp, parse_assign_expression)
	led_reg(lexer.PLUS_EQUALS, assignment_bp, parse_assign_expression)
	led_reg(lexer.MINUS_EQUALS, assignment_bp, parse_assign_expression)
	led_reg(lexer.STAR_EQUALS, assignment_bp, parse_assign_expression)
	led_reg(lexer.SLASH_EQUALS, assignment_bp, parse_assign_expression)

	led_reg(lexer.AND, logical_bp, parse_binary_expression)
	led_reg(lexer.OR, logical_bp, parse_binary_expression)

	led_reg(lexer.LESS, relational_bp, parse_binary_expression)
	led_reg(lexer.LESS_EQUALS, relational_bp, parse_binary_expression)
	led_reg(lexer.GREATER, relational_bp, parse_binary_expression)
	led_reg(lexer.GREATER_EQUALS, relational_bp, parse_binary_expression)
	led_reg(lexer.EQUALS, relational_bp, parse_binary_expression)
	led_reg(lexer.NOT_EQUALS, relational_bp, parse_binary_expression)

	led_reg(lexer.PLUS, additive_bp, parse_binary_expression)
	led_reg(lexer.DASH, additive_bp, parse_binary_expression)

	led_reg(lexer.STAR, multiplicative_bp, parse_binary_expression)
	led_reg(lexer.SLASH, multiplicative_bp, parse_binary_expression)
	led_reg(lexer.PERCENT, multiplicative_bp, parse_binary_expression)

	led_reg(lexer.OPEN_PAREN, call_bp, parse_call_expression)

	led_reg(lexer.OPEN_BRACKET, member_bp, parse_index_expression)
	led_reg(lexer.DOT, member_bp, parse_member_expression)

	led_reg(lexer.PLUS_PLUS, unary_bp, parse_postfix_expression)
	led_reg(lexer.MINUS_MINUS, unary_bp, parse_postfix_expression)

	nud_reg(lexer.NUMBER, parse_primary_expression)
	nud_reg(lexer.STRING, parse_primary_expression)
	nud_reg(lexer.IDENT, parse_primary_expression)
	nud_reg(lexer.OPEN_PAREN, parse_grouping_expression)
	nud_reg(lexer.DASH, parse_prefix_expression)
	nud_reg(lexer.NOT, parse_prefix_expression)
	nud_reg(lexer.OPEN_BRACKET, parse_array_literal)
	nud_reg(lexer.OPEN_CURLY, parse_object_literal)
	nud_reg(lexer.TRUE, parse_primary_expression)
	nud_reg(lexer.FALSE, parse_primary_expression)
	nud_reg(lexer.NULL, parse_primary_expression)
	nud_reg(lexer.TYPEOF, parse_typeof_expression)

	statement_reg(lexer.CONST, parse_var_declaration_statement)
	statement_reg(lexer.LET, parse_var_declaration_statement)
	statement_reg(lexer.SWIM, parse_function_dec_statement)
	statement_reg(lexer.RETURN, parse_return_statement)
	statement_reg(lexer.IF, parse_if_statement)
	statement_reg(lexer.WHILE, parse_while_statement)
	statement_reg(lexer.FOR, parse_for_in_statement)
	statement_reg(lexer.BREAK, parse_break_statement)
	statement_reg(lexer.CONTINUE, parse_continue_statement)
	statement_reg(lexer.FROM, parse_import_statement)
	statement_reg(lexer.CAST, parse_cast_statement)

	type_nud_reg(lexer.IDENT, parse_symbol_type)
	type_nud_reg(lexer.OPEN_BRACKET, parse_array_type)
	type_nud_reg(lexer.FUNCTION, parse_function_type)
}

type parser struct {
	tokens []lexer.Token
	pos    int
}

func (p *parser) currentToken() lexer.Token         { return p.tokens[p.pos] }
func (p *parser) currentTokenKind() lexer.TokenKind { return p.currentToken().Kind }

func (p *parser) advance() lexer.Token {
	token := p.currentToken()
	p.pos++
	return token
}

func (p *parser) hasTokens() bool {
	return p.pos < len(p.tokens) && p.currentTokenKind() != lexer.EOF
}

func (p *parser) parseError(msg string) *lexer.TunaError {
	t := p.currentToken()
	return lexer.NewError(t.Line, t.Column, msg)
}

func (p *parser) expectError(expectedKind lexer.TokenKind, err any) lexer.Token {
	token := p.currentToken()
	if token.Kind != expectedKind {
		if err == nil {
			panic(lexer.NewError(token.Line, token.Column,
				fmt.Sprintf("expected '%s' but got '%s'",
					lexer.TokenKindString(expectedKind), token.Value)))
		}
		panic(err)
	}
	return p.advance()
}

func (p *parser) expect(expectedKind lexer.TokenKind) lexer.Token {
	return p.expectError(expectedKind, nil)
}

func createParser(tokens []lexer.Token) *parser {
	createTokenLookups()
	return &parser{tokens: tokens, pos: 0}
}

func Parse(tokens []lexer.Token) BlockStatement {
	body := make([]Statement, 0)
	p := createParser(tokens)
	for p.hasTokens() {
		body = append(body, parse_statement(p))
	}
	return BlockStatement{Body: body}
}

func parse_statement(p *parser) Statement {
	statement_fn, exists := statement_lu[p.currentTokenKind()]
	var stmt Statement
	if exists {
		stmt = statement_fn(p)
	} else {
		stmt = ExpressionStatement{Expression: parse_expression(p, default_bp)}
	}
	consumeSemicolon(p)
	return stmt
}

func consumeSemicolon(p *parser) {
	if p.currentTokenKind() == lexer.SEMICOLON {
		p.advance()
	}
}

func parse_break_statement(p *parser) Statement {
    p.advance()
    return BreakStatement{}
}
func parse_continue_statement(p *parser) Statement {
    p.advance()
    return ContinueStatement{}
}

func parse_index_expression(p *parser, left Expression, bp BindingPower) Expression {
	p.advance()
	index := parse_expression(p, default_bp)
	p.expect(lexer.CLOSE_BRACKET)
	return IndexExpression{Left: left, Index: index}
}

func parse_body(p *parser) BlockStatement {
	body := []Statement{}
	for p.hasTokens() && p.currentTokenKind() != lexer.SHORE {
		body = append(body, parse_statement(p))
	}
	p.expect(lexer.SHORE)
	return BlockStatement{Body: body}
}

func parse_if_body(p *parser) BlockStatement {
	body := []Statement{}
	for p.hasTokens() &&
		p.currentTokenKind() != lexer.SHORE &&
		p.currentTokenKind() != lexer.ELSE {
		body = append(body, parse_statement(p))
	}
	return BlockStatement{Body: body}
}

func parse_if_statement(p *parser) Statement {
	p.advance()
	condition := parse_expression(p, default_bp)
	then := parse_if_body(p)

	if p.currentTokenKind() == lexer.ELSE {
		p.advance()
		if p.currentTokenKind() == lexer.IF {
			elseIf := parse_if_statement(p)
			elseBlock := BlockStatement{Body: []Statement{elseIf}}
			return IfStatement{Condition: condition, Then: then, Else: &elseBlock}
		}
		elseBlock := parse_if_body(p)
		p.expect(lexer.SHORE)
		return IfStatement{Condition: condition, Then: then, Else: &elseBlock}
	}

	p.expect(lexer.SHORE)
	return IfStatement{Condition: condition, Then: then, Else: nil}
}

func parse_while_statement(p *parser) Statement {
	p.advance()
	condition := parse_expression(p, default_bp)
	body := parse_body(p)
	return WhileStatement{Condition: condition, Body: body}
}

func parse_for_in_statement(p *parser) Statement {
	p.advance()
	iterator := p.expectError(lexer.IDENT, p.parseError("expected iterator name after 'for'")).Value
	p.expectError(lexer.IN, p.parseError("expected 'in' after iterator name"))
	iterable := parse_expression(p, default_bp)
	body := parse_body(p)
	return ForInStatement{Iterator: iterator, Iterable: iterable, Body: body}
}

func parse_function_parameters(p *parser) []FunctionParameter {
	params := []FunctionParameter{}
	p.expect(lexer.OPEN_PAREN)
	for p.hasTokens() && p.currentTokenKind() != lexer.CLOSE_PAREN {
		name := p.expectError(lexer.IDENT, p.parseError("expected parameter name")).Value
		p.expectError(lexer.COLON, p.parseError(fmt.Sprintf("expected ':' after parameter '%s'", name)))
		paramType := parse_type(p, default_bp)
		params = append(params, FunctionParameter{Name: name, Type: paramType})
		if p.currentTokenKind() == lexer.COMMA {
			p.advance()
		}
	}
	p.expect(lexer.CLOSE_PAREN)
	return params
}

func parse_function_dec_statement(p *parser) Statement {
	p.advance()
	name := p.expectError(lexer.IDENT, p.parseError("expected function name after 'swim'")).Value
	params := parse_function_parameters(p)
	p.expectError(lexer.COLON, p.parseError(fmt.Sprintf("expected ':' after parameters of function '%s'", name)))
	returnType := parse_type(p, default_bp)
	body := parse_body(p)
	return FunctionDecStatement{Name: name, Parameters: params, ReturnType: returnType, Body: body}
}

func parse_return_statement(p *parser) Statement {
	p.advance()
	return ReturnStatement{Value: parse_expression(p, default_bp)}
}

func parse_var_declaration_statement(p *parser) Statement {
	var explicitType AstType
	var assignedValue Expression

	keyword := p.advance()
	isConst := keyword.Kind == lexer.CONST
	varName := p.expectError(lexer.IDENT, p.parseError("expected variable name")).Value

	if p.currentTokenKind() == lexer.COLON {
		p.advance()
		explicitType = parse_type(p, default_bp)
	}

	if p.currentTokenKind() == lexer.ASSIGNMENT {
		p.advance()
		assignedValue = parse_expression(p, assignment_bp)
	} else if explicitType == nil {
		panic(lexer.NewError(keyword.Line, keyword.Column,
			fmt.Sprintf("variable '%s' must have a type annotation or an assigned value", varName)))
	}

	if isConst && assignedValue == nil {
		panic(lexer.NewError(keyword.Line, keyword.Column,
			fmt.Sprintf("constant '%s' must be assigned a value", varName)))
	}

	consumeSemicolon(p)
	return VariableDecStatement{
		IsConstant:    isConst,
		VariableName:  varName,
		AssignedValue: assignedValue,
		ExplicitType:  explicitType,
	}
}

func parse_expression(p *parser, bp BindingPower) Expression {
	tokenKind := p.currentTokenKind()
	nud_fn, exists := nud_lu[tokenKind]
	if !exists {
		panic(p.parseError(fmt.Sprintf("unexpected token '%s'", p.currentToken().Value)))
	}
	left := nud_fn(p)

	for bp_lu[p.currentTokenKind()] > bp {
		tokenKind = p.currentTokenKind()
		led_fn, exists := led_lu[tokenKind]
		if !exists {
			panic(p.parseError(fmt.Sprintf("unexpected token '%s'", p.currentToken().Value)))
		}
		left = led_fn(p, left, bp_lu[p.currentTokenKind()])
	}
	return left
}

func parse_primary_expression(p *parser) Expression {
	switch p.currentTokenKind() {
	case lexer.NUMBER:
		number, _ := strconv.ParseFloat(p.advance().Value, 64)
		return NumberExpression{Value: number}
	case lexer.TRUE:
		p.advance()
		return BoolExpression{Value: true}
	case lexer.NULL:
		p.advance()
		return SymbolExpression{Value: "nil"}
	case lexer.FALSE:
		p.advance()
		return BoolExpression{Value: false}
	case lexer.STRING:
		return StringExpression{Value: p.advance().Value}
	case lexer.IDENT:
		return SymbolExpression{Value: p.advance().Value}
	default:
		panic(p.parseError(fmt.Sprintf("unexpected token '%s'", p.currentToken().Value)))
	}
}

func parse_array_literal(p *parser) Expression {
	p.advance()
	elements := []Expression{}
	for p.hasTokens() && p.currentTokenKind() != lexer.CLOSE_BRACKET {
		elements = append(elements, parse_expression(p, assignment_bp))
		if p.currentTokenKind() == lexer.COMMA {
			p.advance()
		}
	}
	p.expect(lexer.CLOSE_BRACKET)
	return ArrayLiteral{Elements: elements}
}

func parse_binary_expression(p *parser, left Expression, bp BindingPower) Expression {
	operatorToken := p.advance()
	right := parse_expression(p, bp)
	return BinaryExpression{Left: left, Operator: operatorToken, Right: right}
}

func parse_assign_expression(p *parser, left Expression, bp BindingPower) Expression {
	operatorToken := p.advance()
	rhs := parse_expression(p, bp)
	return AssignmentExpression{Operator: operatorToken, Value: rhs, Assigne: left}
}

func parse_grouping_expression(p *parser) Expression {
	p.advance()
	exp := parse_expression(p, default_bp)
	p.expect(lexer.CLOSE_PAREN)
	return exp
}

func parse_prefix_expression(p *parser) Expression {
	operatorToken := p.advance()
	rhs := parse_expression(p, default_bp)
	return PrefixExpression{Operator: operatorToken, RightExpression: rhs}
}

func parse_call_expression(p *parser, left Expression, bp BindingPower) Expression {
	p.advance()
	args := []Expression{}
	for p.hasTokens() && p.currentTokenKind() != lexer.CLOSE_PAREN {
		args = append(args, parse_expression(p, assignment_bp))
		if p.currentTokenKind() == lexer.COMMA {
			p.advance()
		}
	}
	p.expect(lexer.CLOSE_PAREN)
	return CallExpression{Callee: left, Arguments: args}
}

func parse_type(p *parser, bp BindingPower) AstType {
	tokenKind := p.currentTokenKind()
	nud_fn, exists := type_nud_lu[tokenKind]
	if !exists {
		panic(p.parseError(fmt.Sprintf("expected a type but got '%s'", p.currentToken().Value)))
	}
	left := nud_fn(p)

	for type_bp_lu[p.currentTokenKind()] > bp {
		tokenKind = p.currentTokenKind()
		led_fn, exists := type_led_lu[tokenKind]
		if !exists {
			panic(p.parseError(fmt.Sprintf("unexpected token in type expression '%s'", p.currentToken().Value)))
		}
		left = led_fn(p, left, type_bp_lu[p.currentTokenKind()])
	}
	return left
}

func parse_symbol_type(p *parser) AstType {
	tok := p.expect(lexer.IDENT)
	name := tok.Value
	validTypes := map[string]bool{
		 "number": true, "string": true, "bool": true,
		 "function": true, "void": true, "null": true, "array": true, "object": true,
	}
	if !validTypes[name] {
		 panic(lexer.NewError(tok.Line, tok.Column,
			  fmt.Sprintf("unknown type '%s': valid types are number, string, bool, function, void, null, array", name)))
	}
	return SymbolType{Name: name}
}

func parse_array_type(p *parser) AstType {
	p.advance()
	p.expect(lexer.CLOSE_BRACKET)
	return ArrayType{Underlying: parse_type(p, default_bp)}
}

func parse_postfix_expression(p *parser, left Expression, bp BindingPower) Expression {
	operatorToken := p.advance()
	return PostfixExpression{Operator: operatorToken, Left: left}
}

func parse_function_type(p *parser) AstType {
	p.advance()
	return SymbolType{Name: "function"}
}

func parse_typeof_expression(p *parser) Expression {
	p.advance()
	expr := parse_expression(p, unary_bp)
	return TypeofExpression{Expr: expr}
}

func parse_object_literal(p *parser) Expression {
	p.advance()
	properties := []ObjectProperty{}
	for p.hasTokens() && p.currentTokenKind() != lexer.CLOSE_CURLY {
		key := p.expectError(lexer.IDENT, p.parseError("expected property name in object literal")).Value
		p.expectError(lexer.COLON, p.parseError(fmt.Sprintf("expected ':' after property '%s'", key)))
		value := parse_expression(p, assignment_bp)
		properties = append(properties, ObjectProperty{Key: key, Value: value})
		if p.currentTokenKind() == lexer.COMMA {
			p.advance()
		}
	}
	p.expect(lexer.CLOSE_CURLY)
	return ObjectLiteral{Properties: properties}
}

func parse_member_expression(p *parser, left Expression, bp BindingPower) Expression {
	p.advance()
	prop := p.expectError(lexer.IDENT, p.parseError("expected property name after '.'")).Value
	return MemberExpression{Object: left, Property: prop}
}

func parse_import_statement(p *parser) Statement {
	p.advance()
	if p.currentTokenKind() != lexer.STRING {
		panic(p.parseError("expected a file path string after 'from'"))
	}
	path := p.advance().Value

	p.expectError(lexer.LET, p.parseError("expected 'catch' after import path"))

	items := []ImportItem{}
	for {
		name := p.expectError(lexer.IDENT, p.parseError("expected export name")).Value
		alias := name
		if p.currentTokenKind() == lexer.AS {
			p.advance()
			alias = p.expectError(lexer.IDENT, p.parseError("expected alias name after 'as'")).Value
		}
		items = append(items, ImportItem{Name: name, Alias: alias})
		if p.currentTokenKind() == lexer.COMMA {
			p.advance()
		} else {
			break
		}
	}
	return ImportStatement{Path: path, Items: items}
}

func parse_cast_statement(p *parser) Statement {
	p.advance()
	tok := p.currentToken()
	var inner Statement
	switch tok.Kind {
	case lexer.SWIM:
		inner = parse_function_dec_statement(p)
	case lexer.LET, lexer.CONST:
		inner = parse_var_declaration_statement(p)
	default:
		panic(p.parseError(fmt.Sprintf("'cast' must be followed by 'swim', 'catch', or 'anchor', got '%s'", tok.Value)))
	}
	return CastStatement{Inner: inner}
}
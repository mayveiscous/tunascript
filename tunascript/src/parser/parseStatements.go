package parser

import (
	"fmt"
	"strconv"
	"tunascript/src/lexer"
)

func parseStatement(p *parser) Statement {
	statementFn, exists := statementLu[p.currentTokenKind()]
	var stmt Statement
	if exists {
		stmt = statementFn(p)
	} else {
		stmt = ExpressionStatement{Expression: parseExpression(p, defaultBp)}
	}
	consumeSemicolon(p)
	return stmt
}

func parseBreakStatement(p *parser) Statement {
	p.advance()
	return BreakStatement{}
}

func parseContinueStatement(p *parser) Statement {
	p.advance()
	return ContinueStatement{}
}

func parseIndexExpression(p *parser, left Expression, bp BindingPower) Expression {
	p.advance()
	index := parseExpression(p, defaultBp)
	p.expect(lexer.CLOSE_BRACKET)
	return IndexExpression{Left: left, Index: index}
}

func parseBody(p *parser) BlockStatement {
	body := []Statement{}
	for p.hasTokens() && p.currentTokenKind() != lexer.SHORE {
		body = append(body, parseStatement(p))
	}
	p.expect(lexer.SHORE)
	return BlockStatement{Body: body}
}

func parseIfBody(p *parser) BlockStatement {
	body := []Statement{}
	for p.hasTokens() &&
		p.currentTokenKind() != lexer.SHORE &&
		p.currentTokenKind() != lexer.ELSE {
		body = append(body, parseStatement(p))
	}
	return BlockStatement{Body: body}
}

func parseIfStatement(p *parser) Statement {
	p.advance()
	condition := parseExpression(p, defaultBp)
	then := parseIfBody(p)

	if p.currentTokenKind() == lexer.ELSE {
		p.advance()
		if p.currentTokenKind() == lexer.IF {
			elseIf := parseIfStatement(p)
			elseBlock := BlockStatement{Body: []Statement{elseIf}}
			return IfStatement{Condition: condition, Then: then, Else: &elseBlock}
		}
		elseBlock := parseIfBody(p)
		p.expect(lexer.SHORE)
		return IfStatement{Condition: condition, Then: then, Else: &elseBlock}
	}

	p.expect(lexer.SHORE)
	return IfStatement{Condition: condition, Then: then, Else: nil}
}

func parseWhileStatement(p *parser) Statement {
	p.advance()
	condition := parseExpression(p, defaultBp)
	body := parseBody(p)
	return WhileStatement{Condition: condition, Body: body}
}

func parseForInStatement(p *parser) Statement {
	p.advance()
	iterator := p.expectError(lexer.IDENT, p.parseError("expected iterator name after 'for'")).Value
	p.expectError(lexer.IN, p.parseError("expected 'in' after iterator name"))
	iterable := parseExpression(p, defaultBp)
	body := parseBody(p)
	return ForInStatement{Iterator: iterator, Iterable: iterable, Body: body}
}

func parseFunctionParameters(p *parser) []FunctionParameter {
	params := []FunctionParameter{}
	p.expect(lexer.OPEN_PAREN)
	for p.hasTokens() && p.currentTokenKind() != lexer.CLOSE_PAREN {
		name := p.expectError(lexer.IDENT, p.parseError("expected parameter name")).Value
		p.expectError(lexer.COLON, p.parseError(fmt.Sprintf("expected ':' after parameter '%s'", name)))
		paramType := parseType(p, defaultBp)
		params = append(params, FunctionParameter{Name: name, Type: paramType})
		if p.currentTokenKind() == lexer.COMMA {
			p.advance()
		}
	}
	p.expect(lexer.CLOSE_PAREN)
	return params
}

func parseFunctionDecStatement(p *parser) Statement {
	p.advance()
	name := p.expectError(lexer.IDENT, p.parseError("expected function name after 'swim'")).Value
	params := parseFunctionParameters(p)
	p.expectError(lexer.COLON, p.parseError(fmt.Sprintf("expected ':' after parameters of function '%s'", name)))
	returnType := parseType(p, defaultBp)
	body := parseBody(p)
	return FunctionDecStatement{Name: name, Parameters: params, ReturnType: returnType, Body: body}
}

func parseReturnStatement(p *parser) Statement {
	p.advance()
	return ReturnStatement{Value: parseExpression(p, defaultBp)}
}

func parseVarDeclarationStatement(p *parser) Statement {
	var explicitType AstType
	var assignedValue Expression

	keyword := p.advance()
	isConst := keyword.Kind == lexer.CONST
	varName := p.expectError(lexer.IDENT, p.parseError("expected variable name")).Value

	if p.currentTokenKind() == lexer.COLON {
		p.advance()
		explicitType = parseType(p, defaultBp)
	}

	if p.currentTokenKind() == lexer.ASSIGNMENT {
		p.advance()
		assignedValue = parseExpression(p, assignmentBp)
	} else if explicitType == nil {
		panic(lexer.NewError(keyword.Line, keyword.Column,
			fmt.Sprintf("variable '%s' must have a type annotation or an assigned value", varName)))
	}

	if isConst && assignedValue == nil {
		panic(lexer.NewError(keyword.Line, keyword.Column,
			fmt.Sprintf("constant '%s' must be assigned a value", varName)))
	}

	return VariableDecStatement{
		IsConstant:    isConst,
		VariableName:  varName,
		AssignedValue: assignedValue,
		ExplicitType:  explicitType,
	}
}

func parseExpression(p *parser, bp BindingPower) Expression {
	tokenKind := p.currentTokenKind()
	nudFn, exists := nudLu[tokenKind]
	if !exists {
		panic(p.parseError(fmt.Sprintf("unexpected token '%s'", p.currentToken().Value)))
	}
	left := nudFn(p)

	for bpLu[p.currentTokenKind()] > bp {
		tokenKind = p.currentTokenKind()
		ledFn, exists := ledLu[tokenKind]
		if !exists {
			panic(p.parseError(fmt.Sprintf("unexpected token '%s'", p.currentToken().Value)))
		}
		left = ledFn(p, left, bpLu[p.currentTokenKind()])
	}
	return left
}

func parsePrimaryExpression(p *parser) Expression {
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

func parseArrayLiteral(p *parser) Expression {
	p.advance()
	elements := []Expression{}
	for p.hasTokens() && p.currentTokenKind() != lexer.CLOSE_BRACKET {
		elements = append(elements, parseExpression(p, assignmentBp))
		if p.currentTokenKind() == lexer.COMMA {
			p.advance()
		}
	}
	p.expect(lexer.CLOSE_BRACKET)
	return ArrayLiteral{Elements: elements}
}

func parseBinaryExpression(p *parser, left Expression, bp BindingPower) Expression {
	operatorToken := p.advance()
	right := parseExpression(p, bp)
	return BinaryExpression{Left: left, Operator: operatorToken, Right: right}
}

func parseAssignExpression(p *parser, left Expression, bp BindingPower) Expression {
	operatorToken := p.advance()
	rhs := parseExpression(p, bp)
	return AssignmentExpression{Operator: operatorToken, Value: rhs, Assignee: left}
}

func parseGroupingExpression(p *parser) Expression {
	p.advance()
	exp := parseExpression(p, defaultBp)
	p.expect(lexer.CLOSE_PAREN)
	return exp
}

func parsePrefixExpression(p *parser) Expression {
	operatorToken := p.advance()
	rhs := parseExpression(p, unaryBp)
	return PrefixExpression{Operator: operatorToken, RightExpression: rhs}
}

func parseCallExpression(p *parser, left Expression, bp BindingPower) Expression {
	p.advance()
	args := []Expression{}
	for p.hasTokens() && p.currentTokenKind() != lexer.CLOSE_PAREN {
		args = append(args, parseExpression(p, assignmentBp))
		if p.currentTokenKind() == lexer.COMMA {
			p.advance()
		}
	}
	p.expect(lexer.CLOSE_PAREN)
	return CallExpression{Callee: left, Arguments: args}
}

func parseType(p *parser, bp BindingPower) AstType {
	tokenKind := p.currentTokenKind()
	nudFn, exists := typeNudLu[tokenKind]
	if !exists {
		panic(p.parseError(fmt.Sprintf("expected a type but got '%s'", p.currentToken().Value)))
	}
	left := nudFn(p)

	for typeBpLu[p.currentTokenKind()] > bp {
		tokenKind = p.currentTokenKind()
		ledFn, exists := typeLedLu[tokenKind]
		if !exists {
			panic(p.parseError(fmt.Sprintf("unexpected token in type expression '%s'", p.currentToken().Value)))
		}
		left = ledFn(p, left, typeBpLu[p.currentTokenKind()])
	}
	return left
}

func parseSymbolType(p *parser) AstType {
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

func parseArrayType(p *parser) AstType {
	p.advance()
	p.expect(lexer.CLOSE_BRACKET)
	return ArrayType{Underlying: parseType(p, defaultBp)}
}

func parsePostfixExpression(p *parser, left Expression, bp BindingPower) Expression {
	operatorToken := p.advance()
	return PostfixExpression{Operator: operatorToken, Left: left}
}

func parseFunctionType(p *parser) AstType {
	p.advance()
	return SymbolType{Name: "function"}
}

func parseTypeofExpression(p *parser) Expression {
	p.advance()
	expr := parseExpression(p, unaryBp)
	return TypeofExpression{Expr: expr}
}

func parseObjectLiteral(p *parser) Expression {
	p.advance()
	properties := []ObjectProperty{}
	for p.hasTokens() && p.currentTokenKind() != lexer.CLOSE_CURLY {
		key := p.expectError(lexer.IDENT, p.parseError("expected property name in object literal")).Value
		p.expectError(lexer.COLON, p.parseError(fmt.Sprintf("expected ':' after property '%s'", key)))
		value := parseExpression(p, assignmentBp)
		properties = append(properties, ObjectProperty{Key: key, Value: value})
		if p.currentTokenKind() == lexer.COMMA {
			p.advance()
		}
	}
	p.expect(lexer.CLOSE_CURLY)
	return ObjectLiteral{Properties: properties}
}

func parseMemberExpression(p *parser, left Expression, bp BindingPower) Expression {
	p.advance()
	prop := p.expectError(lexer.IDENT, p.parseError("expected property name after '.'")).Value
	return MemberExpression{Object: left, Property: prop}
}

func parseImportStatement(p *parser) Statement {
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

func parseCastStatement(p *parser) Statement {
	p.advance()
	tok := p.currentToken()
	var inner Statement
	switch tok.Kind {
	case lexer.SWIM:
		inner = parseFunctionDecStatement(p)
	case lexer.LET, lexer.CONST:
		inner = parseVarDeclarationStatement(p)
	default:
		panic(p.parseError(fmt.Sprintf("'cast' must be followed by 'swim', 'catch', or 'anchor', got '%s'", tok.Value)))
	}
	return CastStatement{Inner: inner}
}
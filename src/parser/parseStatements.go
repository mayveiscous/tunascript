package parser

import (
	"fmt"
	"strconv"
	"tunascript/src/lexer"
)

func parseStatement(p *parser) Statement {
	if p.currentTokenKind() == lexer.IDENT {
		if stmt, ok := tryParseSwapStatement(p); ok {
			consumeSemicolon(p)
			return stmt
		}
	}

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

func tryParseSwapStatement(p *parser) (Statement, bool) {
	if p.pos+1 >= len(p.tokens) || p.tokens[p.pos+1].Kind != lexer.COMMA {
		return nil, false
	}

	targets := []Expression{}
	tok := p.advance()
	targets = append(targets, SymbolExpression{Token: tok, Value: tok.Value})
	for p.currentTokenKind() == lexer.COMMA {
		p.advance()
		targets = append(targets, parseExpression(p, assignmentBp))
	}

	if p.currentTokenKind() != lexer.ASSIGNMENT {
		panic(p.parseError("expected '=' after swap targets"))
	}
	p.advance()

	values := []Expression{}
	values = append(values, parseExpression(p, assignmentBp))
	for p.currentTokenKind() == lexer.COMMA {
		p.advance()
		values = append(values, parseExpression(p, assignmentBp))
	}

	if len(targets) != len(values) {
		panic(p.parseError(fmt.Sprintf(
			"swap mismatch: %d targets but %d values", len(targets), len(values))))
	}

	return SwapStatement{Targets: targets, Values: values}, true
}

func parseBreakStatement(p *parser) Statement {
	tok := p.advance()
	return BreakStatement{Token: tok}
}

func parseContinueStatement(p *parser) Statement {
	tok := p.advance()
	return ContinueStatement{Token: tok}
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
	first := p.expectError(lexer.IDENT, p.parseError("expected iterator name after 'for'")).Value

	keyVar := ""
	iterator := first
	if p.currentTokenKind() == lexer.COMMA {
		p.advance()
		second := p.expectError(lexer.IDENT, p.parseError("expected second iterator name after ','")).Value
		keyVar = first
		iterator = second
	}

	p.expectError(lexer.IN, p.parseError("expected 'in' after iterator name"))
	iterable := parseExpression(p, defaultBp)
	body := parseBody(p)
	return ForInStatement{KeyVar: keyVar, Iterator: iterator, Iterable: iterable, Body: body}
}

func parseFunctionParameters(p *parser) []FunctionParameter {
	params := []FunctionParameter{}
	p.expect(lexer.OPEN_PAREN)
	for p.hasTokens() && p.currentTokenKind() != lexer.CLOSE_PAREN {
		isVariadic := false
		tok := p.currentToken()
		if p.currentTokenKind() == lexer.ELLIPSIS {
			isVariadic = true
			p.advance()
			tok = p.currentToken()
		}
		name := p.expectError(lexer.IDENT, p.parseError("expected parameter name")).Value
		p.expectError(lexer.COLON, p.parseError(fmt.Sprintf("expected ':' after parameter '%s'", name)))
		paramType := parseType(p, defaultBp)
		params = append(params, FunctionParameter{Token: tok, Name: name, Type: paramType, IsVariadic: isVariadic})
		if isVariadic {
			break
		}
		if p.currentTokenKind() == lexer.COMMA {
			p.advance()
		}
	}
	p.expect(lexer.CLOSE_PAREN)
	return params
}

func parseFunctionDecStatement(p *parser) Statement {
	tok := p.advance()
	name := p.expectError(lexer.IDENT, p.parseError("expected function name after 'swim'")).Value
	params := parseFunctionParameters(p)
	p.expectError(lexer.COLON, p.parseError(fmt.Sprintf("expected ':' after parameters of function '%s'", name)))
	returnType := parseType(p, defaultBp)
	body := parseBody(p)
	return FunctionDecStatement{Token: tok, Name: name, Parameters: params, ReturnType: returnType, Body: body}
}

func parseFunctionExpression(p *parser) Expression {
	p.advance()
	params := parseFunctionParameters(p)
	var returnType AstType
	if p.currentTokenKind() == lexer.COLON {
		p.advance()
		returnType = parseType(p, defaultBp)
	}
	body := parseBody(p)
	return FunctionExpression{
		Parameters:	params,
		ReturnType:	returnType,
		Body:		body,
	}
}

func parseReturnStatement(p *parser) Statement {
	tok := p.advance()
	return ReturnStatement{Token: tok, Value: parseExpression(p, defaultBp)}
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

	if p.currentTokenKind() == lexer.ASSIGNMENT || p.currentTokenKind() == lexer.AS {
		p.advance()
		assignedValue = parseExpression(p, assignmentBp)
	} else if explicitType == nil {
		panic(lexer.NewError(p.filePath, keyword.Line, keyword.Column,
			fmt.Sprintf("variable '%s' must have a type annotation or an assigned value", varName)))
	}

	if isConst && assignedValue == nil {
		panic(lexer.NewError(p.filePath, keyword.Line, keyword.Column,
			fmt.Sprintf("constant '%s' must be assigned a value", varName)))
	}

	return VariableDecStatement{
		Token:		keyword,
		IsConstant:	isConst,
		VariableName:	varName,
		AssignedValue:	assignedValue,
		ExplicitType:	explicitType,
	}
}

func parseTryBody(p *parser) BlockStatement {
	body := []Statement{}
	for p.hasTokens() &&
		p.currentTokenKind() != lexer.SHORE &&
		p.currentTokenKind() != lexer.HOOK {
		body = append(body, parseStatement(p))
	}
	return BlockStatement{Body: body}
}

func parseTryStatement(p *parser) Statement {
	p.advance()
	tryBody := parseTryBody(p)
	p.expectError(lexer.HOOK, p.parseError("expected 'hook' after try block"))
	errName := p.expectError(lexer.IDENT, p.parseError("expected error variable name after 'hook'")).Value
	hookBody := parseBody(p)
	return TryStatement{Body: tryBody, ErrName: errName, Hook: hookBody}
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
		tok := p.advance()
		return SymbolExpression{Token: tok, Value: "nil"}
	case lexer.FALSE:
		p.advance()
		return BoolExpression{Value: false}
	case lexer.STRING:
		return StringExpression{Value: p.advance().Value}
	case lexer.IDENT:
		tok := p.advance()
		return SymbolExpression{Token: tok, Value: tok.Value}
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
	builtinTypes := map[string]bool{
		"number":	true, "string": true, "bool": true,
		"function":	true, "void": true, "null": true, "array": true, "object": true,
	}

	if !builtinTypes[name] && (len(name) == 0 || name[0] < 'A' || name[0] > 'Z') {
		panic(lexer.NewError(p.filePath, tok.Line, tok.Column,
			fmt.Sprintf("unknown type '%s': use a built-in type or a school name (must start with uppercase)", name)))
	}
	return SymbolType{Name: name}
}

func parseSchoolStatement(p *parser) Statement {
	p.advance()
	nameTok := p.expectError(lexer.IDENT, p.parseError("expected school name after 'school'"))
	name := nameTok.Value
	if len(name) == 0 || name[0] < 'A' || name[0] > 'Z' {
		panic(lexer.NewError(p.filePath, nameTok.Line, nameTok.Column,
			"school names must start with an uppercase letter"))
	}
	p.expectError(lexer.ASSIGNMENT, p.parseError(fmt.Sprintf("expected '=' after school name '%s'", name)))
	p.expectError(lexer.OPEN_CURLY, p.parseError("expected '{' to begin school body"))

	fields := []SchoolFieldDef{}
	seen := map[string]bool{}
	for p.hasTokens() && p.currentTokenKind() != lexer.CLOSE_CURLY {
		fieldTok := p.expectError(lexer.IDENT, p.parseError("expected field name in school body"))
		fieldName := fieldTok.Value
		if seen[fieldName] {
			panic(lexer.NewError(p.filePath, fieldTok.Line, fieldTok.Column,
				fmt.Sprintf("duplicate field '%s' in school '%s'", fieldName, name)))
		}
		seen[fieldName] = true
		p.expectError(lexer.COLON, p.parseError(fmt.Sprintf("expected ':' after field name '%s'", fieldName)))
		fieldType := parseType(p, defaultBp)
		fields = append(fields, SchoolFieldDef{Name: fieldName, Type: fieldType})
		if p.currentTokenKind() == lexer.COMMA {
			p.advance()
		}
	}
	p.expectError(lexer.CLOSE_CURLY, p.parseError("expected '}' to close school body"))
	return SchoolStatement{Name: name, Fields: fields}
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
	tok := p.advance()
	if p.currentTokenKind() != lexer.STRING {
		panic(p.parseError("expected a file path string after 'from'"))
	}
	path := p.advance().Value

	p.expectError(lexer.LET, p.parseError("expected 'catch' after import path"))

	items := []ImportItem{}
	for {
		itemTok := p.currentToken()
		name := p.expectError(lexer.IDENT, p.parseError("expected export name")).Value
		alias := name
		if p.currentTokenKind() == lexer.AS {
			p.advance()
			alias = p.expectError(lexer.IDENT, p.parseError("expected alias name after 'as'")).Value
		}
		items = append(items, ImportItem{Token: itemTok, Name: name, Alias: alias})
		if p.currentTokenKind() == lexer.COMMA {
			p.advance()
		} else {
			break
		}
	}
	return ImportStatement{Token: tok, Path: path, Items: items}
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

package parser

import (
	"fmt"
	"tunascript/src/lexer"
)

func ledReg(kind lexer.TokenKind, bp BindingPower, fn ledHandler) {
	bpLu[kind] = bp
	ledLu[kind] = fn
}

func nudReg(kind lexer.TokenKind, fn nudHandler) {
	nudLu[kind] = fn
}

func statementReg(kind lexer.TokenKind, fn statementHandler) {
	bpLu[kind] = defaultBp
	statementLu[kind] = fn
}

func typeNudReg(kind lexer.TokenKind, fn typeNudHandler) {
	typeNudLu[kind] = fn
}

func (p *parser) currentToken() lexer.Token { 
	return p.tokens[p.pos] 
}

func (p *parser) currentTokenKind() lexer.TokenKind {
	 return p.currentToken().Kind 
}

func (p *parser) advance() lexer.Token {
	token := p.currentToken()
	p.pos++
	return token
}

func (p *parser) expect(expectedKind lexer.TokenKind) lexer.Token {
	return p.expectError(expectedKind, nil)
}


func (p *parser) hasTokens() bool {
	return p.pos < len(p.tokens) && p.currentTokenKind() != lexer.EOF
}

func consumeSemicolon(p *parser) {
	if p.currentTokenKind() == lexer.SEMICOLON {
		p.advance()
	}
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


func init() {
	ledReg(lexer.ASSIGNMENT,  assignmentBp, parseAssignExpression)
	ledReg(lexer.PLUS_EQUALS,  assignmentBp, parseAssignExpression)
	ledReg(lexer.MINUS_EQUALS, assignmentBp, parseAssignExpression)
	ledReg(lexer.STAR_EQUALS,  assignmentBp, parseAssignExpression)
	ledReg(lexer.SLASH_EQUALS, assignmentBp, parseAssignExpression)

	ledReg(lexer.AND, logicalBp, parseBinaryExpression)
	ledReg(lexer.OR,  logicalBp, parseBinaryExpression)

	ledReg(lexer.LESS,          relationalBp, parseBinaryExpression)
	ledReg(lexer.LESS_EQUALS,   relationalBp, parseBinaryExpression)
	ledReg(lexer.GREATER,       relationalBp, parseBinaryExpression)
	ledReg(lexer.GREATER_EQUALS,relationalBp, parseBinaryExpression)
	ledReg(lexer.EQUALS,        relationalBp, parseBinaryExpression)
	ledReg(lexer.NOT_EQUALS,    relationalBp, parseBinaryExpression)

	ledReg(lexer.PLUS, additiveBp, parseBinaryExpression)
	ledReg(lexer.DASH, additiveBp, parseBinaryExpression)

	ledReg(lexer.STAR,    multiplicativeBp, parseBinaryExpression)
	ledReg(lexer.SLASH,   multiplicativeBp, parseBinaryExpression)
	ledReg(lexer.PERCENT, multiplicativeBp, parseBinaryExpression)

	ledReg(lexer.OPEN_PAREN, callBp, parseCallExpression)

	ledReg(lexer.OPEN_BRACKET, memberBp, parseIndexExpression)
	ledReg(lexer.DOT,          memberBp, parseMemberExpression)

	ledReg(lexer.PLUS_PLUS,   unaryBp, parsePostfixExpression)
	ledReg(lexer.MINUS_MINUS, unaryBp, parsePostfixExpression)

	nudReg(lexer.NUMBER,       parsePrimaryExpression)
	nudReg(lexer.STRING,       parsePrimaryExpression)
	nudReg(lexer.IDENT,        parsePrimaryExpression)
	nudReg(lexer.OPEN_PAREN,   parseGroupingExpression)
	nudReg(lexer.DASH,         parsePrefixExpression)
	nudReg(lexer.NOT,          parsePrefixExpression)
	nudReg(lexer.OPEN_BRACKET, parseArrayLiteral)
	nudReg(lexer.OPEN_CURLY,   parseObjectLiteral)
	nudReg(lexer.TRUE,         parsePrimaryExpression)
	nudReg(lexer.FALSE,        parsePrimaryExpression)
	nudReg(lexer.NULL,         parsePrimaryExpression)
	nudReg(lexer.TYPEOF,       parseTypeofExpression)

	statementReg(lexer.CONST,    parseVarDeclarationStatement)
	statementReg(lexer.LET,      parseVarDeclarationStatement)
	statementReg(lexer.SWIM,     parseFunctionDecStatement)
	statementReg(lexer.RETURN,   parseReturnStatement)
	statementReg(lexer.IF,       parseIfStatement)
	statementReg(lexer.WHILE,    parseWhileStatement)
	statementReg(lexer.FOR,      parseForInStatement)
	statementReg(lexer.BREAK,    parseBreakStatement)
	statementReg(lexer.CONTINUE, parseContinueStatement)
	statementReg(lexer.FROM,     parseImportStatement)
	statementReg(lexer.CAST,     parseCastStatement)

	typeNudReg(lexer.IDENT,        parseSymbolType)
	typeNudReg(lexer.OPEN_BRACKET, parseArrayType)
	typeNudReg(lexer.FUNCTION,     parseFunctionType)
}

func createParser(tokens []lexer.Token) *parser {
	return &parser{tokens: tokens, pos: 0}
}

func Parse(tokens []lexer.Token) BlockStatement {
	body := make([]Statement, 0)
	p := createParser(tokens)
	for p.hasTokens() {
		body = append(body, parseStatement(p))
	}
	return BlockStatement{Body: body}
}
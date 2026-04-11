package lexer

import (
	"fmt"
	"regexp"
)

type TokenKind int

const (
	EOF TokenKind = iota
	NUMBER
	STRING
	IDENT

	TRUE
	FALSE

	PLUS
	DASH
	SLASH
	STAR
	PERCENT

	ASSIGNMENT  // =
	EQUALS      // ==
	NOT         // !
	NOT_EQUALS  // !=

	LESS           // <
	LESS_EQUALS    // <=
	GREATER        // >
	GREATER_EQUALS // >=

	OR
	AND
	DOT

	OPEN_PAREN
	CLOSE_PAREN
	OPEN_BRACKET
	CLOSE_BRACKET
	OPEN_CURLY
	CLOSE_CURLY

	SEMICOLON
	COLON
	COMMA
	QUESTION

	NULL

	PLUS_PLUS    // ++
	MINUS_MINUS  // --
	PLUS_EQUALS  // +=
	MINUS_EQUALS // -=
	STAR_EQUALS  // *=
	SLASH_EQUALS // /=

	// Reserved
	SHORE
	CLASS
	CONST
	IF
	ELSE
	WHILE
	FOR
	FROM
	NEW
	TYPEOF
	IN
	LET
	RETURN
	FUNCTION
	SWIM
	BREAK
	CONTINUE
)

var reserved_lu = map[string]TokenKind{
	"true":     TRUE,
	"false":    FALSE,
	"shore":    SHORE,
	"catch":    LET,
	"anchor":   CONST,
	"if":       IF,
	"else":     ELSE,
	"while":    WHILE,
	"for":      FOR,
	"typeof":   TYPEOF,
	"in":       IN,
	"class":    CLASS,
	"serve":    RETURN,
	"new":      NEW,
	"swim":     SWIM,
	"function": FUNCTION,
	"from":     FROM,
	"and":      AND,
	"or":       OR,
	"nil": NULL,
	"break": BREAK,
	"continue": CONTINUE,
}

type Token struct {
	Value  string
	Kind   TokenKind
	Line   int
	Column int
}

func NewToken(kind TokenKind, value string, line, col int) Token {
	return Token{Kind: kind, Value: value, Line: line, Column: col}
}

func (token Token) IsOneOfMany(expectedTokens ...TokenKind) bool {
	for _, exp := range expectedTokens {
		if exp == token.Kind {
			return true
		}
	}
	return false
}

func TokenKindString(kind TokenKind) string {
	switch kind {
	case EOF:
		return "EOF"
	case NUMBER:
		return "number"
	case STRING:
		return "string"
	case IDENT:
		return "identifier"
	case SHORE:
		return "shore"
	case LET:
		return "catch"
	case RETURN:
		return "serve"
	case NEW:
		return "new"
	case PLUS:
		return "+"
	case DASH:
		return "-"
	case SLASH:
		return "/"
	case STAR:
		return "*"
	case PERCENT:
		return "%"
	case ASSIGNMENT:
		return "="
	case EQUALS:
		return "=="
	case NOT:
		return "!"
	case NOT_EQUALS:
		return "!="
	case LESS:
		return "<"
	case LESS_EQUALS:
		return "<="
	case GREATER:
		return ">"
	case GREATER_EQUALS:
		return ">="
	case OR:
		return "or"
	case AND:
		return "and"
	case DOT:
		return "."
	case FUNCTION:
		return "function"
	case SWIM:
		return "swim"
	case OPEN_PAREN:
		return "("
	case CLOSE_PAREN:
		return ")"
	case OPEN_BRACKET:
		return "["
	case CLOSE_BRACKET:
		return "]"
	case OPEN_CURLY:
		return "{"
	case CLOSE_CURLY:
		return "}"
	case SEMICOLON:
		return ";"
	case COLON:
		return ":"
	case COMMA:
		return ","
	case QUESTION:
		return "?"
	case PLUS_PLUS:
		return "++"
	case MINUS_MINUS:
		return "--"
	case PLUS_EQUALS:
		return "+="
	case MINUS_EQUALS:
		return "-="
	case STAR_EQUALS:
		return "*="
	case SLASH_EQUALS:
		return "/="
	case CLASS:
		return "class"
	case CONST:
		return "anchor"
	case IF:
		return "if"
	case ELSE:
		return "else"
	case WHILE:
		return "while"
	case FOR:
		return "for"
	case TYPEOF:
		return "typeof"
	case IN:
		return "in"
	default:
		return "unknown"
	}
}

type TunaError struct {
	Line    int
	Column  int
	Message string
}

func (e *TunaError) Error() string {
	return fmt.Sprintf("\n\033[31m[TunaScript Error]\033[0m Line %d, Col %d: %s", e.Line, e.Column, e.Message)
}

func NewError(line, col int, msg string) *TunaError {
	return &TunaError{Line: line, Column: col, Message: msg}
}

type regexHandler func(lex *lexer, regex *regexp.Regexp)

type regexPattern struct {
	regex   *regexp.Regexp
	handler regexHandler
}

type lexer struct {
	Tokens   []Token
	source   string
	pos      int
	line     int
	col      int
	patterns []regexPattern
}

func (lex *lexer) advanceN(n int) {
	for i := 0; i < n; i++ {
		if lex.pos < len(lex.source) && lex.source[lex.pos] == '\n' {
			lex.line++
			lex.col = 1
		} else {
			lex.col++
		}
		lex.pos++
	}
}

func (lex *lexer) push(token Token)  { lex.Tokens = append(lex.Tokens, token) }
func (lex *lexer) remainder() string { return lex.source[lex.pos:] }
func (lex *lexer) at_eof() bool      { return lex.pos >= len(lex.source) }

func defaultHandler(kind TokenKind, value string) regexHandler {
	return func(lex *lexer, regex *regexp.Regexp) {
		line, col := lex.line, lex.col
		lex.advanceN(len(value))
		lex.push(NewToken(kind, value, line, col))
	}
}

func numberHandler(lex *lexer, regex *regexp.Regexp) {
	match := regex.FindString(lex.remainder())
	line, col := lex.line, lex.col
	lex.push(NewToken(NUMBER, match, line, col))
	lex.advanceN(len(match))
}

func stringHandler(lex *lexer, regex *regexp.Regexp) {
	match := regex.FindStringIndex(lex.remainder())
	raw := lex.remainder()[match[0]+1 : match[1]-1]
	// Process escape sequences
	var buf []byte
	i := 0
	for i < len(raw) {
		if raw[i] == '\\' && i+1 < len(raw) {
			switch raw[i+1] {
			case '"':
				buf = append(buf, '"')
			case '\\':
				buf = append(buf, '\\')
			case 'n':
				buf = append(buf, '\n')
			case 't':
				buf = append(buf, '\t')
			case 'r':
				buf = append(buf, '\r')
			default:
				buf = append(buf, raw[i], raw[i+1])
			}
			i += 2
		} else {
			buf = append(buf, raw[i])
			i++
		}
	}
	stringLiteral := string(buf)
	line, col := lex.line, lex.col
	lex.push(NewToken(STRING, stringLiteral, line, col))
	lex.advanceN(match[1])
}

func commentHandler(lex *lexer, regex *regexp.Regexp) {
	match := regex.FindString(lex.remainder())
	lex.advanceN(len(match))
}

func skipWhitespace(lex *lexer, regex *regexp.Regexp) {
	match := regex.FindStringIndex(lex.remainder())
	lex.advanceN(match[1])
}

func symbolHandler(lex *lexer, regex *regexp.Regexp) {
	value := regex.FindString(lex.remainder())
	line, col := lex.line, lex.col
	if kind, exists := reserved_lu[value]; exists {
		lex.push(NewToken(kind, value, line, col))
	} else {
		lex.push(NewToken(IDENT, value, line, col))
	}
	lex.advanceN(len(value))
}

func createLexer(source string) *lexer {
	return &lexer{
		pos:    0,
		line:   1,
		col:    1,
		source: source,
		Tokens: make([]Token, 0),
		patterns: []regexPattern{
			{regexp.MustCompile(`\s+`), skipWhitespace},
			{regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*`), symbolHandler},
			{regexp.MustCompile(`[0-9]+(\.[0-9]+)?`), numberHandler},
			{regexp.MustCompile(`"(?:[^"\\]|\\.)*"`), stringHandler},
			{regexp.MustCompile(`><>[^\n]*`), commentHandler},
			{regexp.MustCompile(`\(`), defaultHandler(OPEN_PAREN, "(")},
			{regexp.MustCompile(`\)`), defaultHandler(CLOSE_PAREN, ")")},
			{regexp.MustCompile(`\+\+`), defaultHandler(PLUS_PLUS, "++")},
			{regexp.MustCompile(`--`), defaultHandler(MINUS_MINUS, "--")},
			{regexp.MustCompile(`\+=`), defaultHandler(PLUS_EQUALS, "+=")},
			{regexp.MustCompile(`-=`), defaultHandler(MINUS_EQUALS, "-=")},
			{regexp.MustCompile(`\*=`), defaultHandler(STAR_EQUALS, "*=")},
			{regexp.MustCompile(`==`), defaultHandler(EQUALS, "==")},
			{regexp.MustCompile(`!=`), defaultHandler(NOT_EQUALS, "!=")},
			{regexp.MustCompile(`<=`), defaultHandler(LESS_EQUALS, "<=")},
			{regexp.MustCompile(`>=`), defaultHandler(GREATER_EQUALS, ">=")},
			{regexp.MustCompile(`=`), defaultHandler(ASSIGNMENT, "=")},
			{regexp.MustCompile(`!`), defaultHandler(NOT, "!")},
			{regexp.MustCompile(`<`), defaultHandler(LESS, "<")},
			{regexp.MustCompile(`>`), defaultHandler(GREATER, ">")},
			{regexp.MustCompile(`\+`), defaultHandler(PLUS, "+")},
			{regexp.MustCompile(`-`), defaultHandler(DASH, "-")},
			{regexp.MustCompile(`/=`), defaultHandler(SLASH_EQUALS, "/=")},
			{regexp.MustCompile(`/`), defaultHandler(SLASH, "/")},
			{regexp.MustCompile(`\*`), defaultHandler(STAR, "*")},
			{regexp.MustCompile(`%`), defaultHandler(PERCENT, "%")},
			{regexp.MustCompile(`\.`), defaultHandler(DOT, ".")},
			{regexp.MustCompile(`\[`), defaultHandler(OPEN_BRACKET, "[")},
			{regexp.MustCompile(`\]`), defaultHandler(CLOSE_BRACKET, "]")},
			{regexp.MustCompile(`\{`), defaultHandler(OPEN_CURLY, "{")},
			{regexp.MustCompile(`\}`), defaultHandler(CLOSE_CURLY, "}")},
			{regexp.MustCompile(`;`), defaultHandler(SEMICOLON, ";")},
			{regexp.MustCompile(`:`), defaultHandler(COLON, ":")},
			{regexp.MustCompile(`,`), defaultHandler(COMMA, ",")},
			{regexp.MustCompile(`\?`), defaultHandler(QUESTION, "?")},
		},
	}
}

func Lex(source string) []Token {
	lex := createLexer(source)

	for !lex.at_eof() {
		matched := false
		for _, pattern := range lex.patterns {
			loc := pattern.regex.FindStringIndex(lex.remainder())
			if loc != nil && loc[0] == 0 {
				pattern.handler(lex, pattern.regex)
				matched = true
				break
			}
		}
		if !matched {
			panic(NewError(lex.line, lex.col,
				fmt.Sprintf("unexpected character '%s'", string(lex.source[lex.pos]))))
		}
	}

	lex.push(NewToken(EOF, "EOF", lex.line, lex.col))
	return lex.Tokens
}
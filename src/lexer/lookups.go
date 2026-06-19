package lexer

import (
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
	ELLIPSIS

	NULL

	PLUS_PLUS    // ++
	MINUS_MINUS  // --
	PLUS_EQUALS  // +=
	MINUS_EQUALS // -=
	STAR_EQUALS  // *=
	SLASH_EQUALS // /=

	// Reserved
	SHORE
	CONST
	IF
	ELSE
	WHILE
	FOR
	FROM
	AS
	NEW
	TYPEOF
	IN
	LET
	RETURN
	FUNCTION
	SWIM
	BREAK
	CONTINUE
	CAST
	TRY
	HOOK
	SCHOOL
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
	"cast": CAST,
	"as": AS,
	"try":  TRY,
	"hook": HOOK,
	"school": SCHOOL,
}

var tokenKindStrings = map[TokenKind]string{
	EOF:           "EOF",
	NUMBER:        "number",
	STRING:        "string",
	IDENT:         "identifier",
	SHORE:         "shore",
	LET:           "catch",
	RETURN:        "serve",
	NEW:           "new",
	PLUS:          "+",
	DASH:          "-",
	SLASH:         "/",
	STAR:          "*",
	PERCENT:       "%",
	ASSIGNMENT:    "=",
	EQUALS:        "==",
	NOT:           "!",
	NOT_EQUALS:    "!=",
	LESS:          "<",
	LESS_EQUALS:   "<=",
	GREATER:       ">",
	GREATER_EQUALS: ">=",
	OR:            "or",
	AND:           "and",
	DOT:           ".",
	FUNCTION:      "function",
	SWIM:          "swim",
	OPEN_PAREN:    "(",
	CLOSE_PAREN:   ")",
	OPEN_BRACKET:  "[",
	CLOSE_BRACKET: "]",
	OPEN_CURLY:    "{",
	CLOSE_CURLY:   "}",
	SEMICOLON:     ";",
	COLON:         ":",
	COMMA:         ",",
	QUESTION:      "?",
	PLUS_PLUS:     "++",
	MINUS_MINUS:   "--",
	PLUS_EQUALS:   "+=",
	MINUS_EQUALS:  "-=",
	STAR_EQUALS:   "*=",
	SLASH_EQUALS:  "/=",
	CONST:         "anchor",
	IF:            "if",
	ELSE:          "else",
	WHILE:         "while",
	FOR:           "for",
	TYPEOF:        "typeof",
	IN:            "in",
	CAST:          "cast",
	ELLIPSIS:      "...",
	TRY:  			"try",
	HOOK: 			"hook",
	SCHOOL:			"school",
}

var patterns = []regexPattern{
	// Skipped
	{regexp.MustCompile(`\s+`), skipWhitespace},
	{regexp.MustCompile(`><>[^\n]*`), commentHandler},
	{regexp.MustCompile(`(?s)></.*?/>`), blockCommentHandler},

	// Literals
	{regexp.MustCompile(`"(?:[^"\\]|\\.)*"`), stringHandler},
	{regexp.MustCompile(`[0-9]+(\.[0-9]+)?`), numberHandler},
	{regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*`), symbolHandler},

	// Multi-char operators (must come before their single-char prefixes)
	{regexp.MustCompile(`\+\+`), defaultHandler(PLUS_PLUS, "++")},
	{regexp.MustCompile(`--`),   defaultHandler(MINUS_MINUS, "--")},
	{regexp.MustCompile(`\+=`),  defaultHandler(PLUS_EQUALS, "+=")},
	{regexp.MustCompile(`-=`),   defaultHandler(MINUS_EQUALS, "-=")},
	{regexp.MustCompile(`\*=`),  defaultHandler(STAR_EQUALS, "*=")},
	{regexp.MustCompile(`/=`),   defaultHandler(SLASH_EQUALS, "/=")},
	{regexp.MustCompile(`==`),   defaultHandler(EQUALS, "==")},
	{regexp.MustCompile(`!=`),   defaultHandler(NOT_EQUALS, "!=")},
	{regexp.MustCompile(`<=`),   defaultHandler(LESS_EQUALS, "<=")},
	{regexp.MustCompile(`>=`),   defaultHandler(GREATER_EQUALS, ">=")},
	
	// ellipsis for varidic functions
	{regexp.MustCompile(`\.\.\.`), defaultHandler(ELLIPSIS, "...")},

	// Single-char operators
	{regexp.MustCompile(`=`),  defaultHandler(ASSIGNMENT, "=")},
	{regexp.MustCompile(`!`),  defaultHandler(NOT, "!")},
	{regexp.MustCompile(`<`),  defaultHandler(LESS, "<")},
	{regexp.MustCompile(`>`),  defaultHandler(GREATER, ">")},
	{regexp.MustCompile(`\+`), defaultHandler(PLUS, "+")},
	{regexp.MustCompile(`-`),  defaultHandler(DASH, "-")},
	{regexp.MustCompile(`/`),  defaultHandler(SLASH, "/")},
	{regexp.MustCompile(`\*`), defaultHandler(STAR, "*")},
	{regexp.MustCompile(`%`),  defaultHandler(PERCENT, "%")},
	{regexp.MustCompile(`\.`), defaultHandler(DOT, ".")},

	// Delimiters
	{regexp.MustCompile(`\(`), defaultHandler(OPEN_PAREN, "(")},
	{regexp.MustCompile(`\)`), defaultHandler(CLOSE_PAREN, ")")},
	{regexp.MustCompile(`\[`), defaultHandler(OPEN_BRACKET, "[")},
	{regexp.MustCompile(`\]`), defaultHandler(CLOSE_BRACKET, "]")},
	{regexp.MustCompile(`\{`), defaultHandler(OPEN_CURLY, "{")},
	{regexp.MustCompile(`\}`), defaultHandler(CLOSE_CURLY, "}")},
	{regexp.MustCompile(`;`),  defaultHandler(SEMICOLON, ";")},
	{regexp.MustCompile(`:`),  defaultHandler(COLON, ":")},
	{regexp.MustCompile(`,`),  defaultHandler(COMMA, ",")},
	{regexp.MustCompile(`\?`), defaultHandler(QUESTION, "?")},
}
package lexer

import (
	"regexp"
	"fmt"
)

type lexer struct {
	Tokens		[]Token
	source		string
	pos		int
	line		int
	col		int
	patterns	[]regexPattern
	filePath	string
}

type Token struct {
	Value		string
	Kind		TokenKind
	Line		int
	Column		int
	FilePath	string
}

type regexHandler func(lex *lexer, regex *regexp.Regexp)

type regexPattern struct {
	regex	*regexp.Regexp
	handler	regexHandler
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

func NewToken(kind TokenKind, value string, line, col int, filePath string) Token {
	return Token{
		Value:		value,
		Kind:		kind,
		Line:		line,
		Column:		col,
		FilePath:	filePath,
	}
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
	if s, ok := tokenKindStrings[kind]; ok {
		return s
	}
	return "unknown"
}

func (lex *lexer) push(token Token) {
	lex.Tokens = append(lex.Tokens, token)
}

func (lex *lexer) remainder() string {
	return lex.source[lex.pos:]
}

func (lex *lexer) at_eof() bool {
	return lex.pos >= len(lex.source)
}

func skipWhitespace(lex *lexer, regex *regexp.Regexp) {
	match := regex.FindStringIndex(lex.remainder())
	lex.advanceN(match[1])
}

func createLexer(source string, filePath string) *lexer {
	return &lexer{
		pos:		0,
		line:		1,
		col:		1,
		source:		source,
		Tokens:		make([]Token, 0),
		patterns:	patterns,
		filePath:	filePath,
	}
}

func Lex(source string, filePath string) []Token {
	lex := createLexer(source, filePath)

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
			panic(NewError(lex.filePath, lex.line, lex.col,
				fmt.Sprintf("unexpected character '%s'", string(lex.source[lex.pos]))))
		}
	}

	lex.push(NewToken(EOF, "EOF", lex.line, lex.col, lex.filePath))
	return lex.Tokens
}

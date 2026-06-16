package lexer

import (
	"regexp"
)

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

func commentHandler(lex *lexer, regex *regexp.Regexp) {
	match := regex.FindString(lex.remainder())
	lex.advanceN(len(match))
}

func blockCommentHandler(lex *lexer, regex *regexp.Regexp) {
	match := regex.FindString(lex.remainder())
	lex.advanceN(len(match))
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

func stringHandler(lex *lexer, regex *regexp.Regexp) {
	match := regex.FindStringIndex(lex.remainder())
	raw := lex.remainder()[match[0]+1 : match[1]-1]
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
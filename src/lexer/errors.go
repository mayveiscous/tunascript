package lexer

import "fmt"

type TunaError struct {
	Line    int
	Column  int
	Message string
}

func (e *TunaError) Error() string {
	return fmt.Sprintf("\n\033[31m[Tunascript Error]\033[0m Line %d, Col %d: %s", e.Line, e.Column, e.Message)
}

func NewError(line, col int, msg string) *TunaError {
	return &TunaError{
		Line: line,
		Column: col,
		Message: msg,
	}
}
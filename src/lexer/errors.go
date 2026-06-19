package lexer

import "fmt"

type TunaError struct {
	FilePath string
	Line     int
	Column   int
	Message  string
}

func (e *TunaError) Error() string {
	loc := ""
	if e.FilePath != "" {
		loc = fmt.Sprintf("%s:", e.FilePath)
	}
	return fmt.Sprintf("\n\033[31m[Tunascript Error]\033[0m %sLine %d, Col %d: %s", loc, e.Line, e.Column, e.Message)
}

func NewError(filePath string, line, col int, msg string) *TunaError {
	return &TunaError{
		FilePath: filePath,
		Line:     line,
		Column:   col,
		Message:  msg,
	}
}
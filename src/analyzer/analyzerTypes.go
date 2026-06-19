package analyzer

import "fmt"

type DiagnosticLevel int

const (
	DiagError   DiagnosticLevel = iota
	DiagWarning
	DiagHint
)

type Diagnostic struct {
	Level    DiagnosticLevel
	Message  string
	Line     int
	Column   int
	FilePath string
}

func (d Diagnostic) String() string {
	var prefix string
	switch d.Level {
	case DiagError:
		prefix = "\033[31merror\033[0m"
	case DiagWarning:
		prefix = "\033[33mwarning\033[0m"
	case DiagHint:
		prefix = "\033[36mhint\033[0m"
	}
	loc := ""
	if d.FilePath != "" {
		loc = fmt.Sprintf(" at %s:%d:%d", d.FilePath, d.Line, d.Column)
	} else if d.Line > 0 {
		loc = fmt.Sprintf(" at line %d, col %d", d.Line, d.Column)
	}
	return fmt.Sprintf("\033[1m%s\033[0m%s: %s", prefix, loc, d.Message)
}

type varInfo struct {
	token   tokenPos
	isUsed  bool
	isConst bool
	isParam bool
}

type tokenPos struct {
	line int
	col  int
}

type scope struct {
	inFunction       bool
	inLoop           bool
	variables        map[string]*varInfo
	savedUnreachable bool
}

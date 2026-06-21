package analyzer

import (
	"fmt"
	"strings"
	tunaparser "tunascript/src/parser"
	"tunascript/src/directives"
)

var builtinNames = map[string]bool{
	"bubble":	true,
	"typeOf":	true,
	"toNumber":	true,
	"toString":	true,
	"len":		true,
	"json":		true,
	"math":		true,
	"string":	true,
	"array":	true,
	"tui":		true,
	"os":		true,
	"imui":		true,
}

type Analyzer struct {
	diagnostics	[]Diagnostic
	scopes		[]*scope
	unreachable	bool
	cfg		directives.Config
	filePath	string
}

func Analyze(block tunaparser.BlockStatement, cfg directives.Config, filePath string) []Diagnostic {
	a := &Analyzer{cfg: cfg, filePath: filePath}
	a.pushBlockScope()
	a.analyzeBlock(block)
	a.popScope()
	return a.diagnostics
}

func isContextError(msg string) bool {
	prefixes := []string{
		"'serve' can only be used inside",
		"'break' can only be used inside",
		"'continue' can only be used inside",
		"'break'/'continue' cannot cross",
		"'break' cannot cross",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(msg, p) {
			return true
		}
	}
	return false
}

func (a *Analyzer) add(level DiagnosticLevel, msg string, line, col int) {
	if a.cfg.NonStrict && level == DiagError && isContextError(msg) {
		level = DiagWarning
	}
	if a.cfg.WarnAsError && level == DiagWarning {
		level = DiagError
	}
	a.diagnostics = append(a.diagnostics, Diagnostic{Level: level, Message: msg, Line: line, Column: col, FilePath: a.filePath})
}

func (a *Analyzer) currentScope() *scope {
	if len(a.scopes) == 0 {
		return nil
	}
	return a.scopes[len(a.scopes)-1]
}

func (a *Analyzer) pushFunctionScope() {
	parentFn := false
	if s := a.currentScope(); s != nil {
		parentFn = s.inFunction
	}
	a.scopes = append(a.scopes, &scope{
		inFunction:		true,
		inLoop:			false,
		variables:		map[string]*varInfo{},
		savedUnreachable:	a.unreachable,
	})
	a.unreachable = false
	_ = parentFn
}

func (a *Analyzer) pushLoopScope() {
	parentFn := false
	if s := a.currentScope(); s != nil {
		parentFn = s.inFunction
	}
	a.scopes = append(a.scopes, &scope{
		inFunction:		parentFn,
		inLoop:			true,
		variables:		map[string]*varInfo{},
		savedUnreachable:	a.unreachable,
	})
	a.unreachable = false
}

func (a *Analyzer) pushBlockScope() {
	parentFn := false
	parentLoop := false
	if s := a.currentScope(); s != nil {
		parentFn = s.inFunction
		parentLoop = s.inLoop
	}
	a.scopes = append(a.scopes, &scope{
		inFunction:		parentFn,
		inLoop:			parentLoop,
		variables:		map[string]*varInfo{},
		savedUnreachable:	a.unreachable,
	})
	a.unreachable = false
}

func (a *Analyzer) popScope() {
	s := a.currentScope()
	if s == nil {
		return
	}
	for name, info := range s.variables {
		if !info.isUsed && !isDiscardName(name) {
			kind := "variable"
			if info.isParam {
				kind = "parameter"
			}
			a.add(DiagWarning, fmt.Sprintf("unused %s '%s'", kind, name), info.token.line, info.token.col)
		}
	}
	a.unreachable = s.savedUnreachable
	a.scopes = a.scopes[:len(a.scopes)-1]
}

func (a *Analyzer) declare(name string, line, col int, isConst bool) {
	if s := a.currentScope(); s != nil {
		s.variables[name] = &varInfo{
			token:		tokenPos{line: line, col: col},
			isUsed:		false,
			isConst:	isConst,
			isParam:	false,
		}
	}
}

func (a *Analyzer) declareParam(name string, line, col int) {
	if s := a.currentScope(); s != nil {
		s.variables[name] = &varInfo{
			token:		tokenPos{line: line, col: col},
			isUsed:		false,
			isConst:	false,
			isParam:	true,
		}
	}
}

func (a *Analyzer) resolve(name string) (*varInfo, bool) {
	for i := len(a.scopes) - 1; i >= 0; i-- {
		if info, ok := a.scopes[i].variables[name]; ok {
			info.isUsed = true
			return info, true
		}
	}
	return nil, false
}

func (a *Analyzer) findInAnyScope(name string) *varInfo {
	for i := len(a.scopes) - 1; i >= 0; i-- {
		if info, ok := a.scopes[i].variables[name]; ok {
			return info
		}
	}
	return nil
}

func (a *Analyzer) inFunction() bool {
	if s := a.currentScope(); s != nil {
		return s.inFunction
	}
	return false
}

func (a *Analyzer) inLoop() bool {
	if s := a.currentScope(); s != nil {
		return s.inLoop
	}
	return false
}

func (a *Analyzer) analyzeBlock(block tunaparser.BlockStatement) {
	for _, stmt := range block.Body {
		if a.unreachable {
			line, col := stmtFirstToken(stmt)
			a.add(DiagWarning, "unreachable code", line, col)
		}
		a.analyzeStatement(stmt)
	}
}

func (a *Analyzer) analyzeStatement(stmt tunaparser.Statement) {
	switch s := stmt.(type) {

	case tunaparser.ExpressionStatement:
		a.analyzeExpression(s.Expression)

	case tunaparser.ReturnStatement:
		if !a.inFunction() {
			a.add(DiagError, "'serve' can only be used inside a function", s.Token.Line, s.Token.Column)
		}
		a.analyzeExpression(s.Value)
		a.unreachable = true

	case tunaparser.BreakStatement:
		if !a.inLoop() {
			if a.inFunction() {
				a.add(DiagError, "'break' cannot cross a function boundary", s.Token.Line, s.Token.Column)
			} else {
				a.add(DiagError, "'break' can only be used inside a loop", s.Token.Line, s.Token.Column)
			}
		}
		a.unreachable = true

	case tunaparser.ContinueStatement:
		if !a.inLoop() {
			if a.inFunction() {
				a.add(DiagError, "'continue' cannot cross a function boundary", s.Token.Line, s.Token.Column)
			} else {
				a.add(DiagError, "'continue' can only be used inside a loop", s.Token.Line, s.Token.Column)
			}
		}
		a.unreachable = true

	case tunaparser.VariableDecStatement:
		a.declare(s.VariableName, s.Token.Line, s.Token.Column, s.IsConstant)
		if s.AssignedValue != nil {
			a.analyzeExpression(s.AssignedValue)
		}
		if builtinNames[s.VariableName] {
			a.add(DiagWarning, fmt.Sprintf("overwriting builtin '%s'", s.VariableName), s.Token.Line, s.Token.Column)
		}
		if existing := a.findInAnyScopeButCurrent(s.VariableName); existing != nil {
			a.add(DiagWarning, fmt.Sprintf("variable '%s' shadows outer declaration", s.VariableName), s.Token.Line, s.Token.Column)
		}

	case tunaparser.FunctionDecStatement:
		a.declare(s.Name, s.Token.Line, s.Token.Column, false)
		a.pushFunctionScope()
		for _, p := range s.Parameters {
			a.declareParam(p.Name, p.Token.Line, p.Token.Column)
		}
		a.analyzeBlock(s.Body)
		a.popScope()

	case tunaparser.IfStatement:
		a.analyzeExpression(s.Condition)
		a.pushBlockScope()
		a.analyzeBlock(s.Then)
		a.popScope()
		if s.Else != nil {
			a.pushBlockScope()
			a.analyzeBlock(*s.Else)
			a.popScope()
		}

	case tunaparser.WhileStatement:
		a.analyzeExpression(s.Condition)
		a.pushLoopScope()
		a.analyzeBlock(s.Body)
		a.popScope()

	case tunaparser.ForInStatement:
		a.analyzeExpression(s.Iterable)
		a.pushLoopScope()
		a.declare(s.Iterator, 0, 0, false)
		if s.KeyVar != "" {
			a.declare(s.KeyVar, 0, 0, false)
		}
		a.analyzeBlock(s.Body)
		a.popScope()

	case tunaparser.BlockStatement:
		a.pushBlockScope()
		a.analyzeBlock(s)
		a.popScope()

	case tunaparser.TryStatement:
		a.pushBlockScope()
		a.analyzeBlock(s.Body)
		a.popScope()
		a.pushBlockScope()
		a.declare(s.ErrName, 0, 0, false)
		a.analyzeBlock(s.Hook)
		a.popScope()

	case tunaparser.CastStatement:
		a.analyzeStatement(s.Inner)

	case tunaparser.ImportStatement:
		for _, item := range s.Items {
			a.declare(item.Alias, item.Token.Line, item.Token.Column, false)
		}

	case tunaparser.SwapStatement:
		for _, target := range s.Targets {
			if sym, ok := target.(tunaparser.SymbolExpression); ok {
				if _, found := a.resolve(sym.Value); !found {

				}
			}
		}
		for _, val := range s.Values {
			a.analyzeExpression(val)
		}

	case tunaparser.SchoolStatement:

	default:
		panic(fmt.Sprintf("unknown statement type in analyzer: %T", stmt))
	}
}

func (a *Analyzer) analyzeExpression(expr tunaparser.Expression) {
	switch e := expr.(type) {

	case tunaparser.SymbolExpression:
		if e.Value == "nil" {
			return
		}
		if _, found := a.resolve(e.Value); !found {
			if !builtinNames[e.Value] {
				a.add(DiagError, fmt.Sprintf("undefined variable '%s'", e.Value), e.Token.Line, e.Token.Column)
			}
		}

	case tunaparser.FunctionExpression:
		a.pushFunctionScope()
		for _, p := range e.Parameters {
			a.declareParam(p.Name, p.Token.Line, p.Token.Column)
		}
		a.analyzeBlock(e.Body)
		a.popScope()

	case tunaparser.AssignmentExpression:
		a.analyzeAssignmentTarget(e.Assignee)
		a.analyzeExpression(e.Value)

	case tunaparser.CallExpression:
		a.analyzeExpression(e.Callee)
		for _, arg := range e.Arguments {
			a.analyzeExpression(arg)
		}

	case tunaparser.BinaryExpression:
		a.analyzeExpression(e.Left)
		a.analyzeExpression(e.Right)

	case tunaparser.PrefixExpression:
		a.analyzeExpression(e.RightExpression)

	case tunaparser.PostfixExpression:
		a.analyzeExpression(e.Left)

	case tunaparser.MemberExpression:
		a.analyzeExpression(e.Object)

	case tunaparser.IndexExpression:
		a.analyzeExpression(e.Left)
		a.analyzeExpression(e.Index)

	case tunaparser.ArrayLiteral:
		for _, el := range e.Elements {
			a.analyzeExpression(el)
		}

	case tunaparser.ObjectLiteral:
		for _, prop := range e.Properties {
			a.analyzeExpression(prop.Value)
		}

	case tunaparser.TypeofExpression:
		a.analyzeExpression(e.Expr)

	case tunaparser.NumberExpression, tunaparser.StringExpression, tunaparser.BoolExpression:

	default:
		panic(fmt.Sprintf("unknown expression type in analyzer: %T", expr))
	}
}

func (a *Analyzer) analyzeAssignmentTarget(expr tunaparser.Expression) {
	if sym, ok := expr.(tunaparser.SymbolExpression); ok {
		if info, found := a.resolve(sym.Value); found {
			if info.isConst {
				a.add(DiagError, fmt.Sprintf("cannot reassign constant '%s'", sym.Value), sym.Token.Line, sym.Token.Column)
			}
		} else if builtinNames[sym.Value] {
			a.add(DiagWarning, fmt.Sprintf("overwriting builtin '%s'", sym.Value), sym.Token.Line, sym.Token.Column)
		}

	}

	if mem, ok := expr.(tunaparser.MemberExpression); ok {
		a.analyzeExpression(mem.Object)
	}
	if idx, ok := expr.(tunaparser.IndexExpression); ok {
		a.analyzeExpression(idx.Left)
		a.analyzeExpression(idx.Index)
	}
}

func (a *Analyzer) findInAnyScopeButCurrent(name string) *varInfo {
	for i := len(a.scopes) - 2; i >= 0; i-- {
		if info, ok := a.scopes[i].variables[name]; ok {
			return info
		}
	}
	return nil
}

func isDiscardName(name string) bool {
	return name == "_" || strings.HasPrefix(name, "_") || strings.HasSuffix(name, "_")
}

func stmtFirstToken(stmt tunaparser.Statement) (int, int) {
	switch s := stmt.(type) {
	case tunaparser.ReturnStatement:
		return s.Token.Line, s.Token.Column
	case tunaparser.BreakStatement:
		return s.Token.Line, s.Token.Column
	case tunaparser.ContinueStatement:
		return s.Token.Line, s.Token.Column
	case tunaparser.VariableDecStatement:
		return s.Token.Line, s.Token.Column
	case tunaparser.FunctionDecStatement:
		return s.Token.Line, s.Token.Column
	default:
		return 0, 0
	}
}

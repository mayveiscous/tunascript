package lsp

import (
	tunaparser "tunascript/src/parser"
	"tunascript/src/lexer"
)

type NodeInfo struct {
	Type    string
	Range   Range
	Token   *lexer.Token
	Symbol  string
	IsFunc  bool
	IsCall  bool
	IsMember bool
	Member  string
	ObjExpr tunaparser.Expression
	Expr    tunaparser.Expression
	Stmt    tunaparser.Statement
}

func FindNodeAtPos(ast tunaparser.BlockStatement, pos Position) *NodeInfo {
	var best *NodeInfo

	var walkBlock func(block tunaparser.BlockStatement, depth int)
	var walkExpr func(expr tunaparser.Expression, depth int)
	var walkStmt func(stmt tunaparser.Statement, depth int)

	posAfter := func(tok lexer.Token) Position {
		return Position{Line: tok.Line - 1, Character: tok.Column - 1 + len(tok.Value)}
	}

	contains := func(r Range) bool {
		if pos.Line < r.Start.Line || (pos.Line == r.Start.Line && pos.Character < r.Start.Character) {
			return false
		}
		if r.End.Line == 0 && r.End.Character == 0 {
			return pos.Line >= r.Start.Line
		}
		if pos.Line > r.End.Line || (pos.Line == r.End.Line && pos.Character > r.End.Character) {
			return false
		}
		return true
	}

	tryBest := func(info *NodeInfo) {
		if best == nil {
			best = info
			return
		}
		bestStart := best.Range.Start
		infoStart := info.Range.Start
		if infoStart.Line > bestStart.Line || (infoStart.Line == bestStart.Line && infoStart.Character > bestStart.Character) {
			best = info
		}
	}

	walkBlock = func(block tunaparser.BlockStatement, depth int) {
		for _, stmt := range block.Body {
			walkStmt(stmt, depth)
		}
	}

	walkStmt = func(stmt tunaparser.Statement, depth int) {
		switch s := stmt.(type) {
		case tunaparser.ExpressionStatement:
			walkExpr(s.Expression, depth)

		case tunaparser.VariableDecStatement:
			start := Position{Line: s.Token.Line - 1, Character: s.Token.Column - 1}
			r := Range{Start: start}
			if contains(r) {
				tryBest(&NodeInfo{
					Type:   "variable_decl",
					Range:  r,
					Token:  &s.Token,
					Symbol: s.VariableName,
					Stmt:   stmt,
				})
			}
			if s.AssignedValue != nil {
				walkExpr(s.AssignedValue, depth)
			}

		case tunaparser.FunctionDecStatement:
			start := Position{Line: s.Token.Line - 1, Character: s.Token.Column - 1}
			r := Range{Start: start}
			if contains(r) {
				tryBest(&NodeInfo{
					Type:   "function_decl",
					Range:  r,
					Token:  &s.Token,
					Symbol: s.Name,
					Stmt:   stmt,
				})
			}
			walkBlock(s.Body, depth+1)

		case tunaparser.IfStatement:
			walkExpr(s.Condition, depth)
			walkBlock(s.Then, depth+1)
			if s.Else != nil {
				walkBlock(*s.Else, depth+1)
			}

		case tunaparser.WhileStatement:
			walkExpr(s.Condition, depth)
			walkBlock(s.Body, depth+1)

		case tunaparser.ForInStatement:
			walkExpr(s.Iterable, depth)
			walkBlock(s.Body, depth+1)

		case tunaparser.TryStatement:
			walkBlock(s.Body, depth+1)
			walkBlock(s.Hook, depth+1)

		case tunaparser.BlockStatement:
			walkBlock(s, depth+1)

		case tunaparser.ReturnStatement:
			if s.Value != nil {
				walkExpr(s.Value, depth)
			}

		case tunaparser.CastStatement:
			walkStmt(s.Inner, depth)

		case tunaparser.SchoolStatement:
			start := Position{Line: astStartLine(stmt)}
			r := Range{Start: start}
			if contains(r) {
				tryBest(&NodeInfo{Type: "school_decl", Range: r, Symbol: s.Name, Stmt: stmt})
			}
		}
	}

	walkExpr = func(expr tunaparser.Expression, depth int) {
		switch e := expr.(type) {
		case tunaparser.SymbolExpression:
			start := Position{Line: e.Token.Line - 1, Character: e.Token.Column - 1}
			end := Position{Line: e.Token.Line - 1, Character: e.Token.Column - 1 + len(e.Value)}
			r := Range{Start: start, End: end}
			if contains(r) {
				tryBest(&NodeInfo{
					Type:   "symbol",
					Range:  r,
					Token:  &e.Token,
					Symbol: e.Value,
					Expr:   expr,
				})
			}

		case tunaparser.MemberExpression:
			rightEdge := posAfter(lastTokenOf(expr))
			r := Range{Start: Position{Line: 0}, End: rightEdge}
			if contains(r) {
				tryBest(&NodeInfo{
					Type:     "member_access",
					Range:    r,
					Member:   e.Property,
					ObjExpr:  e.Object,
					Expr:     expr,
				})
			}
			walkExpr(e.Object, depth)

		case tunaparser.CallExpression:
			r := Range{Start: Position{Line: 0}}
			if contains(r) {
				tryBest(&NodeInfo{
					Type:   "call",
					IsCall: true,
					Expr:   expr,
				})
			}
			walkExpr(e.Callee, depth)
			for _, arg := range e.Arguments {
				walkExpr(arg, depth)
			}

		case tunaparser.BinaryExpression:
			walkExpr(e.Left, depth)
			walkExpr(e.Right, depth)

		case tunaparser.AssignmentExpression:
			walkExpr(e.Assignee, depth)
			walkExpr(e.Value, depth)

		case tunaparser.PrefixExpression:
			walkExpr(e.RightExpression, depth)

		case tunaparser.PostfixExpression:
			walkExpr(e.Left, depth)

		case tunaparser.IndexExpression:
			walkExpr(e.Left, depth)
			walkExpr(e.Index, depth)

		case tunaparser.ArrayLiteral:
			for _, el := range e.Elements {
				walkExpr(el, depth)
			}

		case tunaparser.ObjectLiteral:
			for _, prop := range e.Properties {
				walkExpr(prop.Value, depth)
			}

		case tunaparser.FunctionExpression:
			walkBlock(e.Body, depth+1)

		case tunaparser.TypeofExpression:
			walkExpr(e.Expr, depth)

		case tunaparser.NumberExpression, tunaparser.StringExpression, tunaparser.BoolExpression:
		}
	}

	walkBlock(ast, 0)
	return best
}

func lastTokenOf(node any) lexer.Token {
	switch n := node.(type) {
	case tunaparser.SymbolExpression:
		return n.Token
	case tunaparser.MemberExpression:
		return lastTokenOf(n.Object)
	case tunaparser.CallExpression:
		if len(n.Arguments) > 0 {
			return lastTokenOf(n.Arguments[len(n.Arguments)-1])
		}
		return lastTokenOf(n.Callee)
	case tunaparser.BinaryExpression:
		return lastTokenOf(n.Right)
	case tunaparser.NumberExpression:
		return lexer.Token{Line: 0, Column: 0}
	case tunaparser.StringExpression:
		return lexer.Token{Line: 0, Column: 0}
	case tunaparser.BoolExpression:
		return lexer.Token{Line: 0, Column: 0}
	default:
		return lexer.Token{Line: 0, Column: 0}
	}
}

func astStartLine(stmt tunaparser.Statement) int {
	switch s := stmt.(type) {
	case tunaparser.ReturnStatement:
		return s.Token.Line - 1
	case tunaparser.BreakStatement:
		return s.Token.Line - 1
	case tunaparser.ContinueStatement:
		return s.Token.Line - 1
	case tunaparser.VariableDecStatement:
		return s.Token.Line - 1
	case tunaparser.FunctionDecStatement:
		return s.Token.Line - 1
	case tunaparser.ExpressionStatement:
		if sym, ok := s.Expression.(tunaparser.SymbolExpression); ok {
			return sym.Token.Line - 1
		}
		return 0
	default:
		return 0
	}
}

func FindEnclosingScope(root *ScopeNode, pos Position) *ScopeNode {
	if root == nil {
		return nil
	}

	if root.Range.End.Line != 0 || root.Range.End.Character != 0 {
		if pos.Line < root.Range.Start.Line || (pos.Line == root.Range.Start.Line && pos.Character < root.Range.Start.Character) {
			return nil
		}
		if pos.Line > root.Range.End.Line || (pos.Line == root.Range.End.Line && pos.Character > root.Range.End.Character) {
			return nil
		}
	} else if root.Range.Start.Line != 0 || root.Range.Start.Character != 0 {
		if pos.Line < root.Range.Start.Line || (pos.Line == root.Range.Start.Line && pos.Character < root.Range.Start.Character) {
			return nil
		}
	}

	var best *ScopeNode
	for _, child := range root.Children {
		if candidate := FindEnclosingScope(child, pos); candidate != nil {
			if best == nil || candidate.ID > best.ID {
				best = candidate
			}
		}
	}
	if best != nil {
		return best
	}
	return root
}

func FindSymbolDecl(symbols []Symbol, scopes *ScopeNode, name string, pos Position) *Symbol {
	scope := FindEnclosingScope(scopes, pos)
	if scope == nil {
		return nil
	}

	validScopeIDs := map[int]bool{}
	for s := scope; s != nil; s = s.Parent {
		validScopeIDs[s.ID] = true
	}

	var best *Symbol
	for i := range symbols {
		s := &symbols[i]
		if s.Name == name && validScopeIDs[s.ScopeID] {
			if best == nil || s.ScopeID > best.ScopeID || (s.ScopeID == best.ScopeID && s.Range.Start.Line > best.Range.Start.Line) {
				best = s
			}
		}
	}
	return best
}

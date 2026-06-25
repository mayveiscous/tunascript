package lsp

import (
	"fmt"
	tunaparser "tunascript/src/parser"
	"tunascript/src/lexer"
)

func LexSafe(source string, filePath string) (tokens []lexer.Token, errors []lexer.TunaError) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(*lexer.TunaError); ok {
				errors = append(errors, *err)
			} else {
				panic(r)
			}
		}
	}()

	tokens = lexer.Lex(source, filePath)
	return
}

func ParseSafe(tokens []lexer.Token, filePath string) (ast tunaparser.BlockStatement, errors []lexer.TunaError) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case *tunaparser.ParseError:
				ast = v.Partial
				errors = append(errors, *v.Err)
			case *lexer.TunaError:
				errors = append(errors, *v)
			default:
				panic(r)
			}
		}
	}()

	ast = tunaparser.Parse(tokens, filePath)
	return
}

func BuildSymbolTable(ast tunaparser.BlockStatement, fileURI string) []Symbol {
	var symbols []Symbol

	var walkBlock func(block tunaparser.BlockStatement, scopeID int)
	var walkStmt func(stmt tunaparser.Statement, scopeID int)

	nextID := 1

	walkBlock = func(block tunaparser.BlockStatement, scopeID int) {
		for _, stmt := range block.Body {
			walkStmt(stmt, scopeID)
		}
	}

	walkStmt = func(stmt tunaparser.Statement, scopeID int) {
		switch s := stmt.(type) {
		case tunaparser.VariableDecStatement:
			symType := resolveTypeString(s.ExplicitType)
			kind := SymVariable
			if s.IsConstant {
				kind = SymConstant
			}
			symbols = append(symbols, Symbol{
				Name: s.VariableName,
				Kind: kind,
				Type: symType,
				URI:  fileURI,
				Range: Range{
					Start: Position{Line: s.Token.Line - 1, Character: s.Token.Column - 1},
				},
				ScopeID: scopeID,
			})

		case tunaparser.FunctionDecStatement:
			retType := resolveTypeString(s.ReturnType)
			var params []ParamInfo
			for _, p := range s.Parameters {
				params = append(params, ParamInfo{
					Name: p.Name,
					Type: resolveTypeString(p.Type),
				})
			}
			symbols = append(symbols, Symbol{
				Name:   s.Name,
				Kind:   SymFunction,
				Type:   retType,
				Params: params,
				URI:    fileURI,
				Range: Range{
					Start: Position{Line: s.Token.Line - 1, Character: s.Token.Column - 1},
				},
				ScopeID: scopeID,
			})

			fnScopeID := nextID
			nextID++
			for _, p := range s.Parameters {
				symbols = append(symbols, Symbol{
					Name: p.Name,
					Kind: SymParameter,
					Type: resolveTypeString(p.Type),
					URI:  fileURI,
					Range: Range{
						Start: Position{Line: p.Token.Line - 1, Character: p.Token.Column - 1},
					},
					ScopeID: fnScopeID,
				})
			}
			walkBlock(s.Body, fnScopeID)

		case tunaparser.SchoolStatement:
			symbols = append(symbols, Symbol{
				Name: s.Name,
				Kind: SymType,
				Type: "type",
				URI:  fileURI,
				ScopeID: scopeID,
			})
			for _, f := range s.Fields {
				symbols = append(symbols, Symbol{
					Name: f.Name,
					Kind: SymField,
					Type: resolveTypeString(f.Type),
					URI:  fileURI,
					ScopeID: scopeID,
				})
			}

		case tunaparser.IfStatement:
			thenScopeID := nextID
			nextID++
			walkBlock(s.Then, thenScopeID)
			if s.Else != nil {
				elseScopeID := nextID
				nextID++
				walkBlock(*s.Else, elseScopeID)
			}

		case tunaparser.WhileStatement:
			bodyScopeID := nextID
			nextID++
			walkBlock(s.Body, bodyScopeID)

		case tunaparser.ForInStatement:
			bodyScopeID := nextID
			nextID++
			symbols = append(symbols, Symbol{
				Name: s.Iterator,
				Kind: SymIterator,
				Type: "unknown",
				URI:  fileURI,
				ScopeID: bodyScopeID,
			})
			if s.KeyVar != "" {
				symbols = append(symbols, Symbol{
					Name: s.KeyVar,
					Kind: SymIterator,
					Type: "number",
					URI:  fileURI,
					ScopeID: bodyScopeID,
				})
			}
			walkBlock(s.Body, bodyScopeID)

		case tunaparser.TryStatement:
			bodyScopeID := nextID
			nextID++
			walkBlock(s.Body, bodyScopeID)
			hookScopeID := nextID
			nextID++
			symbols = append(symbols, Symbol{
				Name: s.ErrName,
				Kind: SymError,
				Type: "string",
				URI:  fileURI,
				ScopeID: hookScopeID,
			})
			walkBlock(s.Hook, hookScopeID)

		case tunaparser.ImportStatement:
			for _, item := range s.Items {
				symbols = append(symbols, Symbol{
					Name: item.Alias,
					Kind: SymImport,
					Type: "unknown",
					URI:  fileURI,
					ScopeID: scopeID,
				})
			}

		case tunaparser.CastStatement:
			walkStmt(s.Inner, scopeID)

		case tunaparser.BlockStatement:
			blockScopeID := nextID
			nextID++
			walkBlock(s, blockScopeID)
		}
	}

	walkBlock(ast, 0)
	return symbols
}

func BuildScopeTree(ast tunaparser.BlockStatement) *ScopeNode {
	root := &ScopeNode{
		ID:     0,
		Type:   ScopeGlobal,
		Decls:  make(map[string]*Symbol),
		Range:  Range{},
	}

	var walkBlock func(block tunaparser.BlockStatement, parent *ScopeNode)
	var walkStmt func(stmt tunaparser.Statement, parent *ScopeNode)

	nextScopeID := 1

	blockStart := func(block tunaparser.BlockStatement) Position {
		if len(block.Body) > 0 {
			line := astStartLine(block.Body[0])
			return Position{Line: line, Character: 0}
		}
		return Position{}
	}

	walkBlock = func(block tunaparser.BlockStatement, parent *ScopeNode) {
		for _, stmt := range block.Body {
			walkStmt(stmt, parent)
		}
	}

	walkStmt = func(stmt tunaparser.Statement, parent *ScopeNode) {
		switch s := stmt.(type) {
		case tunaparser.FunctionDecStatement:
			fnScope := &ScopeNode{
				ID:     nextScopeID,
				Parent: parent,
				Type:   ScopeFunction,
				Decls:  make(map[string]*Symbol),
				Range:  Range{Start: Position{Line: s.Token.Line - 1, Character: s.Token.Column - 1}},
			}
			nextScopeID++
			parent.Children = append(parent.Children, fnScope)
			for _, p := range s.Parameters {
				fnScope.Decls[p.Name] = &Symbol{Name: p.Name, Kind: SymParameter}
			}
			walkBlock(s.Body, fnScope)

		case tunaparser.IfStatement:
			thenStart := blockStart(s.Then)
			thenScope := &ScopeNode{
				ID:     nextScopeID,
				Parent: parent,
				Type:   ScopeBlock,
				Decls:  make(map[string]*Symbol),
				Range:  Range{Start: thenStart},
			}
			nextScopeID++
			parent.Children = append(parent.Children, thenScope)
			walkBlock(s.Then, thenScope)
			if s.Else != nil {
				elseStart := blockStart(*s.Else)
				elseScope := &ScopeNode{
					ID:     nextScopeID,
					Parent: parent,
					Type:   ScopeBlock,
					Decls:  make(map[string]*Symbol),
					Range:  Range{Start: elseStart},
				}
				nextScopeID++
				parent.Children = append(parent.Children, elseScope)
				walkBlock(*s.Else, elseScope)
			}

		case tunaparser.WhileStatement:
			bodyStart := blockStart(s.Body)
			loopScope := &ScopeNode{
				ID:     nextScopeID,
				Parent: parent,
				Type:   ScopeLoop,
				Decls:  make(map[string]*Symbol),
				Range:  Range{Start: bodyStart},
			}
			nextScopeID++
			parent.Children = append(parent.Children, loopScope)
			walkBlock(s.Body, loopScope)

		case tunaparser.ForInStatement:
			bodyStart := blockStart(s.Body)
			loopScope := &ScopeNode{
				ID:     nextScopeID,
				Parent: parent,
				Type:   ScopeLoop,
				Decls:  make(map[string]*Symbol),
				Range:  Range{Start: bodyStart},
			}
			nextScopeID++
			parent.Children = append(parent.Children, loopScope)
			loopScope.Decls[s.Iterator] = &Symbol{Name: s.Iterator, Kind: SymIterator}
			if s.KeyVar != "" {
				loopScope.Decls[s.KeyVar] = &Symbol{Name: s.KeyVar, Kind: SymIterator}
			}
			walkBlock(s.Body, loopScope)

		case tunaparser.TryStatement:
			bodyStart := blockStart(s.Body)
			bodyScope := &ScopeNode{
				ID:     nextScopeID,
				Parent: parent,
				Type:   ScopeBlock,
				Decls:  make(map[string]*Symbol),
				Range:  Range{Start: bodyStart},
			}
			nextScopeID++
			parent.Children = append(parent.Children, bodyScope)
			walkBlock(s.Body, bodyScope)

			hookStart := blockStart(s.Hook)
			hookScope := &ScopeNode{
				ID:     nextScopeID,
				Parent: parent,
				Type:   ScopeBlock,
				Decls:  make(map[string]*Symbol),
				Range:  Range{Start: hookStart},
			}
			nextScopeID++
			parent.Children = append(parent.Children, hookScope)
			hookScope.Decls[s.ErrName] = &Symbol{Name: s.ErrName, Kind: SymError}
			walkBlock(s.Hook, hookScope)

		case tunaparser.BlockStatement:
			blockStartPos := blockStart(s)
			blockScope := &ScopeNode{
				ID:     nextScopeID,
				Parent: parent,
				Type:   ScopeBlock,
				Decls:  make(map[string]*Symbol),
				Range:  Range{Start: blockStartPos},
			}
			nextScopeID++
			parent.Children = append(parent.Children, blockScope)
			walkBlock(s, blockScope)

		case tunaparser.VariableDecStatement, tunaparser.CastStatement:
			if cast, ok := stmt.(tunaparser.CastStatement); ok {
				walkStmt(cast.Inner, parent)
			}
		}
	}

	for _, stmt := range ast.Body {
		if vd, ok := stmt.(tunaparser.VariableDecStatement); ok {
			root.Decls[vd.VariableName] = &Symbol{Name: vd.VariableName}
		}
		if fd, ok := stmt.(tunaparser.FunctionDecStatement); ok {
			root.Decls[fd.Name] = &Symbol{Name: fd.Name, Kind: SymFunction}
		}
		if sd, ok := stmt.(tunaparser.SchoolStatement); ok {
			root.Decls[sd.Name] = &Symbol{Name: sd.Name, Kind: SymType}
		}
		if imp, ok := stmt.(tunaparser.ImportStatement); ok {
			for _, item := range imp.Items {
				root.Decls[item.Alias] = &Symbol{Name: item.Alias, Kind: SymImport}
			}
		}
		if cast, ok := stmt.(tunaparser.CastStatement); ok {
			switch inner := cast.Inner.(type) {
			case tunaparser.FunctionDecStatement:
				root.Decls[inner.Name] = &Symbol{Name: inner.Name, Kind: SymFunction}
			case tunaparser.VariableDecStatement:
				root.Decls[inner.VariableName] = &Symbol{Name: inner.VariableName}
			}
		}
	}

	walkBlock(ast, root)
	return root
}

func inferSymbolTypes(symbols []Symbol, ast tunaparser.BlockStatement, lib *LanguageLibrary) {
	var walkBlock func(block tunaparser.BlockStatement)
	walkBlock = func(block tunaparser.BlockStatement) {
		for _, stmt := range block.Body {
			switch s := stmt.(type) {
			case tunaparser.VariableDecStatement:
				if s.ExplicitType == nil && s.AssignedValue != nil {
					inferredType := inferExprType(s.AssignedValue, lib, symbols)
					if inferredType != "" {
						for i := range symbols {
							if symbols[i].Name == s.VariableName && symbols[i].Type == "unknown" {
								symbols[i].Type = inferredType
								break
							}
						}
					}
				}
			case tunaparser.FunctionDecStatement:
				walkBlock(s.Body)
			case tunaparser.IfStatement:
				walkBlock(s.Then)
				if s.Else != nil {
					walkBlock(*s.Else)
				}
			case tunaparser.WhileStatement:
				walkBlock(s.Body)
			case tunaparser.ForInStatement:
				walkBlock(s.Body)
			case tunaparser.TryStatement:
				walkBlock(s.Body)
				walkBlock(s.Hook)
			case tunaparser.CastStatement:
				if fd, ok := s.Inner.(tunaparser.FunctionDecStatement); ok {
					walkBlock(fd.Body)
				}
			case tunaparser.BlockStatement:
				walkBlock(s)
			}
		}
	}
	walkBlock(ast)
}

func inferExprType(expr tunaparser.Expression, lib *LanguageLibrary, symbols []Symbol) string {
	switch e := expr.(type) {
	case tunaparser.CallExpression:
		if mem, ok := e.Callee.(tunaparser.MemberExpression); ok {
			if sym, ok := mem.Object.(tunaparser.SymbolExpression); ok {
				if ns, ok := lib.Namespaces[sym.Value]; ok {
					if m, ok := ns.Members[mem.Property]; ok {
						return m.Returns
					}
				}
			}
		}
		if sym, ok := e.Callee.(tunaparser.SymbolExpression); ok {
			for i := range symbols {
				if symbols[i].Name == sym.Value && symbols[i].Kind == SymFunction {
					return symbols[i].Type
				}
			}
		}
	case tunaparser.SymbolExpression:
		for i := range symbols {
			if symbols[i].Name == e.Value {
				return symbols[i].Type
			}
		}
	}
	return ""
}

func resolveTypeString(t tunaparser.AstType) string {
	if t == nil {
		return "unknown"
	}
	switch v := t.(type) {
	case tunaparser.SymbolType:
		return v.Name
	case tunaparser.ArrayType:
		inner := resolveTypeString(v.Underlying)
		if inner != "" {
			return "[]" + inner
		}
		return "array"
	default:
		return fmt.Sprintf("%T", t)
	}
}

package lsp

import (
	"fmt"
	tunaparser "tunascript/src/parser"
)

var keywordDocs = map[string]string{
	"catch":    "Declares a mutable variable: `catch x: number = 0`",
	"anchor":   "Declares a constant: `anchor PI: number = 3.14159`",
	"shore":    "Closes a block opened by `swim`, `if`, `while`, `for`, or `try`.",
	"swim":     "Declares a function: `swim name(param: type): returnType ... shore`",
	"serve":    "Returns a value from the enclosing function.",
	"if":       "Conditional statement.",
	"else":     "Alternate branch for `if`.",
	"while":    "Loops while a condition is truthy.",
	"for":      "Iterates: `for item in someArray ... shore`",
	"in":       "Used with `for` to iterate over an array's elements.",
	"break":    "Exits the nearest enclosing loop.",
	"continue": "Skips to the next iteration of the nearest enclosing loop.",
	"from":     "Begins an import statement.",
	"as":       "Aliases an imported name.",
	"cast":     "Marks a declaration as exported from the module.",
	"typeof":   "Returns the runtime type name of an expression as a string.",
	"try":      "Begins a try block.",
	"hook":     "Introduces the error-handling block of a `try` statement.",
	"true":     "Boolean literal.",
	"false":    "Boolean literal.",
	"nil":      "The null literal.",
	"and":      "Logical AND.",
	"or":       "Logical OR.",
	"school":   "Type declaration.",
}

var builtinHoverDocs = map[string]string{
	"bubble":   "**bubble(...values)** → void\n\nPrints all arguments to stdout, space-separated, followed by a newline.",
	"typeOf":   "**typeOf(value)** → string\n\nReturns the runtime type of a value as a string (e.g. \"number\", \"array\").",
	"toNumber": "**toNumber(value)** → number\n\nConverts a string or bool to a number. Strings must be valid numeric literals.",
	"toString": "**toString(value)** → string\n\nConverts any value to its string representation.",
	"len":      "**len(value)** → number\n\nReturns the length of a string (rune count) or array.",
}

func GetHover(entry *DocumentEntry, pos Position) *Hover {
	node := FindNodeAtPos(entry.AST, pos)
	if node == nil {
		return nil
	}

	switch node.Type {
	case "symbol":
		return hoverSymbol(node, entry, pos)
	case "member_access":
		return hoverMember(node, entry)
	case "variable_decl", "function_decl", "school_decl":
		if node.Symbol != "" {
			return hoverSymbol(node, entry, pos)
		}
		return nil
	default:
		return nil
	}
}

func hoverSymbol(node *NodeInfo, entry *DocumentEntry, pos Position) *Hover {
	name := node.Symbol

	if doc, ok := keywordDocs[name]; ok {
		return &Hover{
			Contents: MarkupContent{
				Kind:  "markdown",
				Value: fmt.Sprintf("**%s** _keyword_\n\n%s", name, doc),
			},
		}
	}

	if doc, ok := builtinHoverDocs[name]; ok {
		return &Hover{
			Contents: MarkupContent{Kind: "markdown", Value: doc},
		}
	}

	if ns, ok := entry.Lib.Namespaces[name]; ok {
		return &Hover{
			Contents: MarkupContent{
				Kind:  "markdown",
				Value: fmt.Sprintf("**%s** namespace\n\n%s", name, ns.Doc),
			},
		}
	}

	sym := FindSymbolDecl(entry.Symbols, entry.Scopes, name, pos)
	if sym != nil {
		return formatSymbolHover(sym)
	}

	return nil
}

func hoverMember(node *NodeInfo, entry *DocumentEntry) *Hover {
	nsName := symbolNameOf(node.ObjExpr)
	if ns, ok := entry.Lib.Namespaces[nsName]; ok {
		if member, ok := ns.Members[node.Member]; ok {
			return &Hover{
				Contents: MarkupContent{
					Kind:  "markdown",
					Value: memberHoverText(nsName, node.Member, member),
				},
			}
		}
	}

	objType := resolveExprType(node.ObjExpr, entry)
	if objType != "" {
		if members, ok := entry.Lib.TypeRegistry[objType]; ok {
			if member, ok := members[node.Member]; ok {
				return &Hover{
					Contents: MarkupContent{
						Kind:  "markdown",
						Value: memberHoverText(objType, node.Member, member),
					},
				}
			}
		}
		if schoolFields, ok := resolveSchoolFields(objType, entry); ok {
			if member, ok := schoolFields[node.Member]; ok {
				return &Hover{
					Contents: MarkupContent{
						Kind:  "markdown",
						Value: fmt.Sprintf("**%s**: %s\n\nfield of %s", node.Member, member.Returns, objType),
					},
				}
			}
		}
	}

	return nil
}

func formatSymbolHover(sym *Symbol) *Hover {
	var value string
	switch sym.Kind {
	case SymFunction:
		params := ""
		for i, p := range sym.Params {
			if i > 0 {
				params += ", "
			}
			params += fmt.Sprintf("%s: %s", p.Name, p.Type)
		}
		value = fmt.Sprintf("**swim %s(%s): %s**\n\nfunction, declared at line %d, col %d",
			sym.Name, params, sym.Type, sym.Range.Start.Line+1, sym.Range.Start.Character+1)
	case SymType:
		value = fmt.Sprintf("**%s**\n\ntype, declared at line %d, col %d",
			sym.Name, sym.Range.Start.Line+1, sym.Range.Start.Character+1)
	default:
		value = fmt.Sprintf("**%s**: %s\n\n%s, declared at line %d, col %d",
			sym.Name, sym.Type, symName(sym.Kind), sym.Range.Start.Line+1, sym.Range.Start.Character+1)
	}

	return &Hover{
		Contents: MarkupContent{Kind: "markdown", Value: value},
	}
}

func memberHoverText(nsOrType, memberName string, member LibMember) string {
	if member.IsConst {
		return fmt.Sprintf("**%s.%s** → %s\n\n%s", nsOrType, memberName, member.Returns, member.Doc)
	}
	params := ""
	for i, p := range member.Params {
		if i > 0 {
			params += ", "
		}
		params += p
	}
	return fmt.Sprintf("**%s.%s(%s)** → %s\n\n%s", nsOrType, memberName, params, member.Returns, member.Doc)
}

func resolveExprType(expr tunaparser.Expression, entry *DocumentEntry) string {
	switch e := expr.(type) {
	case tunaparser.SymbolExpression:
		sym := FindSymbolDecl(entry.Symbols, entry.Scopes, e.Value, Position{Line: e.Token.Line - 1})
		if sym != nil {
			return sym.Type
		}
		return ""
	case tunaparser.CallExpression:
		if sym, ok := e.Callee.(tunaparser.SymbolExpression); ok {
			if _, ok := entry.Lib.Namespaces[sym.Value]; ok {
				return ""
			}
			fnSym := FindSymbolDecl(entry.Symbols, entry.Scopes, sym.Value, Position{Line: sym.Token.Line - 1})
			if fnSym != nil {
				return fnSym.Type
			}
		}
		if mem, ok := e.Callee.(tunaparser.MemberExpression); ok {
			if ns, ok := entry.Lib.Namespaces[symbolNameOf(mem.Object)]; ok {
				if m, ok := ns.Members[mem.Property]; ok {
					return m.Returns
				}
			}
			objType := resolveExprType(mem.Object, entry)
			if members, ok := entry.Lib.TypeRegistry[objType]; ok {
				if m, ok := members[mem.Property]; ok {
					return m.Returns
				}
			}
		}
		return ""
	case tunaparser.MemberExpression:
		nsName := symbolNameOf(e.Object)
		if ns, ok := entry.Lib.Namespaces[nsName]; ok {
			if m, ok := ns.Members[e.Property]; ok {
				return m.Returns
			}
		}
		objType := resolveExprType(e.Object, entry)
		if members, ok := entry.Lib.TypeRegistry[objType]; ok {
			if m, ok := members[e.Property]; ok {
				return m.Returns
			}
		}
		return ""
	default:
		return ""
	}
}

func symbolNameOf(expr tunaparser.Expression) string {
	if sym, ok := expr.(tunaparser.SymbolExpression); ok {
		return sym.Value
	}
	return ""
}

func resolveSchoolFields(typeName string, entry *DocumentEntry) (map[string]LibMember, bool) {
	for i, sym := range entry.Symbols {
		if sym.Name == typeName && sym.Kind == SymType {
			fields := map[string]LibMember{}
			for j := i + 1; j < len(entry.Symbols); j++ {
				s := entry.Symbols[j]
				if s.Kind != SymField {
					break
				}
				fields[s.Name] = LibMember{Returns: s.Type, IsConst: true}
			}
			if len(fields) > 0 {
				return fields, true
			}
			return nil, false
		}
	}
	return nil, false
}

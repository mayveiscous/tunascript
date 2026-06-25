package lsp

import (
	"strings"
	"unicode"
)

func GetSignatureHelp(entry *DocumentEntry, pos Position) *SignatureHelp {
	lines := strings.Split(entry.Text, "\n")
	if pos.Line >= len(lines) {
		return nil
	}
	line := lines[pos.Line]
	if pos.Character > len(line) {
		pos.Character = len(line)
	}

	caller := findCallContext(line, pos.Character, entry)
	if caller == nil {
		return nil
	}

	label := caller.name + "("
	params := caller.params
	paramInfos := make([]ParameterInformation, len(params))

	for i, p := range params {
		if i > 0 {
			label += ", "
		}
		label += p
		paramInfos[i] = ParameterInformation{Label: p}
	}
	label += ") → " + caller.returns

	return &SignatureHelp{
		Signatures: []SignatureInformation{
			{
				Label:      label,
				Parameters: paramInfos,
			},
		},
		ActiveSignature: 0,
		ActiveParameter: caller.activeParam,
	}
}

type callContext struct {
	name       string
	params     []string
	returns    string
	activeParam int
}

func findCallContext(line string, character int, entry *DocumentEntry) *callContext {
	if character > len(line) {
		character = len(line)
	}
	if character <= 0 {
		return nil
	}
	openParen := -1
	for i := character - 1; i >= 0; i-- {
		if line[i] == ')' {
			depth := 1
			i--
			for i >= 0 && depth > 0 {
				if line[i] == ')' {
					depth++
				} else if line[i] == '(' {
					depth--
				}
				i--
			}
		} else if line[i] == '(' {
			openParen = i
			break
		}
	}
	if openParen == -1 {
		return nil
	}

	depth := 1
	closeParen := -1
	for i := openParen + 1; i < len(line); i++ {
		if line[i] == '(' {
			depth++
		} else if line[i] == ')' {
			depth--
			if depth == 0 {
				closeParen = i
				break
			}
		}
	}
	if closeParen != -1 && character >= closeParen {
		return nil
	}

	funcStart := openParen - 1
	for funcStart >= 0 && (unicode.IsLetter(rune(line[funcStart])) || unicode.IsDigit(rune(line[funcStart])) || line[funcStart] == '_' || line[funcStart] == '.') {
		funcStart--
	}
	funcName := strings.TrimSpace(line[funcStart+1 : openParen])

	activeParam := countCommas(line, openParen, character)

	if ns := extractNamespace(funcName); ns != "" {
		memberName := funcName[len(ns)+1:]
		if nsObj, ok := entry.Lib.Namespaces[ns]; ok {
			if m, ok := nsObj.Members[memberName]; ok {
				return &callContext{
					name:       funcName,
					params:     m.Params,
					returns:    m.Returns,
					activeParam: activeParam,
				}
			}
		}
		return nil
	}

	if dot := strings.LastIndex(funcName, "."); dot != -1 {
		objName := funcName[:dot]
		memberName := funcName[dot+1:]
		sym := FindSymbolDecl(entry.Symbols, entry.Scopes, objName, Position{Line: openParen - 1})
		if sym != nil {
			if members, ok := entry.Lib.TypeRegistry[sym.Type]; ok {
				if m, ok := members[memberName]; ok {
					return &callContext{
						name:       funcName,
						params:     m.Params,
						returns:    m.Returns,
						activeParam: activeParam,
					}
				}
			}
			if fields, ok := resolveSchoolFields(sym.Type, entry); ok {
				if m, ok := fields[memberName]; ok {
					return &callContext{
						name:       funcName,
						params:     m.Params,
						returns:    m.Returns,
						activeParam: activeParam,
					}
				}
			}
		}
		return nil
	}

	builtinParams := map[string][]string{
		"bubble":   {"...values"},
		"typeOf":   {"value"},
		"toNumber": {"value"},
		"toString": {"value"},
		"len":      {"value"},
	}
	builtinReturns := map[string]string{
		"bubble": "void", "typeOf": "string", "toNumber": "number",
		"toString": "string", "len": "number",
	}
	if params, ok := builtinParams[funcName]; ok {
		return &callContext{
			name:       funcName,
			params:     params,
			returns:    builtinReturns[funcName],
			activeParam: activeParam,
		}
	}

	sym := FindSymbolDecl(entry.Symbols, entry.Scopes, funcName, Position{Line: openParen - 1})
	if sym != nil && sym.Kind == SymFunction {
		paramStrs := make([]string, len(sym.Params))
		for i, p := range sym.Params {
			paramStrs[i] = p.Name + ": " + p.Type
		}
		return &callContext{
			name:       funcName,
			params:     paramStrs,
			returns:    sym.Type,
			activeParam: activeParam,
		}
	}

	return nil
}

func extractNamespace(name string) string {
	dot := strings.LastIndex(name, ".")
	if dot == -1 {
		return ""
	}
	ns := name[:dot]
	if _, ok := defaultLib.Namespaces[ns]; ok {
		return ns
	}
	return ""
}

func countCommas(line string, openParen int, character int) int {
	depth := 0
	count := 0
	for i := openParen + 1; i < character; i++ {
		switch line[i] {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				count++
			}
		}
	}
	return count
}

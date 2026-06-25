package lsp

import (
	"regexp"
	"strings"
)

var (
	reTypeColon     = regexp.MustCompile(`:\s*([a-zA-Z_]\w*)?$`)
	reSwimParams    = regexp.MustCompile(`\bswim\s+\w+\s*\([^)]*$`)
	reSwimReturn    = regexp.MustCompile(`\bswim\s+\w+\s*\([^)]*\)\s*$`)
	reCatchAnnot    = regexp.MustCompile(`^\s*(catch|anchor)\s+\w+\s*$`)
	reNamespaceDot  = regexp.MustCompile(`([a-zA-Z_]\w*)\.(\w*)$`)
	reWord          = regexp.MustCompile(`[a-zA-Z_]\w*$`)
	defaultLib      = NewLanguageLibrary()
)

type CompletionContext int

const (
	CtxIdentifier CompletionContext = iota
	CtxMember
	CtxNamespace
	CtxTypeAnnotation
	CtxNone
)

type ctxResult struct {
	context CompletionContext
	partial string
	objName string
}

func GetCompletions(entry *DocumentEntry, pos Position) *CompletionList {
	lines := strings.Split(entry.Text, "\n")
	if pos.Line >= len(lines) {
		return &CompletionList{IsIncomplete: false, Items: []CompletionItem{}}
	}

	line := lines[pos.Line]
	if pos.Character > len(line) {
		pos.Character = len(line)
	}
	beforeCursor := line[:pos.Character]
	ctx := detectContext(beforeCursor)

	var items []CompletionItem

	switch ctx.context {
	case CtxNamespace:
		items = completeNamespace(ctx, entry.Lib)
	case CtxMember:
		items = completeMember(ctx, entry, pos)
	case CtxTypeAnnotation:
		items = completeTypeAnnotation(ctx, entry)
	default:
		items = completeIdentifier(ctx, entry, pos)
	}

	if items == nil {
		items = []CompletionItem{}
	}
	return &CompletionList{IsIncomplete: false, Items: items}
}

func detectContext(before string) ctxResult {
	m := reTypeColon.FindStringSubmatch(before)
	if m != nil {
		colonIdx := strings.LastIndex(before, ":")
		beforeColon := strings.TrimRight(before[:colonIdx], " \t")
		if reSwimParams.MatchString(beforeColon) || reSwimReturn.MatchString(beforeColon) || reCatchAnnot.MatchString(beforeColon) {
			partial := ""
			if m[1] != "" {
				partial = m[1]
			}
			return ctxResult{context: CtxTypeAnnotation, partial: partial}
		}
		return ctxResult{context: CtxNone}
	}

	m = reNamespaceDot.FindStringSubmatch(before)
	if m != nil {
		if _, ok := defaultLib.Namespaces[m[1]]; ok {
			return ctxResult{context: CtxNamespace, partial: m[2], objName: m[1]}
		}
		return ctxResult{context: CtxMember, partial: m[2], objName: m[1]}
	}

	m = reWord.FindStringSubmatch(before)
	if m != nil {
		return ctxResult{context: CtxIdentifier, partial: m[0]}
	}

	return ctxResult{context: CtxNone}
}

func completeIdentifier(ctx ctxResult, entry *DocumentEntry, pos Position) []CompletionItem {
	items := []CompletionItem{}
	partial := ctx.partial

	keywordNames := []string{
		"catch", "anchor", "shore", "swim", "serve", "if", "else",
		"while", "for", "in", "break", "continue", "from", "as",
		"cast", "typeof", "try", "hook", "true", "false", "nil",
		"and", "or", "school",
	}
	for _, name := range keywordNames {
		if partial != "" && !strings.HasPrefix(name, partial) {
			continue
		}
		items = append(items, CompletionItem{
			Label:  name,
			Kind:   CompKeyword,
			Detail: "keyword",
		})
	}

	builtins := map[string]struct{ params, returns string }{
		"bubble":   {"...values", "void"},
		"typeOf":   {"value", "string"},
		"toNumber": {"value", "number"},
		"toString": {"value", "string"},
		"len":      {"value", "number"},
	}
	for name, b := range builtins {
		if partial != "" && !strings.HasPrefix(name, partial) {
			continue
		}
		items = append(items, CompletionItem{
			Label:  name,
			Kind:   CompFunction,
			Detail: name + "(" + b.params + ") → " + b.returns,
		})
	}

	nsNames := []string{"math", "string", "array", "tui", "os", "json", "imui"}
	for _, name := range nsNames {
		if partial != "" && !strings.HasPrefix(name, partial) {
			continue
		}
		items = append(items, CompletionItem{
			Label: name,
			Kind:  CompModule,
			Detail: name + " namespace",
		})
	}

	visibleScopeIDs := map[int]bool{}
	scope := FindEnclosingScope(entry.Scopes, pos)
	if scope != nil {
		for s := scope; s != nil; s = s.Parent {
			visibleScopeIDs[s.ID] = true
		}
	} else {
		visibleScopeIDs[0] = true
	}

	seen := map[string]bool{}
	for _, sym := range entry.Symbols {
		if !visibleScopeIDs[sym.ScopeID] {
			continue
		}
		if sym.Kind == SymField || sym.Kind == SymProperty || sym.Kind == SymError {
			continue
		}
		if seen[sym.Name] {
			continue
		}
		seen[sym.Name] = true
		if partial != "" && !strings.HasPrefix(sym.Name, partial) {
			continue
		}
		items = append(items, toCompletionItem(sym))
	}

	return items
}

func completeNamespace(ctx ctxResult, lib *LanguageLibrary) []CompletionItem {
	ns, ok := lib.Namespaces[ctx.objName]
	if !ok {
		return nil
	}

	items := []CompletionItem{}
	for name, member := range ns.Members {
		if ctx.partial != "" && !strings.HasPrefix(name, ctx.partial) {
			continue
		}
		items = append(items, memberToItem(name, member))
	}
	return items
}

func completeMember(ctx ctxResult, entry *DocumentEntry, pos Position) []CompletionItem {
	sym := FindSymbolDecl(entry.Symbols, entry.Scopes, ctx.objName, pos)
	if sym == nil {
		return nil
	}

	if members, ok := entry.Lib.TypeRegistry[sym.Type]; ok {
		items := []CompletionItem{}
		for name, member := range members {
			if ctx.partial != "" && !strings.HasPrefix(name, ctx.partial) {
				continue
			}
			items = append(items, memberToItem(name, member))
		}
		return items
	}

	if fields, ok := resolveSchoolFields(sym.Type, entry); ok {
		items := []CompletionItem{}
		for name, member := range fields {
			if ctx.partial != "" && !strings.HasPrefix(name, ctx.partial) {
				continue
			}
			items = append(items, memberToItem(name, member))
		}
		return items
	}

	return nil
}

func completeTypeAnnotation(ctx ctxResult, entry *DocumentEntry) []CompletionItem {
	seen := map[string]bool{}
	items := []CompletionItem{}

	for _, t := range entry.Lib.ValidTypes {
		if seen[t] {
			continue
		}
		seen[t] = true
		if ctx.partial != "" && !strings.HasPrefix(t, ctx.partial) {
			continue
		}
		items = append(items, CompletionItem{
			Label:  t,
			Kind:   CompTypeParameter,
			Detail: "built-in type",
		})
	}

	for _, sym := range entry.Symbols {
		if sym.Kind != SymType {
			continue
		}
		if seen[sym.Name] {
			continue
		}
		seen[sym.Name] = true
		if ctx.partial != "" && !strings.HasPrefix(sym.Name, ctx.partial) {
			continue
		}
		items = append(items, CompletionItem{
			Label:  sym.Name,
			Kind:   CompStruct,
			Detail: "custom type",
		})
	}

	return items
}

func toCompletionItem(sym Symbol) CompletionItem {
	kind := CompVariable
	switch sym.Kind {
	case SymFunction:
		kind = CompFunction
		params := ""
		for i, p := range sym.Params {
			if i > 0 {
				params += ", "
			}
			params += p.Name + ": " + p.Type
		}
		return CompletionItem{
			Label:  sym.Name,
			Kind:   kind,
			Detail: "swim " + sym.Name + "(" + params + "): " + sym.Type,
		}
	case SymConstant:
		kind = CompConstant
	case SymType:
		kind = CompStruct
	case SymImport:
		kind = CompReference
	case SymParameter:
		kind = CompTypeParameter
	}

	return CompletionItem{
		Label:  sym.Name,
		Kind:   kind,
		Detail: symName(sym.Kind) + ": " + sym.Type,
	}
}

func memberToItem(name string, member LibMember) CompletionItem {
	if member.IsConst {
		return CompletionItem{
			Label:  name,
			Kind:   CompProperty,
			Detail: name + ": " + member.Returns,
		}
	}
	params := ""
	for i, p := range member.Params {
		if i > 0 {
			params += ", "
		}
		params += p
	}
	return CompletionItem{
		Label:  name,
		Kind:   CompMethod,
		Detail: name + "(" + params + ")",
	}
}

func symName(k SymbolKind) string {
	switch k {
	case SymVariable:
		return "variable"
	case SymConstant:
		return "constant"
	case SymFunction:
		return "function"
	case SymParameter:
		return "parameter"
	case SymIterator:
		return "iterator"
	case SymImport:
		return "import"
	case SymType:
		return "type"
	case SymField:
		return "field"
	default:
		return "symbol"
	}
}

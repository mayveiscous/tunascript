package lsp

func GetDefinition(entry *DocumentEntry, pos Position) *Location {
	node := FindNodeAtPos(entry.AST, pos)
	if node == nil {
		return nil
	}

	if node.Type != "symbol" {
		return nil
	}

	sym := FindSymbolDecl(entry.Symbols, entry.Scopes, node.Symbol, pos)
	if sym == nil {
		return nil
	}

	return &Location{
		URI:   sym.URI,
		Range: sym.Range,
	}
}

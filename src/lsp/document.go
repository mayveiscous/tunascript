package lsp

import (
	tunaparser "tunascript/src/parser"
	"tunascript/src/lexer"
)

type DocumentEntry struct {
	URI      string
	Version  int
	Text     string
	Tokens   []lexer.Token
	AST      tunaparser.BlockStatement
	Symbols  []Symbol
	Scopes   *ScopeNode
	Diags    []Diagnostic
	Types    map[string]string
	Lib      *LanguageLibrary
}

type DocumentStore struct {
	docs map[string]*DocumentEntry
}

func NewDocumentStore() *DocumentStore {
	return &DocumentStore{
		docs: make(map[string]*DocumentEntry),
	}
}

func (ds *DocumentStore) Open(uri, text string, version int) {
	entry := &DocumentEntry{
		URI:     uri,
		Version: version,
		Text:    text,
	}
	ds.analyze(entry)
	ds.docs[uri] = entry
}

func (ds *DocumentStore) Update(uri, text string, version int) {
	entry, ok := ds.docs[uri]
	if !ok {
		entry = &DocumentEntry{URI: uri}
		ds.docs[uri] = entry
	}
	entry.Version = version
	entry.Text = text
	ds.analyze(entry)
}

func (ds *DocumentStore) Close(uri string) {
	delete(ds.docs, uri)
}

func (ds *DocumentStore) Get(uri string) *DocumentEntry {
	return ds.docs[uri]
}

func (ds *DocumentStore) analyze(entry *DocumentEntry) {
	tokens, _ := LexSafe(entry.Text, entry.URI)
	entry.Tokens = tokens

	ast, _ := ParseSafe(tokens, entry.URI)
	entry.AST = ast

	entry.Diags = []Diagnostic{}

	entry.Symbols = BuildSymbolTable(ast, entry.URI)
	entry.Scopes = BuildScopeTree(ast)
	entry.Lib = NewLanguageLibrary()
	inferSymbolTypes(entry.Symbols, entry.AST, entry.Lib)
}

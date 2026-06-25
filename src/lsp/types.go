package lsp

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

type DiagnosticSeverity int

const (
	DiagError   DiagnosticSeverity = 1
	DiagWarning DiagnosticSeverity = 2
	DiagInfo    DiagnosticSeverity = 3
	DiagHint    DiagnosticSeverity = 4
)

type Diagnostic struct {
	Range    Range              `json:"range"`
	Severity DiagnosticSeverity `json:"severity"`
	Message  string             `json:"message"`
	Source   string             `json:"source"`
}

type CompletionItemKind int

const (
	CompText          CompletionItemKind = 1
	CompMethod        CompletionItemKind = 2
	CompFunction      CompletionItemKind = 3
	CompConstructor   CompletionItemKind = 4
	CompField         CompletionItemKind = 5
	CompVariable      CompletionItemKind = 6
	CompClass         CompletionItemKind = 7
	CompInterface     CompletionItemKind = 8
	CompModule        CompletionItemKind = 9
	CompProperty      CompletionItemKind = 10
	CompUnit          CompletionItemKind = 11
	CompValue         CompletionItemKind = 12
	CompEnum          CompletionItemKind = 13
	CompKeyword       CompletionItemKind = 14
	CompSnippet       CompletionItemKind = 15
	CompColor         CompletionItemKind = 16
	CompFile          CompletionItemKind = 17
	CompReference     CompletionItemKind = 18
	CompFolder        CompletionItemKind = 19
	CompEnumMember    CompletionItemKind = 20
	CompConstant      CompletionItemKind = 21
	CompStruct        CompletionItemKind = 22
	CompEvent         CompletionItemKind = 23
	CompOperator      CompletionItemKind = 24
	CompTypeParameter CompletionItemKind = 25
)

type InsertTextFormat int

const (
	InsertPlainText InsertTextFormat = 1
	InsertSnippet   InsertTextFormat = 2
)

type CompletionItem struct {
	Label            string           `json:"label"`
	Kind             CompletionItemKind `json:"kind,omitempty"`
	Detail           string           `json:"detail,omitempty"`
	Documentation    string           `json:"documentation,omitempty"`
	InsertText       string           `json:"insertText,omitempty"`
	InsertTextFormat InsertTextFormat `json:"insertTextFormat,omitempty"`
}

type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type ParameterInformation struct {
	Label         string        `json:"label"`
	Documentation string        `json:"documentation,omitempty"`
}

type SignatureInformation struct {
	Label         string                  `json:"label"`
	Documentation string                  `json:"documentation,omitempty"`
	Parameters    []ParameterInformation  `json:"parameters,omitempty"`
}

type SignatureHelp struct {
	Signatures      []SignatureInformation `json:"signatures"`
	ActiveSignature int                    `json:"activeSignature"`
	ActiveParameter int                    `json:"activeParameter"`
}

type JSONRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCNotification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type SymbolKind int

const (
	SymVariable  SymbolKind = iota
	SymConstant
	SymFunction
	SymParameter
	SymIterator
	SymImport
	SymError
	SymType
	SymField
	SymProperty
)

type ParamInfo struct {
	Name string
	Type string
}

type Symbol struct {
	Name      string
	Kind      SymbolKind
	Type      string
	Doc       string
	Params    []ParamInfo
	URI       string
	Range     Range
	ScopeID   int
}

type ScopeType int

const (
	ScopeGlobal ScopeType = iota
	ScopeFunction
	ScopeBlock
	ScopeLoop
)

type ScopeNode struct {
	ID       int
	Parent   *ScopeNode
	Children []*ScopeNode
	Decls    map[string]*Symbol
	Type     ScopeType
	Range    Range
	Symbol   *Symbol
}

type DeclarationIndex struct {
	ByFile map[string][]*Symbol
}

package lsp

func GetDiagnostics(entry *DocumentEntry) []Diagnostic {
	return entry.Diags
}

package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type handlerFunc func(JSONRPCRequest) (any, error)
type notifHandler func(JSONRPCNotification)

type methodHandler struct {
	req  handlerFunc
	notif notifHandler
}

type Server struct {
	store       *DocumentStore
	handlers    map[string]methodHandler
	reader      *bufio.Reader
	writer      io.Writer
	initialized bool
	shutdown    bool
}

func (s *Server) log(msg string) {
	fmt.Fprintf(os.Stderr, "[lsp] %s\n", msg)
}

func NewServer() *Server {
	s := &Server{
		store:    NewDocumentStore(),
		reader:   bufio.NewReader(os.Stdin),
		writer:   os.Stdout,
		handlers: make(map[string]methodHandler),
	}

	s.handlers["initialize"] = methodHandler{req: s.handleInitialize}
	s.handlers["initialized"] = methodHandler{notif: s.handleInitialized}
	s.handlers["shutdown"] = methodHandler{req: s.handleShutdown}
	s.handlers["exit"] = methodHandler{notif: s.handleExit}
	s.handlers["textDocument/didOpen"] = methodHandler{notif: s.handleDidOpen}
	s.handlers["textDocument/didChange"] = methodHandler{notif: s.handleDidChange}
	s.handlers["textDocument/didClose"] = methodHandler{notif: s.handleDidClose}
	s.handlers["textDocument/completion"] = methodHandler{req: s.handleCompletion}
	s.handlers["textDocument/hover"] = methodHandler{req: s.handleHover}
	s.handlers["textDocument/definition"] = methodHandler{req: s.handleDefinition}
	s.handlers["textDocument/signatureHelp"] = methodHandler{req: s.handleSignatureHelp}

	return s
}

func (s *Server) Run() error {
	s.log("server starting")

	buf := make([]byte, 0, 65536)

	for {
		msg, err := s.readMessage(buf[:0])
		if err != nil {
			if err == io.EOF {
				s.log("stdin closed, exiting")
				return nil
			}
			s.log(fmt.Sprintf("read error: %v", err))
			continue
		}

		var base struct {
			JSONRPC string `json:"jsonrpc"`
			ID      any    `json:"id,omitempty"`
			Method  string `json:"method"`
		}
		if err := json.Unmarshal(msg, &base); err != nil {
			s.log(fmt.Sprintf("parse error: %v", err))
			continue
		}

		if base.ID != nil {
			var req JSONRPCRequest
			if err := json.Unmarshal(msg, &req); err != nil {
				s.log(fmt.Sprintf("request parse error: %v", err))
				s.sendError(base.ID, -32700, "Parse error")
				continue
			}
			s.dispatchRequest(req)
		} else {
			var notif JSONRPCNotification
			if err := json.Unmarshal(msg, &notif); err != nil {
				s.log(fmt.Sprintf("notification parse error: %v", err))
				continue
			}
			s.dispatchNotification(notif)
		}
	}
}

func (s *Server) dispatchRequest(req JSONRPCRequest) {
	if s.shutdown {
		s.sendError(req.ID, -32800, "Server is shutting down")
		return
	}

	handler, ok := s.handlers[req.Method]
	if !ok || handler.req == nil {
		s.sendError(req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method))
		return
	}

	result, err := handler.req(req)
	if err != nil {
		s.sendError(req.ID, -32603, err.Error())
		return
	}

	s.sendResponse(req.ID, result)
}

func (s *Server) dispatchNotification(notif JSONRPCNotification) {
	handler, ok := s.handlers[notif.Method]
	if !ok || handler.notif == nil {
		return
	}
	handler.notif(notif)
}

func (s *Server) readMessage(buf []byte) ([]byte, error) {
	header, err := s.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	header = strings.TrimSuffix(header, "\r\n")
	header = strings.TrimSuffix(header, "\n")

	if !strings.HasPrefix(header, "Content-Length:") {
		return nil, fmt.Errorf("expected Content-Length header, got: %s", header)
	}

	length, err := strconv.Atoi(strings.TrimSpace(header[16:]))
	if err != nil {
		return nil, fmt.Errorf("invalid Content-Length: %v", err)
	}

	empty, err := s.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if empty != "\r\n" && empty != "\n" {
		return nil, fmt.Errorf("expected empty line after header, got: %s", empty)
	}

	body := make([]byte, length)
	_, err = io.ReadFull(s.reader, body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (s *Server) sendRaw(data []byte) {
	fmt.Fprintf(s.writer, "Content-Length: %d\r\n\r\n", len(data))
	s.writer.Write(data)
}

func (s *Server) sendResponse(id any, result any) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	data, _ := json.Marshal(resp)
	s.sendRaw(data)
}

func (s *Server) sendError(id any, code int, message string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &JSONRPCError{Code: code, Message: message},
	}
	data, _ := json.Marshal(resp)
	s.sendRaw(data)
}

func (s *Server) sendNotification(method string, params any) {
	notif := JSONRPCNotification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	data, _ := json.Marshal(notif)
	s.sendRaw(data)
}

type serverCapabilities struct {
	TextDocumentSync   int                    `json:"textDocumentSync"`
	CompletionProvider *completionOptions     `json:"completionProvider,omitempty"`
	HoverProvider      bool                   `json:"hoverProvider"`
	DefinitionProvider bool                   `json:"definitionProvider"`
	SignatureHelpProvider *signatureHelpOptions `json:"signatureHelpProvider,omitempty"`
}

type completionOptions struct {
	TriggerCharacters   []string `json:"triggerCharacters"`
	ResolveProvider     bool     `json:"resolveProvider"`
}

type signatureHelpOptions struct {
	TriggerCharacters   []string `json:"triggerCharacters"`
	RetriggerCharacters []string `json:"retriggerCharacters"`
}

type initializeResult struct {
	Capabilities serverCapabilities `json:"capabilities"`
}

type initializeParams struct {
	ProcessID int               `json:"processId"`
	RootURI   string            `json:"rootUri"`
	Capabilities map[string]any `json:"capabilities"`
}

func (s *Server) handleInitialize(req JSONRPCRequest) (any, error) {
	return initializeResult{
		Capabilities: serverCapabilities{
			TextDocumentSync:   1,
			CompletionProvider: &completionOptions{
				TriggerCharacters: []string{".", ":"},
				ResolveProvider:   false,
			},
			HoverProvider:      true,
			DefinitionProvider: true,
			SignatureHelpProvider: &signatureHelpOptions{
				TriggerCharacters:   []string{"(", ","},
				RetriggerCharacters: []string{","},
			},
		},
	}, nil
}

func (s *Server) handleInitialized(notif JSONRPCNotification) {
	s.initialized = true
	s.log("client initialized")
}

func (s *Server) handleShutdown(req JSONRPCRequest) (any, error) {
	s.shutdown = true
	return nil, nil
}

func (s *Server) handleExit(notif JSONRPCNotification) {
	s.log("server exiting")
	os.Exit(0)
}

type didOpenParams struct {
	TextDocument struct {
		URI     string `json:"uri"`
		LanguageID string `json:"languageId"`
		Version int    `json:"version"`
		Text    string `json:"text"`
	} `json:"textDocument"`
}

type didChangeParams struct {
	TextDocument struct {
		URI     string `json:"uri"`
		Version int    `json:"version"`
	} `json:"textDocument"`
	ContentChanges []struct {
		Text string `json:"text"`
	} `json:"contentChanges"`
}

type didCloseParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
}

func (s *Server) handleDidOpen(notif JSONRPCNotification) {
	var params didOpenParams
	mustUnmarshal(notif.Params, &params)

	doc := params.TextDocument
	s.store.Open(doc.URI, doc.Text, doc.Version)
	entry := s.store.Get(doc.URI)
	if entry != nil {
		s.log(fmt.Sprintf("opened %s (version %d) symbols=%d", doc.URI, doc.Version, len(entry.Symbols)))
	}

	s.publishDiagnostics(doc.URI)
}

func (s *Server) handleDidChange(notif JSONRPCNotification) {
	var params didChangeParams
	mustUnmarshal(notif.Params, &params)

	for _, change := range params.ContentChanges {
		s.store.Update(params.TextDocument.URI, change.Text, params.TextDocument.Version)
	}
	s.log(fmt.Sprintf("updated %s (version %d)", params.TextDocument.URI, params.TextDocument.Version))

	s.publishDiagnostics(params.TextDocument.URI)
}

func (s *Server) handleDidClose(notif JSONRPCNotification) {
	var params didCloseParams
	mustUnmarshal(notif.Params, &params)

	s.store.Close(params.TextDocument.URI)

	s.sendNotification("textDocument/publishDiagnostics", map[string]any{
		"uri":         params.TextDocument.URI,
		"diagnostics": []Diagnostic{},
	})
	s.log(fmt.Sprintf("closed %s", params.TextDocument.URI))
}

func (s *Server) publishDiagnostics(uri string) {
	entry := s.store.Get(uri)
	if entry == nil {
		return
	}

	diags := GetDiagnostics(entry)
	s.sendNotification("textDocument/publishDiagnostics", map[string]any{
		"uri":         uri,
		"diagnostics": diags,
	})
}

type textDocumentParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
}

type positionParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Position Position `json:"position"`
}

func (s *Server) handleCompletion(req JSONRPCRequest) (any, error) {
	var params positionParams
	mustUnmarshal(req.Params, &params)

	entry := s.store.Get(params.TextDocument.URI)
	if entry == nil {
		return nil, nil
	}

	result := GetCompletions(entry, params.Position)
	if result != nil {
		s.log(fmt.Sprintf("completion at %d:%d returned %d items", params.Position.Line, params.Position.Character, len(result.Items)))
	}
	return result, nil
}

func (s *Server) handleHover(req JSONRPCRequest) (any, error) {
	var params positionParams
	mustUnmarshal(req.Params, &params)

	entry := s.store.Get(params.TextDocument.URI)
	if entry == nil {
		return nil, nil
	}

	return GetHover(entry, params.Position), nil
}

func (s *Server) handleDefinition(req JSONRPCRequest) (any, error) {
	var params positionParams
	mustUnmarshal(req.Params, &params)

	entry := s.store.Get(params.TextDocument.URI)
	if entry == nil {
		return nil, nil
	}

	result := GetDefinition(entry, params.Position)
	return result, nil
}

func (s *Server) handleSignatureHelp(req JSONRPCRequest) (any, error) {
	var params positionParams
	mustUnmarshal(req.Params, &params)

	entry := s.store.Get(params.TextDocument.URI)
	if entry == nil {
		return nil, nil
	}

	result := GetSignatureHelp(entry, params.Position)
	return result, nil
}

func mustUnmarshal(raw any, target any) {
	data, err := json.Marshal(raw)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		panic(err)
	}
}

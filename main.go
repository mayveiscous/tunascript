package main

import (
	"fmt"
	"os"
	"tunascript/src/lexer"
	tunaparser "tunascript/src/parser"
	"tunascript/src/interpreter"
	"tunascript/src/analyzer"
	"tunascript/src/directives"
	"tunascript/src/repl"
	"tunascript/src/lsp"
)

func runFile(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: tuna <file.tuna>")
		os.Exit(1)
	}

	filePath := args[0]

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[31m[Tunascript Error]\033[0m Could not read file '%s': %v\n", filePath, err)
		os.Exit(1)
	}

	source := string(bytes)
	cfg := directives.Extract(source)

	defer func() {
		if r := recover(); r != nil {
			if tunaErr, ok := r.(*lexer.TunaError); ok {
				fmt.Fprintln(os.Stderr, tunaErr.Error())
			} else {
				fmt.Fprintf(os.Stderr, "\n\033[31m[Tunascript Runtime Error]\033[0m %v\n", r)
			}
			os.Exit(1)
		}
	}()

	tokens := lexer.Lex(source, filePath)
	tree := tunaparser.Parse(tokens, filePath)

	diagnostics := analyzer.Analyze(tree, cfg, filePath)
	hasErrors := false
	for _, d := range diagnostics {
		fmt.Fprintln(os.Stderr, d.String())
		if d.Level == analyzer.DiagError {
			hasErrors = true
		}
	}
	if hasErrors {
		os.Exit(1)
	}

	interpreter.Interpret(tree, filePath, cfg)
}

const Version = "0.1.0"

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Printf("Tunascript v%s\n", Version)
		printUsage()
		return
	}

	command := args[0]

	switch command {
	case "run":
		runFile(args[1:])
	case "serve":
		repl.Start()
	case "lsp":
		server := lsp.NewServer()
		if err := server.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "LSP server error: %v\n", err)
			os.Exit(1)
		}
	default:
		runFile(args)
	}
}

func printUsage() {
	fmt.Printf(`Usage:
  tuna                   Print version and usage
  tuna <file.tuna>       Run a script
  tuna run <file.tuna>   Run a script
  tuna serve             Start interactive REPL
  tuna lsp               Start LSP language server
`)
}
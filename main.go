package main

import (
	"fmt"
	"os"
	"test-go/src/interpreter"
	"test-go/src/lexer"
	tunaparser "test-go/src/parser"
)

func runFile(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: tunascript swim <file.tuna>")
		os.Exit(1)
	}

	filePath := args[0]

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[31m[TunaScript Error]\033[0m Could not read file '%s': %v\n", filePath, err)
		os.Exit(1)
	}

	source := string(bytes)

	defer func() {
		if r := recover(); r != nil {
			if tunaErr, ok := r.(*lexer.TunaError); ok {
				fmt.Fprintln(os.Stderr, tunaErr.Error())
			} else {
				fmt.Fprintf(os.Stderr, "\n\033[31m[TunaScript Runtime Error]\033[0m %v\n", r)
			}
			os.Exit(1)
		}
	}()

	tokens := lexer.Lex(source)
	tree := tunaparser.Parse(tokens)
	interpreter.Interpret(tree, os.Args[1])
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printUsage()
		return
	}

	command := args[0]
	commandArgs := args[1:]

	switch command {
	case "run":
		runFile(commandArgs)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`TunaScript CLI
Usage:
  tunascript run <file.tuna>   Run a script
  tunascript lex <file.tuna>    Print tokens (not implemented)
  tunascript serve              Start REPL (not implemented)`)
}

/*

case "serve":
	runREPL(commandArgs)
case "lex":
	runLexer(commandArgs)

*/
package main

import (
	"fmt"
	"os"
	"tunascript/src/lexer"
	tunaparser "tunascript/src/parser"
	"tunascript/src/interpreter"
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

	tokens := lexer.Lex(source)
	tree := tunaparser.Parse(tokens)
	interpreter.Interpret(tree, filePath)
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printUsage()
		return
	}

	command := args[0]

	switch command {
	case "run":
		runFile(args[1:])
	default:
		runFile(args)
	}
}

func printUsage() {
	fmt.Println(`Tunascript CLI
Usage:
  tuna <file.tuna>   Run a script
  tuna serve              Start REPL (not implemented)`)
}
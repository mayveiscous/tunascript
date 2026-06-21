package repl

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"tunascript/src/lexer"
	tunaparser "tunascript/src/parser"
	"tunascript/src/interpreter"
	"tunascript/src/directives"
)

func Start() {
	fmt.Println("Tunascript REPL — type 'exit' to quit")
	scanner := bufio.NewScanner(os.Stdin)
	buf := strings.Builder{}
	cfg := directives.Config{}

	env := interpreter.NewREPLEnvironment()

	for {
		lineCount := strings.Count(buf.String(), "\n")
		if buf.Len() == 0 {
			fmt.Print("> ")
		} else {
			fmt.Printf("%d| ", lineCount+1)
		}

		if !scanner.Scan() {
			break
		}

		line := scanner.Text()

		if strings.TrimSpace(line) == "exit" {
			break
		}

		if strings.TrimSpace(line) == "" && buf.Len() > 0 {
			source := buf.String()
			buf.Reset()
			execREPL(source, env, cfg)
			continue
		}

		buf.WriteString(line)
		buf.WriteByte('\n')

		if !isComplete(buf.String()) {
			continue
		}

		source := buf.String()
		buf.Reset()
		execREPL(source, env, cfg)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "REPL error: %v\n", err)
	}
}

func execREPL(source string, env *interpreter.Environment, cfg directives.Config) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case *lexer.TunaError:
				fmt.Fprintln(os.Stderr, v.Error())
			case string:
				fmt.Fprintln(os.Stderr, v)
			default:
				fmt.Fprintf(os.Stderr, "%v\n", v)
			}
		}
	}()

	tokens := lexer.Lex(source, "<repl>")
	tree := tunaparser.Parse(tokens, "<repl>")

	interpreter.SetCurrentFile("<repl>")

	for _, stmt := range tree.Body {
		interpreter.EvaluateStatement(stmt, env, interpreter.ExecContext{})
	}
}

var openers = map[rune]rune{
	'(':	')',
	'[':	']',
	'{':	'}',
}

var closers = map[rune]bool{
	')':	true,
	']':	true,
	'}':	true,
}

func isComplete(src string) bool {
	stack := []rune{}
	inStr := false
	for _, ch := range src {
		if ch == '"' {
			inStr = !inStr
			continue
		}
		if inStr {
			continue
		}
		if closer, ok := openers[ch]; ok {
			stack = append(stack, closer)
		} else if closers[ch] {
			if len(stack) == 0 || stack[len(stack)-1] != ch {
				stack = append(stack, ch)
			} else {
				stack = stack[:len(stack)-1]
			}
		}
	}
	return len(stack) == 0 && !inStr
}

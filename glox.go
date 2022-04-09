package glox

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
)

type Runtime struct {
	hadError bool
}

func NewRuntime() *Runtime {
	return &Runtime{
		hadError: false,
	}
}

func (r *Runtime) Run(args []string) {
	if len(args) > 1 {
		fmt.Println("Usage: glox [script]")
		os.Exit(64)
	} else if len(args) == 1 {
		r.RunFile(args[0])
	} else {
		r.RunPrompt()
	}
}

func (r *Runtime) RunFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(fmt.Sprintf("error reading file: %s", err.Error()))
		return
	}

	r.run(string(data))

	if r.hadError {
		os.Exit(65)
	}
}

func (r *Runtime) RunPrompt() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(">>> ")

		// Scans a line from the standard input
		scanner.Scan()
		line := scanner.Text()

		if line == "" {
			break
		}

		r.run(line)
		r.hadError = false
	}
}

func (r *Runtime) Error(line int, message string) {
	r.report(line, "", message)
}

func (r *Runtime) run(source string) {
	scanner := NewScanner(bytes.NewBuffer([]byte(source)), r)
	tokens := scanner.ScanTokens()

	parser := NewParser(tokens, r)
	expr := parser.Parse()

	printer := &AstPrinter{}
	if r.hadError {
		return
	}
	
	fmt.Println(printer.Print(expr))

	// for _, token := range tokens {
	// 	if token.Type == String {
	// 		fmt.Println(token.ToString())
	// 	}
	// }
}

func (r *Runtime) report(line int, where string, message string) {
	errMessage := fmt.Sprintf("[line %d] Error%s: %s", line, where, message)
	r.hadError = true
	fmt.Println(errMessage)
}

func (r *Runtime) tokenError(token Token, message string) {
	if token.Type == Eof {
		r.report(token.Line, " at end ", message)
	} else {
		r.report(token.Line, " at '" + token.Lexeme + "'", message)
	}
}

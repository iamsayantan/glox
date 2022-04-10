package glox

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
)

var interpreter *Interpreter

type Runtime struct {
	hadError        bool
	hadRuntimeError bool
}

func NewRuntime() *Runtime {
	r := &Runtime{
		hadError: false,
	}

	interpreter = NewInterpreter(r)
	return r
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

	if r.hadRuntimeError {
		os.Exit(70)
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

	if r.hadError {
		return
	}

	interpreter.Interpret(expr)
}

func (r *Runtime) report(line int, where string, message string) {
	errMessage := fmt.Sprintf("[line %d] Error%s: %s", line, where, message)
	r.hadError = true
	fmt.Println(errMessage)
}

func (r *Runtime) runtimeError(err error) {
	runErr := err.(*RuntimeError)
	fmt.Printf("%s \n[line %d ]\n", runErr.Error(), runErr.token.Line)
	r.hadRuntimeError = true
}

func (r *Runtime) tokenError(token Token, message string) {
	if token.Type == Eof {
		r.report(token.Line, " at end ", message)
	} else {
		r.report(token.Line, " at '"+token.Lexeme+"'", message)
	}
}

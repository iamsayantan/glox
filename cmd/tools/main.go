package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/iamsayantan/glox"
	"github.com/iamsayantan/glox/tools"
)

func main() {
	args := os.Args[1:]
	err := tools.GenerateAst(args)
	if err != nil {
		if errors.Is(err, tools.ErrInvalidArgumentList) {
			fmt.Println("Usage: generate_ast <output dir>")
			os.Exit(64)
		}

		fmt.Println("Error generating AST: ", err.Error())
	}

	ap := glox.AstPrinter{}
	ast, _ := ap.Print()
	fmt.Println(ast)
}

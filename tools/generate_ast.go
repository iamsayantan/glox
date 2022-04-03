package tools

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

var (
	// ErrInvalidArgumentList is returned when the arguments count do not match the expected count
	ErrInvalidArgumentList = errors.New("invalid arguments provided")
)

func GenerateAst(args []string) error {
	if len(args) != 1 {
		return ErrInvalidArgumentList
	}

	outputDir := args[0]
	err := defineAst(outputDir, "Expr", []string{
		"Binary : Left Expr, Operator Token, Right Expr",
		"Grouping : Expression Expr",
		"Literal : Value interface{}",
		"Unary : Operator Token, Right Expr",
	})

	if err != nil {
		return err
	}

	return nil
}

func defineAst(outputDir, baseName string, astTypes []string) error {
	path := outputDir + "/" + strings.ToLower(baseName) + ".go"

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(f)

	w.WriteString("package glox\n\n")
	w.WriteString("type " + baseName + " struct {\n")
	w.WriteString("}\n\n")

	for _, astType := range astTypes {
		typeName := strings.Trim(strings.Split(astType, ":")[0], " ")
		fields := strings.Trim(strings.Split(astType, ":")[1], " ")
		defineType(w, baseName, typeName, fields)
	}

	err = w.Flush()

	if err != nil {
		return err
	}

	return nil
}

func defineType(w *bufio.Writer, baseName, typeName, fieldList string) {
	w.WriteString("type " + typeName + " struct { \n")
	w.WriteString("    " + baseName + "\n")

	fields := strings.Split(fieldList, ", ")
	for _, field := range fields {
		w.WriteString("    " + field + "\n")
	}

	w.WriteString("}\n\n")
}
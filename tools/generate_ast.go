package tools

import (
	"bufio"
	"errors"
	"fmt"
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
	// err := defineAst(outputDir, "Expr", []string{
	// 	"Binary : Left Expr, Operator Token, Right Expr",
	// 	"Grouping : Expression Expr",
	// 	"Literal : Value interface{}",
	// 	"Unary : Operator Token, Right Expr",
	// })

	err := defineAst(outputDir, "Stmt", []string{
		"Expression : Expression Expr",
		"Print : Expression Expr",
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
	w.WriteString("type " + baseName + " interface {\n")
	w.WriteString("    Accept(visitor Visitor" + baseName +") (interface{}, error)\n")
	w.WriteString("}\n\n")

	defineVisitor(w, baseName, astTypes)

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

func defineVisitor(w *bufio.Writer, baseName string, astTypes []string) {
	w.WriteString("type " + baseName + "Visitor interface {\n")
	for _, astType := range astTypes {
		typeName := strings.Trim(
			strings.Split(astType, ":")[0],
			" ",
		)
		w.WriteString(fmt.Sprintf("    Visit%sExpr(expr *%s) (interface{}, error)\n", typeName, typeName))
	}

	w.WriteString("}\n\n")
}

func defineType(w *bufio.Writer, baseName, typeName, fieldList string) {
	w.WriteString("type " + typeName + " struct { \n")

	fields := strings.Split(fieldList, ", ")
	for _, field := range fields {
		w.WriteString("    " + field + "\n")
	}

	w.WriteString("}\n\n")

	// define the Accept method so it implements the base interface
	typeAsParam := strings.ToLower(string([]rune(typeName)[0])) // the first character from the type will be used as receiver parameter

	w.WriteString(fmt.Sprintf("func (%s *%s) Accept(visitor %sVisitor) (interface{}, error) {\n", typeAsParam, typeName, baseName))
	w.WriteString(fmt.Sprintf("    return visitor.Visit%sExpr(%s)\n", typeName, typeAsParam))
	w.WriteString("}\n\n")
}

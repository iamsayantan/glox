package glox

import (
	"fmt"
	"strings"
)

type AstPrinter struct{}

func (ap *AstPrinter) Print() string {
	exp := &Binary{
		Left: &Unary{
			Operator: NewToken(Minus, "-", nil, 1),
			Right:    &Literal{Value: 123},
		},
		Operator: NewToken(Star, "*", nil, 1),
		Right: &Grouping{
			Expression: &Literal{Value: 45.67},
		},
	}
	
	return exp.Accept(ap).(string)
}

func (ap *AstPrinter) VisitBinaryExpr(expr *Binary) interface{} {
	return ap.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)
}

func (ap *AstPrinter) VisitGroupingExpr(expr *Grouping) interface{} {
	return ap.parenthesize("group", expr.Expression)
}

func (ap *AstPrinter) VisitLiteralExpr(expr *Literal) interface{} {
	if expr.Value == nil {
		return "nil"
	}

	return fmt.Sprintf("%v", expr.Value)
}

func (ap *AstPrinter) VisitUnaryExpr(expr *Unary) interface{} {
	return ap.parenthesize(expr.Operator.Lexeme, expr.Right)
}

func (ap *AstPrinter) parenthesize(name string, exprs ...Expr) string {
	s := strings.Builder{}
	s.WriteString("(" + name)

	for _, expr := range exprs {
		s.WriteString(" ")
		s.WriteString(expr.Accept(ap).(string))
	}

	s.WriteString(")")
	return s.String()
}

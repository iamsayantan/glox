package glox

type AstPrinter struct{}

// func (ap *AstPrinter) Print(expr Expr) (string, error) {
// 	val, err := expr.Accept(ap)
// 	if err != nil {
// 		return "", err
// 	}

// 	return val.(string), nil
// }

// func (ap *AstPrinter) VisitBinaryExpr(expr *Binary) (interface{}, error) {
// 	return ap.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right), nil
// }

// func (ap *AstPrinter) VisitGroupingExpr(expr *Grouping) (interface{}, error) {
// 	return ap.parenthesize("group", expr.Expression), nil
// }

// func (ap *AstPrinter) VisitLiteralExpr(expr *Literal) (interface{}, error) {
// 	if expr.Value == nil {
// 		return "nil", nil
// 	}

// 	return fmt.Sprintf("%v", expr.Value), nil
// }

// func (ap *AstPrinter) VisitUnaryExpr(expr *Unary) (interface{}, error) {
// 	return ap.parenthesize(expr.Operator.Lexeme, expr.Right), nil
// }

// func (ap *AstPrinter) parenthesize(name string, exprs ...Expr) string {
// 	s := strings.Builder{}
// 	s.WriteString("(" + name)

// 	for _, expr := range exprs {
// 		s.WriteString(" ")
// 		val,_ := expr.Accept(ap)
// 		s.WriteString(val.(string))
// 	}

// 	s.WriteString(")")
// 	return s.String()
// }

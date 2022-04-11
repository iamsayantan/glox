package glox

// Stmt is the interface for lox statements. There are no place in the grammar
// where both expressions and statements are allowed. E.g. the both operands for
// the + operator must be expressions, the body of while loop is always statements.
// Making a separate interface for statements will forbid us to pass statements
// where an expression was required or vice versa.
type Stmt interface {
	Accept(visitor StmtVisitor) error
}

type StmtVisitor interface {
	VisitExpressionExpr(expr *Expression) error
	VisitPrintExpr(expr *Print) error
}

type Expression struct {
	Expression Expr
}

func (e *Expression) Accept(visitor StmtVisitor) error {
	return visitor.VisitExpressionExpr(e)
}

type Print struct {
	Expression Expr
}

func (p *Print) Accept(visitor StmtVisitor) error {
	return visitor.VisitPrintExpr(p)
}

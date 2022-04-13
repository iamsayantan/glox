package glox

type Expr interface {
	Accept(visitor Visitor) (interface{}, error)
}

type Visitor interface {
	VisitAssignExpr(expr *Assign) (interface{}, error)
	VisitBinaryExpr(expr *Binary) (interface{}, error)
	VisitGroupingExpr(expr *Grouping) (interface{}, error)
	VisitLiteralExpr(expr *Literal) (interface{}, error)
	VisitUnaryExpr(expr *Unary) (interface{}, error)
	VisitVarExpr(expr *VarExpr) (interface{}, error)
}

type Assign struct {
	Name  Token
	Value Expr
}

func (a *Assign) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitAssignExpr(a)
}

type Binary struct {
	Left     Expr
	Operator Token
	Right    Expr
}

func (b *Binary) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitBinaryExpr(b)
}

type Grouping struct {
	Expression Expr
}

func (g *Grouping) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitGroupingExpr(g)
}

type Literal struct {
	Value interface{}
}

func (l *Literal) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitLiteralExpr(l)
}

type Unary struct {
	Operator Token
	Right    Expr
}

func (u *Unary) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitUnaryExpr(u)
}

type VarExpr struct {
	Name Token
}

func (v *VarExpr) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitVarExpr(v)
}

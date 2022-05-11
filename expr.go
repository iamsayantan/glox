package glox

type Expr interface {
	Accept(visitor Visitor) (interface{}, error)
}

type Visitor interface {
	VisitAssignExpr(expr *Assign) (interface{}, error)
	VisitLogicalExpr(expr *Logical) (interface{}, error)
	VisitBinaryExpr(expr *Binary) (interface{}, error)
	VisitCallExpr(expr *Call) (interface{}, error)
	VisitGroupingExpr(expr *Grouping) (interface{}, error)
	VisitLiteralExpr(expr *Literal) (interface{}, error)
	VisitUnaryExpr(expr *Unary) (interface{}, error)
	VisitVarExpr(expr *VarExpr) (interface{}, error)
	VisitGetExpr(expr *GetExpr) (interface{}, error)
	VisitSetExpr(expr *SetExpr) (interface{}, error)
	VisitThisExpr(expr *ThisExpr) (interface{}, error)
	VisitSuperExpr(expr *SuperExpr) (interface{}, error)
}

type Assign struct {
	Name  Token
	Value Expr
}

func (a *Assign) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitAssignExpr(a)
}

type Logical struct {
	Left Expr
	Operator Token
	Right Expr
}

func (l *Logical) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitLogicalExpr(l)
}

type Binary struct {
	Left     Expr
	Operator Token
	Right    Expr
}

func (b *Binary) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitBinaryExpr(b)
}

type Call struct {
	Callee Expr
	Paren Token
	Arguments []Expr
}

func (c *Call) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitCallExpr(c)
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

type GetExpr struct {
	Object Expr
	Name Token
}

func (g *GetExpr) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitGetExpr(g)
}

type SetExpr struct {
	Object Expr
	Name Token
	Value Expr
}

func (se *SetExpr) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitSetExpr(se)
}

type ThisExpr struct {
	Keyword Token
}

func (th *ThisExpr) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitThisExpr(th)
}

type SuperExpr struct {
	Keyword Token
	Method Token
}

func (se *SuperExpr) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitSuperExpr(se)
}
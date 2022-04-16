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
	VisitBlockStmt(stmt *Block) error
	VisitExpressionExpr(expr *Expression) error
	VisitPrintExpr(expr *Print) error
	VisitVarStmt(expr *VarStmt) error
	VisitIfStmt(stmt *IfStmt) error
	VisitWhileStmt(stmt *WhileStmt) error
	VisitFunctionStmt(stmt *FunctionStmt) error
	VisitReturnStmt(stmt *ReturnStmt) error
}

type Block struct {
	Statements []Stmt
}

func (b *Block) Accept(visitor StmtVisitor) error {
	return visitor.VisitBlockStmt(b)
}

type Expression struct {
	Expression Expr
}

func (e *Expression) Accept(visitor StmtVisitor) error {
	return visitor.VisitExpressionExpr(e)
}

type FunctionStmt struct {
	Name Token
	Params []Token
	Body []Stmt
}

func (f *FunctionStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitFunctionStmt(f)
}

type IfStmt struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

func (i *IfStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitIfStmt(i)
}


type Print struct {
	Expression Expr
}

func (p *Print) Accept(visitor StmtVisitor) error {
	return visitor.VisitPrintExpr(p)
}

type VarStmt struct {
	Name        Token
	Initializer Expr
}

func (v *VarStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitVarStmt(v)
}

type WhileStmt struct {
	Condition Expr
	Body Stmt
}

func (w *WhileStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitWhileStmt(w)
}

type ReturnStmt struct {
	Keyword Token
	Value Expr
}

func (r *ReturnStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitReturnStmt(r)
}
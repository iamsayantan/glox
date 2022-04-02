package glox

type Expr struct {
}

type Binary struct {
	Left     Expr
	Operator Token
	Right    Expr
}

type Grouping struct {
	Expression Expr
}

type Literal struct {
	Value interface{}
}

type Unary struct {
	Operator Token
	Right    Expr
}

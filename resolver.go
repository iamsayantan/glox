package glox

import (
	"github.com/iamsayantan/glox/util"
)

type FunctionType int

type ClassType int

const (
	FunctionTypeNone FunctionType = iota
	FunctionTypeFunction
	FunctionTypeMethod
	FunctionTypeInitializer
)

const (
	ClassTypeNone ClassType = iota
	ClassTypeClass
)

type Resolver struct {
	interpreter *Interpreter
	// scopes keeps track of the stack of scopes currently in scope. Each element
	// in the stack is a map representing a new block scope. Keys, like in
	// environment is the variable name, the value is boolean used to track if we
	// have finished resolving the variable's initializer. The scope stack only keep
	// tracks of the block scopes, variables declared in the top level are not tracked
	// by the resolver since they are more dynamic in Lox. While resolving a variable if
	// we don't find it in the stack of global scopes, we assume it must be global.
	scopes util.Stack[map[string]bool]

	currentFunction FunctionType
	currentClass    ClassType

	runtime *Runtime
}

func NewResolver(i *Interpreter, runtime *Runtime) *Resolver {
	stack := util.NewStack[map[string]bool]()
	return &Resolver{interpreter: i, scopes: *stack, runtime: runtime, currentFunction: FunctionTypeNone, currentClass: ClassTypeNone}
}

// VisitAssignExpr resolves an assignment expression, first we resolve the expression for
// the assigned value in case it also contains references to other variables. Then we use
// our existing resolveLocal() method to resolve the variable that's being assigned to.
func (r *Resolver) VisitAssignExpr(expr *Assign) (interface{}, error) {
	_, err := r.resolveExpr(expr.Value)
	if err != nil {
		return nil, err
	}

	r.resolveLocal(expr, expr.Name)
	return nil, nil
}

func (r *Resolver) VisitLogicalExpr(expr *Logical) (interface{}, error) {
	// Since static analysis does no control flow or short circuiting, logical expression is
	// exactly same as other binary operators.
	r.resolveExpr(expr.Left)
	r.resolveExpr(expr.Right)

	return nil, nil
}

func (r *Resolver) VisitBinaryExpr(expr *Binary) (interface{}, error) {
	r.resolveExpr(expr.Left)
	r.resolveExpr(expr.Right)

	return nil, nil
}

func (r *Resolver) VisitCallExpr(expr *Call) (interface{}, error) {
	r.resolveExpr(expr.Callee)

	for _, argument := range expr.Arguments {
		r.resolveExpr(argument)
	}

	return nil, nil
}

func (r *Resolver) VisitGroupingExpr(expr *Grouping) (interface{}, error) {
	r.resolveExpr(expr.Expression)

	return nil, nil
}

func (r *Resolver) VisitLiteralExpr(expr *Literal) (interface{}, error) {
	// A literal does not mention any variables and does not contain any subexpression.
	// So there is no work to do here.
	return nil, nil
}

func (r *Resolver) VisitUnaryExpr(expr *Unary) (interface{}, error) {
	r.resolveExpr(expr.Right)

	return nil, nil
}

// VisitVarExpr resolves a variable. Variable declaration and function declaration
// writes to the scope maps, those maps are read when we read the variable. First we check
// to  see if the variable is being accessed inside it's own initializer. If the variable
// exists in the current scope but its value is false, that means we have declared it but
// not yet defined it. We report that error.
func (r *Resolver) VisitVarExpr(expr *VarExpr) (interface{}, error) {
	if !r.scopes.IsEmpty() {
		scope, err := r.scopes.Peek()
		if err == nil {
			if val, ok := scope[expr.Name.Lexeme]; ok && !val {
				r.runtime.tokenError(expr.Name, "Can't read local variable in its own initializer.")
			}
		}
	}

	r.resolveLocal(expr, expr.Name)
	return nil, nil
}

func (r *Resolver) VisitClassStmt(stmt *ClassStmt) error {
	r.declare(stmt.Name)
	r.define(stmt.Name)

	if stmt.Superclass != nil && stmt.Superclass.Name.Lexeme == stmt.Name.Lexeme {
		r.runtime.tokenError(stmt.Superclass.Name, "A class can't inherit from itself.")
	}

	if stmt.Superclass != nil {
		r.resolveExpr(stmt.Superclass)
	}

	enclosingClass := r.currentClass
	r.currentClass = ClassTypeClass
	// we resolve "this" exactly like any other local variable, using "this" as the name.
	// Before we start resolving the method bodies, we push a new scope and define "this"
	// in it as any other variable. Then when we are done, we discard the surrounding scope.
	r.beginScope()

	scope, err := r.scopes.Peek()
	if err != nil {
		return err
	}

	scope["this"] = true

	for _, method := range stmt.Methods {
		declaration := FunctionTypeMethod
		if method.Name.Lexeme == "init" {
			declaration = FunctionTypeInitializer
		}

		r.resolveFunction(method, declaration)
	}

	r.endScope()

	r.currentClass = enclosingClass
	return nil
}

func (r *Resolver) VisitThisExpr(expr *ThisExpr) (interface{}, error) {
	if r.currentClass == ClassTypeNone {
		r.runtime.tokenError(expr.Keyword, "Can't use 'this' outside of a class.")
		return nil, nil
	}

	r.resolveLocal(expr, expr.Keyword)
	return nil, nil
}

func (r *Resolver) VisitGetExpr(expr *GetExpr) (interface{}, error) {
	return r.resolveExpr(expr.Object)
}

func (r *Resolver) VisitSetExpr(expr *SetExpr) (interface{}, error) {
	_, err := r.resolveExpr(expr.Value)
	if err != nil {
		return nil, err
	}

	_, err = r.resolveExpr(expr.Object)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// VisitBlockStmt will visit a block statement which will create a new lexical scope,
// traverse the statements inside the block and then discard the scope.
func (r *Resolver) VisitBlockStmt(stmt *Block) error {
	r.beginScope()
	err := r.resolveStatements(stmt.Statements)
	if err != nil {
		return err
	}

	r.endScope()
	return nil
}

func (r *Resolver) VisitExpressionExpr(expr *Expression) error {
	r.resolveExpr(expr.Expression)
	return nil
}

func (r *Resolver) VisitPrintExpr(expr *Print) error {
	r.resolveExpr(expr.Expression)
	return nil
}

// VisitVarStmt resolves a variable declaration statement. Resolving a variable statement
// will add a new entry to the current innermost scopes map. As we visit expression we need
// to know if we are inside the initializer for some variable. We do that by splitting binding
// in two steps, the first is declaring it.
func (r *Resolver) VisitVarStmt(stmt *VarStmt) error {
	r.declare(stmt.Name)
	if stmt.Initializer != nil {
		_, err := r.resolveExpr(stmt.Initializer)
		if err != nil {
			return err
		}
	}

	r.define(stmt.Name)
	return nil
}

// VisitIfStmt resolves an if statement. It has one expression for its condition and one or two
// statements for the branches. The resolution is different from interpretetion here, when we
// resolve an if statement, there is no control flow. We resolve the condition and both the
// branches.
func (r *Resolver) VisitIfStmt(stmt *IfStmt) error {
	r.resolveExpr(stmt.Condition)
	r.resolveStmt(stmt.ThenBranch)
	if stmt.ElseBranch != nil {
		r.resolveStmt(stmt.ElseBranch)
	}

	return nil
}

// VisitWhileStmt will resolve a while statement. It resolves both the condition and the body
// exactly once.
func (r *Resolver) VisitWhileStmt(stmt *WhileStmt) error {
	r.resolveExpr(stmt.Condition)
	r.resolveStmt(stmt.Body)

	return nil
}

// VisitFunctionStmt resolves a function declaration. Functions both bind names and introduce
// a scope. The name of the function itself is bound in the surrounding scope where the function
// is declared. When we step into the function's body, we also bind its parameters into the inner
// function scope.
func (r *Resolver) VisitFunctionStmt(stmt *FunctionStmt) error {
	// We declare and define the name of the function in the current scope. Unlike variables, though
	// we define the name eagerly, before resolving the function's body. This lets a function recursively
	// refer to itself inside its own body.
	r.declare(stmt.Name)
	r.define(stmt.Name)

	r.resolveFunction(stmt, FunctionTypeFunction)
	return nil
}

func (r *Resolver) VisitReturnStmt(stmt *ReturnStmt) error {
	if r.currentFunction == FunctionTypeNone {
		r.runtime.tokenError(stmt.Keyword, "Can't return from top-level code")
	}

	if stmt.Value != nil {
		if r.currentFunction == FunctionTypeInitializer {
			r.runtime.tokenError(stmt.Keyword, "Can't return a value from initializer.")
			return nil
		}

		r.resolveExpr(stmt.Value)
	}

	return nil
}

func (r *Resolver) resolveStatements(statements []Stmt) error {
	for _, stmt := range statements {
		err := r.resolveStmt(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Resolver) resolveStmt(statement Stmt) error {
	return statement.Accept(r)
}

func (r *Resolver) resolveExpr(expr Expr) (interface{}, error) {
	return expr.Accept(r)
}

// beginScope creates a new scope and pushes it into the stack.
func (r *Resolver) beginScope() {
	r.scopes.Push(make(map[string]bool))
}

func (r *Resolver) endScope() {
	r.scopes.Pop()
}

// declare adds a variable to the innermost scope so that it shadows any outer
// one and so we know that the variable exists. We mark it as "not ready yet"
// by binding the name as false in the scope map.
func (r *Resolver) declare(name Token) {
	if r.scopes.IsEmpty() {
		return
	}

	scope, _ := r.scopes.Peek()

	// when we declare a variable in a local scope, we already know the names of
	// every previously declared variables in that same scope. If we see collision
	// we report an error.
	if _, ok := scope[name.Lexeme]; ok {
		r.runtime.tokenError(name, "Already a variable with this name in this scope")
	}

	scope[name.Lexeme] = false
}

// define marks a variable as ready for use. This essentially means that the
// variable is fully initialized.
func (r *Resolver) define(name Token) {
	if r.scopes.IsEmpty() {
		return
	}

	scope, _ := r.scopes.Peek()
	scope[name.Lexeme] = true
}

// resolveLocal resolves a variable in the stack of local scopes. We start at the innermost
// scope and work our way outwards, looking at each map for a matching name. If we find it
// we resolve it, passing in the number of scopes between the current innermost scope and the
// scope where the variable was found. If we walk thorough all the scopes and never find the
// variable, we assume its global.
func (r *Resolver) resolveLocal(expr Expr, name Token) {
	for i := r.scopes.Size() - 1; i >= 0; i-- {
		val, _ := r.scopes.Get(i)
		if _, ok := val[name.Lexeme]; ok {
			r.interpreter.resolve(expr, r.scopes.Size()-1-i)
			return
		}
	}
}

// resolveFunction resolves a function's body. It creates a new scope for the body and then binds
// variables for each of the function's parameters. Once that's done, it resolves the function's
// body in the scope. The difference from how interpreter handles is that, at runtime, declaring
// a function doesn't do anything to the function's body, the body doesn't get touched until the
// function is called. But in static analysis we immediately traverse into the body.
func (r *Resolver) resolveFunction(function *FunctionStmt, funcType FunctionType) {
	// We stash the previous value of the field in a local variable first. As Lox has local functions,
	// we can nest function declaration arbitrarily deeply. We need to track not just we are in a
	// function, but how many we're in.
	enclosingFunction := r.currentFunction
	r.currentFunction = funcType

	r.beginScope()
	for _, param := range function.Params {
		r.declare(param)
		r.define(param)
	}

	r.resolveStatements(function.Body)
	r.endScope()

	r.currentFunction = enclosingFunction
}

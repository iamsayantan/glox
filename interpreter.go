package glox

import (
	"fmt"

	"github.com/iamsayantan/glox/tools"
)

type Interpreter struct {
	runtime     *Runtime
	globals     *Environment
	environment *Environment
	locals      map[Expr]int
}

func NewInterpreter(runtime *Runtime) *Interpreter {
	global := NewEnvironment(nil)
	global.Define("clock", Clock{})
	return &Interpreter{runtime: runtime, environment: global, globals: global, locals: make(map[Expr]int)}
}

type RuntimeError struct {
	token   Token
	message string
}

func (r *RuntimeError) Error() string {
	return r.message
}

func NewRuntimeError(token Token, message string) error {
	return &RuntimeError{token: token, message: message}
}

type ReturnErr struct {
	Value interface{}
}

func (re *ReturnErr) Error() string {
	return ""
}

func NewReturn(value interface{}) *ReturnErr {
	return &ReturnErr{Value: value}
}

func (i *Interpreter) Interpret(statements []Stmt) {
	for _, stmt := range statements {
		err := i.execute(stmt)
		if err != nil {

			i.runtime.runtimeError(err)
			return
		}
	}
}

func (i *Interpreter) execute(stmt Stmt) error {
	err := stmt.Accept(i)
	if err != nil {
		return err
	}

	return nil
}

func (i *Interpreter) VisitClassStmt(stmt *ClassStmt) error {
	i.environment.Define(stmt.Name.Lexeme, nil)
	klass := NewLoxClass(stmt.Name.Lexeme)
	i.environment.Assign(stmt.Name, klass)

	return nil
}

func (i *Interpreter) VisitGetExpr(expr *GetExpr) (interface{}, error) {
	object, err := i.evaluate(expr.Object)
	if err != nil {
		return nil, err
	}

	if loxInstance, ok := object.(*LoxInstance); ok {
		return loxInstance.Get(expr.Name)
	}

	return nil, NewRuntimeError(expr.Name, "Only instances have properties")
}

func (i *Interpreter) VisitSetExpr(expr *SetExpr) (interface{}, error) {
	object, err := i.evaluate(expr.Object)
	if err != nil {
		return nil, err
	}

	loxInstance, ok := object.(*LoxInstance)
	if !ok {
		return nil, NewRuntimeError(expr.Name, "Only instances have fields")
	}

	value, err := i.evaluate(expr.Value)
	if err != nil {
		return nil, err
	}

	loxInstance.Set(expr.Name, value)
	return value, nil
}

func (i *Interpreter) VisitBlockStmt(stmt *Block) error {
	return i.executeBlock(stmt.Statements, NewEnvironment(i.environment))
}

func (i *Interpreter) executeBlock(statements []Stmt, env *Environment) error {
	previousEnv := i.environment

	i.environment = env
	for _, stmt := range statements {
		err := i.execute(stmt)
		if err != nil {
			i.environment = previousEnv
			return err
		}
	}

	i.environment = previousEnv
	return nil
}

// VisitVarStmt interprets an variable declaration. If the variable has an
// initialization part, we first evaluate it, otherwise we store the default
// nil value for it. Thus it allows us to define an uninitialized variable.
// Like other dynamically typed languages, we just assign nil if the variable
// is not initialized.
func (i *Interpreter) VisitVarStmt(expr *VarStmt) error {
	var val interface{}
	var err error
	if expr.Initializer != nil {
		val, err = i.evaluate(expr.Initializer)
		if err != nil {
			return err
		}
	}

	i.environment.Define(expr.Name.Lexeme, val)
	return nil
}

func (i *Interpreter) VisitWhileStmt(stmt *WhileStmt) error {
	for {
		condition, err := i.evaluate(stmt.Condition)
		if err != nil {
			return err
		}

		if i.isTruthy(condition) {
			err := i.execute(stmt.Body)
			if err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (i *Interpreter) VisitVarExpr(expr *VarExpr) (interface{}, error) {
	return i.lookupVariable(expr.Name, expr)
}

// VisitAssignExpr evaluates the right hand side expression to get the value and then stores it in the
// named variable. We use Assign method on the environment which only updates existing variable and is
// not allowed to create new variable. This method returns the assigned value because assignment is an
// expression and can be nested inside other expression.
// var a = 1;
// print a = 2; // "2"
func (i *Interpreter) VisitAssignExpr(expr *Assign) (interface{}, error) {
	val, err := i.evaluate(expr.Value)
	if err != nil {
		return nil, err
	}

	distance, ok := i.locals[expr]
	if ok {
		i.environment.AssignAt(distance, expr.Name, val)
	} else {
		err = i.environment.Assign(expr.Name, val)
		if err != nil {
			return nil, err
		}
	}

	return val, nil
}

// VisitExpressionExpr interprets expression statements. As statements do not
// produce any value, we are discarding the expression generated from evaluating
// the statement's expression.
func (i *Interpreter) VisitExpressionExpr(expr *Expression) error {
	_, err := i.evaluate(expr.Expression)
	if err != nil {
		return err
	}

	return nil
}

// VisitLogicalExpr evaluates a logical expression. Here we evaluate the left operand first,
// and we look at its value to check if we can short circuit. If not and only then we evaluate
// the right operand.
// Another interesting thing is we are returning the value with appropriate truthiness.
func (i *Interpreter) VisitLogicalExpr(expr *Logical) (interface{}, error) {
	left, err := i.evaluate(expr.Left)
	if err != nil {
		return nil, err
	}

	if expr.Operator.Type == Or {
		if i.isTruthy(left) {
			return left, nil
		}
	} else {
		if !i.isTruthy(left) {
			return left, nil
		}
	}

	return i.evaluate(expr.Right)
}

func (i *Interpreter) VisitIfStmt(stmt *IfStmt) error {
	condition, err := i.evaluate(stmt.Condition)
	if err != nil {
		return err
	}

	if i.isTruthy(condition) {
		err := i.execute(stmt.ThenBranch)
		if err != nil {
			return err
		}
	} else if stmt.ElseBranch != nil {
		err := i.execute(stmt.ElseBranch)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Interpreter) VisitPrintExpr(expr *Print) error {
	val, err := i.evaluate(expr.Expression)
	if err != nil {
		return err
	}

	fmt.Println(i.stringify(val))
	return nil
}

func (i *Interpreter) VisitReturnStmt(stmt *ReturnStmt) error {
	var value interface{}
	var err error

	if stmt.Value != nil {
		value, err = i.evaluate(stmt.Value)
		if err != nil {
			return nil
		}
	}

	return &ReturnErr{Value: value}
}

func (i *Interpreter) stringify(val interface{}) string {
	if val == nil {
		return "nil"
	}

	if tools.IsFloat64(val) {
		return fmt.Sprintf("%d", int(val.(float64)))
	}

	return fmt.Sprint(val)
}

func (i *Interpreter) VisitBinaryExpr(expr *Binary) (interface{}, error) {
	left, err := i.evaluate(expr.Left)
	if err != nil {
		return nil, err
	}

	right, err := i.evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type {
	case Greater:
		err := i.checkNumberOperandBoth(expr.Operator, left, right)
		if err != nil {
			return nil, err
		}

		return left.(float64) > right.(float64), nil
	case GreaterEqual:
		err := i.checkNumberOperandBoth(expr.Operator, left, right)
		if err != nil {
			return nil, err
		}

		return left.(float64) >= right.(float64), nil
	case Less:
		err := i.checkNumberOperandBoth(expr.Operator, left, right)
		if err != nil {
			return nil, err
		}

		return left.(float64) < right.(float64), nil
	case LessEqual:
		err := i.checkNumberOperandBoth(expr.Operator, left, right)
		if err != nil {
			return nil, err
		}

		return left.(float64) <= right.(float64), nil
	case BangEqual:
		return !(left == right), nil
	case EqualEqual:
		return left == right, nil
	case Minus:
		err := i.checkNumberOperandBoth(expr.Operator, left, right)
		if err != nil {
			return nil, err
		}

		return left.(float64) - right.(float64), nil
	case Plus:
		// plus (+) handles both string concatenation and arithmetic addition.
		if tools.IsString(left) && tools.IsString(right) {
			return left.(string) + right.(string), nil
		}

		if tools.IsFloat64(left) && tools.IsFloat64(right) {
			return left.(float64) + right.(float64), nil
		}

		return nil, NewRuntimeError(expr.Operator, "The both operands must be either string or number")
	case Slash:
		err := i.checkNumberOperandBoth(expr.Operator, left, right)
		if err != nil {
			return nil, err
		}

		return left.(float64) / right.(float64), nil
	case Star:
		err := i.checkNumberOperandBoth(expr.Operator, left, right)
		if err != nil {
			return nil, err
		}

		return left.(float64) * right.(float64), nil
	}

	// unreachable
	return nil, nil
}

// VisitCallExpr interprts function call tree node. First we evaluate the expression for the
// callee, typically this expression is just an identifier that looks up the function by its
// name, but it could be anything. Then we evaluate each of the arguments in order and store
// them in a list. To call a function we cast the callee to the LoxCallable interface and call
// the Call() method on it. The go representation of any lox object that can be called like an
// function will implement this interface.
func (i *Interpreter) VisitCallExpr(expr *Call) (interface{}, error) {
	callee, err := i.evaluate(expr.Callee)
	if err != nil {
		return nil, err
	}

	arguments := make([]interface{}, 0)
	for _, argument := range expr.Arguments {
		ag, err := i.evaluate(argument)
		if err != nil {
			return nil, err
		}

		arguments = append(arguments, ag)
	}

	function, ok := callee.(LoxCallable)
	if !ok {
		return nil, NewRuntimeError(expr.Paren, "Can only call function and classes")
	}

	if len(arguments) != function.Arity() {
		return nil, NewRuntimeError(expr.Paren, fmt.Sprintf("Expected %d arguments but got %d", function.Arity(), len(arguments)))
	}

	return function.Call(i, arguments)
}

// VisitFunctionStmt interprets a function syntax node. We take FunctionStmt syntax node, which
// is a compile time representation of the function - and convert it to its runtime representation.
// Here that's LoxFunction that wraps the syntax node. Here we also bind the resulting object to
// a new variable. So after creating LoxFunction, we create a new binding in the current environment
// and store a reference to it there.
func (i *Interpreter) VisitFunctionStmt(stmt *FunctionStmt) error {
	// When we create the LoxFunction, we capture the current environment. This is the env that is
	// active when the function is declared, not when it's called.
	function := NewLoxFunction(stmt, i.environment)
	i.environment.Define(stmt.Name.Lexeme, function)
	return nil
}

// VisitGroupingExpr evaluates the grouping expressions, the node that we get from
// using parenthesis around an expression. The grouping node has reference to the
// inner expression, so to evaluate it we recursively evaluate the inner subexpression.
func (i *Interpreter) VisitGroupingExpr(expr *Grouping) (interface{}, error) {
	return i.evaluate(expr.Expression)
}

// VisitLiteralExpr converts the literal tree node created during parsing to the
// runtime value. Which simply pulls the literal value back from the Token created
// during scanning.
func (i *Interpreter) VisitLiteralExpr(expr *Literal) (interface{}, error) {
	return expr.Value, nil
}

// VisitUnaryExpr evaluates the unary tree node. Unary expression have single subexpression that
// we need to evaluate first.
func (i *Interpreter) VisitUnaryExpr(expr *Unary) (interface{}, error) {
	// this will evaluate recursively for expressions like !!true, the right operand will be
	// evaluated first before evaluating the operator.
	right, err := i.evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type {
	case Bang:
		return !i.isTruthy(right), nil
	case Minus:
		if err := i.checkNumberOperand(expr.Operator, right); err != nil {
			return nil, err
		}

		return -right.(float64), nil
	}

	// unreachable.
	return nil, nil
}

// evaluate is a helper method that sends the expression back to the interpreter's visitor
// implementation.
func (i *Interpreter) evaluate(expr Expr) (interface{}, error) {
	return expr.Accept(i)
}

// isTruthy is a helper method that determines the truthfulness of a value. In lox the boolean value
// false and nil is considered falsy and everything else truthy.
func (i *Interpreter) isTruthy(val interface{}) bool {
	if val == nil {
		return false
	}

	switch val := val.(type) {
	case bool:
		return val
	}

	return true
}

func (i *Interpreter) checkNumberOperand(operator Token, operand interface{}) error {
	if tools.IsFloat64(operand) {
		return nil
	}

	return NewRuntimeError(operator, "Operand must me a number")
}

func (i *Interpreter) checkNumberOperandBoth(operator Token, left, right interface{}) error {
	if tools.IsFloat64(left) && tools.IsFloat64(right) {
		return nil
	}

	return NewRuntimeError(operator, "Both operands must be numbers")
}

func (i *Interpreter) resolve(expr Expr, depth int) {
	i.locals[expr] = depth
}

// lookupVariable resolves a variable. First we look up the resolved distance in the local map. Remember
// we only resolved local variables, globals are treated differently and don't end up in the map. So, if
// we don't find it in the local map, then it must be in the global environment.
func (i *Interpreter) lookupVariable(name Token, expr Expr) (interface{}, error) {
	distance, ok := i.locals[expr]
	if ok {
		return i.environment.GetAt(distance, name.Lexeme), nil
	} else {
		return i.globals.Get(name)
	}
}

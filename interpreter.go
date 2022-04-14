package glox

import (
	"fmt"

	"github.com/iamsayantan/glox/tools"
)

type Interpreter struct {
	runtime     *Runtime
	environment *Environment
}

func NewInterpreter(runtime *Runtime) *Interpreter {
	return &Interpreter{runtime: runtime, environment: NewEnvironment(nil),}
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

func (i *Interpreter) VisitVarExpr(expr *VarExpr) (interface{}, error) {
	val, err := i.environment.Get(expr.Name)
	if err != nil {
		return nil, err
	}

	return val, nil
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

	err = i.environment.Assign(expr.Name, val)
	if err != nil {
		return nil, err
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

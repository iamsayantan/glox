package glox

import (
	"fmt"

	"github.com/iamsayantan/glox/tools"
)

type Interpreter struct {
	runtime *Runtime
}

func NewInterpreter(runtime *Runtime) *Interpreter {
	return &Interpreter{runtime: runtime}
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

func (i *Interpreter) Interpret(expr Expr) {
	val, err := i.evaluate(expr)
	if err != nil {

		i.runtime.runtimeError(err)
		return
	}

	fmt.Println(i.stringify(val))
}

func (i *Interpreter) stringify(val interface{}) string {
	if val == nil {
		return "nil"
	}

	if tools.IsFloat64(val) {
		return fmt.Sprintf("%d", int(val.(float64)))
	}

	return fmt.Sprintf("%v", val)
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

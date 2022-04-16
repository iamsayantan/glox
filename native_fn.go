package glox

import "time"

type Clock struct{}

func (c Clock) Call(interpreter *Interpreter, arguments []interface{}) (interface{}, error) {
	return float64(time.Now().Unix()), nil
}

func (c Clock) Arity() int {
	return 0
}

func (c Clock) String() string {
	return "<native fn>"
}

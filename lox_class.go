package glox

import "errors"

var ErrMethodNotFound = errors.New("method not found with the given name")

type LoxClass struct {
	Name string
	methods map[string]LoxFunction
}

func NewLoxClass(name string, methods map[string]LoxFunction) *LoxClass {
	return &LoxClass{Name: name, methods: methods}
}

func (lc *LoxClass) String() string {
	return lc.Name
}

func (lc *LoxClass) Call(ip *Interpreter, arguments []interface{}) (interface{}, error) {
	instance := NewLoxInstance(lc)
	return instance, nil
}

func (lc *LoxClass) Arity() int {
	return 0
}

func (lc *LoxClass) findMethod(name string) (LoxFunction, error) {
	if method, ok := lc.methods[name]; ok {
		return method, nil
	}

	return LoxFunction{}, ErrMethodNotFound
}
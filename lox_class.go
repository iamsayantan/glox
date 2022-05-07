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

	// When a class is called, and the lox instance is created, we look for an "init" method,
	// If we find it, we immediately bind and invoke it just like normal method call. The
	// argument list is forwarded along.
	initializer, err := lc.findMethod("init")
	if err == nil {
		initializer.Bind(instance).Call(ip, arguments)
	}


	return instance, nil
}

// Arity returns the arity of the class. If there is an initializer, that method's arity determines
// how many arguments users must pass to call the class. But the initializer is not required though,
// in that case the arity is zero.
func (lc *LoxClass) Arity() int {
	initializer, err := lc.findMethod("init")
	if err == nil {
		return initializer.Arity()
	}

	return 0
}

func (lc *LoxClass) findMethod(name string) (LoxFunction, error) {
	if method, ok := lc.methods[name]; ok {
		return method, nil
	}

	return LoxFunction{}, ErrMethodNotFound
}
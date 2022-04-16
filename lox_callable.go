package glox

// LoxCallable interface should be implemented by any lox object that can be called like
// a function.
type LoxCallable interface {
	// Call is the method that is called to evaluate the function. We pass in the
	// interpreter in case the implementing object needs it and the list of arguments.
	// The implementing object should return the evaluated value as return parameter.
	Call(interpreter *Interpreter, arguments []interface{}) (interface{}, error)

	// Arity is the number of arguments a function expects. It's used to check if the
	// number of arguments passed to the function matches the number of arguments the
	// function expects.
	Arity() int
}

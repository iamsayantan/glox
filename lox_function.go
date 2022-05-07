package glox

// LoxFunction is the representation of the lox function in terms of the interpreter. 
// This struct also implements the LoxCallable interface so the runtime can call this
// function.
type LoxFunction struct {
	declaration *FunctionStmt
	closure *Environment
}

func NewLoxFunction(declaration *FunctionStmt, closure *Environment) LoxCallable {
	return LoxFunction{declaration: declaration, closure: closure}
}

// Call will execute the function body with the arguments passed to it. The parameters are 
// core to a function, a function encapsulates its parameters. No other code outside the 
// function should see them. This means each function gets its own environment. And this 
// environment is generated at runtime during the function call. Then it walks the parameters
// and argument lists and for each pair it creates a new variable with the parameter's name 
// and binds it to the argument's value.
func (lf LoxFunction) Call(interpreter *Interpreter, arguments []interface{}) (interface{}, error) {
	env := NewEnvironment(lf.closure)
	for i, param := range lf.declaration.Params {
		env.Define(param.Lexeme, arguments[i])
	}

	err := interpreter.executeBlock(lf.declaration.Body, env)
	if err != nil {
		if runE, ok := err.(*ReturnErr); ok {
			return runE.Value, nil
		}

		return nil, err
	}

	return  nil, nil
}


func (lf LoxFunction) Arity() int {
	return len(lf.declaration.Params)
}

func (lf LoxFunction) String() string {
	return "<fn " + lf.declaration.Name.Lexeme + ">"
}

func (lf LoxFunction) Bind(instance *LoxInstance) LoxFunction {
	env := NewEnvironment(lf.closure)
	env.Define("this", instance)
	return NewLoxFunction(lf.declaration, env).(LoxFunction)
}
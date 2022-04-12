package glox

type Environment struct {
	// values uses string for the keys and not Token because token represents
	// a unit of code at a specific place in the source text, but when it comes
	// to variables, all identifier tokens using the same name should refer to 
	// the same variable (ignorig scope for now).
	values map[string]interface{}
}

func NewEnvironment() *Environment {
	return &Environment{values: make(map[string]interface{}, 0)}
}

func (e *Environment) Define(name string, value interface{}) {
	e.values[name] = value
}

func (e *Environment) Get(name Token) (interface{}, error) {
	val, ok := e.values[name.Lexeme]
	if !ok {
		return nil, NewRuntimeError(name, "Undefined variable '" + name.Lexeme + "'")
	}

	return val, nil
}

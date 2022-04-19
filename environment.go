package glox

type Environment struct {
	// values uses string for the keys and not Token because token represents
	// a unit of code at a specific place in the source text, but when it comes
	// to variables, all identifier tokens using the same name should refer to
	// the same variable (ignorig scope for now).
	values map[string]interface{}

	// enclosing works as the parent of this Environment. For the global scope,
	// this should be null breaking the chain. But for each local scope, we must
	// enclose the parent scope.
	enclosing *Environment
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{values: make(map[string]interface{}, 0), enclosing: parent}
}

// Define defines a new variable in the current innermost scope.
func (e *Environment) Define(name string, value interface{}) {
	e.values[name] = value
}

// Get looks up a variable in the environment. It starts by looking into the innermost
// environment and goes up till it reaches the global scope.
func (e *Environment) Get(name Token) (interface{}, error) {
	val, ok := e.values[name.Lexeme]
	if ok {
		return val, nil
	}

	if e.enclosing != nil {
		return e.enclosing.Get(name)
	}

	return nil, NewRuntimeError(name, "Undefined variable '"+name.Lexeme+"'")
}

// Assign will assign value to the variable. If the variable is not available in the current
// environment, it will try to assign it recursively to the out environments until it reaches
// the global environment.
func (e *Environment) Assign(name Token, value interface{}) error {
	_, ok := e.values[name.Lexeme]

	if ok {
		e.values[name.Lexeme] = value
		return nil
	}

	if e.enclosing != nil {
		return e.enclosing.Assign(name, value)
	}

	return NewRuntimeError(name, "Undefined variable '"+name.Lexeme+"'.")
}

// GetAt will get the exact environment where the variable is defined in the environment chain and
// return the value.
func (e *Environment) GetAt(distance int, name string) interface{} {
	return e.ancestor(distance).values[name]
}

// AssignAt walks fixed numbers of steps and stuffs the variable into that map.
func (e *Environment) AssignAt(distance int, name Token, value interface{}) {
	e.ancestor(distance).values[name.Lexeme] = value
}

// ancestor walks a fixed number of hops up the parent chain and returns the environment there.
func (e *Environment) ancestor(distance int) *Environment {
	env := e
	for i := 0; i < distance; i++ {
		env = env.enclosing
	}

	return env
}

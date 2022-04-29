package glox

type LoxInstance struct {
	klass *LoxClass
	fields map[string]interface{}
}

func NewLoxInstance(klass *LoxClass) *LoxInstance {
	return &LoxInstance{klass: klass, fields: make(map[string]interface{})}
}

func (li *LoxInstance) String() string {
	return li.klass.Name + " instance"
}

func (li *LoxInstance) Get(name Token) (interface{}, error) {
	if val, ok := li.fields[name.Lexeme]; ok {
		return val, nil
	}

	return nil, NewRuntimeError(name, "Undefined property '" + name.Lexeme + "'")
} 

func (li *LoxInstance) Set(name Token, value interface{}) {
	li.fields[name.Lexeme] = value
}
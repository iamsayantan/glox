package glox

type LoxClass struct {
	Name string
}

func NewLoxClass(name string) *LoxClass {
	return &LoxClass{Name: name}
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
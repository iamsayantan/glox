package glox 

import "fmt"

type Token struct {
	Type    TokenType
	Lexeme  string
	Literal interface{}
	Line    int
}

func NewToken(tokenType TokenType, lexeme string, literal interface{}, line int) Token {
	return Token{
		tokenType,
		lexeme,
		literal,
		line,
	}
}

func (t Token) ToString() string {
	return fmt.Sprintf("%v %s %s", t.Type, t.Lexeme, t.Literal)
}

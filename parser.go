package glox

type Parser struct {
	// tokens is the list of tokens
	tokens []Token
	// current points to the next token to be consumed
	current int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens:  tokens,
		current: 0,
	}
}

// expression parses the grammar
// expression --> equality
func (p *Parser) expression() Expr {
	return p.equality()
}

// equality parses the grammar. It matches an equality and anything of higher precedence.
// equality --> comparison ( ("==" | "!=") comparison )*
func (p *Parser) equality() Expr {
	expr := p.comparison()

	// if the control goes into this for loop, it means we have found
	// a == or != operator and we are parsing an equality expression.
	// Note that if equality does not match any equality operator, it 
	// essentially calls and returns comparison().
	for p.match(Bang, BangEqual) {
		// we grab the operator that has been consumed by match
		operator := p.previous()

		// calling comparison again to grab the right side of the operator
		right := p.comparison()

		// then we combine the operator and the two operands to a new Binary
		// syntax tree node.
		expr = &Binary{expr, operator, right}

		// Now we loop around to parse expression like this a == b == c == d == e.
		// With each new iteration we create a new Binary expression with the previous
		// expression as the left operand.
	}

	return expr
}

// comparison matches a comparison expression or anything of higher precedence.
// comparison --> term ( (">" | ">=" | "<" | "<=") term )*
func (p *Parser) comparison() Expr {

}

// match checks to see if the current token has any of the given
// types provided as parameter, if it matches it consumes the token
// and returns true. Otherwise it leaves the current token alone
// and return false.
func (p *Parser) match(tokenTypes ...TokenType) bool {
	for _, tokenType := range tokenTypes {
		if p.check(tokenType) {
			p.advance()
			return true
		}
	}

	return false
}

// check method returns if the current token matches the given type.
// It does not consume the token though, just looks at it.
func (p *Parser) check(tokenType TokenType) bool {
	if p.isAtEnd() {
		return false
	}

	return p.peek().Type == tokenType
}

// advance consumes the current token and returns it.
func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.current++
	}

	return p.previous()
}

// isAtEnd checks if we have run out of tokens to parse.
func (p *Parser) isAtEnd() bool {
	return p.peek().Type == Eof
}

// peek returns the current token we are yet to consume.
func (p *Parser) peek() Token {
	return p.tokens[p.current]
}

// previous returns the most recent token that has been consumed.
func (p *Parser) previous() Token {
	return p.tokens[p.current-1]
}

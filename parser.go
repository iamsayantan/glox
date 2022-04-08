package glox

type Parser struct {
	// tokens is the list of tokens
	tokens []Token
	// current points to the next token to be consumed
	current int

	runtime *Runtime
}

type ParseError struct {
	message string
}

func NewParseError(message string) error {
	return ParseError{message: message}
}

func (pe ParseError) Error() string {
	return pe.message
}

func NewParser(tokens []Token, runtime *Runtime) *Parser {
	return &Parser{
		tokens:  tokens,
		current: 0,
		runtime: runtime,
	}
}

func (p *Parser) Parse() Expr {
	expr, err := p.expression()
	if err != nil {
		return nil
	}

	return expr
}

// expression parses the grammar
// expression --> equality
func (p *Parser) expression() (Expr, error) {
	return p.equality()
}

// equality parses the grammar. It matches an equality and anything of higher precedence.
// equality --> comparison ( ("==" | "!=") comparison )*
func (p *Parser) equality() (Expr, error) {
	expr, err := p.comparison()
	if err != nil {
		return nil, err
	}

	// if the control goes into this for loop, it means we have found
	// a == or != operator and we are parsing an equality expression.
	// Note that if equality does not match any equality operator, it
	// essentially calls and returns comparison().
	for p.match(Bang, BangEqual) {
		// we grab the operator that has been consumed by match
		operator := p.previous()

		// calling comparison again to grab the right side of the operator
		right, err := p.comparison()
		if err != nil {
			return nil, err
		}

		// then we combine the operator and the two operands to a new Binary
		// syntax tree node.
		expr = &Binary{expr, operator, right}

		// Now we loop around to parse expression like this a == b == c == d == e.
		// With each new iteration we create a new Binary expression with the previous
		// expression as the left operand.
	}

	return expr, nil
}

// comparison matches a comparison expression or anything of higher precedence.
// comparison --> term ( (">" | ">=" | "<" | "<=") term )*
func (p *Parser) comparison() (Expr, error) {
	expr, err := p.term()
	if err != nil {
		return nil, err
	}

	for p.match(Greater, GreaterEqual, Less, LessEqual) {
		operator := p.previous()
		right, err := p.term()

		if err != nil {
			return nil, err
		}

		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

// term matches a term expression or anything of higher precedence.
// term --> factor ( ( "-" | "+" ) factor )*
func (p *Parser) term() (Expr, error) {
	expr, err := p.factor()
	if err != nil {
		return nil, err
	}

	for p.match(Plus, Minus) {
		operator := p.previous()
		right, err := p.factor()

		if err != nil {
			return nil, err
		}

		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

// factor parses a factor expression or anything of higher precedence.
// factor --> unary ( ( "/" | "*" ) unary )*
func (p *Parser) factor() (Expr, error) {
	expr, err := p.unary()

	if err != nil {
		return nil, err
	}

	for p.match(Slash, Star) {
		operator := p.previous()
		right, err := p.unary()

		if err != nil {
			return nil, err
		}

		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

// unary parses an unary expression and primary expression.
// unary --> ( "!" | "-" ) unary
//			 | primary
func (p *Parser) unary() (Expr, error) {
	if p.match(Bang, Minus) {
		operator := p.previous()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}

		return &Unary{Operator: operator, Right: right}, nil
	}

	return p.primary()
}

// primary parses the primary expressions, these are of highest level of precedence.
// primary --> NUMBER | STRING | "true" | "false" | "nil"
//            | "(" expression ")"
func (p *Parser) primary() (Expr, error) {
	if p.match(False) {
		return &Literal{Value: false}, nil
	}

	if p.match(True) {
		return &Literal{Value: true}, nil
	}

	if p.match(Nil) {
		return &Literal{Value: nil}, nil
	}

	if p.match(String, Number) {
		return &Literal{Value: p.previous().Literal}, nil
	}

	// if we find a '(' token during parsing, we must find a ')' too
	// after the expression, otherwise its an error.
	if p.match(LeftParen) {
		expression, err := p.expression()
		if err != nil {
			return nil, err
		}

		_, err = p.consume(RightParen, "Expect ')' after expression.")
		if err != nil {
			return nil, err
		}

		return &Grouping{Expression: expression}, nil
	}

	// The parser has descent down from the initial expression grammer to
	// all the way to primary expression. If the token does not match any
	// of the cases for primary, that means we are sitting on a token that
	// can't start an expression. We need to handle that error too.

	return nil, p.error(p.peek(), "Expect Expression")
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

func (p *Parser) consume(tokenType TokenType, message string) (Token, error) {
	if p.check(tokenType) {
		return p.advance(), nil
	}

	return Token{}, NewParseError(message)
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

func (p *Parser) error(token Token, message string) error {
	p.runtime.tokenError(token, message)
	return NewParseError(message)
}

// synchronize synchronizes the parser state in case of encountering an error.
// We want to discard tokens until we are right at the beginning of the next statement.
// After a semicolon, we are probably finished with a statement. And also most statements
// start with a key word 'for', 'if', 'var', 'return' etc, when the next token is any
// of those, we are probably starting a statement.
func (p *Parser) synchronize() {
	p.advance()

	for !p.isAtEnd() {
		if p.previous().Type == Semicolon {
			return
		}

		switch p.peek().Type {
		case Class, Fun, Var, For, If, While, Print, Return:
			return
		}

		p.advance()
	}
}

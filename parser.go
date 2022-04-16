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

func (p *Parser) Parse() []Stmt {
	statements := make([]Stmt, 0)
	for !p.isAtEnd() {
		expr, err := p.declaration()
		if err != nil {
			return nil
		}

		statements = append(statements, expr)
	}

	return statements
}

// declaration parses declaration statements. Any place where a declaration is allowed also
// allowes non declaring statements, so the declaration rule falls through the statement.
// declaration is called repeatedly when parsing a series of statements. If we get any error
// while parsing, the parser tries to recover using synchronize and continue parsing the next
// statements.
// declaration --> varDecl
// 				   | funcDeclaration
// 				   | statement
func (p *Parser) declaration() (Stmt, error) {
	if p.match(Fun) {
		stmt, err := p.function("function")
		if err != nil {
			return nil, err
		}

		return stmt, nil
	}

	if p.match(Var) {
		stmt, err := p.varDeclaration()
		if err != nil {
			p.synchronize()
			return nil, nil
		}

		return stmt, nil
	}

	return p.statement()
}

// function parses grammar for function declaration. Since we already matched and consumed
// fun keyword, we don't need to do that again. Next we parse the list of parameters with 
// the parentheses wrapped around them. The outer if condition handles the zero parameter
// case and the inner for loop parses parameters as long as we find commas to separate them.
// We consume the { at the  beginning of the body before calling block, as block() assumes
// brace token has already been consumed. And this way we cal provide a more precise error
// message if the brace is not provided.
func (p *Parser) function(kind string) (Stmt, error) {
	name, err := p.consume(Identifiers, "Expect " + kind + " name")
	if err != nil {
		return nil, err
	}

	_, err = p.consume(LeftParen, "Expect '(' after " + kind + " name")
	if err != nil {
		return nil, err
	}

	parameters := make([]Token, 0)
	if !p.check(RightParen) {
		for {
			if len(parameters) > 255 {
				p.error(p.peek(), "Can't have more than 255 parameters")
			}

			param, err := p.consume(Identifiers, "Expect parameter name")
			if err != nil {
				return nil, err
			}

			parameters = append(parameters, param)
			if !p.match(Comma) {
				break
			}
		}
	}

	_, err = p.consume(RightParen, "Expect ')' after parameters")
	if err != nil {
		return nil, err
	}

	_, err = p.consume(LeftBrace, "Expect '{' before " + kind + " body")
	if err != nil {
		return nil, err
	}

	body, err := p.block()
	if err != nil {
		return nil, err
	}

	return &FunctionStmt{Name: name, Body: body, Params: parameters}, nil
}

// varDeclaration parses variable declaration syntax. When the parser matches a var
// keyword, this method is used to parse that statement.
// varDecl        → "var" IDENTIFIER ( "=" expression )? ";" ;
func (p *Parser) varDeclaration() (Stmt, error) {
	name, err := p.consume(Identifiers, "Expect a variable name")
	if err != nil {
		return nil, err
	}

	var expr Expr
	if p.match(Equal) {
		expr, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	_, err = p.consume(Semicolon, "Expect a ';' after variable declaration")
	if err != nil {
		return nil, err
	}

	return &VarStmt{Name: name, Initializer: expr}, nil
}

// statement parses statements, a program can have multiple statements. Statements are
// of two types, print statement and expression statement.
// statement --> exprStmt
//				| printStmt
func (p *Parser) statement() (Stmt, error) {
	if p.match(If) {
		return p.ifStatement()
	}

	if p.match(PRINT) {
		return p.printStatement()
	}

	if p.match(While) {
		return p.whileStatement()
	}

	if p.match(For) {
		return p.forStatement()
	}

	if p.match(Return) {
		return p.returnStatement()
	}

	if p.match(LeftBrace) {
		stmt, err := p.block()
		if err != nil {
			return nil, err
		}

		return &Block{Statements: stmt}, nil
	}

	return p.expressionStatement()
}

// returnStatement will parse a return statement. After fetching the previously consumed return 
// keyword, we look for a value expression. As many different tokens can start an expression, it's
// hard to tell if return value is present. So instead, we look for it's absence. Since semicolon
// can't begin an expression, if the next token is that, we know there must not be a value.
func (p *Parser) returnStatement() (Stmt, error) {
	keyword := p.previous()
	var value Expr
	var err error
	
	if !p.check(Semicolon) {
		value, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	_, err = p.consume(Semicolon, "Expect ';' after return value")
	return &ReturnStmt{Keyword: keyword, Value: value}, nil
}

func (p *Parser) forStatement() (Stmt, error) {
	_, err := p.consume(LeftParen, "Expect '(' after 'for'")
	if err != nil {
		return nil, err
	}

	var initializer Stmt = nil
	var condition Expr = nil
	var increment Expr = nil

	// If the token following the '(' is a semicolon, then the initializer
	// has been omitted. Otherwise we check for the var keyword to see if
	// it's a variable declaration. If none of those is matched, it must be
	// an expression.
	if p.match(Semicolon) {
		// no need to do anything, initializer already is nil
	} else if p.match(Var) {
		initializer, err = p.varDeclaration()
		if err != nil {
			return nil, err
		}
	} else {
		initializer, err = p.expressionStatement()
		if err != nil {
			return nil, err
		}
	}

	if !p.check(Semicolon) {
		condition, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	_, err = p.consume(Semicolon, "Expect ';' after loop condition")
	if err != nil {
		return nil, err
	}

	if !p.check(RightParen) {
		increment, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	_, err = p.consume(RightParen, "Expect ')' after for clause")
	if err != nil {
		return nil, err
	}

	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	// if increment is not nil, it executes after body in each iteration of the loop.
	// And as the increment expression in the for loop does not produce any value, we
	// convert it to an expression statement.
	if increment != nil {
		body = &Block{
			Statements: []Stmt{body, &Expression{Expression: increment}},
		}
	}

	// If the condition is omitted, we put in true to make it an infinite loop.
	if condition == nil {
		condition = &Literal{Value: True}
	}

	// Now we take the condition and body and make it a primitive while loop.
	body = &WhileStmt{Condition: condition, Body: body}

	// Now if we have an initializer, it runs once before the body of the loop. We do that
	// by creating a block that runs the initializer and then executes the loop.
	if initializer != nil {
		body = &Block{Statements: []Stmt{initializer, body}}
	}

	return body, nil
}

func (p *Parser) whileStatement() (Stmt, error) {
	_, err := p.consume(LeftParen, "Expect '(' after 'while'")
	if err != nil {
		return nil, err
	}

	condition, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, err = p.consume(RightParen, "Expect ')' after condition")
	if err != nil {
		return nil, err
	}

	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	return &WhileStmt{Condition: condition, Body: body}, nil
}

func (p *Parser) ifStatement() (Stmt, error) {
	// The parenthesis around the if statement is only half useful. We need some kind of delimiter between
	// the condition and the then statement, otherwise the parser can't tell when it has reached the end
	// of the condition. But the opening parenthesis in the if condition doesn't do anything useful, it's
	// only there because otherwise we'd end up with unbalanced parenthesis. Go requires the statement to
	// be braced block, so the '{' acts as the end of the condition.
	_, err := p.consume(LeftParen, "Expected '(' after 'if'")
	if err != nil {
		return nil, err
	}

	condition, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, err = p.consume(RightParen, "Expect ')' after if condition.")
	if err != nil {
		return nil, err
	}

	thenBranch, err := p.statement()
	if err != nil {
		return nil, err
	}

	var elseBranch Stmt = nil
	if p.match(Else) {
		elseBranch, err = p.statement()
		if err != nil {
			return nil, err
		}
	}

	return &IfStmt{Condition: condition, ThenBranch: thenBranch, ElseBranch: elseBranch}, nil
}

// block parses a block of statements when it encounters a '{'.
func (p *Parser) block() ([]Stmt, error) {
	statements := make([]Stmt, 0)

	for !p.check(RightBrace) && !p.isAtEnd() {
		stmt, err := p.declaration()
		if err != nil {
			return nil, err
		}

		statements = append(statements, stmt)
	}

	_, err := p.consume(RightBrace, "Expect '}' after block")
	if err != nil {
		return nil, err
	}

	return statements, nil
}

// printStatement parses a print statement. Since the print keyword is
// already consumed by the match method earlier, we just parse the
// subsequent expression, consume the terminating semicolon and emit the
// syntax tree.
// printStmt --> "print" expression ";"
func (p *Parser) printStatement() (Stmt, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, err = p.consume(Semicolon, "Expect ; after value.")
	if err != nil {
		return nil, err
	}

	return &Print{Expression: expr}, nil
}

// expressionStatement parses expression statements. It kind of acts like a
// fallthrough condition. If we can't match with any known statements, we
// assume it's a expression statement.
// exprStmt --> expression ";";
func (p *Parser) expressionStatement() (Stmt, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, err = p.consume(Semicolon, "Expect ; after value.")
	if err != nil {
		return nil, err
	}

	return &Expression{Expression: expr}, nil
}

// expression parses the grammar
// expression --> assignment
func (p *Parser) expression() (Expr, error) {
	return p.assignment()
}

// assignment parses an assignment expression. First we parse the left hand side, which can be
// any expression of higher precedence. If we find an '=',  we parse the right hand side and
// then wrap it all up in an assignment tree node. The difference from other binary expressions
// is that we don't loop to build up a sequence of same operator. Since assignment is right associative
// we call assignment() recursively to parse the right hand side.
// assignment --> IDENTIFIER "=" assignment
// 				  | logic_or
func (p *Parser) assignment() (Expr, error) {
	expr, err := p.or()
	if err != nil {
		return nil, err
	}

	if p.match(Equal) {
		equals := p.previous()
		value, err := p.assignment()

		if err != nil {
			return nil, err
		}

		// Before we create a assignment node, we look at the left hand side expression and figure out
		// what kind of assignment target it is. If the left hand side is not a valid assignment target
		// we report a syntax error. This makes sure that we report an error on code like a + b = c.
		if variable, ok := expr.(*VarExpr); ok {
			name := variable.Name
			return &Assign{Name: name, Value: value}, nil
		} else {
			p.error(equals, "Invalid assignment target")
			return nil, nil
		}
	}

	return expr, nil
}

func (p *Parser) or() (Expr, error) {
	expr, err := p.and()
	if err != nil {
		return nil, err
	}

	for p.match(Or) {
		operator := p.previous()
		right, err := p.and()
		if err != nil {
			return nil, err
		}

		expr = &Logical{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

func (p *Parser) and() (Expr, error) {
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}

	for p.match(And) {
		operator := p.previous()
		right, err := p.equality()
		if err != nil {
			return nil, err
		}

		expr = &Logical{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
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
//			 | call
func (p *Parser) unary() (Expr, error) {
	if p.match(Bang, Minus) {
		operator := p.previous()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}

		return &Unary{Operator: operator, Right: right}, nil
	}

	return p.call()
}

// call parses a function call grammar. This rule matches a primary expression followed by
// zero or more function calls. If there is no parenthesis this matches a bare primary expression.
// The * in the grammar allows calls like fn(1)(2)(3) function calls.
// call --> primary ( "(" arguments? ")")*;
func (p *Parser) call() (Expr, error) {
	expr, err := p.primary()
	if err != nil {
		return nil, err
	}

	for {
		if p.match(LeftParen) {
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}
		} else {
			break
		}
	}

	return expr, nil
}

// finishCall is a helper that parses the function arguments. This is more or less
// the grammar for arguments. Except we also check the zero argument condition. If
// we find the ')' as the next token, we don't parse any expression.
// arguments --> expression ( "," expression )*;
func (p *Parser) finishCall(callee Expr) (Expr, error) {
	arguments := make([]Expr, 0)
	if !p.check(RightParen) {
		for {
			expr, err := p.expression()
			if err != nil {
				return nil, err
			}

			if len(arguments) >= 255 {
				p.error(p.peek(), "Can't have more than 255 arguments.")
			}

			arguments = append(arguments, expr)
			if !p.match(Comma) {
				break
			}
		}
	}

	paren, err := p.consume(RightParen, "Expect ')' after arguments")
	if err != nil {
		return nil, err
	}

	return &Call{Callee: callee, Paren: paren, Arguments: arguments}, nil
}

// primary parses the primary expressions, these are of highest level of precedence.
// primary --> NUMBER | STRING | "true" | "false" | "nil"
//            | "(" expression ")"
//            | IDENTIFIER;
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

	if p.match(Identifiers) {
		return &VarExpr{Name: p.previous()}, nil
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

	return Token{}, p.error(p.peek(), message)
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
		case Class, Fun, Var, For, If, While, PRINT, Return:
			return
		}

		p.advance()
	}
}

package glox

import (
	"bytes"
	"fmt"
	"strconv"
	"unicode"
)

type Scanner struct {
	source      *bytes.Buffer
	sourceRunes []rune
	tokens      []Token
	keywords    map[string]TokenType

	start   int
	current int
	line    int

	runtime *Runtime
}

func NewScanner(source *bytes.Buffer, runtime *Runtime) *Scanner {
	keywords := map[string]TokenType{
		"and":    And,
		"class":  Class,
		"else":   Else,
		"false":  False,
		"for":    For,
		"fun":    Fun,
		"if":     If,
		"nil":    Nil,
		"or":     Or,
		"print":  PRINT,
		"return": Return,
		"super":  Super,
		"this":   This,
		"true":   True,
		"var":    Var,
		"while":  While,
	}

	return &Scanner{
		source:      source,
		sourceRunes: bytes.Runes(source.Bytes()),
		tokens:      make([]Token, 0),
		keywords:    keywords,
		start:       0,
		current:     0,
		line:        1,
		runtime:     runtime,
	}
}

func (sc *Scanner) ScanTokens() []Token {
	for !sc.isAtEnd() {
		// We are at the begining of the next lexeme.
		sc.start = sc.current
		sc.scanToken()
	}

	sc.tokens = append(sc.tokens, NewToken(Eof, "", nil, sc.line))
	return sc.tokens
}

func (sc *Scanner) scanToken() {
	c, _, _ := sc.advance()
	switch c {
	case '(':
		sc.addToken(LeftParen, nil)
	case ')':
		sc.addToken(RightParen, nil)
	case '{':
		sc.addToken(LeftBrace, nil)
	case '}':
		sc.addToken(RightBrace, nil)
	case ',':
		sc.addToken(Comma, nil)
	case '.':
		sc.addToken(Dot, nil)
	case '-':
		sc.addToken(Minus, nil)
	case '+':
		sc.addToken(Plus, nil)
	case ';':
		sc.addToken(Semicolon, nil)
	case '*':
		sc.addToken(Star, nil)
	case ' ', '\r', '\t':
	case '\n':
		sc.line++
	case '!':
		if sc.match('=') {
			sc.addToken(BangEqual, nil)
		} else {
			sc.addToken(Bang, nil)
		}
	case '=':
		if sc.match('=') {
			sc.addToken(EqualEqual, nil)
		} else {
			sc.addToken(Equal, nil)
		}
	case '<':
		if sc.match('=') {
			sc.addToken(LessEqual, nil)
		} else {
			sc.addToken(Less, nil)
		}
	case '>':
		if sc.match('=') {
			sc.addToken(GreaterEqual, nil)
		} else {
			sc.addToken(Greater, nil)
		}
	case '/':
		if sc.match('/') {
			// A comment goes on until the end of line.
			for sc.peek() != '\n' && !sc.isAtEnd() {
				sc.advance()
			}
		} else {
			sc.addToken(Slash, nil)
		}
	case '"':
		sc.scanString()
	default:
		if sc.isDigit(c) {
			sc.scanNumber()
		} else if sc.isAlpha(c) {
			sc.scanIdentifier()
		} else {
			sc.runtime.Error(sc.line, fmt.Sprintf("Unexpected character %c", c))
		}
	}
}

func (sc *Scanner) scanString() {
	for sc.peek() != '"' && !sc.isAtEnd() {
		if sc.peek() == '\n' {
			sc.line++
		}

		sc.advance()
	}

	if sc.isAtEnd() {
		sc.runtime.Error(sc.line, "Unterminated string")
		return
	}

	// the closing "
	sc.advance()

	// Trim the surrounding quotes and just take the string literal.
	val := sc.sourceRunes[sc.start+1 : sc.current-1]

	sc.addToken(String, string(val))
}

func (sc *Scanner) scanNumber() {
	for sc.isDigit(sc.peek()) {
		sc.advance()
	}

	// Look for a fractional part
	if sc.peek() == '.' && sc.isDigit(sc.peekNext()) {
		// consume the .
		sc.advance()

		// consume the digits of the fractional part
		for sc.isDigit(sc.peek()) {
			sc.advance()
		}
	}

	num, _ := strconv.ParseFloat(string(sc.sourceRunes[sc.start:sc.current]), 64)
	sc.addToken(Number, num)
}

func (sc *Scanner) scanIdentifier() {
	for sc.isAlphaNumeric(sc.peek()) {
		sc.advance()
	}

	// After scanning the identifier, we need to check if this is a reserved keyword.
	text := sc.sourceRunes[sc.start:sc.current]
	tokenType, ok := sc.keywords[string(text)]

	if !ok {
		tokenType = Identifiers
	}

	sc.addToken(tokenType, nil)
}

func (sc *Scanner) isAtEnd() bool {
	return sc.source.Len() == 0
}

func (sc *Scanner) advance() (rune, int, error) {
	sc.current += 1

	return sc.source.ReadRune()
}

func (sc *Scanner) match(expected rune) bool {
	if sc.isAtEnd() {
		return false
	}

	if sc.peek() != expected {
		return false
	}

	sc.advance()
	return true
}

func (sc *Scanner) peek() rune {
	if sc.isAtEnd() {
		return 0
	}

	return sc.sourceRunes[sc.current]
}

func (sc *Scanner) peekNext() rune {
	if sc.current >= len(sc.sourceRunes) {
		return 0
	}

	return sc.sourceRunes[sc.current+1]
}

func (sc *Scanner) isDigit(r rune) bool {
	return unicode.IsDigit(r)
}

func (sc *Scanner) isAlpha(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func (sc *Scanner) isAlphaNumeric(r rune) bool {
	return sc.isAlpha(r) || sc.isDigit(r)
}

func (sc *Scanner) addToken(tokenType TokenType, literal interface{}) {
	text := string(sc.sourceRunes[sc.start:sc.current])
	sc.tokens = append(sc.tokens, NewToken(tokenType, text, literal, sc.line))
}

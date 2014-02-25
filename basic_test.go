package lexinator

import (
	"testing"
	"unicode"
)

const (
	TokenSymbol TokenType = iota
	TokenEquals
	TokenNumber
	TokenSemi
	TokenString
)

func symbolState(l *Lexer) StateFn {
	l.Run(unicode.IsSpace)
	l.Ignore()
	if l.Peek() == Eof {
		return l.EOF()
	}
	ok := l.ExceptAnyRun(" =")
	if ok == 0 {
		return l.Errorf("Expected symbol!")
	}
	l.Emit(TokenSymbol)
	return afterSymbolState
}

func afterSymbolState(l *Lexer) StateFn {
	l.Run(unicode.IsSpace)
	l.Ignore()
	if l.AcceptAnyOne("=") == false {
		return l.Errorf("Expected operator '='!")
	}
	l.Emit(TokenEquals)
	return afterOperatorState
}

func afterOperatorState(l *Lexer) StateFn {
	l.Run(unicode.IsSpace)
	l.Ignore()
	if l.AcceptAnyOne("0123456789") {
		l.AcceptAnyRun("0123456789")
		l.Emit(TokenNumber)
		return semiState
	}
	if l.AcceptAnyOne("\"") {
		return stringState
	}
	return l.Errorf("Expected constant number or string!")
}

func stringState(l *Lexer) StateFn {
	for {
		l.ExceptAnyRun("\"\\")
		if l.Peek() == Eof {
			return l.Errorf("EOF in the middle of a string!")
		}
		if l.AcceptAnyOne("\"") {
			l.Emit(TokenString)
			return semiState
		}
		if l.AcceptAnyOne("\\") {
			if l.AcceptAnyOne("\"\\") == false {
				return l.Errorf("Expected a known escape character (\\ or \"), instead of: %c", l.Next())
			}
		}
	}
	panic("not reached")
}

func semiState(l *Lexer) StateFn {
	l.Run(unicode.IsSpace)
	l.Ignore()
	if l.AcceptAnyOne(";") {
		l.Emit(TokenSemi)
		return symbolState
	}
	return l.Errorf("Expected semicolon, instead of: %c", l.Next())
}

func TestLexer(t *testing.T) {
	var text = `
foo = 500;
barbaz="Hello world";
`
	tokens := []Token{
		Token{TokenSymbol, "foo", "anonymous", 2},
		Token{TokenEquals, "=", "anonymous", 2},
		Token{TokenNumber, "500", "anonymous", 2},
		Token{TokenSemi, ";", "anonymous", 2},
		Token{TokenSymbol, "barbaz", "anonymous", 3},
		Token{TokenEquals, "=", "anonymous", 3},
		Token{TokenString, "\"Hello world\"", "anonymous", 3},
		Token{TokenSemi, ";", "anonymous", 3},
		Token{TokenEOF, "EOF", "anonymous", 4},
		Token{TokenStopped, "", "anonymous", 4},
	}
	l := New("anonymous", text, symbolState)
	for i, expected := range tokens {
		token := l.Token()
		if token != expected {
			t.Fatalf("Token %d invalid: %#v expected %#v", i, token, expected)
		}
	}
}

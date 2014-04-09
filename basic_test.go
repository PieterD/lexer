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
	if l.Eof() {
		return l.EmitEof()
	}
	ok := l.ExceptRun(" =")
	if ok == 0 {
		return l.Errorf("Expected symbol!")
	}
	l.Emit(TokenSymbol)
	return afterSymbolState
}

func afterSymbolState(l *Lexer) StateFn {
	l.Run(unicode.IsSpace)
	l.Ignore()
	if l.Accept("=") == false {
		return l.Errorf("Expected operator '='!")
	}
	l.Emit(TokenEquals)
	return afterOperatorState
}

func afterOperatorState(l *Lexer) StateFn {
	l.Run(unicode.IsSpace)
	l.Ignore()
	if l.Accept("0123456789") {
		l.AcceptRun("0123456789")
		l.Emit(TokenNumber)
		return semiState
	}
	if l.Accept("\"") {
		return stringState
	}
	return l.Errorf("Expected constant number or string!")
}

func stringState(l *Lexer) StateFn {
	for {
		l.ExceptRun("\"\\")
		if l.Eof() {
			return l.Errorf("EOF in the middle of a string!")
		}
		if l.Accept("\"") {
			l.Emit(TokenString)
			return semiState
		}
		if l.Accept("\\") {
			if l.Accept("\"\\") == false {
				return l.Errorf("Expected a known escape character (\\ or \"), instead of: %c", l.Next())
			}
		}
	}
	panic("not reached")
}

func semiState(l *Lexer) StateFn {
	l.Run(unicode.IsSpace)
	l.Ignore()
	if l.Accept(";") {
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
		Token{TokenEmpty, "", "", 0},
	}
	l := New("anonymous", text, symbolState)
	for i, expected := range tokens {
		token := l.Token()
		if token != expected {
			t.Fatalf("Token %d invalid: %#v expected %#v", i, token, expected)
		}
	}
	// Test Go
	l = New("anonymous", text, symbolState)
	tokenchan := l.Go()
	for i, expected := range tokens {
		token := <-tokenchan
		if token != expected {
			t.Fatalf("Token %d invalid: %#v expected %#v", i, token, expected)
		}
	}
}

func TestLexTestBig(t *testing.T) {
	lt := LexTest(t, symbolState, `
hello="world";
num=500;
`)
	lt.Expect(TokenSymbol, "hello", 2)
	lt.Expect(TokenEquals, "=", 2)
	lt.Expect(TokenString, "\"world\"", 2)
	lt.Expect(TokenSemi, ";", 2)
	lt.Expect(TokenSymbol, "num", 3)
	lt.Expect(TokenEquals, "=", 3)
	lt.Expect(TokenNumber, "500", 3)
	lt.Expect(TokenSemi, ";", 3)
	lt.Expect(TokenEOF, "EOF", 4)
	lt.Expect(TokenEmpty, "", 0)
	lt.Expect(TokenEmpty, "", 0)
}

func TestLexTestSmall(t *testing.T) {
	s := afterOperatorState
	LexTest(t, s, `"hello"`).Expect(TokenString, `"hello"`, 1)
	LexTest(t, s, `"he\"llo"`).Expect(TokenString, `"he\"llo"`, 1)
	LexTest(t, s, `"he\!llo"`).Expect(TokenError, `Expected a known escape character (\ or "), instead of: !`, 1)
	LexTest(t, s, `"hello`).Expect(TokenError, `EOF in the middle of a string!`, 1)
}


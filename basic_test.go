package lexer_test

import (
	"testing"
	"unicode"

	"github.com/PieterD/lexer"
	"github.com/PieterD/lexer/lextest"
)

const (
	TokenSymbol lexer.TokenType = iota
	TokenEquals
	TokenNumber
	TokenSemi
	TokenString
)

func symbolState(l *lexer.LexInner) lexer.StateFn {
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

func afterSymbolState(l *lexer.LexInner) lexer.StateFn {
	l.Run(unicode.IsSpace)
	l.Ignore()
	if l.Accept("=") == false {
		return l.Errorf("Expected operator '='!")
	}
	l.Emit(TokenEquals)
	return afterOperatorState
}

func afterOperatorState(l *lexer.LexInner) lexer.StateFn {
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

func stringState(l *lexer.LexInner) lexer.StateFn {
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
}

func semiState(l *lexer.LexInner) lexer.StateFn {
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
	tokens := []lexer.Token{
		lexer.Token{TokenSymbol, "foo", "anonymous", 2},
		lexer.Token{TokenEquals, "=", "anonymous", 2},
		lexer.Token{TokenNumber, "500", "anonymous", 2},
		lexer.Token{TokenSemi, ";", "anonymous", 2},
		lexer.Token{TokenSymbol, "barbaz", "anonymous", 3},
		lexer.Token{TokenEquals, "=", "anonymous", 3},
		lexer.Token{TokenString, "\"Hello world\"", "anonymous", 3},
		lexer.Token{TokenSemi, ";", "anonymous", 3},
		lexer.Token{lexer.TokenEOF, "EOF", "anonymous", 3},
		lexer.Token{lexer.TokenEmpty, "", "", 0},
	}
	l := lexer.New("anonymous", text, symbolState)
	it := l.Iterate()
	if l.Iterate() != nil {
		t.Fatalf("Second Iterate should return nil")
	}
	for i, expected := range tokens {
		token := it.Token()
		if token != expected {
			t.Fatalf("Token %d invalid: %#v expected %#v", i, token, expected)
		}
	}
	// Test Go
	l = lexer.New("anonymous", text, symbolState)
	tokenchan := l.Go()
	if l.Go() != nil {
		t.Fatalf("Second go should return nil")
	}
	for i, expected := range tokens {
		token := <-tokenchan
		if token != expected {
			t.Fatalf("Token %d invalid: %#v expected %#v", i, token, expected)
		}
	}
}

func TestLexTestBig(t *testing.T) {
	lt := lextest.NewTester(t, symbolState, `
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
	lt.Expect(lexer.TokenEOF, "EOF", 3)
	lt.End()
	lt.End()
}

func TestLexTestSmall(t *testing.T) {
	s := afterOperatorState
	lextest.NewTester(t, s, `"hello"`).Expect(TokenString, `"hello"`, 1)
	lextest.NewTester(t, s, `"he\"llo"`).Expect(TokenString, `"he\"llo"`, 1)
	lextest.NewTester(t, s, `"he\!llo"`).Error(`Expected a known escape character (\ or "), instead of: !`, 1)
	lextest.NewTester(t, s, `"hello`).Error(`EOF in the middle of a string!`, 1)
}

func TestLexTooManyEmits(t *testing.T) {
	tl := lextest.NewTester(t, generateWarningState, "")
	tl.Warning("warning", 1)
	tl.Error("Too many emits in a single stat function", 1)
}

func generateWarningState(l *lexer.LexInner) lexer.StateFn {
	l.Warningf("warning")
	return tooManyEmitsState
}

func tooManyEmitsState(l *lexer.LexInner) lexer.StateFn {
	for i := 0; i <= lexer.MaxEmitsInFunction; i++ {
		l.Emit(1)
	}
	return nil
}

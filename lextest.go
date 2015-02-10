package lexinator

import (
	"testing"
)

type LexTester struct {
	l *Lexer
	t *testing.T
	n int
}

// Testing lexers involves some boiler plate.
// LexTest returns a struct value that can be used to easily
// test your lexer for correctness.
func NewTester(t *testing.T, f StateFn, text string) *LexTester {
	l := New("testing", text, f)
	return &LexTester{l, t, 0}
}

// Succeeds if the next token has the given type, value and line.
// Calls t.Fatalf with an error otherwise.
func (lt *LexTester) Expect(typ TokenType, val string, line int) *LexTester {
	lt.n++
	tok := lt.l.Token()
	if tok.Typ != typ {
		lt.t.Fatalf("Token %d: expected typ %d, got %d", lt.n, typ, tok.Typ)
	}
	if tok.Val != val {
		lt.t.Fatalf("Token %d: expected val %s, got %s", lt.n, val, tok.Val)
	}
	if tok.Line != line {
		lt.t.Fatalf("Token %d: expected line %d, got %d", lt.n, line, tok.Line)
	}
	return lt
}

// Succeeds if the next token is the empty token.
func (lt *LexTester) End() {
	tok := lt.l.Token()
	if tok.Typ != TokenEmpty {
		lt.t.Fatalf("Token %d: expected typ 0, got %d", lt.n, tok.Typ)
	}
	if tok.Val != "" {
		lt.t.Fatalf("Token %d: expected empty val, got %d", lt.n, tok.Typ)
	}
	if tok.Line != 0 {
		lt.t.Fatalf("Token %d: expected line 0, got %d", lt.n, tok.Typ)
	}
}

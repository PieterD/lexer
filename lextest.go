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

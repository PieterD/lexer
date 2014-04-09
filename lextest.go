package lexinator

import (
	"testing"
)

type lexTest struct {
	l *Lexer
	t *testing.T
	n int
}

// Testing lexers involves some boiler plate.
// LexTest returns a struct value that can be used to easily
// test your lexer for correctness.
func LexTest (t *testing.T, f StateFn, text string) *lexTest {
	l := New("testing", text, f)
	return &lexTest{l, t, 0}
}

func (lt *lexTest) Expect(typ TokenType, val string, line int) *lexTest {
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


package lexinator

import (
	"testing"
)

type Tester struct {
	l Lexinator
	t *testing.T
	n int
}

// Testing lexers involves some boiler plate.
// LexTest returns a struct value that can be used to easily
// test your lexer for correctness.
func NewTester(t *testing.T, f StateFn, text string) *Tester {
	l := New("testing", text, f)
	return &Tester{l, t, 0}
}

// Succeeds if the next token has the given type, value and line.
// Calls t.Fatalf with an error otherwise.
func (lt *Tester) Expect(typ TokenType, val string, line int) *Tester {
	lt.n++
	tok := lt.l.Token()
	if tok.Typ != typ || tok.Val != val || tok.Line != line {
		lt.t.Fatalf("Token %d: expected [typ:%d line:%d val:'%s'] got [typ:%d line:%d val:'%s']", lt.n,
			typ, line, val, tok.Typ, tok.Line, tok.Val)
	}
	return lt
}

// Succeeds if the next token is an error with the given value and line.
func (lt *Tester) Error(val string, line int) *Tester {
	return lt.Expect(TokenError, val, line)
}

// Succeeds if the next token is the empty token.
func (lt *Tester) End() {
	tok := lt.l.Token()
	if tok.Typ != TokenEmpty || tok.Val != "" || tok.Line != 0 {
		lt.t.Fatalf("Token %d: expected end token, got [typ:%d line:%d val:'%s']", lt.n,
			tok.Typ, tok.Line, tok.Val)
	}
}

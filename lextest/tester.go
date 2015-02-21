package lextest

import (
	"testing"

	"github.com/PieterD/lexer"
)

type Tester struct {
	it *lexer.Iterator
	t  *testing.T
	n  int
}

// Testing lexers involves some boiler plate.
// LexTest returns a struct value that can be used to easily
// test your lexer for correctness.
func NewTester(t *testing.T, f lexer.StateFn, text string) *Tester {
	it := lexer.New("testing", text, f).Iterate()
	return &Tester{it, t, 0}
}

// Succeeds if the next token has the given type, value and line.
// Calls t.Fatalf with an error otherwise.
func (lt *Tester) Expect(typ lexer.TokenType, val string, line int) *Tester {
	lt.n++
	tok := lt.it.Token()
	if tok.Typ != typ || tok.Val != val || tok.Line != line {
		lt.t.Logf("Token %d:      got [typ:%2d line:%3d val:'%s']", lt.n, tok.Typ, tok.Line, tok.Val)
		lt.t.Logf("Token %d: expected [typ:%2d line:%3d val:'%s']", lt.n, typ, line, val)
		lt.t.Fatalf("Token %d Expect failed", lt.n)
	}
	return lt
}

// Succeeds if the next token is a warning with the given value and line.
func (lt *Tester) Warning(val string, line int) *Tester {
	return lt.Expect(lexer.TokenWarning, val, line)
}

// Succeeds if the next token is an error with the given value and line.
func (lt *Tester) Error(val string, line int) *Tester {
	return lt.Expect(lexer.TokenError, val, line)
}

// Succeeds if the next token is the empty token.
func (lt *Tester) End() *Tester {
	return lt.Expect(lexer.TokenEmpty, "", 0)
}

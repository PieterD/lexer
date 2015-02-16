package lexer

import "testing"

func TestTokenString(t *testing.T) {
	tok := Token{Typ: TokenError, Val: "errorvalue"}
	if tok.String() != "Error: errorvalue" {
		t.Fatalf("TokenError 'errorvalue' expected, got %s", tok.String())
	}

	tok = Token{Typ: TokenWarning, Val: "warningvalue"}
	if tok.String() != "Warning: warningvalue" {
		t.Fatalf("TokenWarning 'warningvalue' expected, got %s", tok.String())
	}

	tok = Token{Typ: TokenEmpty, Val: "emptyvalue"}
	if tok.String() != "Empty Token" {
		t.Fatalf("TokenEmpty 'Empty Token' expected, got %s", tok.String())
	}

	tok = Token{Typ: TokenEOF, Val: "eofvalue"}
	if tok.String() != "EOF" {
		t.Fatalf("TokenEmpty 'EOF' expected, got %s", tok.String())
	}

	tok = Token{Typ: 99, Val: "short"}
	if tok.String() != "\"short\"" {
		t.Fatalf("Short token expected, got %s", tok.String())
	}

	tok = Token{Typ: 99, Val: "much longer value here"}
	if tok.String() != "\"much longer val\"..." {
		t.Fatalf("Long token expected, got %s", tok.String())
	}
}

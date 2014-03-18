package lexinator

import "fmt"

// Tokens are emitted by the lexer. They contained a (usually) user-defined
// Typ, the Value of the token, and the Filename and Line number where the
// token was generated.
type Token struct {
	Typ  TokenType
	Val  string
	File string
	Line int
}

// TokenType is an integer representing the type of token that has been emitted.
// Most TokenTypes will be userdefined, and those that are must be greater than 0.
type TokenType int

const (
	// TokenEmpty is the TokenType with value 0.
	// Any zero-valued token will have this as its Typ.
	TokenEmpty TokenType = -iota
	// TokenStopped is contained by tokens returned by Lexer.Token when
	// the lexer has stopped (by an error, or Eof)
	TokenStopped
	// TokenError is the Typ for errors reported by, for example, Lexer.Errorf.
	TokenError
	// TokenEOF is returned once per file, when the end of file has been reached.
	TokenEOF
)

// Return a simple string representation of the value contained within the token.
func (i Token) String() string {
	switch i.Typ {
	case TokenError:
		return i.Val
	case TokenEOF:
		return "EOF"
	}
	if len(i.Val) > 10 {
		return fmt.Sprintf("%.15q...", i.Val)
	}
	return fmt.Sprintf("%q", i.Val)
}

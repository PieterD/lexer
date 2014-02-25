package lexinator

import "fmt"

type Token struct {
	Typ  TokenType
	Val  string
	File string
	Line int
}

type TokenType int

const (
	TokenStopped TokenType = -1 - iota
	TokenError
	TokenEOF
)

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

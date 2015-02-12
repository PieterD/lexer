package lexinator

// Lexinator is the external type which emits tokens.
type Lexinator struct {
	lexer *LexInner
}

// Create a new lexer.
func New(name string, input string, start_state StateFn) Lexinator {
	var ln Lexinator
	ln.lexer = new(LexInner)
	l := ln.lexer
	l.tokens = make(chan Token, 10)
	l.state = start_state
	l.name = name
	l.input = input
	l.mark.line = 1
	l.prev.line = 1
	return ln
}

// Spawn a goroutine which keeps sending tokens on the returned channel,
// until TokenEmpty would be encountered.
func (ln Lexinator) Go() <-chan Token {
	l := ln.lexer
	l.going = true
	go func() {
		defer close(l.tokens)
		for {
			l.state = l.state(l)
			if l.state == nil {
				return
			}
		}
	}()
	return l.tokens
}

// Get a Token from the Lexer.
// Please note that only 10 tokens can be emitted in a single lex function.
// If you wish to emite more per function, use the Go method.
func (ln Lexinator) Token() Token {
	l := ln.lexer
	if l.going {
		return Token{}
	}

	for {
		select {
		case token, ok := <-l.tokens:
			if !ok {
				return Token{TokenEmpty, "", "", 0}
			}
			return token
		default:
			l.state = l.state(l)
			if l.state == nil {
				close(l.tokens)
			}
		}
	}
	panic("not reached")
}

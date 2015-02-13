package lexer

// The maximum number of emits in a single state function when using Token.
// If this number has been reached, Token returns a StateError.
// If you wish to emit more than this, use the Go method to read tokens
// off the channel directly.
const MaxEmitsInFunction = 10

// Lexer is the external type which emits tokens.
type Lexer struct {
	lexer *LexInner
	going bool
}

// Create a new lexer.
func New(name string, input string, start_state StateFn) *Lexer {
	ln := new(Lexer)
	ln.lexer = new(LexInner)
	l := ln.lexer
	l.tokens = make(chan Token, MaxEmitsInFunction)
	l.state = start_state
	l.name = name
	l.input = input
	l.mark.line = 1
	l.prev.line = 1
	return ln
}

// Spawn a goroutine which keeps sending tokens on the returned channel,
// until TokenEmpty would be encountered.
// If Go has already been called, it will return nil.
func (ln *Lexer) Go() <-chan Token {
	if ln.going {
		return nil
	}
	ln.going = true
	l := ln.lexer
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
// Please note that only 10 tokens can be emitted in a single state function.
// If you wish to emit more per function, use the Go method.
func (ln *Lexer) Token() (token Token) {
	if ln.going {
		return Token{TokenEmpty, "", "", 0}
	}
	l := ln.lexer

	defer func() {
		err := recover()
		if err == errTooManyEmits {
			token = Token{TokenError, errTooManyEmits.Error(), l.name, l.mark.line}
		}
	}()

	for {
		var ok bool
		select {
		case token, ok = <-l.tokens:
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

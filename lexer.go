package lexer

// The maximum number of emits in a single state function when using Token.
// If this number has been reached, Token returns a StateError.
// If you wish to emit more than this, use the Go method to read tokens
// off the channel directly.
const MaxEmitsInFunction = 10

// Generates tokens asynchronously. See Lexer.Go
type Channel <-chan Token

// Generates tokens synchronously. See Lexer.Iterate
type Iterator struct {
	l *LexInner
}

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
// If Go or Iterate has already been called, it will return nil.
func (ln *Lexer) Go() Channel {
	if ln.going {
		return nil
	}
	ln.going = true
	l := ln.lexer
	l.async = true
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

// Where Go starts a goroutine, Iterate returns an iterator.
// When using an Iterator, only MaxEmitsInFunction emits may be done
// in any single state function, or an error will be reported.
// If Go or Iterate has already been called, it will return nil.
func (ln *Lexer) Iterate() *Iterator {
	if ln.going {
		return nil
	}
	ln.going = true
	return &Iterator{ln.lexer}
}

// Get a Token from the Lexer.
// Please note that only 10 tokens can be emitted in a single state function.
// If you wish to emit more per function, use the Go method.
func (it Iterator) Token() (token Token) {
	l := it.l

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
}

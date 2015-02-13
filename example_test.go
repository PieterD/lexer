package lexer_test

import (
	"fmt"
	"unicode"

	"github.com/PieterD/lexer"
)

const (
	tokenComment lexer.TokenType = 1 + iota
	tokenVariable
	tokenAssign
	tokenNumber
	tokenString
)

func ExampleLexer() {
	text := `
/* comment */
pie=314
// comment
string = "Hello world!"
`
	l := lexer.New("filename", text, state_base)
	tokenchan := l.Go()
	for token := range tokenchan {
		fmt.Printf("%s:%d [%d]\"%s\"\n", token.File, token.Line, token.Typ, token.Val)
	}
	// Output: filename:2 [1]"/* comment */"
	// filename:3 [2]"pie"
	// filename:3 [3]"="
	// filename:4 [4]"314"
	// filename:6 [1]"// comment"
	// filename:7 [2]"string"
	// filename:7 [3]"="
	// filename:7 [5]""Hello world!""
	// filename:8 [-3]"EOF"
}

// Start parsing with this.
func state_base(l *lexer.LexInner) lexer.StateFn {
	// Ignore all whitespace.
	l.Run(unicode.IsSpace)
	l.Ignore()
	if l.String("//") {
		// We're remembering the '//' here so it gets included in the Emit
		// contained in state_comment_line.
		return state_comment_line
	}
	if l.String("/*") {
		return state_comment_block(state_base)
	}
	if l.Eof() {
		return l.EmitEof()
	}
	// It's not a comment or Eof, so it must be a variable name.
	return state_variable
}

// Parse a line comment.
func state_comment_line(l *lexer.LexInner) lexer.StateFn {
	// Eat up everything until end of line (or Eof)
	l.ExceptRun("\n")
	l.Emit(tokenComment)
	// Consume the end of line. If we reached Eof, this does nothing.
	l.Accept("\n")
	// Ignore that last newline
	l.Ignore()
	return state_base
}

// Parse a block comment.
// Since block comments may appear in different states,
// instead of defining the usual StateFn we define a function that
// returns a statefn, which in turn will return the parent state
// after its parsing is done.
func state_comment_block(parent lexer.StateFn) lexer.StateFn {
	return func(l *lexer.LexInner) lexer.StateFn {
		if !l.Find("*/") {
			// If closing statement couldn't be found, emit an error.
			// Errorf always returns nil, so parsing is done after this.
			return l.Errorf("Couldn't find end of block comment")
		}
		l.String("*/")
		l.Emit(tokenComment)
		return parent
	}
}

// Parse a variable name
func state_variable(l *lexer.LexInner) lexer.StateFn {
	if l.AcceptRun("abcdefghijklmnopqrstuvwxyz") == 0 {
		return l.Errorf("Invalid variable name")
	}
	l.Emit(tokenVariable)

	return state_operator
}

// Parse an assignment operator
func state_operator(l *lexer.LexInner) lexer.StateFn {
	l.Run(unicode.IsSpace)
	l.Ignore()
	if l.Accept("=") {
		l.Emit(tokenAssign)
		return state_value
	}
	return l.Errorf("Only '=' is a valid operator")
}

// Parse a value
func state_value(l *lexer.LexInner) lexer.StateFn {
	l.Run(unicode.IsSpace)
	l.Ignore()
	if l.AcceptRun("0123456789") > 0 {
		l.Emit(tokenNumber)
		return state_base
	}
	if l.Accept("\"") {
		return state_string
	}
	return l.Errorf("Unidentified value")
}

// Parse a string
func state_string(l *lexer.LexInner) lexer.StateFn {
	for {
		l.ExceptRun("\"\\")
		// Now we're either at a ", a \, or Eof.
		if l.Accept("\"") {
			l.Emit(tokenString)
			return state_base
		}
		if l.Accept("\\") {
			if !l.Accept("nrt\"'\\") {
				return l.Errorf("Invalid escape sequence: \"\\%c\"", l.Last())
			}
		}
		if l.Eof() {
			return l.Errorf("No closing '\"' found")
		}
	}
}

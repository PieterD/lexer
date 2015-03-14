package lexer

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

var errTooManyEmits = errors.New("Too many emits in a single stat function")

// This is returned by next when there are no more characters to read.
const Eof rune = -1

// This is returned when a bad rune is encountered.
const Err rune = utf8.RuneError

// LexInner is the inner type which is used within StateFn to do the actual lexing.
type LexInner struct {
	tokens chan Token
	state  StateFn
	name   string
	input  string
	mark   Mark
	prev   Mark
	async  bool
}

// The Mark type (used by Mark and Unmark) can be used to save
// the current state of the lexer, and restore it later.
type Mark struct {
	pos     int
	line    int
	start   int
	width   int
	replace *Replacer
}

func (mark Mark) rpos() int {
	return mark.pos - mark.start
}

type Replacer struct {
	start, end Mark
	with       string
}

// StateFn is a function that takes a LexInner and returns a StateFn.
type StateFn func(*LexInner) StateFn

// Return the length of the token gathered so far.
func (l *LexInner) Len() int {
	return l.mark.pos - l.mark.start
}

// Get the string of the token gathered so far.
func (l *LexInner) Get() string {
	str := l.input[l.mark.start:l.mark.pos]
	return str
}

// Get the last character accepted into the token.
func (l *LexInner) Last() rune {
	if l.Len() == 0 {
		return Eof
	}
	r, _ := utf8.DecodeLastRuneInString(l.Get())
	return r
}

// Store the state of the lexer.
func (l *LexInner) Mark() Mark {
	return l.mark
}

// Recover the state of the lexer.
func (l *LexInner) Unmark(mark Mark) {
	l.mark = mark
}

// Replace the text from the start Mark to the current position with the given string.
// With may be a different length than the string being replaced, but this change
// will not be reflected by functions like Len and Get.
// Call ReplaceGet to get the token including its replaces. This is how it will be sent by Emit.
// The replace is part of the current Mark, so Unmarking to before a replace was done will
// remove the replace.
func (l *LexInner) Replace(start Mark, with string) {
	if with != "" {
		l.mark.replace = &Replacer{start, l.mark, with}
	}
}

// Get the current token with all replaces included.
// This can be expensive, if you have many replaces.
// Without any replaces, it is identical to Get.
func (l *LexInner) ReplaceGet() string {
	return replaceGet(l.Get(), l.mark.replace)
}

func replaceGet(cur string, replace *Replacer) string {
	if replace == nil {
		return cur
	}
	start := replace.start
	end := replace.end
	return replaceGet(cur[:start.rpos()]+replace.with+cur[end.rpos():], replace.end.replace)
}

// Emit a token with the given type and string.
func (l *LexInner) EmitString(typ TokenType, str string) {
	tok := Token{typ, str, l.name, l.mark.line}
	if l.async {
		l.tokens <- tok
	} else {
		select {
		case l.tokens <- tok:
		default:
			panic(errTooManyEmits)
		}
	}
}

// Emit the gathered token, given its type.
// Emits the result of ReplaceGet, then calls Ignore.
func (l *LexInner) Emit(typ TokenType) {
	l.EmitString(typ, l.ReplaceGet())
	l.Ignore()
}

// Emit a token of type TokenEOF.
// Returns nil.
func (l *LexInner) EmitEof() StateFn {
	l.EmitString(TokenEOF, "EOF")
	return nil
}

// Emit an Error token.
// Like EmitEof, Errorf returns nil.
func (l *LexInner) Errorf(format string, args ...interface{}) StateFn {
	l.EmitString(TokenError, fmt.Sprintf(format, args...))
	return nil
}

// Emit a Warning token.
func (l *LexInner) Warningf(format string, args ...interface{}) {
	l.EmitString(TokenWarning, fmt.Sprintf(format, args...))
}

// Return true if the lexer has reached the end of the file.
func (l *LexInner) Eof() bool {
	if l.mark.pos >= len(l.input) {
		return true
	}
	return false
}

// Read a single character.
// If there are no more characters, it will return Eof.
// If a non-utf8 character is read, it will return Err.
func (l *LexInner) Next() (char rune) {
	if l.Eof() {
		l.mark.width = 0
		char = Eof
		return Eof
	}
	char, l.mark.width = utf8.DecodeRuneInString(l.input[l.mark.pos:])
	l.mark.pos += l.mark.width
	if char == '\n' {
		l.mark.line++
	}
	return char
}

// Read n characters.
// Returns the number of characters read.
// If it returns less than n, it will have reached EOF.
func (l *LexInner) Skip(n int) int {
	i := 0
	for i = 0; i < n; i++ {
		if l.Next() == Eof {
			return i
		}
	}
	return i
}

// Undo the last Next.
// This is probably won't work after calling any other lexer functions.
// If you need to undo more, use Mark and Unmark.
func (l *LexInner) Back() {
	if l.Last() == '\n' {
		l.mark.line--
	}
	l.mark.pos -= l.mark.width
	l.mark.width = 0
}

// Spy on the upcoming rune.
func (l *LexInner) Peek() rune {
	char := l.Next()
	l.Back()
	return char
}

// Ignore everything gathered about the token so far.
// Also removes any Replaces.
func (l *LexInner) Ignore() {
	l.mark.start = l.mark.pos
	l.mark.width = 0
	l.mark.replace = nil
}

// Retry everything since starting this token.
func (l *LexInner) Retry() {
	l.mark.pos = l.mark.start
	l.mark.width = 0
}

// Attempt to read a string.
// Only if the entire string is successfully accepted does it return true.
// If only a part of the string was matched, none of it is.
func (l *LexInner) String(valid string) bool {
	if strings.HasPrefix(l.input[l.mark.pos:], valid) {
		l.mark.line += strings.Count(valid, "\n")
		l.mark.pos += len(valid)
		l.mark.width = len(valid)
		return true
	}
	return false
}

// Accepts things until the first occurence of the given string.
// The string itself is not accepted.
func (l *LexInner) Find(valid string) bool {
	idx := strings.Index(l.input[l.mark.pos:], valid)
	if idx >= 0 {
		l.mark.line += strings.Count(l.input[l.mark.pos:l.mark.pos+idx], "\n")
		l.mark.pos += idx
		l.mark.width = idx
		return true
	}
	return false
}

// Accept a single character and return true if f returns true.
// Otherwise, do nothing and return false.
func (l *LexInner) One(f func(rune) bool) bool {
	if f(l.Next()) {
		return true
	}
	l.Back()
	return false
}

// Reads characters and feeds them to the given function, and keeps reading until it returns false.
func (l *LexInner) Run(f func(rune) bool) (acceptnum int) {
	for l.One(f) {
		acceptnum++
	}
	return
}

// Accepts any whitespace (unicode.IsSpace), except for whitespace in except.
// For instance, Whitespace("\n") will accept all whitespace except newlines.
// Returns the number of runes read.
func (l *LexInner) Whitespace(except string) (acceptnum int) {
	return l.Run(func(r rune) bool {
		return unicode.IsSpace(r) && !strings.ContainsRune(except, r)
	})
}

func acceptAny(valid string) func(rune) bool {
	return func(char rune) bool {
		return (strings.IndexRune(valid, char) >= 0)
	}
}

// Read one character, but only if it is one of the characters in the given string.
func (l *LexInner) Accept(valid string) bool {
	return l.One(acceptAny(valid))
}

// Read as many characters as possible, but only characters that exist in the given string.
func (l *LexInner) AcceptRun(valid string) (acceptnum int) {
	return l.Run(acceptAny(valid))
}

func not(in func(rune) bool) func(rune) bool {
	return func(char rune) bool {
		if char == Eof {
			return false
		}
		return !in(char)
	}
}

// Read one character, but only if it is NOT one of the characters in the given string.
// If Eof is reached, Except fails regardless of what the given string is.
func (l *LexInner) Except(valid string) bool {
	return l.One(not(acceptAny(valid)))
}

// Read as many characters as possible, but only characters that do NOT exist in the given string.
// If Eof is reached, ExceptRun stops as though it found a successful character.
// Thus, ExceptRun("") accepts everything until Eof.
func (l *LexInner) ExceptRun(valid string) (acceptnum int) {
	return l.Run(not(acceptAny(valid)))
}

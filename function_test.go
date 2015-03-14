package lexer

import (
	"testing"
	"unicode/utf8"
)

func TestBasic(t *testing.T) {
	ln := New("test", "ABtestingXYXYZ\nline2", nil)
	l := ln.lexer
	start := l.Mark()
	if l.Next() != 'A' || l.Next() != 'B' {
		t.Fatalf("Next didn't return expected value")
	}
	l.Back()
	stored := l.Mark()
	if l.Next() != 'B' {
		t.Fatalf("Back followed by Next didn't return the same character")
	}
	if l.Len() != 2 {
		t.Fatalf("Len returned wrong value")
	}
	if l.Get() != "AB" {
		t.Fatalf("Get returned wrong string")
	}
	if l.Last() != 'B' {
		t.Fatalf("Last returned wrong character")
	}
	if l.Peek() != 't' {
		t.Fatalf("Peek returned wrong character")
	}
	if !l.String("testing") || l.Get() != "ABtesting" {
		t.Fatalf("String failed")
	}
	l.Unmark(stored)
	if l.Next() != 'B' {
		t.Fatalf("Next returned wrong string")
	}
	l.Retry()
	if !l.Find("testing") || l.Get() != "AB" {
		t.Fatalf("Find failed")
	}
	l.Ignore()
	if l.Len() != 0 {
		t.Fatalf("Ignore failed")
	}
	if l.AcceptRun("tse") != 4 || l.Get() != "test" {
		t.Fatalf("AcceptAnyRun failed")
	}
	stored = l.Mark()
	if l.ExceptRun("XY") != 3 || l.Get() != "testing" {
		t.Fatalf("ExceptAnyRun failed")
	}
	l.Ignore()
	if l.Except("XYXY") {
		t.Fatalf("ExceptAnyOne succeeded")
	}
	if l.Find("spoopty") {
		t.Fatalf("Find succeeded")
	}
	if l.String("spoopty") {
		t.Fatalf("String succeeded")
	}
	if !l.Find("line2") {
		t.Fatalf("Find failed")
	}
	if l.mark.line != 2 {
		t.Fatalf("Wrong line")
	}
	l.Unmark(stored)
	if l.mark.line != 1 {
		t.Fatalf("Wrong line")
	}
	//if l.ExceptRun("Q") == 0 {
	//	t.Fatalf("Bad ExceptRun")
	//}
	l.Unmark(start)
	if !l.String("ABtestingXYXYZ\nline2") {
		t.Fatalf("String failed")
	}
	if l.Last() != '2' {
		t.Fatalf("Expected '2' from Last, got '%c'", l.Last())
	}
	if !l.Eof() {
		t.Fatalf("Expected Eof to return true")
	}
	if l.Next() != Eof {
		t.Fatalf("Expected Next to return Eof")
	}
	if l.Accept("\n") {
		t.Fatalf("Did not expect a successful accept")
	}
	if l.Except("\n") {
		t.Fatalf("Did not expect a successful except")
	}
	l.Unmark(start)
	l.ExceptRun("")
	if l.Next() != Eof {
		t.Fatalf("Expected Eof after ExceptRun(\"\")")
	}
	l.Ignore()
	if l.Last() != Eof {
		t.Fatalf("Expected Eof from Last, got %c", l.Last())
	}
	l.Unmark(start)
	n := l.Skip(2)
	if n != 2 {
		t.Fatalf("Expected Skip(2) to return 2, got %d", n)
	}
	if l.Get() != "AB" {
		t.Fatalf("Expected Skip(2) to return 'AB', got %s", l.Get())
	}
	l.ExceptRun("\n")
	l.Next()
	l.Ignore()
	n = l.Skip(6)
	if n != 5 {
		t.Fatalf("Expected short skip to return 5, got %d", n)
	}
	if l.Get() != "line2" {
		t.Fatalf("Expected skip to read 'line2', got %s", l.Get())
	}
}

func TestBadRune(t *testing.T) {
	ln := New("test", "\xff", nil)
	l := ln.lexer
	r := l.Next()
	if r != utf8.RuneError {
		t.Fatalf("Expected RuneError, got: %d", r)
	}
}

func TestWhitespace(t *testing.T) {
	ln := New("test", "\t\thello \r\t\nworld\r  ", nil)
	l := ln.lexer
	n := l.Whitespace("")
	if n != 2 {
		t.Fatalf("Expected 2, got %d", n)
	}
	if !l.String("hello") {
		t.Fatalf("Failed to get 'hello'")
	}
	n = l.Whitespace("\n")
	if n != 3 {
		t.Fatalf("Expected 3, got %d", n)
	}
	ch := l.Next()
	if ch != '\n' {
		t.Fatalf("Expected '\\n', got %d", ch)
	}
	if !l.String("world") {
		t.Fatalf("Failed to get 'world'")
	}
	n = l.Whitespace("\n")
	if n != 3 {
		t.Fatalf("Expected 3, got %d", n)
	}
	ch = l.Next()
	if ch != Eof {
		t.Fatalf("Expected Eof, got %d", ch)
	}
}

func TestReplace(t *testing.T) {
	ln := New("test", "Hello, world!", nil)
	l := ln.lexer
	start := l.Mark()
	if !l.String("Hello") {
		t.Fatalf("Expected 'Hello'")
	}
	mark := l.Mark()
	if !l.String(", ") {
		t.Fatalf("Expected ', '")
	}
	// Without any replaces
	if l.ReplaceGet() != "Hello, " {
		t.Fatalf("Expected 'Hello, '")
	}

	l.Replace(mark, "!")
	mark = l.Mark()
	if l.Next() != 'w' {
		t.Fatalf("Expected 'w'")
	}
	l.Replace(mark, "W")
	if !l.String("orld!") {
		t.Fatalf("Expected 'orld!")
	}
	if l.ReplaceGet() != "Hello!World!" {
		t.Fatalf("Expected 'Hello!World!', got '%s'", l.ReplaceGet())
	}

	// Undo that last replace
	l.Unmark(mark)
	if !l.String("world!") {
		t.Fatalf("Expected 'world!'")
	}
	if l.ReplaceGet() != "Hello!world!" {
		t.Fatalf("Expected 'Hello!world!', got '%s'", l.ReplaceGet())
	}

	// Back to the start, prepend something.
	l.Unmark(start)
	l.Replace(start, "Start, ")
	if !l.String("Hello") {
		t.Fatalf("Expected 'Hello'")
	}
	if l.ReplaceGet() != "Start, Hello" {
		t.Fatalf("Expected 'Start, Hello' got '%s''", l.ReplaceGet())
	}

	// Append something
	l.Unmark(start)
	if !l.String("Hello") {
		t.Fatalf("Expected 'Hello'")
	}
	l.Replace(l.Mark(), ", end!")
	if l.ReplaceGet() != "Hello, end!" {
		t.Fatalf("Expected 'Hello, end!' got '%s'", l.ReplaceGet())
	}
}

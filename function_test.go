package lexinator

import (
	"testing"
)

func TestBasic(t *testing.T) {
	l := New("test", "ABtestingXYXYZ\nline2", nil)
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
	if l.Next() != Eof {
		t.Fatalf("Expected Eof")
	}
	if l.Accept("\n") {
		t.Fatalf("Did not expect a successful accept")
	}
	if l.Except("\n") {
		t.Fatalf("Did not expect a successful except")
	}
}

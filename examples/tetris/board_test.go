package tetris

import (
	"strings"
	"testing"
)

func TestBoardDimensions(t *testing.T) {
	var b Board
	if len(b) != BoardRows {
		t.Fatalf("expected %d rows, got %d", BoardRows, len(b))
	}
	if len(b[0]) != BoardCols {
		t.Fatalf("expected %d cols, got %d", BoardCols, len(b[0]))
	}
}

func TestEmptyBoardPrintsTwentyRowsOfTenDots(t *testing.T) {
	var b Board
	got := b.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != BoardRows {
		t.Fatalf("expected %d rows in output, got %d", BoardRows, len(lines))
	}
	for i, line := range lines {
		if line != strings.Repeat(".", BoardCols) {
			t.Errorf("row %d: expected %q, got %q",
				i, strings.Repeat(".", BoardCols), line)
		}
	}
}

func TestFilledCellRendersColouredBlock(t *testing.T) {
	var b Board
	b[0][0] = 1
	out := b.String()
	first := strings.SplitN(out, "\n", 2)[0]
	if !strings.Contains(first, "\x1b[36m") {
		t.Errorf("expected cyan ANSI prefix in row 0, got %q", first)
	}
	if !strings.Contains(first, "█") {
		t.Errorf("expected block glyph in row 0, got %q", first)
	}
	if !strings.Contains(first, ansiReset) {
		t.Errorf("expected ANSI reset after coloured cell, got %q", first)
	}
}

func TestOutOfRangeCellRendersAsEmpty(t *testing.T) {
	var b Board
	b[0][0] = 99
	first := strings.SplitN(b.String(), "\n", 2)[0]
	if first[0] != '.' {
		t.Errorf("expected dot for out-of-range cell, got %q", first[:1])
	}
}

func TestStringHasTrailingNewlinePerRow(t *testing.T) {
	var b Board
	out := b.String()
	if strings.Count(out, "\n") != BoardRows {
		t.Errorf("expected %d newlines, got %d",
			BoardRows, strings.Count(out, "\n"))
	}
}

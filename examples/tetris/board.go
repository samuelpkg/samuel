// Package tetris implements a minimal text-mode Tetris.
package tetris

import "strings"

// Board dimensions for the standard Tetris playing field.
const (
	BoardRows = 20
	BoardCols = 10
)

// Cell holds the colour id of a single board square.
// A zero value means the square is empty; 1-7 map to the
// seven tetromino colour ids (I, O, T, S, Z, L, J).
type Cell byte

// Board models the playing field as a fixed [BoardRows][BoardCols] grid.
type Board [BoardRows][BoardCols]Cell

// ansiColors maps each non-empty Cell value to an ANSI SGR colour code.
// Index 0 is unused (empty cells render as a dot without colour).
var ansiColors = [8]string{
	"",
	"\x1b[36m", // 1 I — cyan
	"\x1b[33m", // 2 O — yellow
	"\x1b[35m", // 3 T — magenta
	"\x1b[32m", // 4 S — green
	"\x1b[31m", // 5 Z — red
	"\x1b[34m", // 6 L — blue
	"\x1b[37m", // 7 J — white
}

const ansiReset = "\x1b[0m"

// String renders the board to ANSI-coloured text. Empty cells become
// a dot; filled cells become a coloured block character. Each row is
// terminated by a newline.
func (b *Board) String() string {
	var sb strings.Builder
	sb.Grow(BoardRows * (BoardCols*8 + 1))
	for row := 0; row < BoardRows; row++ {
		for col := 0; col < BoardCols; col++ {
			writeCell(&sb, b[row][col])
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// writeCell appends the rendered representation of a single cell.
func writeCell(sb *strings.Builder, c Cell) {
	if c == 0 || int(c) >= len(ansiColors) {
		sb.WriteByte('.')
		return
	}
	sb.WriteString(ansiColors[c])
	sb.WriteRune('█')
	sb.WriteString(ansiReset)
}

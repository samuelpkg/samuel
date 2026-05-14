package tetris

import "time"

// TickInterval is the gravity cadence: one descent every 500 ms.
const TickInterval = 500 * time.Millisecond

// Active is a tetromino positioned on the board by the (row, col) of its
// 4×4 bounding-box top-left corner. Row and col may be negative while the
// piece spawns above the playfield.
type Active struct {
	Piece    Tetromino
	Row, Col int
}

// Game holds the playing state: the board, the active piece, the running
// score, and a GameOver flag set once a piece cannot spawn. The zero value
// is a fresh game with an empty board and no active piece configured.
type Game struct {
	Board    Board
	Active   Active
	Score    int
	GameOver bool
}

// Tick advances gravity by one step. If GameOver is already set, Tick is
// a no-op and returns false. Otherwise: if the active piece can descend
// by a row, it descends and Tick returns false; if not, the piece locks,
// any complete rows clear, the score gains 10×cleared², and Tick returns
// true to signal that a new piece should be spawned.
func (g *Game) Tick() bool {
	if g.GameOver {
		return false
	}
	next := g.Active
	next.Row++
	if g.fits(next) {
		g.Active = next
		return false
	}
	g.lock()
	cleared := g.clearLines()
	g.Score += 10 * cleared * cleared
	return true
}

// Spawn places piece at the given (row, col) bounding-box top-left. If
// the position is blocked — even one filled bit overlaps the stack or
// falls outside the board — Spawn sets GameOver and returns false without
// touching g.Active. On success it updates g.Active and returns true.
func (g *Game) Spawn(piece Tetromino, row, col int) bool {
	a := Active{Piece: piece, Row: row, Col: col}
	if !g.fits(a) {
		g.GameOver = true
		return false
	}
	g.Active = a
	return true
}

// fits reports whether a would sit entirely on empty, in-bounds cells.
// Filled bits of the piece must land inside the board and on Cell == 0.
func (g *Game) fits(a Active) bool {
	bits := Bitmask(a.Piece)
	for i := 0; i < 16; i++ {
		if bits&(1<<i) == 0 {
			continue
		}
		r := a.Row + i/4
		c := a.Col + i%4
		if r < 0 || r >= BoardRows || c < 0 || c >= BoardCols {
			return false
		}
		if g.Board[r][c] != 0 {
			return false
		}
	}
	return true
}

// lock stamps the active piece into the board using its palette colour.
// Bits that fall outside the board are skipped — useful while a piece is
// still partially above the playfield at spawn time.
func (g *Game) lock() {
	bits := Bitmask(g.Active.Piece)
	colour := Cell(g.Active.Piece.Shape + 1)
	for i := 0; i < 16; i++ {
		if bits&(1<<i) == 0 {
			continue
		}
		r := g.Active.Row + i/4
		c := g.Active.Col + i%4
		if r < 0 || r >= BoardRows || c < 0 || c >= BoardCols {
			continue
		}
		g.Board[r][c] = colour
	}
}

// clearLines removes every fully-occupied row, shifts the rows above
// downward to fill the gaps, and returns how many rows were cleared.
func (g *Game) clearLines() int {
	cleared := 0
	write := BoardRows - 1
	for read := BoardRows - 1; read >= 0; read-- {
		if rowFull(&g.Board[read]) {
			cleared++
			continue
		}
		g.Board[write] = g.Board[read]
		write--
	}
	for write >= 0 {
		g.Board[write] = [BoardCols]Cell{}
		write--
	}
	return cleared
}

// rowFull reports whether every cell in row is non-empty.
func rowFull(row *[BoardCols]Cell) bool {
	for _, c := range row {
		if c == 0 {
			return false
		}
	}
	return true
}

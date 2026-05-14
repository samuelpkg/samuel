package tetris

import (
	"testing"
	"time"
)

func TestTickInterval_Is500ms(t *testing.T) {
	if TickInterval != 500*time.Millisecond {
		t.Fatalf("expected 500ms, got %v", TickInterval)
	}
}

func TestTick_DescendsWhenClear(t *testing.T) {
	g := Game{Active: Active{Piece: Tetromino{Shape: ShapeI}, Row: 0, Col: 3}}
	if locked := g.Tick(); locked {
		t.Fatalf("expected no lock on empty board")
	}
	if g.Active.Row != 1 {
		t.Fatalf("expected Row=1 after descent, got %d", g.Active.Row)
	}
	if g.Score != 0 {
		t.Fatalf("expected score 0, got %d", g.Score)
	}
}

func TestTick_LocksOnFloor(t *testing.T) {
	// Horizontal I at bbox Row=18 → cells at board row 19 (floor).
	// Next descent would push cells off the board → lock.
	g := Game{Active: Active{Piece: Tetromino{Shape: ShapeI}, Row: 18, Col: 3}}
	locked := g.Tick()
	if !locked {
		t.Fatalf("expected lock at floor")
	}
	for c := 3; c <= 6; c++ {
		if g.Board[19][c] != Cell(ShapeI+1) {
			t.Fatalf("expected I colour at row 19 col %d, got %d", c, g.Board[19][c])
		}
	}
	if g.Score != 0 {
		t.Fatalf("expected score 0 (no full row), got %d", g.Score)
	}
}

func TestTick_LocksOnStack(t *testing.T) {
	var g Game
	// Block the I's descent without filling the row — keep one cell empty so
	// clearLines is a no-op and the lock row stays put.
	g.Board[19][3] = Cell(ShapeO + 1)
	// I at Row=17 → cells at board row 18; descend would put cells at row 19 (occupied).
	g.Active = Active{Piece: Tetromino{Shape: ShapeI}, Row: 17, Col: 3}
	if locked := g.Tick(); !locked {
		t.Fatalf("expected lock on stack")
	}
	for c := 3; c <= 6; c++ {
		if g.Board[18][c] != Cell(ShapeI+1) {
			t.Fatalf("expected I colour at row 18 col %d, got %d", c, g.Board[18][c])
		}
	}
}

// TestTick_TetrisScores160 is the task 1.4 acceptance test: dropping a
// horizontal I onto a fully-occupied bottom slab scores 160 and leaves
// the I's lock row exposed as the new bottom row.
func TestTick_TetrisScores160(t *testing.T) {
	var g Game
	for r := 16; r <= 19; r++ {
		for c := 0; c < BoardCols; c++ {
			g.Board[r][c] = Cell(ShapeO + 1)
		}
	}
	// I bbox at Row=14 → cells at row 15. Descend would land at row 16 (full) → lock.
	g.Active = Active{Piece: Tetromino{Shape: ShapeI}, Row: 14, Col: 3}

	if locked := g.Tick(); !locked {
		t.Fatalf("expected lock against filled slab")
	}
	if g.Score != 160 {
		t.Fatalf("expected score 160 for tetris, got %d", g.Score)
	}
	// After 4 rows cleared, the I's row shifts to the new bottom.
	for c := 0; c < BoardCols; c++ {
		want := Cell(0)
		if c >= 3 && c <= 6 {
			want = Cell(ShapeI + 1)
		}
		if g.Board[19][c] != want {
			t.Fatalf("row 19 col %d: want %d, got %d", c, want, g.Board[19][c])
		}
	}
	// The row above (next-most-recent) should be empty — the slab is gone.
	for c := 0; c < BoardCols; c++ {
		if g.Board[18][c] != 0 {
			t.Fatalf("row 18 col %d: want empty, got %d", c, g.Board[18][c])
		}
	}
}

func TestClearLines_ScoresMatchTable(t *testing.T) {
	cases := []struct {
		rows  int
		score int
	}{
		{1, 10},
		{2, 40},
		{3, 90},
		{4, 160},
	}
	for _, tc := range cases {
		var g Game
		for r := BoardRows - tc.rows; r < BoardRows; r++ {
			for c := 0; c < BoardCols; c++ {
				g.Board[r][c] = Cell(ShapeO + 1)
			}
		}
		cleared := g.clearLines()
		if cleared != tc.rows {
			t.Fatalf("rows=%d: expected %d cleared, got %d", tc.rows, tc.rows, cleared)
		}
		g.Score += 10 * cleared * cleared
		if g.Score != tc.score {
			t.Fatalf("rows=%d: expected score %d, got %d", tc.rows, tc.score, g.Score)
		}
	}
}

func TestClearLines_ShiftsRowsAboveDown(t *testing.T) {
	var g Game
	g.Board[5][0] = Cell(ShapeT + 1) // sentinel two rows above the cleared row
	for c := 0; c < BoardCols; c++ {
		g.Board[19][c] = Cell(ShapeO + 1)
	}
	if cleared := g.clearLines(); cleared != 1 {
		t.Fatalf("expected 1 row cleared, got %d", cleared)
	}
	if g.Board[6][0] != Cell(ShapeT+1) {
		t.Fatalf("expected sentinel shifted to row 6, got %d", g.Board[6][0])
	}
	if g.Board[5][0] != 0 {
		t.Fatalf("expected row 5 cleared after shift, got %d", g.Board[5][0])
	}
	if g.Board[19][0] != 0 {
		t.Fatalf("expected bottom row empty after clear, got %d", g.Board[19][0])
	}
}

func TestClearLines_NoFullRowsLeavesBoardUntouched(t *testing.T) {
	var g Game
	g.Board[10][4] = Cell(ShapeZ + 1)
	g.Board[19][0] = Cell(ShapeO + 1) // row 19 has only one filled cell
	before := g.Board
	if cleared := g.clearLines(); cleared != 0 {
		t.Fatalf("expected 0 cleared, got %d", cleared)
	}
	if g.Board != before {
		t.Fatalf("expected board unchanged when no full rows")
	}
}

func TestSpawn_PlacesPieceOnEmptyBoard(t *testing.T) {
	var g Game
	if !g.Spawn(Tetromino{Shape: ShapeI}, 0, 3) {
		t.Fatalf("expected spawn to succeed on empty board")
	}
	if g.GameOver {
		t.Fatalf("expected GameOver=false after successful spawn")
	}
	if g.Active.Row != 0 || g.Active.Col != 3 {
		t.Fatalf("expected Active at (0,3), got (%d,%d)", g.Active.Row, g.Active.Col)
	}
	if g.Active.Piece.Shape != ShapeI {
		t.Fatalf("expected Active piece I, got %d", g.Active.Piece.Shape)
	}
}

// TestSpawn_GameOverWhenColumnZeroFull is the task 1.5 acceptance test:
// filling column 0 to the top and spawning an I in that column triggers
// game-over within one tick.
func TestSpawn_GameOverWhenColumnZeroFull(t *testing.T) {
	var g Game
	for r := 0; r < BoardRows; r++ {
		g.Board[r][0] = Cell(ShapeO + 1)
	}
	// Horizontal I rotation 0 fills bbox row 1 cols 0..3. Spawn at (0, 0)
	// puts a filled cell at board (1, 0), which is in the stacked column.
	if g.Spawn(Tetromino{Shape: ShapeI}, 0, 0) {
		t.Fatalf("expected spawn to fail when column 0 is full")
	}
	if !g.GameOver {
		t.Fatalf("expected GameOver flag set after blocked spawn")
	}
}

func TestSpawn_LeavesActiveUntouchedWhenBlocked(t *testing.T) {
	var g Game
	for r := 0; r < BoardRows; r++ {
		g.Board[r][0] = Cell(ShapeO + 1)
	}
	prev := Active{Piece: Tetromino{Shape: ShapeT}, Row: 5, Col: 4}
	g.Active = prev
	if g.Spawn(Tetromino{Shape: ShapeI}, 0, 0) {
		t.Fatalf("expected spawn to fail")
	}
	if g.Active != prev {
		t.Fatalf("expected Active unchanged after failed spawn, got %+v", g.Active)
	}
}

func TestTick_NoopAfterGameOver(t *testing.T) {
	var g Game
	g.GameOver = true
	g.Active = Active{Piece: Tetromino{Shape: ShapeI}, Row: 0, Col: 3}
	if locked := g.Tick(); locked {
		t.Fatalf("expected Tick to return false after game over")
	}
	if g.Active.Row != 0 {
		t.Fatalf("expected Active.Row unchanged, got %d", g.Active.Row)
	}
	if g.Score != 0 {
		t.Fatalf("expected Score unchanged, got %d", g.Score)
	}
}

func TestFits_RejectsOutOfBoundsAndOccupied(t *testing.T) {
	var g Game
	below := Active{Piece: Tetromino{Shape: ShapeI}, Row: 19, Col: 3} // cells at row 20
	if g.fits(below) {
		t.Fatalf("expected fits=false for piece below floor")
	}
	left := Active{Piece: Tetromino{Shape: ShapeI}, Row: 0, Col: -1} // cells at cols -1..2
	if g.fits(left) {
		t.Fatalf("expected fits=false for piece off the left edge")
	}
	g.Board[5][4] = Cell(ShapeO + 1)
	overlap := Active{Piece: Tetromino{Shape: ShapeI}, Row: 4, Col: 3} // cells at row 5 cols 3-6
	if g.fits(overlap) {
		t.Fatalf("expected fits=false when overlapping an occupied cell")
	}
	ok := Active{Piece: Tetromino{Shape: ShapeI}, Row: 4, Col: 3}
	g.Board[5][4] = 0
	if !g.fits(ok) {
		t.Fatalf("expected fits=true when path is clear")
	}
}

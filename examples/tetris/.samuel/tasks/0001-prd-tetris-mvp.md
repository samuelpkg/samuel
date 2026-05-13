# PRD 0001 — Tetris MVP

Implement a minimal text-mode Tetris in Go. This PRD is intentionally
small so it doubles as a smoke fixture for `samuel run init --prd` —
the inline `### N.M Title` headings under `## Tasks` exercise the
parser that landed in v2.0.0-rc.12.

## Goals

1. Playable tetromino loop on a fixed 10×20 grid.
2. Score line clears.
3. No external runtime dependencies; pure stdlib.

## Tasks

### 1.1 Define the board model

Model the playing field as a `[20][10]Cell` where `Cell` is `byte`
(0 = empty, 1-7 = tetromino color id). Add `func (b *Board) String() string`
that renders the board to ANSI for stdout.

**Acceptance**: a freshly constructed board prints 20 rows of 10 dots.

### 1.2 Implement the seven tetrominoes

Encode I, O, T, S, Z, L, J as 4×4 bitmasks with 4 rotation states each.
Centralize rotation in a single `Rotate(t Tetromino) Tetromino` so the
rest of the loop never indexes the rotation tables.

**Acceptance**: rotating a piece four times returns the original
bitmask byte-for-byte.

### 1.3 Wire keyboard input

Read raw key events from stdin (left/right/down/space for hard-drop,
'r' to rotate, 'q' to quit). Use `golang.org/x/term` only if it's
already in `go.sum` — otherwise spin a tiny `termios` wrapper inline.

**Acceptance**: pressing the same key twice in the test harness emits
two input events.

### 1.4 Implement gravity and line clearing

Tick every 500ms. On each tick, descend the active piece by one row.
On lock, sweep complete rows, shift everything above down, and add
`10 × cleared²` to the score (single = 10, double = 40, triple = 90,
tetris = 160).

**Acceptance**: dropping a horizontal I onto a fully-occupied bottom
row scores 160 and leaves the next-most-recent row exposed.

### 1.5 Game-over detection

When the spawn position of a new piece is blocked, stop the tick
loop, print the final score, and exit cleanly.

**Acceptance**: filling column 0 to the top and spawning an I in that
column triggers game-over within one tick.

## Out of Scope

- Hold-piece slot.
- Ghost piece.
- T-spin scoring.
- Wall kicks (use simple-rotation only).

These are reasonable v2 features but would inflate this PRD past the
"smoke fixture" budget.

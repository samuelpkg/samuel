package tetris

import (
	"bufio"
	"errors"
	"io"
)

// Event is one recognised player input event.
type Event byte

// Recognised input events. EventNone is returned alongside an error
// when the input stream did not yield a meaningful event.
const (
	EventNone Event = iota
	EventLeft
	EventRight
	EventDown
	EventHardDrop
	EventRotate
	EventQuit
)

// ErrUnknownKey signals that the next byte (or escape sequence) did not
// map to any game event. The caller may keep reading to consume the
// following event without aborting the input loop.
var ErrUnknownKey = errors.New("tetris: unknown key")

// Reader decodes a stream of raw keyboard bytes into game events. The
// underlying stream must deliver bytes as the terminal produces them —
// call EnableRawMode on the controlling tty before reading os.Stdin.
type Reader struct {
	br *bufio.Reader
}

// NewReader wraps r so successive Read calls return one Event each.
func NewReader(r io.Reader) *Reader {
	return &Reader{br: bufio.NewReader(r)}
}

// Read returns the next recognised Event. It returns io.EOF when the
// stream is exhausted, or ErrUnknownKey when an unrecognised byte or
// partial escape sequence is encountered.
func (r *Reader) Read() (Event, error) {
	b, err := r.br.ReadByte()
	if err != nil {
		return EventNone, err
	}
	if ev, ok := singleByteEvent(b); ok {
		return ev, nil
	}
	if b == 0x1b {
		return r.readEscape()
	}
	return EventNone, ErrUnknownKey
}

// singleByteEvent maps printable / control keys to their event.
func singleByteEvent(b byte) (Event, bool) {
	switch b {
	case ' ':
		return EventHardDrop, true
	case 'r', 'R':
		return EventRotate, true
	case 'q', 'Q':
		return EventQuit, true
	}
	return EventNone, false
}

// readEscape decodes the rest of a CSI sequence (ESC [ X) into a
// direction event. Anything else is reported as ErrUnknownKey.
func (r *Reader) readEscape() (Event, error) {
	b, err := r.br.ReadByte()
	if err != nil {
		return EventNone, err
	}
	if b != '[' {
		return EventNone, ErrUnknownKey
	}
	b, err = r.br.ReadByte()
	if err != nil {
		return EventNone, err
	}
	switch b {
	case 'B':
		return EventDown, nil
	case 'C':
		return EventRight, nil
	case 'D':
		return EventLeft, nil
	}
	return EventNone, ErrUnknownKey
}

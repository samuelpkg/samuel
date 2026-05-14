package tetris

import (
	"errors"
	"io"
	"strings"
	"testing"
)

// TestReader_SameKeyTwiceEmitsTwoEvents is the PRD's acceptance criterion
// for task 1.3: pressing the same key twice yields two distinct events.
func TestReader_SameKeyTwiceEmitsTwoEvents(t *testing.T) {
	r := NewReader(strings.NewReader("qq"))
	for i := 0; i < 2; i++ {
		ev, err := r.Read()
		if err != nil {
			t.Fatalf("read %d: unexpected error: %v", i, err)
		}
		if ev != EventQuit {
			t.Errorf("read %d: got %v, want EventQuit", i, ev)
		}
	}
	if _, err := r.Read(); !errors.Is(err, io.EOF) {
		t.Fatalf("third read: want EOF, got %v", err)
	}
}

func TestReader_RotateAcceptsLowerAndUpper(t *testing.T) {
	r := NewReader(strings.NewReader("rR"))
	for i := 0; i < 2; i++ {
		ev, err := r.Read()
		if err != nil {
			t.Fatalf("read %d: unexpected error: %v", i, err)
		}
		if ev != EventRotate {
			t.Errorf("read %d: got %v, want EventRotate", i, ev)
		}
	}
}

func TestReader_SpaceIsHardDrop(t *testing.T) {
	r := NewReader(strings.NewReader("  "))
	for i := 0; i < 2; i++ {
		ev, err := r.Read()
		if err != nil {
			t.Fatalf("read %d: unexpected error: %v", i, err)
		}
		if ev != EventHardDrop {
			t.Errorf("read %d: got %v, want EventHardDrop", i, ev)
		}
	}
}

func TestReader_ArrowKeys(t *testing.T) {
	cases := []struct {
		input string
		want  Event
	}{
		{"\x1b[D", EventLeft},
		{"\x1b[C", EventRight},
		{"\x1b[B", EventDown},
	}
	for _, tc := range cases {
		r := NewReader(strings.NewReader(tc.input))
		ev, err := r.Read()
		if err != nil {
			t.Fatalf("input %q: unexpected error: %v", tc.input, err)
		}
		if ev != tc.want {
			t.Errorf("input %q: got %v, want %v", tc.input, ev, tc.want)
		}
	}
}

func TestReader_ConsecutiveArrowsBothDecode(t *testing.T) {
	r := NewReader(strings.NewReader("\x1b[D\x1b[D"))
	for i := 0; i < 2; i++ {
		ev, err := r.Read()
		if err != nil {
			t.Fatalf("read %d: unexpected error: %v", i, err)
		}
		if ev != EventLeft {
			t.Errorf("read %d: got %v, want EventLeft", i, ev)
		}
	}
}

func TestReader_MixedSequenceDecodesEachEvent(t *testing.T) {
	r := NewReader(strings.NewReader("\x1b[D r q"))
	want := []Event{EventLeft, EventHardDrop, EventRotate, EventHardDrop, EventQuit}
	for i, w := range want {
		ev, err := r.Read()
		if err != nil {
			t.Fatalf("read %d: unexpected error: %v", i, err)
		}
		if ev != w {
			t.Errorf("read %d: got %v, want %v", i, ev, w)
		}
	}
}

func TestReader_UnknownKeyReturnsErrUnknownKey(t *testing.T) {
	r := NewReader(strings.NewReader("x"))
	ev, err := r.Read()
	if !errors.Is(err, ErrUnknownKey) {
		t.Fatalf("want ErrUnknownKey, got err=%v", err)
	}
	if ev != EventNone {
		t.Errorf("got %v, want EventNone", ev)
	}
}

func TestReader_UnknownEscapeReturnsErrUnknownKey(t *testing.T) {
	// ESC [ A is "up arrow" — not bound to any game event.
	r := NewReader(strings.NewReader("\x1b[A"))
	ev, err := r.Read()
	if !errors.Is(err, ErrUnknownKey) {
		t.Fatalf("want ErrUnknownKey, got err=%v", err)
	}
	if ev != EventNone {
		t.Errorf("got %v, want EventNone", ev)
	}
}

func TestReader_EmptyStreamReturnsEOF(t *testing.T) {
	r := NewReader(strings.NewReader(""))
	_, err := r.Read()
	if !errors.Is(err, io.EOF) {
		t.Fatalf("want io.EOF, got %v", err)
	}
}

func TestReader_PartialEscapeReturnsEOF(t *testing.T) {
	// Bare ESC with no trailing CSI byte — stream ends mid-sequence.
	r := NewReader(strings.NewReader("\x1b"))
	_, err := r.Read()
	if !errors.Is(err, io.EOF) {
		t.Fatalf("want io.EOF, got %v", err)
	}
}

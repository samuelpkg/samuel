//go:build darwin || linux

package tetris

import (
	"fmt"
	"os"
	"os/exec"
)

// EnableRawMode puts the controlling terminal into cbreak mode so input
// is delivered byte-at-a-time with no echo. The returned restore func
// reverts the terminal to its default cooked mode; callers should defer
// it before reading from os.Stdin.
//
// This is a tiny inline termios wrapper (via stty(1)) — no external Go
// module is pulled in, matching the project's zero-dependency posture.
func EnableRawMode() (restore func() error, err error) {
	if err := runStty("cbreak", "-echo"); err != nil {
		return nil, fmt.Errorf("tetris: enable raw mode: %w", err)
	}
	return func() error {
		if err := runStty("sane"); err != nil {
			return fmt.Errorf("tetris: restore terminal: %w", err)
		}
		return nil
	}, nil
}

// runStty shells out to stty(1) with the given args. Inheriting os.Stdin
// is required so stty acts on the controlling tty.
func runStty(args ...string) error {
	cmd := exec.Command("stty", args...)
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

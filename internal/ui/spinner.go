package ui

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

// Spinner wraps bubbles/spinner for non-bubbletea callers. It runs
// its own tick goroutine and erases itself on Stop. Safe to Stop
// multiple times.
//
// We avoid spinning up bubbletea for the common case of a one-off
// "downloading…" indicator in a non-interactive command; PRD 0006's
// Charm UI pass will adopt full bubbletea where it makes sense.
type Spinner struct {
	model   spinner.Model
	message string
	out     io.Writer
	done    chan struct{}
	once    sync.Once
	wg      sync.WaitGroup
}

// NewSpinner returns a spinner with the given message.
func NewSpinner(message string) *Spinner {
	m := spinner.New()
	m.Spinner = spinner.Dot
	m.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#58A6FF"))
	return &Spinner{
		model:   m,
		message: message,
		out:     os.Stdout,
		done:    make(chan struct{}),
	}
}

// Start begins the animation. It does not block.
func (s *Spinner) Start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(120 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-s.done:
				// Erase the last line.
				fmt.Fprint(s.out, "\r\033[K")
				return
			case <-ticker.C:
				s.model, _ = s.model.Update(s.model.Tick())
				fmt.Fprintf(s.out, "\r%s %s", s.model.View(), s.message)
			}
		}
	}()
}

// Stop halts the animation. Safe to call multiple times.
func (s *Spinner) Stop() {
	s.once.Do(func() {
		close(s.done)
		s.wg.Wait()
	})
}

// Success stops and prints a success line.
func (s *Spinner) Success(msg string) {
	s.Stop()
	Success("%s", msg)
}

// Error stops and prints an error line.
func (s *Spinner) Error(msg string) {
	s.Stop()
	Error("%s", msg)
}

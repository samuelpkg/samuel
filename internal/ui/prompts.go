package ui

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/huh"
	"golang.org/x/term"
)

// IsTerminal reports whether stdin is attached to an interactive
// terminal. Prompts fall back to a non-interactive scan path when
// false (e.g. CI, piped input).
func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// Confirm runs a yes/no prompt with a default. When stdin is not a
// terminal it falls back to printing the question and reading a line,
// matching the previous inline-scanln behavior so scripted callers
// keep working.
func Confirm(title, description string, defaultYes bool) (bool, error) {
	if !IsTerminal() {
		return scanConfirm(title, defaultYes), nil
	}
	answer := defaultYes
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Description(description).
				Affirmative("Yes").
				Negative("No").
				Value(&answer),
		),
	)
	if err := form.Run(); err != nil {
		return false, err
	}
	return answer, nil
}

// Select shows a single-choice list. Returns the chosen value or the
// zero string if cancelled.
func Select(title, description string, options []string) (string, error) {
	if len(options) == 0 {
		return "", fmt.Errorf("ui.Select: no options")
	}
	if !IsTerminal() {
		return options[0], nil
	}
	var chosen string
	opts := make([]huh.Option[string], 0, len(options))
	for _, o := range options {
		opts = append(opts, huh.NewOption(o, o))
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Description(description).
				Options(opts...).
				Value(&chosen),
		),
	)
	if err := form.Run(); err != nil {
		return "", err
	}
	return chosen, nil
}

// MultiSelect shows a native multi-select list. Returns the slice of
// selected values. v1 hand-rolled this with promptui; huh ships it
// natively (RFD/PRD 0006 §1).
func MultiSelect(title, description string, options []string, defaults []string) ([]string, error) {
	if len(options) == 0 {
		return nil, nil
	}
	if !IsTerminal() {
		return defaults, nil
	}
	chosen := append([]string(nil), defaults...)
	opts := make([]huh.Option[string], 0, len(options))
	for _, o := range options {
		opts = append(opts, huh.NewOption(o, o))
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(title).
				Description(description).
				Options(opts...).
				Value(&chosen),
		),
	)
	if err := form.Run(); err != nil {
		return nil, err
	}
	return chosen, nil
}

// promptInput is overridable by tests.
var promptInput io.Reader = os.Stdin

func scanConfirm(title string, defaultYes bool) bool {
	def := "y/N"
	if defaultYes {
		def = "Y/n"
	}
	fmt.Fprintf(stdout, "%s [%s] ", title, def)
	var ans string
	_, _ = fmt.Fscanln(promptInput, &ans)
	switch ans {
	case "":
		return defaultYes
	case "y", "Y", "yes", "YES", "Yes":
		return true
	default:
		return false
	}
}

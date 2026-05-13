// Package ui is Samuel v2's human-output layer: lipgloss colour tokens
// plus the six-category vocabulary v1 settled on (success / error /
// warn / info / bold / dim) and a handful of list-and-table helpers.
//
// Output rules (preserved from v1):
//   - Errors go to stderr; everything else goes to stdout.
//   - Color is on by default. Pipes, NO_COLOR, and the --no-color
//     persistent flag disable it.
//   - Output adapts to terminal-light/dark via lipgloss adaptive colour.
//
// The JSON envelope lives in json.go; subcommands branch on
// commands.JSONMode(cmd) before calling any of the helpers here.
package ui

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
)

// Symbols used by the small built-in glyph set. They render fine in
// every modern terminal and degrade gracefully when colour is off.
const (
	SuccessSymbol = "✓"
	ErrorSymbol   = "✗"
	WarnSymbol    = "⚠"
	InfoSymbol    = "→"
	PendingSymbol = "○"
	ActiveSymbol  = "●"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#1A7F37", Dark: "#3FB950"})
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#A40E26", Dark: "#FF6A69"})
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#9A6700", Dark: "#D29922"})
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0969DA", Dark: "#58A6FF"})
	boldStyle    = lipgloss.NewStyle().Bold(true)
	dimStyle     = lipgloss.NewStyle().Faint(true)
	headerStyle  = lipgloss.NewStyle().Bold(true).MarginTop(1).MarginBottom(1)

	// Default sinks; tests can override via SetWriters.
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
)

// SetWriters overrides the package-level stdout/stderr sinks. Used by
// tests to capture output. Pass nil to keep a stream untouched.
func SetWriters(out, errw io.Writer) {
	if out != nil {
		stdout = out
	}
	if errw != nil {
		stderr = errw
	}
}

// DisableColors turns off colored output globally. The --no-color
// flag, NO_COLOR env var, and non-tty stdout all drive this on.
func DisableColors() { lipgloss.SetColorProfile(0) }

// Success prints a green success line to stdout.
func Success(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(stdout, successStyle.Render(SuccessSymbol+" "+msg))
}

// Error prints a red error line to stderr.
func Error(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(stderr, errorStyle.Render(ErrorSymbol+" "+msg))
}

// Warn prints a yellow warning line to stdout.
func Warn(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(stdout, warnStyle.Render(WarnSymbol+" "+msg))
}

// Info prints a cyan info line to stdout.
func Info(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(stdout, infoStyle.Render(InfoSymbol+" "+msg))
}

// Print writes a plain line to stdout without styling.
func Print(format string, args ...any) {
	fmt.Fprintf(stdout, format+"\n", args...)
}

// Bold prints a bold line to stdout.
func Bold(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(stdout, boldStyle.Render(msg))
}

// Dim prints a faint line to stdout.
func Dim(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(stdout, dimStyle.Render(msg))
}

// Header prints a top-level section title with surrounding blank
// lines.
func Header(title string) { fmt.Fprintln(stdout, headerStyle.Render(title)) }

// Section prints a subsection header (boldface, no margin).
func Section(title string) { fmt.Fprintf(stdout, "\n%s:\n", boldStyle.Render(title)) }

func indent(n int) string {
	if n <= 0 {
		return ""
	}
	out := make([]byte, n*2)
	for i := range out {
		out[i] = ' '
	}
	return string(out)
}

// ListItem prints a plain list item.
func ListItem(level int, format string, args ...any) {
	fmt.Fprintf(stdout, "%s%s\n", indent(level), fmt.Sprintf(format, args...))
}

// SuccessItem prints a green-checked list item.
func SuccessItem(level int, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(stdout, "%s%s\n", indent(level), successStyle.Render(SuccessSymbol+" "+msg))
}

// WarnItem prints a yellow warning list item.
func WarnItem(level int, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(stdout, "%s%s\n", indent(level), warnStyle.Render(WarnSymbol+" "+msg))
}

// ErrorItem prints a red error list item to stdout. (Item helpers
// stay on stdout so they group with their parent section; standalone
// Error() goes to stderr.)
func ErrorItem(level int, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(stdout, "%s%s\n", indent(level), errorStyle.Render(ErrorSymbol+" "+msg))
}

// TableRow prints a key/value row used by version-style output.
func TableRow(key, value string) {
	fmt.Fprintf(stdout, "  %-20s %s\n", key+":", value)
}

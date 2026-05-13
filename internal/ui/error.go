package ui

import (
	stderrors "errors"
	"fmt"

	"github.com/ar4mirez/samuel/internal/errors"
)

// RenderError emits the multi-line CLI rendering of a structured
// *errors.Error to stderr. Non-structured errors fall back to a
// single-line "Error: ..." print so callers can wrap every CLI error
// in this one helper.
func RenderError(err error) {
	if err == nil {
		return
	}
	var oe *errors.Error
	if !stderrors.As(err, &oe) {
		Error("%s", err.Error())
		return
	}
	// Multi-line render. We bypass the helper styling for the lines
	// after the first so the rendering matches v1 verbatim and stays
	// stable for shell-grep workflows.
	fmt.Fprintln(stderr, errorStyle.Render("✗ "+oe.Problem))
	if oe.Cause != "" {
		fmt.Fprintf(stderr, "  %s %s\n", dimStyle.Render("Cause:"), oe.Cause)
	}
	if oe.Fix != "" {
		fmt.Fprintf(stderr, "  %s   %s\n", dimStyle.Render("Fix:"), oe.Fix)
	}
	if oe.DocsURL != "" {
		fmt.Fprintf(stderr, "  %s  %s\n", dimStyle.Render("Docs:"), oe.DocsURL)
	}
	if oe.Path != "" {
		fmt.Fprintf(stderr, "  %s  %s\n", dimStyle.Render("Path:"), oe.Path)
	}
}

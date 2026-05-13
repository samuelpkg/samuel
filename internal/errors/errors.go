// Package errors defines Samuel's structured error type used by every
// subsystem instead of bare errors.New / fmt.Errorf strings. It carries
// enough context for the CLI to render actionable feedback to users and
// for `samuel doctor` to translate failures into fix hints.
//
// Ported verbatim from samuel_v1/internal/orchestrator/errors.go with
// the package renamed and docs URL retargeted to the v2 docs site. The
// shape (six fields + a wrapped chain) is unchanged so any v1 component
// that ports forward keeps working.
//
// The CLI renders these across multiple lines in interactive mode:
//
//	Error: cannot register MCP server
//	  Cause: claude not on PATH
//	  Fix:   install claude or add it to PATH
//	  Docs:  https://ar4mirez.github.io/samuel/docs/errors/SAM-MCP-001
//
// Error() returns a single-line string suitable for logs and tests.
//
// # Error code namespace
//
// Codes follow SAM-<area>-<NNN>. Reserved subsystem ranges (extend here
// as new subsystems land):
//
//	SAM-LOCK-001 … SAM-LOCK-099 — flock acquisition + lock file
//	SAM-CFG-001  … SAM-CFG-099  — samuel.toml / samuel.lock parsing
//	SAM-TOON-001 … SAM-TOON-099 — TOON encoding/decoding
//	SAM-PLUG-001 … SAM-PLUG-099 — plugin manifest + lifecycle
//	SAM-MCP-001  … SAM-MCP-099  — MCP server registration
//	SAM-CLI-001  … SAM-CLI-099  — generic CLI / flag failures
package errors

import (
	"errors"
	"fmt"
)

// Error is the structured error type. Every subsystem returns *Error
// (or wraps one) so the CLI can render multi-line guidance.
type Error struct {
	// Component identifies which subsystem produced the error (e.g.
	// "lock", "config", "plugin", "samuel-skills").
	Component string
	// Problem describes what failed in one short sentence.
	Problem string
	// Cause is the underlying root cause, often a wrapped error string.
	Cause string
	// Fix is the recommended remediation. Should be copy-paste-able
	// when possible.
	Fix string
	// DocsURL points to a documentation page covering this error class.
	// Optional but encouraged.
	DocsURL string
	// Recoverable signals whether the user can fix this themselves
	// (true) vs. needing to file a bug (false).
	Recoverable bool
	// Path is the filesystem path involved, when relevant.
	Path string
	// wrapped preserves the original error chain for errors.Is /
	// errors.As traversal.
	wrapped error
}

// Error formats the structured fields into a single line suitable for
// logging. Multi-line rendering for the CLI is handled by internal/ui.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Component, e.Problem, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Component, e.Problem)
}

// Unwrap returns the wrapped error so errors.Is and errors.As work
// across the *Error boundary. Safe to call on a nil receiver.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.wrapped
}

// Wrap returns a copy of e with err preserved as the underlying cause.
// If Cause is empty it is populated from err.Error(). The receiver is
// not mutated, so a single *Error template can be Wrap'd repeatedly.
func (e *Error) Wrap(err error) *Error {
	if e == nil {
		return nil
	}
	cp := *e
	cp.wrapped = err
	if cp.Cause == "" && err != nil {
		cp.Cause = err.Error()
	}
	return &cp
}

// IsRecoverable reports whether err carries Recoverable=true, treating
// non-*Error values as non-recoverable. It traverses joined and wrapped
// errors via errors.As.
func IsRecoverable(err error) bool {
	var oe *Error
	if errors.As(err, &oe) {
		return oe.Recoverable
	}
	return false
}

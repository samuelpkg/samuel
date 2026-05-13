package commands

import (
	stderrors "errors"
	"os"

	"github.com/samuelpkg/samuel/internal/builtins"
	"github.com/samuelpkg/samuel/internal/components/samuel"
	"github.com/samuelpkg/samuel/internal/errors"
	"github.com/samuelpkg/samuel/internal/orchestrator"
	"github.com/samuelpkg/samuel/internal/plugin"
	"github.com/samuelpkg/samuel/internal/ui"
)

// buildOrchestrator constructs the v2 orchestrator with the framework's
// concrete plugins. As of PRD 0002 the only plugin is the embedded
// samuel-builtins component; PRD 0003 will append translator and skill
// plugins discovered from samuel.toml.
//
// homeDir is the override for the advisory lock + the install target.
// Pass "" to use os.UserHomeDir() (production). version is what
// SamuelComponent reports via Detect — the binary version is the
// builtins version since the content is embedded.
func buildOrchestrator(homeDir, version string) *orchestrator.Orchestrator {
	sc := samuel.New(builtins.FS(), homeDir, version)
	o := orchestrator.New(sc)
	if homeDir != "" {
		o.WithHomeDir(homeDir)
	}
	return o
}

// installOptionsFromCmd translates the persistent --verbose flag and
// command-local flags into plugin.InstallOptions. Centralized so init
// and any future install command stay consistent.
func installOptionsFromCmd(dryRun, force, verbose bool) plugin.InstallOptions {
	return plugin.InstallOptions{
		DryRun:  dryRun,
		Force:   force,
		Verbose: verbose,
		Stdout:  os.Stdout,
	}
}

// renderStructuredError prints a structured *errors.Error in the
// shape promised by the v2 error standard:
//
//	Error: <Problem>
//	  Cause: <Cause>
//	  Fix:   <Fix>
//	  Docs:  <DocsURL>
//
// Falls back to a plain Error: line when err is not an *errors.Error.
// Returns the input err so callers can `return renderStructuredError(err)`.
func renderStructuredError(err error) error {
	if err == nil {
		return nil
	}
	var oe *errors.Error
	if !stderrors.As(err, &oe) {
		ui.Error("%v", err)
		return err
	}
	ui.Error("%s", oe.Problem)
	if oe.Cause != "" {
		ui.ListItem(1, "Cause: %s", oe.Cause)
	}
	if oe.Fix != "" {
		ui.ListItem(1, "Fix:   %s", oe.Fix)
	}
	if oe.DocsURL != "" {
		ui.ListItem(1, "Docs:  %s", oe.DocsURL)
	}
	if oe.Path != "" {
		ui.ListItem(1, "Path:  %s", oe.Path)
	}
	return err
}

// renderInstallResults prints a one-line per-plugin summary of what the
// orchestrator did. Used by `samuel init` after Install returns.
func renderInstallResults(results []plugin.InstallResult) {
	for _, r := range results {
		switch {
		case r.Skipped:
			ui.Dim("  - %s: skipped", r.Component)
		case r.AlreadyInstalled:
			ui.Dim("  ✓ %s: already installed", r.Component)
		case len(r.Mutations) > 0:
			ui.SuccessItem(1, "%s: installed (%d changes)", r.Component, len(r.Mutations))
		default:
			ui.SuccessItem(1, "%s: ok", r.Component)
		}
	}
}

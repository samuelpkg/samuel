package commands

import (
	stderrors "errors"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/samuelpkg/samuel/internal/config"
	"github.com/samuelpkg/samuel/internal/sync"
	"github.com/samuelpkg/samuel/internal/translator/claude"
	"github.com/samuelpkg/samuel/internal/ui"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Regenerate per-folder AGENTS.md",
	Long: `Walk the project tree and refresh per-folder AGENTS.md files.

Files without the autogen marker are treated as user-customized and
skipped unless --force. Use --dry-run to preview without writing.

Examples:
  samuel sync                  # update auto-generated files only
  samuel sync --force          # overwrite user-customized files
  samuel sync --dry-run        # preview without writes
  samuel sync --max-depth 3    # cap walk depth
  samuel sync --json           # machine-readable envelope`,
	RunE: runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().Bool("dry-run", false, "Preview changes without writing")
	syncCmd.Flags().Bool("force", false, "Overwrite user-customized files")
	syncCmd.Flags().Int("max-depth", -1, "Maximum walk depth (-1 = unlimited)")
}

func runSync(cmd *cobra.Command, _ []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	force, _ := cmd.Flags().GetBool("force")
	maxDepth, _ := cmd.Flags().GetInt("max-depth")

	cwd, err := os.Getwd()
	if err != nil {
		return renderStructuredError(err)
	}

	// Smart bare invocation: when run without a samuel.toml, preview
	// the walk so the user sees what would happen even before init.
	initialized := true
	if _, err := config.Load(cwd); err != nil {
		if stderrors.Is(err, config.ErrNotFound) {
			initialized = false
			dryRun = true
			ui.Warn("Not initialized. Showing preview only. Run `samuel init` first.")
		}
	}

	res, err := sync.SyncFolderContext(sync.Options{
		RootDir:   cwd,
		MaxDepth:  maxDepth,
		Force:     force,
		DryRun:    dryRun,
		Overrides: overridesFromConfig(cwd),
	})
	if err != nil {
		return renderStructuredError(err)
	}

	mirror := runClaudeMirror(cwd, maxDepth, force, dryRun)

	if JSONMode(cmd) {
		ui.PrintJSON(commandPath(cmd), map[string]any{
			"initialized": initialized,
			"created":     res.Created,
			"updated":     res.Updated,
			"skipped":     res.Skipped,
			"errors":      stringifyErrors(res.Errors),
			"counts": map[string]int{
				"created": len(res.Created),
				"updated": len(res.Updated),
				"skipped": len(res.Skipped),
				"errors":  len(res.Errors),
			},
			"claude_mirror": claudeMirrorJSON(mirror),
		})
		return nil
	}

	renderSyncHuman(cwd, res, dryRun)
	renderClaudeMirrorHuman(cwd, mirror)
	return nil
}

// runClaudeMirror invokes the built-in AGENTS.md → CLAUDE.md translator
// when enabled in the project config. Returns nil when disabled or when
// the project has no samuel.toml yet — sync's smart-bare-invocation
// path needs to stay side-effect-free.
func runClaudeMirror(root string, maxDepth int, force, dryRun bool) *claude.Result {
	cfg, err := config.Load(root)
	if err != nil {
		return nil
	}
	if cfg.Translators == nil || cfg.Translators.Claude == nil || !cfg.Translators.Claude.Enabled {
		return nil
	}
	res, err := claude.Mirror(claude.Options{
		RootDir:       root,
		MaxDepth:      maxDepth,
		Force:         force,
		DryRun:        dryRun,
		SkipOverrides: overridesFromConfig(root),
	})
	if err != nil {
		return &claude.Result{Errors: []error{err}}
	}
	return res
}

func claudeMirrorJSON(r *claude.Result) map[string]any {
	if r == nil {
		return nil
	}
	return map[string]any{
		"created": r.Created,
		"updated": r.Updated,
		"skipped": r.Skipped,
		"errors":  stringifyErrors(r.Errors),
		"counts": map[string]int{
			"created": len(r.Created),
			"updated": len(r.Updated),
			"skipped": len(r.Skipped),
			"errors":  len(r.Errors),
		},
	}
}

func renderClaudeMirrorHuman(root string, r *claude.Result) {
	if r == nil {
		return
	}
	if len(r.Created)+len(r.Updated)+len(r.Skipped)+len(r.Errors) == 0 {
		return
	}
	ui.Print("")
	ui.Bold("Claude translator")
	for _, p := range r.Created {
		rel, _ := filepath.Rel(root, p)
		ui.SuccessItem(1, "create %s", rel)
	}
	for _, p := range r.Updated {
		rel, _ := filepath.Rel(root, p)
		ui.SuccessItem(1, "update %s", rel)
	}
	for _, p := range r.Skipped {
		rel, _ := filepath.Rel(root, p)
		ui.Dim("  - skip %s (user-customized)", rel)
	}
	for _, e := range r.Errors {
		ui.ErrorItem(1, "%v", e)
	}
}

// overridesFromConfig loads samuel.toml from cwd and pulls the [sync.*]
// blocks (when they land). Today the sections aren't represented in
// config.Config — this returns an empty Overrides — but the call site
// is wired so the future migration is mechanical.
func overridesFromConfig(cwd string) sync.Overrides {
	_, _ = config.Load(cwd)
	return sync.Overrides{}
}

func stringifyErrors(errs []error) []string {
	if len(errs) == 0 {
		return nil
	}
	out := make([]string, 0, len(errs))
	for _, e := range errs {
		out = append(out, e.Error())
	}
	return out
}

func renderSyncHuman(root string, res *sync.Result, dryRun bool) {
	if dryRun {
		ui.Bold("Samuel sync (dry-run)")
	} else {
		ui.Bold("Samuel sync")
	}
	for _, p := range res.Created {
		rel, _ := filepath.Rel(root, p)
		ui.SuccessItem(1, "create %s", rel)
	}
	for _, p := range res.Updated {
		rel, _ := filepath.Rel(root, p)
		ui.SuccessItem(1, "update %s", rel)
	}
	for _, p := range res.Skipped {
		rel, _ := filepath.Rel(root, p)
		ui.Dim("  - skip %s (user-customized)", rel)
	}
	for _, e := range res.Errors {
		ui.ErrorItem(1, "%v", e)
	}
	ui.Print("")
	ui.Bold("Summary: %d created, %d updated, %d skipped, %d errors",
		len(res.Created), len(res.Updated), len(res.Skipped), len(res.Errors))
}

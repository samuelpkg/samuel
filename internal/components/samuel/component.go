// Package samuel implements Samuel's first concrete plugin: the
// "samuel-builtins" component. It syncs the framework's embedded
// skills (ralph, create-skill, sync, generate-agents-md) from an
// embed.FS into ~/.samuel/builtins/.
//
// Differences from v1's SamuelComponent:
//
//   - Target moved to ~/.samuel/builtins/ per RFD 0009 (the framework
//     is agent-agnostic; no writes to legacy assistant home paths).
//   - No project-level symlink: v2's .samuel/builtins is populated by
//     the init command writing a local copy (per the PRD's "Recommend
//     copy" decision in §10).
//   - Returns plugin.Mutation values (with kinds renamed to the v2
//     enum) instead of v1's orchestrator.Mutation.
//
// Atomicity, idempotency (content-hash), and the path-traversal defense
// in syncFS are ported verbatim from v1 — those properties were earned
// the hard way and reverting them would re-introduce known bugs.
package samuel

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/samuelpkg/samuel/internal/errors"
	"github.com/samuelpkg/samuel/internal/plugin"
)

// Name is the stable identifier this component reports via Plugin.Name.
const Name = "samuel-builtins"

// globalDir is the directory under $HOME that hosts the framework's
// built-in skill tree. v2 lives under .samuel/ exclusively — no writes
// land in legacy assistant home paths (RFD 0009).
const globalDir = ".samuel/builtins"

// Component is the concrete plugin.Plugin that syncs Samuel's embedded
// built-in skills into the user's home directory.
type Component struct {
	// Source is the read-only filesystem of skill content. In
	// production this is builtins.FS(); tests inject fstest.MapFS or
	// os.DirFS for hermetic, fast assertions.
	Source fs.FS

	// homeDir overrides the user's home for testing. Empty falls
	// through to os.UserHomeDir().
	homeDir string

	// version is what Detect reports as the installed version. In
	// practice this is the Samuel binary version — skills are
	// embedded, so the binary IS the source of truth.
	version string
}

// New constructs a samuel-builtins component.
//
//   - source: fs.FS rooted at the skills tree (top-level dirs = skill names)
//   - homeDir: empty for production; tests pass a temp dir
//   - version: stable identifier reported by Detect (binary version)
func New(source fs.FS, homeDir, version string) *Component {
	return &Component{Source: source, homeDir: homeDir, version: version}
}

// Name returns the stable component identifier.
func (c *Component) Name() string { return Name }

// Manifest returns a synthetic manifest for the builtin tier.
// `kind = "builtin"` distinguishes this from skill/wasm/oci plugins.
func (c *Component) Manifest() plugin.Manifest {
	return plugin.Manifest{
		Name:    Name,
		Version: c.version,
		Kind:    plugin.KindBuiltin,
		Summary: "Samuel's embedded built-in skill bundle (ralph, create-skill, sync, generate-agents-md)",
		Source:  "embedded",
	}
}

func (c *Component) home() (string, error) {
	if c.homeDir != "" {
		return c.homeDir, nil
	}
	return os.UserHomeDir()
}

// GlobalPath returns the absolute path to ~/.samuel/builtins/.
func (c *Component) GlobalPath() (string, error) {
	h, err := c.home()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, globalDir), nil
}

// Detect reports whether the global skill tree exists and is non-empty.
// An empty directory does NOT count as installed — that state happens
// when a previous Install crashed mid-sync and left a husk behind.
func (c *Component) Detect(ctx context.Context) (plugin.DetectResult, error) {
	path, err := c.GlobalPath()
	if err != nil {
		return plugin.DetectResult{}, err
	}
	info, statErr := os.Stat(path)
	if statErr != nil || !info.IsDir() {
		return plugin.DetectResult{Installed: false, Path: path}, nil
	}
	entries, readErr := os.ReadDir(path)
	if readErr != nil || len(entries) == 0 {
		return plugin.DetectResult{Installed: false, Path: path}, nil
	}
	return plugin.DetectResult{
		Installed: true,
		Version:   c.version,
		Path:      path,
	}, nil
}

// Install syncs Source into ~/.samuel/builtins/. Idempotent: if the
// on-disk content hash matches Source's, the call is a no-op.
//
// Atomicity: write to a sibling tmp dir, then rename onto the target.
// On failure, the tmp dir is removed without touching the live tree.
// If a previous tree exists, it is renamed aside as a `.bak-<hash>`
// backup that is removed only after the successful swap.
func (c *Component) Install(ctx context.Context, opts plugin.InstallOptions) (plugin.InstallResult, error) {
	res := plugin.InstallResult{Component: Name}
	if c.Source == nil {
		return res, &errors.Error{
			Component:   Name,
			Problem:     Name + " component has no Source fs.FS configured",
			Cause:       "constructor was called with nil source",
			Recoverable: false,
		}
	}

	target, err := c.GlobalPath()
	if err != nil {
		return res, (&errors.Error{
			Component:   Name,
			Problem:     "cannot resolve builtins install path",
			Recoverable: true,
		}).Wrap(err)
	}

	desired, err := hashFS(c.Source)
	if err != nil {
		return res, (&errors.Error{
			Component:   Name,
			Problem:     "cannot hash builtin source",
			Recoverable: true,
		}).Wrap(err)
	}
	current, _ := hashTree(target)
	if desired == current && !opts.Force {
		res.AlreadyInstalled = true
		return res, nil
	}

	if opts.DryRun {
		return res, nil
	}

	parent := filepath.Dir(target)
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return res, (&errors.Error{
			Component:   Name,
			Problem:     "cannot create parent directory for builtins",
			Path:        parent,
			Recoverable: true,
		}).Wrap(err)
	}

	tmp, err := os.MkdirTemp(parent, "builtins.tmp-")
	if err != nil {
		return res, (&errors.Error{
			Component:   Name,
			Problem:     "cannot create staging dir for builtins",
			Path:        parent,
			Recoverable: true,
		}).Wrap(err)
	}
	cleanupTmp := func() { _ = os.RemoveAll(tmp) }

	if err := syncFS(c.Source, tmp); err != nil {
		cleanupTmp()
		return res, err
	}

	var backup string
	if _, statErr := os.Stat(target); statErr == nil {
		backup = target + ".bak-" + shortHash(desired)
		if err := os.Rename(target, backup); err != nil {
			cleanupTmp()
			return res, (&errors.Error{
				Component:   Name,
				Problem:     "cannot move existing builtins out of the way",
				Path:        target,
				Recoverable: true,
			}).Wrap(err)
		}
	}
	if err := os.Rename(tmp, target); err != nil {
		if backup != "" {
			_ = os.Rename(backup, target)
		}
		cleanupTmp()
		return res, (&errors.Error{
			Component:   Name,
			Problem:     "cannot rename staged builtins into place",
			Path:        target,
			Recoverable: true,
		}).Wrap(err)
	}
	if backup != "" {
		_ = os.RemoveAll(backup)
	}

	res.Mutations = append(res.Mutations, plugin.Mutation{
		Kind:        plugin.MutationDirCreated,
		Path:        target,
		Description: "synced samuel builtins to " + target,
		Reverse: func(context.Context) error {
			return os.RemoveAll(target)
		},
	})
	return res, nil
}

// Check reports samuel-builtins health: the global tree must be present
// and non-empty.
func (c *Component) Check(ctx context.Context) plugin.HealthStatus {
	d, err := c.Detect(ctx)
	if err != nil {
		return plugin.HealthStatus{
			Component: Name,
			OK:        false,
			Message:   "cannot detect samuel builtins: " + err.Error(),
			FixHint:   "samuel init --force",
		}
	}
	if !d.Installed {
		return plugin.HealthStatus{
			Component: Name,
			OK:        false,
			Message:   "samuel builtins not synced to " + d.Path,
			FixHint:   "samuel doctor --fix",
		}
	}
	msg := "samuel builtins synced"
	if d.Version != "" {
		msg = "samuel builtins " + d.Version + " synced"
	}
	return plugin.HealthStatus{Component: Name, OK: true, Message: msg}
}

// Uninstall removes the global tree when opts.Global or opts.All is
// set. Project-scoped state (`.samuel/builtins/` inside the project)
// is the init command's responsibility, not this component's.
func (c *Component) Uninstall(ctx context.Context, opts plugin.UninstallOptions) (plugin.UninstallResult, error) {
	res := plugin.UninstallResult{Component: Name}
	doGlobal := opts.Global || opts.All
	if !doGlobal {
		res.Skipped = true
		return res, nil
	}
	if opts.DryRun {
		return res, nil
	}
	target, err := c.GlobalPath()
	if err != nil {
		return res, err
	}
	if _, statErr := os.Stat(target); statErr != nil {
		// Nothing to remove — idempotent uninstall.
		return res, nil
	}
	if err := os.RemoveAll(target); err != nil {
		return res, (&errors.Error{
			Component:   Name,
			Problem:     "cannot remove global samuel builtins",
			Path:        target,
			Recoverable: true,
		}).Wrap(err)
	}
	res.Mutations = append(res.Mutations, plugin.Mutation{
		Kind:        plugin.MutationDirCreated,
		Path:        target,
		Description: "removed global samuel builtins",
	})
	return res, nil
}

// syncFS copies every file from src into dst, preserving directory
// structure. Path traversal defense: every entry must pass
// filepath.IsLocal — a malicious or buggy fs.FS yielding paths like
// ../../etc/passwd is rejected with a structured *Error.
func syncFS(src fs.FS, dst string) error {
	return fs.WalkDir(src, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == "." {
			return os.MkdirAll(dst, 0o700)
		}
		if !filepath.IsLocal(p) {
			return &errors.Error{
				Component: Name,
				Problem:   "path traversal attempt in builtin source",
				Cause:     "fs.FS yielded a non-local path: " + p,
				Path:      p,
			}
		}
		out := filepath.Join(dst, filepath.FromSlash(p))
		if d.IsDir() {
			return os.MkdirAll(out, 0o700)
		}
		f, err := src.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()
		w, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
		if err != nil {
			return err
		}
		defer w.Close()
		if _, err := io.Copy(w, f); err != nil {
			return err
		}
		return nil
	})
}

// hashFS returns a deterministic content hash of src's tree. Used by
// Install's idempotency check to skip work when the live tree already
// matches the desired tree.
func hashFS(src fs.FS) (string, error) {
	h := sha256.New()
	err := fs.WalkDir(src, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == "." {
			return nil
		}
		if !filepath.IsLocal(p) {
			return fmt.Errorf("non-local path in builtin source: %s", p)
		}
		fmt.Fprintf(h, "P:%s\n", p)
		if d.IsDir() {
			return nil
		}
		f, err := src.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := io.Copy(h, f); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// hashTree returns the same deterministic hash as hashFS but operates
// on a real on-disk tree.
func hashTree(root string) (string, error) {
	h := sha256.New()
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == root {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		fmt.Fprintf(h, "P:%s\n", filepath.ToSlash(rel))
		if d.IsDir() {
			return nil
		}
		f, err := os.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := io.Copy(h, f); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func shortHash(full string) string {
	if len(full) < 8 {
		return full
	}
	return strings.ToLower(full[:8])
}

// Compile-time guarantee that *Component satisfies plugin.Plugin.
var _ plugin.Plugin = (*Component)(nil)

// Package orchestrator coordinates installation, detection, health
// checking, and uninstallation of the Plugins that make up a Samuel
// project. Concurrent invocations across processes are serialized by
// the cross-process advisory flock in internal/lock.
//
// The orchestrator is the v2 plugin loader pattern (RFD 0005):
// regardless of plugin kind (builtin / skill / wasm / oci), the
// lifecycle protocol is the same and the rollback semantics are the
// same. Ported from samuel_v1/internal/orchestrator with the contract
// renamed to plugin.Plugin and the lock path moved to ~/.samuel/.
package orchestrator

import (
	"context"
	stderrors "errors"
	"fmt"
	"os"
	"time"

	"github.com/ar4mirez/samuel/internal/errors"
	"github.com/ar4mirez/samuel/internal/lock"
	"github.com/ar4mirez/samuel/internal/plugin"
)

// rollbackTimeout caps the fresh context the orchestrator builds for
// rollback. The install context may be canceled exactly at the moment
// of failure; running cleanup on a separate, bounded context means a
// canceled install does not also abort the undo work.
const rollbackTimeout = 30 * time.Second

// Component is the namespace used in structured errors produced by
// this package (lock acquisition, home-dir resolution).
const Component = "orchestrator"

// Orchestrator coordinates a Plugin list's lifecycle. The zero value
// is unusable — call New.
type Orchestrator struct {
	plugins []plugin.Plugin
	homeDir string // root for the advisory lock; defaults to $HOME
}

// New constructs an Orchestrator that runs the given plugins in
// declared order on Install and reverse order on Uninstall.
func New(plugins ...plugin.Plugin) *Orchestrator {
	return &Orchestrator{plugins: plugins}
}

// WithHomeDir overrides the home directory used to locate the advisory
// lock file. Tests point this at a t.TempDir() to avoid colliding with
// a real user install.
func (o *Orchestrator) WithHomeDir(home string) *Orchestrator {
	o.homeDir = home
	return o
}

// Plugins returns the registered plugins in declared order. The slice
// is shared, not copied — callers MUST NOT mutate it.
func (o *Orchestrator) Plugins() []plugin.Plugin { return o.plugins }

// Install runs each Plugin.Install in declared order. On any failure
// every applied Mutation (including the failing plugin's partial
// mutations) is reversed in LIFO order on a fresh rollback context.
//
// DryRun skips lock acquisition entirely — creating the lock file or
// writing a PID into it would itself be a state mutation, which DryRun
// must not produce.
func (o *Orchestrator) Install(ctx context.Context, opts plugin.InstallOptions) ([]plugin.InstallResult, error) {
	if !opts.DryRun {
		release, err := o.acquireLock()
		if err != nil {
			return nil, err
		}
		defer release()
	}

	results := make([]plugin.InstallResult, 0, len(o.plugins))
	applied := make([]plugin.Mutation, 0, len(o.plugins)*4)

	for _, p := range o.plugins {
		res, ierr := p.Install(ctx, opts)
		if res.Component == "" {
			res.Component = p.Name()
		}
		if ierr != nil {
			// Defense-in-depth: include the failing plugin's partial
			// mutations in the rollback queue even though plugins are
			// contracted to stage atomically.
			applied = append(applied, res.Mutations...)
			results = append(results, res)
			return results, o.rollbackOnFailure(p.Name(), ierr, applied)
		}
		results = append(results, res)
		applied = append(applied, res.Mutations...)
	}
	return results, nil
}

// Uninstall runs Plugin.Uninstall in reverse-of-install order. It is
// BEST-EFFORT: a failure in one plugin does not stop later plugins
// from running. All errors are collected and returned via errors.Join.
func (o *Orchestrator) Uninstall(ctx context.Context, opts plugin.UninstallOptions) ([]plugin.UninstallResult, error) {
	if !opts.DryRun {
		release, err := o.acquireLock()
		if err != nil {
			return nil, err
		}
		defer release()
	}

	results := make([]plugin.UninstallResult, 0, len(o.plugins))
	var errs []error
	for i := len(o.plugins) - 1; i >= 0; i-- {
		p := o.plugins[i]
		res, uerr := p.Uninstall(ctx, opts)
		if res.Component == "" {
			res.Component = p.Name()
		}
		results = append(results, res)
		if uerr != nil {
			errs = append(errs, fmt.Errorf("uninstall %s: %w", p.Name(), uerr))
		}
	}
	if len(errs) > 0 {
		return results, stderrors.Join(errs...)
	}
	return results, nil
}

// Doctor runs Check on every plugin. It does NOT acquire the lock
// (Check is read-only by contract), so concurrent Doctor calls are
// always safe.
func (o *Orchestrator) Doctor(ctx context.Context) []plugin.HealthStatus {
	out := make([]plugin.HealthStatus, 0, len(o.plugins))
	for _, p := range o.plugins {
		s := p.Check(ctx)
		if s.Component == "" {
			s.Component = p.Name()
		}
		out = append(out, s)
	}
	return out
}

// rollbackOnFailure runs Reverse on every applied mutation in LIFO
// order on a fresh, bounded context. When rollback itself fails the
// joined result is wrapped in a non-recoverable *Error so callers using
// errors.IsRecoverable see the right answer (without the wrap, errors.As
// would walk the joined tree and surface the install side's Recoverable
// flag — possibly true — even though manual cleanup is now required).
func (o *Orchestrator) rollbackOnFailure(name string, installErr error, applied []plugin.Mutation) error {
	rbCtx, cancel := context.WithTimeout(context.Background(), rollbackTimeout)
	defer cancel()
	rbErr := o.rollback(rbCtx, applied)
	wrapped := fmt.Errorf("install %s: %w", name, installErr)
	if rbErr == nil {
		return wrapped
	}
	joined := stderrors.Join(wrapped, fmt.Errorf("rollback: %w", rbErr))
	return (&errors.Error{
		Component:   Component,
		Problem:     "install failed and rollback also failed",
		Fix:         "inspect ~/.samuel state manually before retrying — automated cleanup did not complete",
		DocsURL:     "https://ar4mirez.github.io/samuel/docs/errors/SAM-ROLLBACK-001",
		Recoverable: false,
	}).Wrap(joined)
}

// rollback walks muts in reverse and runs each Reverse. Errors are
// collected and joined; rollback continues even on partial failure so
// the worst case is "best-effort cleanup ran, here is what failed"
// rather than "stuck halfway, no further attempts."
func (o *Orchestrator) rollback(ctx context.Context, muts []plugin.Mutation) error {
	var errs []error
	for i := len(muts) - 1; i >= 0; i-- {
		m := muts[i]
		if m.Reverse == nil {
			continue
		}
		if err := m.Reverse(ctx); err != nil {
			errs = append(errs, fmt.Errorf("rollback %s: %w", m.Path, err))
		}
	}
	return stderrors.Join(errs...)
}

func (o *Orchestrator) acquireLock() (func(), error) {
	home, err := o.resolveHome()
	if err != nil {
		return nil, err
	}
	return lock.Acquire(home)
}

func (o *Orchestrator) resolveHome() (string, error) {
	if o.homeDir != "" {
		return o.homeDir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", (&errors.Error{
			Component:   Component,
			Problem:     "cannot determine home directory",
			Fix:         "set HOME environment variable",
			Recoverable: true,
		}).Wrap(err)
	}
	return home, nil
}

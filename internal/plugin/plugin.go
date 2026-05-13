// Package plugin defines the universal Plugin contract every
// installable unit in Samuel v2 satisfies: built-in framework
// components, text skills, WASM modules, and OCI containers. The
// interface is the foundation for the v2 plugin loader and is the
// subject of RFD 0005.
//
// This package ships interface + companion types only; the three
// kinds (SkillPlugin, WasmPlugin, OciPlugin) and the built-in syncer
// land in PRD 0003 / PRD 0002. The empty struct stubs here are
// compile-time placeholders that exist so the rest of the framework
// can begin importing the kind-specific types.
package plugin

import (
	"context"
	"io"
)

// Plugin is the contract satisfied by every installable unit. It
// extends v1's Component with one new method, Manifest, that returns
// the parsed samuel-plugin.toml.
type Plugin interface {
	// Name returns the stable identifier used in logs, errors, and the
	// installed-plugin manifest. Examples: "go-guide", "ralph",
	// "claude-runner".
	Name() string

	// Manifest returns the parsed samuel-plugin.toml. The framework
	// reads it for capability checks, version-range resolution, and
	// dependency-graph construction.
	Manifest() Manifest

	// Detect inspects the system and reports whether the plugin is
	// installed, what version is present, and where it lives. It MUST
	// NOT mutate state.
	Detect(ctx context.Context) (DetectResult, error)

	// Install brings the plugin to the desired state. It MUST be
	// idempotent. On failure, the plugin is responsible for staging
	// changes so callers can roll back via InstallResult.Mutations.
	Install(ctx context.Context, opts InstallOptions) (InstallResult, error)

	// Check reports the current health for `samuel doctor`. It MUST
	// NOT mutate state and MUST be safe to call without the install
	// lock.
	Check(ctx context.Context) HealthStatus

	// Uninstall reverses Install. Idempotent — uninstalling an absent
	// plugin is a no-op.
	Uninstall(ctx context.Context, opts UninstallOptions) (UninstallResult, error)
}

// Kind classifies a plugin tier.
type Kind string

const (
	KindBuiltin Kind = "builtin"
	KindSkill   Kind = "skill"
	KindWasm    Kind = "wasm"
	KindOci     Kind = "oci"
)

// Manifest models the samuel-plugin.toml shape (RFD 0001 + RFD 0003).
// Fields land progressively; this snapshot covers what PRD 0003 wires
// in the registry/loader.
type Manifest struct {
	Name         string   `toml:"name"`
	Version      string   `toml:"version"`
	Kind         Kind     `toml:"kind"`
	Summary      string   `toml:"summary,omitempty"`
	Homepage     string   `toml:"homepage,omitempty"`
	License      string   `toml:"license,omitempty"`
	Authors      []string `toml:"authors,omitempty"`
	Capabilities []string `toml:"capabilities,omitempty"`
	MinSamuel    string   `toml:"min_samuel,omitempty"`
	MaxSamuel    string   `toml:"max_samuel,omitempty"`
	Source       string   `toml:"source,omitempty"`
	Digest       string   `toml:"digest,omitempty"`
}

// DetectResult captures the current state of a plugin on the host.
type DetectResult struct {
	Installed bool
	Version   string
	Path      string
}

// InstallOptions configures how Install runs.
type InstallOptions struct {
	DryRun  bool
	Force   bool
	Verbose bool
	Stdout  io.Writer
}

// UninstallOptions configures how Uninstall runs.
type UninstallOptions struct {
	DryRun  bool
	Verbose bool
	Stdout  io.Writer
	// Project removes only project-local artifacts.
	Project bool
	// Global removes only user-scoped artifacts (~/.samuel/...).
	Global bool
	// All removes both (mutually exclusive with Project/Global; the
	// CLI validates at the boundary).
	All bool
}

// InstallResult records what Install actually changed. Mutations are
// executed in declared order; the loader rolls them back in reverse
// on a partial failure.
type InstallResult struct {
	Component        string
	Mutations        []Mutation
	AlreadyInstalled bool
	Skipped          bool
}

// UninstallResult mirrors InstallResult for the reverse direction.
type UninstallResult struct {
	Component string
	Mutations []Mutation
	Skipped   bool
}

// Mutation describes one state change. Plugins emit these in
// chronological order; the loader rolls back in reverse.
type Mutation struct {
	Kind        MutationKind
	Path        string
	Description string
	// Reverse undoes this mutation. Required for rollback. MUST be
	// safe to call multiple times.
	Reverse func(context.Context) error
}

// MutationKind classifies state changes for telemetry and rollback.
type MutationKind string

const (
	MutationFileWritten    MutationKind = "file_written"
	MutationSymlinkCreated MutationKind = "symlink_created"
	MutationDirCreated     MutationKind = "dir_created"
	MutationCommandRun     MutationKind = "command_run"
	MutationGitClone       MutationKind = "git_clone"
	MutationOciPull        MutationKind = "oci_pull"
	MutationWasmCache      MutationKind = "wasm_cache"
)

// HealthStatus is what Check returns; the orchestrator rolls all
// statuses into `samuel doctor` output.
type HealthStatus struct {
	Component string
	OK        bool
	Message   string
	FixHint   string
}

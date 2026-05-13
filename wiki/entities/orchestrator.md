---
title: orchestrator package (Orchestrator, Component, Lock, Error)
type: entity
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-orchestrator]
tags: [v1, orchestrator, rescue]
---

# orchestrator package

The lifecycle-coordination layer of v1. Highest-engineering-quality subsystem in v1.

## Public surface

### `Orchestrator`

```go
type Orchestrator struct { ... }

func New(components ...Component) *Orchestrator
func (o *Orchestrator) WithHomeDir(home string) *Orchestrator
func (o *Orchestrator) Install(ctx, InstallOptions) ([]InstallResult, error)
func (o *Orchestrator) Uninstall(ctx, UninstallOptions) ([]UninstallResult, error)
func (o *Orchestrator) Doctor(ctx) []HealthStatus
```

- `Install` runs components in declared order with rollback-on-failure.
- `Uninstall` runs components in **reverse** order, best-effort, errors joined.
- `Doctor` runs `Check` on each, no lock acquired (read-only).
- `WithHomeDir` enables hermetic tests against a temp dir.

### `Component` interface

```go
type Component interface {
    Name() string
    Detect(ctx) (DetectResult, error)
    Install(ctx, InstallOptions) (InstallResult, error)
    Check(ctx) HealthStatus
    Uninstall(ctx, UninstallOptions) (UninstallResult, error)
}
```

Contract:
- **Idempotent**: reinstall current = no-op (`AlreadyInstalled=true`).
- **Detect / Check MUST NOT mutate.** Pure read.
- **Install must stage atomically.** On error, return mutations applied so far; caller rolls back.

### `Mutation` log

```go
type Mutation struct {
    Kind        MutationKind  // file_written | symlink_created | dir_created | command_run | git_clone
    Path        string
    Description string
    Reverse     func(ctx) error  // called LIFO on rollback
}
```

### Structured `Error`

```go
type Error struct {
    Component   string
    Problem     string  // one-line description
    Cause       string  // root cause / wrapped err string
    Fix         string  // copy-pasteable remediation
    DocsURL     string  // optional documentation
    Recoverable bool    // can the user fix this themselves?
    Path        string  // filesystem path involved
}

func (e *Error) Wrap(err error) *Error
func IsRecoverable(err error) bool
```

Rendered in interactive CLI as a multi-line block:
```
Error: Cannot register gbrain MCP server
  Cause: gbrain not found on PATH
  Fix:   bun install -g gbrain
  Docs:  https://samuel.dev/docs/errors/SAM-MCP-001
```

### Lock

flock(2) at `~/.claude/.samuel.lock`. `O_CLOEXEC` to keep child processes from inheriting. Persistent file (never removed — removing races with new acquirers).

- Acquire blocks Install / Uninstall.
- Skipped in `DryRun` (lock file creation would count as mutation).
- Holder hint reads first 32 bytes, validates PID, otherwise reports "unknown".

## Cross-platform

`lock_unix.go` has the real implementation, gated by `//go:build unix`. `lock_other.go` provides a no-op fallback for non-Unix builds.

## v2 implications #rescue

Port the package nearly verbatim. Refinements:

- Rename `Component` → `Plugin` or `Module` (the v2 thing).
- Add a `Manifest()` method returning the plugin's TOML manifest (name, version, capabilities, etc.) — for v2's [[concepts/versioning-compatibility]] checks.
- Lock path moves from `~/.claude/.samuel.lock` to `~/.samuel/lock`.
- The `Doctor` method becomes `samuel doctor` core; plugins can implement Check to participate.
- The Mutation log becomes the basis of `samuel.lock` — every plugin install adds entries that uninstall can reverse.

## Related

- [[entities/component-samuel]]
- [[entities/component-gstack-gbrain]]
- [[concepts/component-lifecycle]]
- [[concepts/structured-errors]]
- [[synthesis/orchestrator-as-plugin-loader]]

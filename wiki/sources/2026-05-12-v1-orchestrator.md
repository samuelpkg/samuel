---
title: v1 Orchestrator (component lifecycle + rollback)
type: source
created: 2026-05-12
updated: 2026-05-12
sources: []
tags: [v1, orchestrator]
---

# v1 Orchestrator

Ingest pass 4. The component-lifecycle layer — install/uninstall/health-check coordination for Samuel's curated bundle.

## Files

- `samuel_v1/internal/orchestrator/orchestrator.go` (236 lines) — `Orchestrator` type, `Install`, `Uninstall`, `Doctor`, rollback logic, lock acquisition
- `samuel_v1/internal/orchestrator/component.go` (187) — `Component` interface, `DetectResult`, `InstallOptions`, `InstallResult`, `Mutation`, `MutationKind`, `HealthStatus`
- `samuel_v1/internal/orchestrator/component_samuel.go` (577) — `SamuelComponent`: sync embedded skills to `~/.claude/skills/samuel/` + project symlink
- `samuel_v1/internal/orchestrator/component_gstack.go` (291) — `GstackComponent`: clone+setup `garrytan/gstack` at pinned SHA
- `samuel_v1/internal/orchestrator/component_gbrain.go` (283) — `GbrainComponent`: register `gbrain` MCP server via `claude mcp add`
- `samuel_v1/internal/orchestrator/errors.go` (87) — structured `Error` type
- `samuel_v1/internal/orchestrator/lock_unix.go` (142) — flock(2) advisory lock
- `samuel_v1/internal/orchestrator/lock_other.go` (26) — non-Unix fallback (build tag)

## Key claims

### Component interface (`component.go`)

```go
type Component interface {
    Name() string
    Detect(ctx) (DetectResult, error)     // read-only state inspection
    Install(ctx, InstallOptions) (InstallResult, error)
    Check(ctx) HealthStatus                // read-only health for `samuel doctor`
    Uninstall(ctx, UninstallOptions) (UninstallResult, error)
}
```

- **Idempotency required**. Reinstalling current is a no-op (`AlreadyInstalled=true`).
- **Detect and Check MUST NOT mutate** — enforced by contract, not types.
- **Install must stage atomically** — failures get rolled back via returned `InstallResult.Mutations`.
- Component name constants: `NameGstack`, `NameGbrain`, `NameSamuelSkills`, `NameOrchestrator`. (`component.go:22-29`)

### Mutation/Reverse log

```go
type Mutation struct {
    Kind        MutationKind   // file_written | symlink_created | dir_created | command_run | git_clone
    Path        string
    Description string
    Reverse     func(ctx) error  // required, called on rollback
}
```

Every state change emits a `Mutation` with its reversal closure. Orchestrator collects them in chronological order, runs them in **reverse LIFO** on failure. (`component.go:142-172`)

### Orchestrator lifecycle (`orchestrator.go`)

`Install(ctx, opts)` loop:

```
for c in components (in declared order):
    if shouldSkip(c, opts):
        record skipped; continue
    res, err = c.Install(ctx, opts)
    if err:
        applied += res.Mutations            # include partial mutations
        return results, rollbackOnFailure(c, err, applied)
    applied += res.Mutations
```

- **Order matters**: gstack → gbrain → samuel-skills. (Declared via `orchestrator.New(...)`.)
- **Rollback runs on a fresh context** (`rollbackTimeout = 30s`) so a cancelled install doesn't also kill cleanup. (`orchestrator.go:14-15`, `159-175`)
- **If rollback also fails**, the joined error is wrapped in `*Error{Recoverable: false, DocsURL: SAM-ROLLBACK-001}` — without this wrapper, `errors.As` would walk into the install side's `Recoverable` flag and possibly mis-report. (`orchestrator.go:165-174`)

`Uninstall(ctx, opts)` runs components in **reverse** of install order. **Best-effort** — failures collected via `errors.Join`, not aborted. (`orchestrator.go:109-147`)

`Doctor(ctx)` runs `Check` on every component. No lock acquired (Check is read-only). (`orchestrator.go:97-107`)

### Lock (`lock_unix.go`)

flock(2) advisory lock at `~/.claude/.samuel.lock`. The file persists across runs; the body holds the PID as a diagnostic hint.

- **`O_CLOEXEC`** so child processes (gstack setup, claude mcp) don't inherit the lock and accidentally hold it past Samuel's exit. (`lock_unix.go:48-50`)
- **`LOCK_EX | LOCK_NB`** — exclusive, non-blocking. EWOULDBLOCK/EAGAIN → "another samuel process is running" with PID hint. (`lock_unix.go:61-78`)
- **Lock file never removed** — removing would race with another acquirer. Persistent file + kernel-managed flock is the safe combination. (`lock_unix.go:106-112`)
- **Holder hint** read with 32-byte limit and PID validation — defense against malicious or crashed processes planting blobs. (`lock_unix.go:125-142`)
- **DryRun skips lock acquisition entirely** — creating the lock file counts as mutation, violating the no-state-change contract. (`orchestrator.go:60-66`)

### Structured Error (`errors.go`)

```go
type Error struct {
    Component   string
    Problem     string   // one-line description
    Cause       string   // root cause
    Fix         string   // copy-pasteable remediation
    DocsURL     string   // optional documentation link
    Recoverable bool
    Path        string   // filesystem path involved
    wrapped     error    // chain for errors.Is/As
}
```

Rendered by the CLI as:
```
Error: Cannot register gbrain MCP server
  Cause: gbrain not found on PATH
  Fix:   bun install -g gbrain
  Docs:  https://samuel.dev/docs/errors/SAM-MCP-001
```

`Wrap(err)` preserves the chain. `IsRecoverable(err)` does `errors.As` and reads the flag. (`errors.go`)

### SamuelComponent (`component_samuel.go`)

Syncs `embed.FS` (from pass 1) → `~/.claude/skills/samuel/` + creates project symlink at `<project>/.claude/skills/samuel/`.

Idempotency via **content hash**: SHA-256 over `P:<path>\n<bytes>` for every entry in both source and target. Skip work if hashes match AND symlink is OK. (`component_samuel.go:142-157`, `499-568`)

**Atomic swap**:
1. Stage sync into sibling tmp dir (`samuel.tmp-<rand>`)
2. If target exists, rename to `<target>.bak-<shortHash>`
3. Rename tmp → target
4. On success: drop backup. On failure: restore backup.

(`component_samuel.go:173-223`)

**Path traversal defense** in `syncFS`: `filepath.IsLocal(p)` check before any write. Rejects fs.FS paths like `../etc/passwd`. (`component_samuel.go:459-497`)

**Symlink conflict matrix** (`component_samuel.go:381-431`):
- missing → create
- symlink to target → no-op
- symlink elsewhere → remove + recreate
- real file/dir → **refuse**, return Recoverable error (don't clobber user data)

### GstackComponent (`component_gstack.go`)

Composes external `github.com/garrytan/gstack` into Samuel.

- **Pinned SHA**: `gstackPinnedSHA = "e8893a18b18e32ebd63a21f6915337868249ebe1"`. Bumping is a deliberate Samuel release event with documented procedure (`gstack_component.go:13-23`).
- **Process**: `git clone --quiet <url> <path>` → `git checkout --quiet <sha>` → run `<path>/setup --team --quiet --host claude`.
- Detect uses `git rev-parse --short HEAD`; matches prefix-only via `matchesShortSHA` (case-insensitive).
- **Uninstall is intentionally no-op** — gstack is user-owned and may be shared with other tools. Diagnostic message printed instead. (`component_gstack.go:275-281`)
- Refuses to overwrite a different SHA without `--force` — surfaces a clear error.

### GbrainComponent (`component_gbrain.go`)

Registers `gbrain` as a Claude Code MCP server.

- **Does NOT install gbrain itself** — user installs via `bun add -g gbrain` or `npm install -g gbrain`.
- Uses `claude mcp add -s user gbrain <bin> serve` (scope: user). (`component_gbrain.go:131-133`)
- Pre-mutation gate: check `gbrain` on PATH AND `claude` on PATH BEFORE any mutation — returns Recoverable error if either missing. (`component_gbrain.go:94-115`)
- Idempotency via `claude mcp get gbrain` (exits non-zero if absent — treated as "not registered"). (`component_gbrain.go:268-283`)
- Uninstall: `claude mcp remove -s user gbrain`. Treats "not found" output as idempotent success. (`component_gbrain.go:218-266`)
- `gbrainExec` and `gbrainLookPath` are package-level vars overridable in tests — clean test injection. (`component_gbrain.go:12-19`)

## Assessment

- **Credibility**: high.
- **Engineering quality**: high. This is the highest-quality subsystem in v1. The orchestrator + Component pattern, the Mutation/Reverse log, the structured errors, the flock semantics, the atomic swap — all well-engineered.
- **The pattern generalizes**. This same shape can drive v2's plugin install/uninstall lifecycle.

## v2 implications

### `#rescue` (highest confidence in any pass so far)

- **Component interface** — port verbatim. Rename: `Component` → `Plugin` (or keep `Component` for internal-built-ins and `Plugin` as a subtype).
- **Mutation + Reverse log** — port verbatim. Plugins emit mutations; samuel runs them in reverse on rollback or uninstall.
- **Structured `Error` type** — port verbatim. Every plugin error gets `Problem/Cause/Fix/DocsURL`.
- **flock(2) advisory locking** — port verbatim. Same lock path semantics.
- **`O_CLOEXEC` on lock fd** — keep this. Child processes (sandbox containers, wasm runtimes) must not inherit the lock.
- **DryRun = no lock acquired** — keep the contract.
- **Atomic swap** (sibling tmp + rename + backup restore) — keep. Use for any plugin install that writes to a known directory.
- **Content-hash idempotency** — keep. Plugins should hash their payload and skip work when current.
- **Best-effort uninstall** with `errors.Join` — keep.
- **Pinned SHA pattern for external deps** — keep, generalize to "plugin manifest declares pinned version + content hash".

### `#refactor`

- **SamuelComponent → v2 framework bootstrap**. Drop the symlink (v2 might have a different layout) but keep the embed→disk sync pattern for built-in skills.
- **GstackComponent and GbrainComponent → questionable.** Both are v1 product opinions. Worth a user decision (see "Open questions for the user" below). If they survive, they become plugins, not built-ins.

### `#drop`

- Hardcoded ordering of components — v2 should compute install order from a plugin dependency graph.
- Component name constants in source — derived from plugin manifest names.

### `#open` — pending user decisions

1. **Does gstack survive v2?** v1 composes it as a core part of the bundle. Is "git stack" workflow part of v2's identity, or was it a v1 product opinion?
2. **Does gbrain survive v2?** Same question. v1 ships an MCP-server registration as part of `samuel init`. v2 either keeps it as a built-in or makes it a plugin.

If both drop, v2's "core install" is just: framework binary + built-in skills synced to disk + lockfile setup. Plugins handle everything else.

## Related pages

- [[entities/orchestrator]]
- [[entities/component-samuel]]
- [[entities/component-gstack-gbrain]]
- [[concepts/component-lifecycle]]
- [[concepts/structured-errors]]
- [[synthesis/orchestrator-as-plugin-loader]]

---
title: Component lifecycle (Detect / Install / Check / Uninstall + Mutation log)
type: concept
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-orchestrator]
tags: [v1, rescue, v2-decision]
---

# Component lifecycle

The interface every orchestrated piece implements. This is the pattern v2's plugin loader should adopt verbatim.

## The four methods

```
Detect (read-only)   → "what's currently on disk?"
Install              → "bring it to the desired state, atomically, with rollback support"
Check  (read-only)   → "is it healthy right now?"
Uninstall            → "reverse Install, best-effort"
```

## Contract

- **Idempotent.** Install on a current install = no-op (return `AlreadyInstalled=true`). Uninstall on an absent install = no-op.
- **Read-only Detect and Check.** No side effects, ever. Doctor doesn't acquire the lock because of this.
- **Atomic staging.** Install must hold mutations until the swap. On error, return what was applied — caller rolls back.
- **Mutations are reversible.** Every state change emits a `Mutation{Kind, Path, Description, Reverse}`. `Reverse` MUST be safe to call multiple times.

## The mutation/reverse log

The state-change accounting that makes rollback automatic:

```go
type Mutation struct {
    Kind        MutationKind
    Path        string
    Description string
    Reverse     func(ctx) error
}
```

Each component emits these in chronological order. The orchestrator concatenates them across components. On failure (or uninstall), it walks them in reverse LIFO calling `Reverse(ctx)`.

### Why LIFO

If Install creates `~/foo/` then `~/foo/bar/`, uninstall must remove `~/foo/bar/` before `~/foo/` — directory must be empty to remove. LIFO ensures the order is right.

### Rollback context separation

Rollback runs on a **fresh** `context.WithTimeout(rollbackTimeout)`, not the install ctx. Reasoning:

> If the user Ctrl-C'd the install, that cancellation should kill the install but **not** the cleanup. Cleanup needs to run to completion regardless of whether the user wanted the install to keep going.

This subtle but important detail is the difference between "graceful failure" and "stuck halfway, nothing to recover."

## Best-effort uninstall

Uninstall doesn't stop on first failure. Each component's error is collected via `errors.Join`. Worst case for the user: "most things uninstalled, here's a list of the failures and what to clean up manually."

Compare to abort-on-first-error: "uninstall failed at component 3 of 5, no idea what state we're in." Much worse UX.

## Order matters

Install order is declared at orchestrator construction:

```go
orchestrator.New(
    NewGstackComponent(""),
    NewGbrainComponent(""),
    NewSamuelComponent(skills.MustFS(), "", "", version),
)
```

Uninstall walks **reverse**. Dependencies install bottom-up, uninstall top-down. Sound.

## v2 application #v2-decision

Adopt verbatim. Plugins implement the same shape:

```go
type Plugin interface {
    Name() string
    Manifest() PluginManifest                                  // new for v2: TOML manifest
    Detect(ctx) (DetectResult, error)
    Install(ctx, InstallOptions) (InstallResult, error)
    Check(ctx) HealthStatus
    Uninstall(ctx, UninstallOptions) (UninstallResult, error)
}
```

Built-in components (samuel-skills, methodology hooks) implement directly in Go. WASM plugins implement via an adapter that translates the interface to wazero host functions. OCI plugins implement via an adapter that translates to container lifecycle commands.

The pattern is the same across all three [[concepts/plugin-format]] tiers.

## Dependency ordering in v2

v1 hardcodes install order in `orchestrator.New(...)`. v2 should compute it from the dependency graph declared in each plugin manifest:

```toml
[requires]
go-runtime = "^1.0"
samuel-skills = "*"
```

Topological sort. Cycles → error before any install starts.

## Open

- Versioned Install — can we run `Install` to upgrade an existing install? v1 uses Force; v2 might want an explicit `Upgrade` method.
- Migration hooks — when a plugin moves from v1 to v2 internally, can it run a migration as part of Install? Or is that always user-explicit (`samuel migrate <plugin>`)?
- Health severity — `HealthStatus.OK` is a bool. Real-world health is "healthy / degraded / failing". Add a `Severity` enum?

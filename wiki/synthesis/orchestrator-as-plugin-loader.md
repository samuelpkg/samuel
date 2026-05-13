---
title: Orchestrator as plugin loader
type: synthesis
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-orchestrator]
tags: [v2, v2-decision, plugin-loader]
---

# Orchestrator as plugin loader

The single biggest engineering takeaway from pass 4: v1's orchestrator pattern is already the right shape for v2's plugin loader. Don't redesign — port and extend.

## The mapping

| v1 concept | v2 concept |
|---|---|
| `Orchestrator` | `PluginManager` (or keep the name) |
| `Component` interface | `Plugin` interface (Detect/Install/Check/Uninstall) |
| `Mutation + Reverse` | `samuel.lock` entries (each install records mutations; uninstall reverses them) |
| `Error{Problem/Cause/Fix/DocsURL}` | Plugin SDK error type, used framework-wide |
| flock(2) at `~/.claude/.samuel.lock` | flock(2) at `~/.samuel/lock` |
| Atomic swap (tmp dir + rename + backup restore) | Plugin payload install path |
| Pinned SHA for gstack | Plugin manifest version pin (`version = "1.4.2"`) |
| Rollback context separation | Same — survives plugin Ctrl-C cleanly |
| Best-effort uninstall with `errors.Join` | Same — applies to multi-plugin uninstall |
| `Doctor` running `Check` per component | `samuel doctor` aggregating per-plugin health |
| Order via `New(components...)` | Order computed from manifest dependency graph |

## What v2 adds on top

### `Manifest()` method

Plugins return their TOML manifest. The loader uses this for:

- Capability checks (does the user grant `filesystem.write:/workspace`?)
- Version range validation (does the plugin support this framework version?)
- Dependency resolution (topological sort across plugins)
- Sigstore signature verification

### Three plugin kinds, one interface

Per [[concepts/plugin-format]]:

- **Skill plugins** — implement `Install` as "copy text to disk." No execution.
- **WASM plugins** — `Install` instantiates wazero module, `Check` calls a health export.
- **OCI plugins** — `Install` pulls image, `Check` calls a health endpoint inside a running container.

All three implement the same `Plugin` interface. The framework doesn't care which kind it's talking to.

### Dependency-graph install order

v1 hardcodes order via `New(gstack, gbrain, samuelSkills)`. v2 reads `[requires]` from each manifest:

```toml
[requires]
go-runtime = "^1.0"
```

Topological sort. Cycles error before any install starts. Missing dependencies prompt to install them first.

### Plugin migration

When a plugin moves from v1.x to v2.0 internally, the loader should run a one-time migration. Strawman:

```toml
[migrations]
"1.x->2.0" = "migrate-1-to-2.wasm"
```

Loader invokes the migration WASM before the new version's `Install`. Out-of-scope for v2.0 release; sketch the path now.

## What stays the same

- **Idempotency contract** — same as v1.
- **Read-only Detect / Check** — same.
- **Mutation log + LIFO Reverse** — same.
- **DryRun skips lock** — same.
- **Atomic staging** — same.
- **Structured errors** — same.

## Implementation order for v2

Suggested build order:

1. Port `errors.go` + `lock_unix.go` first — pure utilities with high reuse.
2. Port `Component` interface + `Orchestrator` core. Substitute `Plugin` naming.
3. Port `SamuelComponent` as the **framework-self-install** built-in (samuel binary syncing its own bundled skills).
4. Add `Manifest()` + manifest parsing (`samuel-plugin.toml`).
5. Build the three plugin-kind adapters (skill / WASM / OCI) on top of `Plugin`.
6. Add dependency resolver.
7. Add Sigstore verification.

That's roughly the bottom-up build path for v2's plugin layer.

## Open

- Naming: `Component` (v1 internal) → `Plugin` (v2 generic) or keep separate names for built-in vs external? Suggest one name (`Plugin`) with a `Builtin bool` field — same interface, less special-casing.
- Plugin namespace at install time — `~/.samuel/plugins/<plugin-name>/` or `~/.samuel/plugins/<plugin-name>@<version>/`? The latter lets multiple versions coexist for testing; the former is simpler. v1 used unversioned paths.
- Hot reload — can a plugin be reloaded without restarting Samuel? Probably not for v2.0 (complex), but worth noting as a future capability.

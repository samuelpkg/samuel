---
title: GstackComponent + GbrainComponent (composed externals)
type: entity
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-orchestrator]
tags: [v1, orchestrator, drop]
---

# GstackComponent + GbrainComponent

Two components for **composed external tools** that v1 wires into the Samuel bundle.

**Status: `#drop` in v2** (decision 2026-05-12). Neither survives the rebuild. Both were v1 product opinions, not framework essentials. May be reintroduced later as plugins, or pieces extracted as skills.

## GstackComponent

Composes [`github.com/garrytan/gstack`](https://github.com/garrytan/gstack) — a "git stack" workflow tool.

### Process

```
1. git clone --quiet https://github.com/garrytan/gstack <home>/.claude/skills/gstack
2. git checkout --quiet <gstackPinnedSHA>
3. <home>/.claude/skills/gstack/setup --team --quiet --host claude
```

### Pinned SHA pattern

```go
const gstackPinnedSHA = "e8893a18b18e32ebd63a21f6915337868249ebe1"
```

Bumping is a deliberate release event with documented procedure:
1. `git ls-remote https://github.com/garrytan/gstack HEAD` for new SHA
2. Run integration test of `samuel init` against new SHA in clean container
3. Update CHANGELOG with delta

Refuses to overwrite a different SHA without `--force` — surfaces a clear error so the user can decide.

### Uninstall = no-op

Intentional. gstack is **user-owned and shared** with other tools (Codex skill adapters, Cursor integrations). Removing it from `samuel uninstall` would silently break unrelated workflows.

User who wants gstack gone runs `rm -rf ~/.claude/skills/gstack` themselves.

## GbrainComponent

Registers [`gbrain`](https://github.com/) as a Claude Code MCP server. v1 does NOT install gbrain — user is responsible for `bun add -g gbrain` or `npm install -g gbrain`.

### Process

```
1. (Detect): exec.LookPath("gbrain") + claude mcp get gbrain
2. (Install): claude mcp add -s user gbrain <bin> serve
3. (Uninstall): claude mcp remove -s user gbrain
```

Idempotency: `claude mcp get gbrain` exits non-zero when absent → treated as "not registered".

### Pre-mutation gate

Before invoking `claude mcp add`:

- `gbrain` must be on PATH (or override via `--gbrain-binary`)
- `claude` must be on PATH

If either is missing: structured Recoverable error with install guidance. **No partial-install state** — fails before the first mutation.

### Test injection

```go
var gbrainExec    = exec.CommandContext
var gbrainLookPath = exec.LookPath
```

Tests override these package-level vars to inject fakes. CI runners (where `claude` is not installed) don't fail the suite.

## v2 implications #drop

Both dropped. They're orthogonal to "Rails for coding assistants." Git-stack workflow is one valid choice; gbrain is one valid memory-MCP choice. v2's hub model says "let users pick" — built-ins should be only universal cross-tool primitives.

v2's `samuel init` becomes much smaller as a result: framework binary + built-in skills + lockfile setup. That's it.

### Future-plugin sketch (optional, not v2.0)

If reintroduced later:

- `samuel-gstack` plugin: declares `requires: ["git"]`, installs to `~/.samuel/plugins/gstack/`, composes gstack as a registered "stack manager" capability.
- `samuel-gbrain` plugin: declares `requires: ["gbrain", "claude"]` on PATH, registers MCP server via `claude mcp add`.

The orchestrator pattern they were built on **is** the right shape for plugins. Lift the pattern; leave the specific components behind.

### Pinned-SHA pattern survives

The pinned-SHA-with-documented-bump-procedure pattern is the right model for any plugin that wraps an external repo. Codify it in the plugin manifest spec ([[concepts/versioning-compatibility]]).

## Related

- [[entities/orchestrator]]
- [[concepts/component-lifecycle]]
- [[concepts/structured-errors]]

---
title: JSON mode everywhere
type: concept
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-commands]
tags: [v1, rescue, automation]
---

# JSON mode everywhere

Every Samuel command can emit machine-parseable JSON. The mechanism is small, the discipline is consistent. Worth porting as a v2 invariant.

## The shape

`--json` is a persistent root flag. `JSONMode(cmd)` returns true if the user passed it on the current command OR on the root.

Every handler has the same pattern:

```go
if JSONMode(cmd) {
    ui.PrintJSON("subcommand-path", map[string]interface{}{...})
    return nil
}

// Human output
ui.Success("...")
return nil
```

The envelope (from [[entities/ui-package]]):

```json
{
  "schemaVersion": 3,
  "command": "<invoked command path>",
  "success": true,
  "data": { /* command-specific */ },
  "error": "..."
}
```

- Schema-versioned. Consumers branch on `schemaVersion`.
- Stdout for success. Stderr for errors (`PrintJSONError`).
- Indent 2 spaces — pretty-printed by default. Acceptable; piping to `jq` doesn't care.
- The `command` field reflects the **invoked** command path (so legacy aliases show what the user typed, not the redirected target).

## What every command exposes

- `samuel ls --json` → `{installed: {languages, frameworks, workflows, skills}}`
- `samuel skill list --json` → `{total, skills: [{name, description, valid, errors}]}`
- `samuel doctor --json` → `{healthy, passed, failed, checks: [{name, passed, message, fixable}]}`
- `samuel run tasks --json` → task list
- `samuel run status --json` → progress summary
- `samuel add --json` → `{type, name, path, installed}` or `{alreadyInstalled: true}`
- `samuel init --json` → `{version, path, languages, frameworks, workflows}`
- `samuel admin config get k --json` → `{key, value}`

Consistent enough to write automation against. CI scripts, dashboards, Cursor/Claude Code integrations.

## Why this matters

Samuel is the kind of tool that:
- Runs in CI (need machine-readable status).
- Runs from IDE integrations (Cursor sidebar, Claude Code commands).
- Runs from other CLIs (`gh` integration, scripted workflows).
- Sometimes runs from agents (an agent invokes `samuel run status` to check progress).

Without `--json`, every consumer scrapes human output. Brittle.

With `--json`, every consumer reads structured data. Stable.

## v2 application #rescue

**Mandatory invariant**: every v2 command implements `--json`. Code review rejects PRs that don't.

- Same envelope shape. Bump `schemaVersion` to 4 if anything changes.
- Same stdout/stderr split.
- Plugin SDK exposes `cmd.JSONMode()` + `cmd.PrintJSON(...)` helpers. WASM plugins emit JSON-structured output via host functions.
- Consider **JSON Lines** as an additional mode for long-running commands. `samuel run start --json` should emit one JSON line per iteration, not one giant JSON at the end.

## Open

- `--json` vs `--output json`: latter is more conventional (gh, kubectl). Likely the right call for v2, with `--json` as a shortcut alias.
- Per-command JSON schema docs — generate from Go structs? Likely worth it for the public-facing commands.
- Streaming JSON Lines mode for long-running ops — `--output jsonl`?

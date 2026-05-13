---
title: Claude Code PreToolUse hooks (agent-boundary enforcement)
type: concept
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-cli-entry-runtime]
tags: [v1, pattern, translator-plugin]
---

# Claude Code PreToolUse hooks

A pattern v1 uses to enforce prerequisites **at the agent's tool boundary**. Much stronger than asking nicely in CLAUDE.md or relying on the agent to "remember" project rules.

## The mechanism (Claude Code feature)

Claude Code's `.claude/settings.json` can register hooks that fire before specific tools are used:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Skill",                                       // tool name to gate
        "hooks": [
          {
            "type": "command",
            "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/check-gstack.sh\""
          }
        ]
      }
    ]
  }
}
```

When the agent tries to use the matched tool, Claude Code runs the command first. The hook's **stdout JSON** controls what happens:

```json
{
  "permissionDecision": "deny",
  "message": "..."
}
```

- `deny` → tool call blocked. Stderr message shown to user.
- `{}` → allow (no opinion).

Hook stderr gets shown to the user. Use it for actionable install instructions.

## v1 use case: gate Skill usage on gstack

v1's `check-gstack.sh`:

1. Checks `~/.claude/skills/gstack/bin/` exists.
2. If missing, emits stderr install instructions + denies the tool call.
3. Otherwise allows.

The result: an agent that tries to use any Skill in a project without gstack installed gets blocked **at the tool call**, before any code runs. The user sees the install command in stderr. Strong defensive UX.

## Why this matters for v2

Even though gstack drops in v2, the **pattern** is worth keeping. Lots of v2 use cases want this kind of enforcement:

- Deny `Read` on paths outside `/workspace` (sandbox escape detection).
- Deny `Bash` commands matching dangerous patterns (`rm -rf /`, `curl ... | sh`).
- Audit-log every `Write` to a sensitive directory.
- Block `Skill` usage when a required plugin isn't installed (mirrors v1's gstack gate).
- Require user confirmation for `Write` to specific paths.

These are all things v2's translator plugins can install when they own `.claude/settings.json`.

## v2 implementation #v2-decision

- **Framework doesn't own `.claude/settings.json`.** That's the `claude-translator` plugin's filesystem.
- **The `claude-translator` plugin** can install hooks at install time. It owns the schema for what hooks it supports.
- **Samuel exposes a hook-emission helper** — the framework knows how to format `{permissionDecision, message}` JSON. Translator plugins use the helper so the JSON shape is correct.
- **Each agent has its own hook system.** Codex has different surfaces. Cursor different again. Per-agent translator plugins own their tool-specific enforcement.

## Cross-agent gaps (what doesn't translate)

- **OpenAI Codex** has its own settings/agents.md format. No equivalent PreToolUse hook (yet).
- **Cursor** uses `.cursor/rules/*.md` for guidance but no programmable hook.
- **Continue / Cline / etc.** vary.

Samuel can't guarantee tool-boundary enforcement on every agent. The translator plugin for each agent does what it can.

## Open

- Hook authoring UX: should translator plugins ship hook scripts in Bash, Go, or WASM? Bash is universally available but tool-specific to Unix. Go binaries are portable but require a build step. WASM via wazero is portable + sandboxed. Suggest: framework provides a small hook-runner that can invoke any of the three; translator plugin author picks per script.
- Hook authoring API: a Go SDK for plugin authors writing v2 hooks. Helpers for path checks, env reads, JSON emission.
- Aggregation: when multiple plugins want to register hooks on the same tool (e.g., samuel-claude + a user-installed audit-plugin both gate `Write`), how do their JSON outputs merge? Last-deny-wins? All-must-allow? Open.

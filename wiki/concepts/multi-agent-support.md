---
title: Multi-agent support
type: concept
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-config-sync]
tags: [v1, rescue, v2-decision]
---

# Multi-agent support

v1 already wires Claude, Codex, Copilot, Gemini, and Kiro through one sandbox interface. This is the cross-tool primitive v2's thin core should ship — and v1's pattern is the right starting point.

## The v1 pattern (`docker.go`)

One abstraction (`DockerSandboxRunConfig.Agent`). Five agent values. Per-agent prompt translation in `GetAgentArgs`:

```go
switch aiTool {
case "claude":  -p <content> --dangerously-skip-permissions
case "codex":   --dangerously-bypass-approvals-and-sandbox exec <content>
case "amp":     --prompt-file <path>
default:        <path>
}
```

Each agent has a different CLI shape. Samuel normalizes upstream so the rest of the system speaks one language.

## What's good about it

- **Single source of truth** for prompts. The same content drives every agent.
- **Per-tool quirks isolated** in one switch. The orchestrator and CLI don't know about `--dangerously-skip-permissions`.
- **Env-var allowlist** is per-agent-friendly. `ANTHROPIC_API_KEY` for Claude, `OPENAI_API_KEY` for Codex, `AMP_API_KEY` for Amp — all forwarded conditionally.

## What needs to change for v2

- The switch in `GetAgentArgs` is going to grow as more agents land. Make it a registry.
- "Agent" should be a plugin kind (third-party agents — opencode, aider, anthropic-cli-plus — can ship as plugins).
- Capability model: each agent plugin declares what env vars it needs, what flags it understands, what its sandbox image is.

## v2 decision #v2-decision

- **Multi-agent is core.** Samuel is "Rails for coding assistants" plural by definition. The agent-adapter contract is part of the framework.
- **Built-in adapters first**: ship adapters for the agents v1 already supports (Claude, Codex, Copilot, Gemini, Kiro). They live in the binary so `samuel run claude` works out of the box.
- **External adapters as plugins**: new agents come as OCI plugins (since they wrap a binary) or as small Go-binary plugins (TBD whether we need a Go-binary kind alongside WASM/OCI).

## Agent adapter shape (sketch)

```toml
# A built-in adapter looks the same as a plugin one — just shipped in-binary.
name = "claude"
kind = "agent"
version = "1.0.0"

[agent]
binary = "claude"                    # what to exec
default_image = "node:lts"           # sandbox base
env_allowlist = ["ANTHROPIC_API_KEY", "AI_TOOL", "TERM"]
prompt_mode = "stdin-content"        # how to feed a prompt: stdin-content | file-arg | content-arg
prompt_arg = "-p"                    # flag for content-arg mode
extra_args = ["--dangerously-skip-permissions"]
```

The four `prompt_mode` values cover the four cases in `GetAgentArgs`. Plugins implement the same shape.

## Related

- [[entities/docker-sandbox]] — v1 implementation
- [[concepts/plugin-format]] — where agent plugins fit (OCI tier)
- [[synthesis/positioning-rails-for-coding-assistants]] — why this is core

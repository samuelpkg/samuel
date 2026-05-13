---
title: core/docker.go — v1 sandbox layer
type: entity
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-config-sync]
tags: [v1, sandbox, rescue]
---

# core/docker.go

The v1 sandbox layer. This is the seed of v2's [[concepts/plugin-format]] sandbox model — v1 already does the multi-agent + sandbox pattern.

## Three modes

- `SandboxNone` — run agent on host.
- `SandboxDocker` — run agent in a plain Docker container.
- `SandboxDockerSandbox` — use Docker Desktop's experimental Sandbox plugin (`docker sandbox run`). Auth, networking, and persistence managed by Docker.

## Multi-agent support #rescue

`DockerSandboxRunConfig.Agent` accepts:

- `claude` (default)
- `codex`
- `copilot`
- `gemini`
- `kiro`

Five agents already wired in v1. See [[concepts/multi-agent-support]].

## Env var allowlist

Only these forward into the sandbox container, only if set on host:

```
ANTHROPIC_API_KEY, OPENAI_API_KEY, AMP_API_KEY,
AI_TOOL, PAUSE_SECONDS, MAX_CONSECUTIVE_FAILURES, TERM
```

Clean — no accidental secret leakage from arbitrary env.

## Image validation

Regex: `^[a-zA-Z0-9][a-zA-Z0-9._\-/]*(:[a-zA-Z0-9._\-]+)?(@sha256:[a-f0-9]{64})?$`

Rejects shell metacharacters, absolute paths, leading dots. Defense against injection in `docker run <image>`.

Default image: `node:lts`. Mount point: `/workspace`.

## Per-agent prompt translation #rescue

`GetAgentArgs(aiTool, promptPath)`:

| Agent | Translation |
|---|---|
| `claude` | Read prompt file content → `-p <content> --dangerously-skip-permissions` |
| `codex` | Read prompt file content → `--dangerously-bypass-approvals-and-sandbox exec <content>` |
| `amp` | `--prompt-file <path>` |
| (default) | Pass `<path>` as positional |

This pattern — one prompt source, per-tool argument shape — is exactly the "cross-tool abstraction" v2's thin core should ship. Generalize beyond agents: any per-tool translation (prompts, config, output parsing) flows through this same shape.

## Availability checks

- `CheckDockerAvailable()` — `exec.LookPath("docker")` + `docker info` (5s timeout).
- `CheckDockerSandboxAvailable()` — `docker sandbox version`.

Both fail with actionable messages ("install Docker", "start Docker Desktop", "install Docker Desktop with Sandbox support").

## v2 implications

### `#rescue`

- Multi-agent abstraction (5 agents already supported).
- Env var allowlist pattern.
- Image validation regex.
- Per-agent prompt translation.
- Availability checks with actionable error messages.

### `#refactor`

- Generalize from "docker / docker-sandbox" to OCI runtime detection — match v2's `Podman → Docker → other` order.
- Hard-coded image (`node:lts`) should come from a plugin manifest, not config.
- The `AgentArgs` slice in `DockerSandboxRunConfig` is fine, but the per-agent logic in `GetAgentArgs` will grow — should become a registry of "agent adapter" plugins in v2.

### Bridge to v2 design

`docker.go` does what [[concepts/plugin-format]] calls "OCI plugin / coding-assistant execution" today, just narrower. v2 generalizes the same shape:

| v1 | v2 |
|---|---|
| Docker / Docker Sandbox | Any OCI runtime via detection (Podman → Docker → …) |
| Hardcoded image | Image declared in plugin manifest |
| 5 hardcoded agents | Agent adapters as plugins |
| Env var allowlist (constant) | Allowlist + capability declarations |
| `/workspace` mount | `/workspace`, `/skills`, `/plugin/config`, `/samuel-bridge` |

## Related

- [[concepts/plugin-format]]
- [[concepts/multi-agent-support]]

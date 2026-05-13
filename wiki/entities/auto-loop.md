---
title: auto_loop.go — RunAutoLoop + InvokeAgent
type: entity
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-auto-mode]
tags: [v1, auto-mode, rescue]
---

# auto_loop.go

The orchestration core of auto-mode.

## RunAutoLoop pseudocode

```
for i = 1..cfg.MaxIterations:
    prd = LoadAutoPRD(cfg.PRDPath)              # reload each iteration
    if prd.GetNextTask() == nil:
        OnIterEnd(i, nil); return nil           # all done
    
    OnIterStart(i, "implementation")
    
    PrepareProgressContext(cfg.ProjectDir)      # regen progress-context.md, rotate progress.md
    GenerateProjectSnapshot(cfg.ProjectDir)     # regen project-snapshot.md
    GenerateTaskContext(cfg.ProjectDir, prd, false)  # regen task-context.md
    
    err = InvokeAgent(cfg)
    if err:
        consecutiveFailures++
        OnIterEnd(i, err)
        if consecutiveFailures >= cfg.MaxConsecFails:
            return error("aborting after N failures")
    else:
        consecutiveFailures = 0
        OnIterEnd(i, nil)
    
    sleep cfg.PauseSecs
```

Notes:
- prd reloaded **every** iteration — the agent edits prd.json mid-loop, so cached state would go stale.
- Three pre-compute steps happen before the agent runs. Each writes a small md file the agent reads first thing.
- Failure backoff: 3 consecutive failures abort with `"Check AI tool auth/config"` hint.
- `OnIterStart` / `OnIterEnd` callbacks drive CLI progress UI.

## LoopConfig

```go
type LoopConfig struct {
  ProjectDir     string
  PRDPath        string
  PromptPath     string
  AITool         string
  MaxIterations  int
  Sandbox        string  // "none" | "docker" | "docker-sandbox"
  SandboxImage   string
  SandboxTpl     string
  PauseSecs      int     // default 2, env PAUSE_SECONDS
  MaxConsecFails int     // default 3, env MAX_CONSECUTIVE_FAILURES
  OnIterStart    func(iter int, iterType string)
  OnIterEnd      func(iter int, err error)
}
```

`NewLoopConfig(projectDir, prd)` derives from a loaded prd + env overrides.

## InvokeAgent

```
InvokeAgent(cfg):
    if not IsValidAITool(cfg.AITool):
        return error("refused to invoke invalid AI tool")
    switch cfg.Sandbox:
        case "docker-sandbox":  invokeAgentDockerSandbox(cfg)
        case "docker":          invokeAgentDocker(cfg)
        default:                invokeAgentLocal(cfg)
```

- **Allowlist check first.** This is the defense against `prd.json` injection — if an agent edits its own `ai_tool` to `"rm -rf /"`, the next `InvokeAgent` refuses. (`auto_loop.go:105-110`)
- Image regex validation before `docker run`. (`auto_loop.go:160-164`)
- `requiresPromptContent(aiTool)`: for `claude` / `codex`, the prompt file content is read on host and passed as a content arg (the file isn't in the container without an extra mount).

### Docker run arg shape (`buildDockerRunArgs`)

```
docker run --rm --init -i
    --user $UID:$GID
    -v <projectDir>:/workspace
    -w /workspace
    <-e KEY=VAL ...>          # from getAIToolEnvVars allowlist
    <image>
    <aiTool> <agentArgs...>
```

User mapping (`--user $UID:$GID`) ensures files written in the container are owned by the host user. Defaults: image `node:lts`, mount `/workspace`.

### Docker-sandbox path

Uses `docker sandbox run` (Docker Desktop's experimental Sandbox plugin). Configured via `DockerSandboxRunConfig`:

```go
DockerSandboxRunConfig{
    Agent:     cfg.AITool,
    WorkDir:   cfg.ProjectDir,
    Template:  cfg.SandboxTpl,
    AgentArgs: agentArgs,
}
```

Built into `docker sandbox run [--name X] [--template Y] <agent> <workdir> -- <agentArgs>` by `BuildDockerSandboxArgs` in [[entities/docker-sandbox]].

## v2 implications

### `#rescue`

- Loop shape (pre-compute → invoke → backoff → sleep).
- AI-tool allowlist check before exec.
- Image regex validation before exec.
- User mapping in `docker run` (`--user $UID:$GID`).
- Env-var allowlist forwarding.
- Consecutive-failure abort with actionable error message.
- `OnIterStart` / `OnIterEnd` callback hooks — exactly the kind of extension point v2 [[concepts/methodology-default-plus-plugin]] needs.

### `#refactor`

- The `switch cfg.Sandbox` dispatches three hardcoded modes. v2 replaces with [[concepts/plugin-format]] runtime detection.
- `IsValidAITool` checks a hardcoded list. v2 drives it from installed agent-adapter plugins.
- `requiresPromptContent` switch grows per agent. v2 moves to per-agent plugin metadata (`prompt_mode = "content-arg"`).

### `#open`

- Hook surface for v2 — exact names + payload shapes. Sketched in [[synthesis/auto-mode-v2-design]].

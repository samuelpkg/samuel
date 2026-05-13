---
title: Prompt template variables (v2)
type: concept
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-auto-mode]
tags: [v2, v2-decision, auto-mode]
---

# Prompt template variables

What Samuel exposes to the methodology prompt template engine (Go `text/template`).

## Goals

- Cover everything v1's prompts say (paths, config, guardrails) without hardcoding strings.
- Allow per-project overrides via `samuel.toml`.
- Allow plugins to inject named context (e.g. a `python-plugin` adds a `Python` block with pyproject details).
- Keep secrets and full file contents out.

## Top-level template context

```go
type PromptContext struct {
    Samuel       SamuelInfo          // framework identity
    Project      ProjectInfo         // current project
    Methodology  MethodologyInfo     // active workflow (e.g. "ralph")
    Iteration    IterationInfo       // this iteration's metadata
    Config       MethodologyConfig   // resolved methodology config
    Guardrails   GuardrailsConfig    // project-level guardrails
    Paths        PathsInfo           // runtime file paths
    State        StateSnapshot       // current task + counts (read-only)
    Mode         ModeInfo            // implementation vs discovery
    Hooks        HooksInfo           // registered plugin hooks
    Plugins      map[string]any      // plugin-contributed namespaces
}
```

## Sections in detail

### `Samuel`

```go
type SamuelInfo struct {
    Version  string  // "2.0.0"
    Binary   string  // /usr/local/bin/samuel
}
```

### `Project`

```go
type ProjectInfo struct {
    Name      string  // from samuel.toml or dir basename
    Root      string  // absolute project path
    Branch    string  // current git branch (or "" if not a git repo)
    Detected  []string // languages detected (from sync analyzer)
}
```

### `Methodology`

```go
type MethodologyInfo struct {
    Name   string  // "ralph", "tdd-strict", ...
    Source string  // "built-in" or "plugin:my-methodology"
}
```

### `Iteration`

```go
type IterationInfo struct {
    Number          int    // 1, 2, 3, ...
    Max             int    // MaxIterations
    Type            string // "implementation" | "discovery"
    LastDiscoveryAt int    // iteration number of last discovery (pilot)
}
```

### `Config`

The full resolved `[methodology.<name>]` block from `samuel.toml`, plus defaults:

```go
type MethodologyConfig struct {
    Agent           string
    MaxIterations   int
    PauseSecs       int
    MaxConsecFails  int
    QualityChecks   []string
    Pilot           PilotConfig    // optional, see below
    Context         ContextLimits  // pre-compute thresholds
}
```

### `Guardrails`

Promoted out of prompt-text literals. Sourced from `[methodology.<name>.guardrails]`:

```go
type GuardrailsConfig struct {
    MaxFunctionLines  int       // 50
    MaxFileLines      int       // 300
    RequireTests      bool      // true
    CommitConvention  string    // "conventional"
    BranchConvention  string    // optional
    Extra             []string  // freeform additional rules
}
```

Used in templates like:

```
Keep functions ≤{{.Guardrails.MaxFunctionLines}} lines,
files ≤{{.Guardrails.MaxFileLines}} lines.
{{if .Guardrails.RequireTests}}Write tests for all new code.{{end}}
```

### `Paths`

All paths agent might reference. Relative to project root.

```go
type PathsInfo struct {
    SamuelDir           string // .samuel/
    RunDir              string // .samuel/run/  (or .samuel/<methodology>/)
    PRDFile             string // .samuel/run/prd.json
    ProgressFile        string // .samuel/run/progress.md
    TaskContextFile     string // .samuel/run/task-context.md
    ProgressContextFile string // .samuel/run/progress-context.md
    SnapshotFile        string // .samuel/run/project-snapshot.md
    AgentsMD            string // AGENTS.md (canonical, framework-managed)
    // NOTE: Tool-specific paths like CLAUDE.md, .cursor/rules/, .codex/AGENTS.md
    // are NOT exposed by the framework. Translator plugins (claude-translator,
    // cursor-translator, etc.) own their own filesystem and add their paths
    // under ctx.Plugins["<plugin-name>"] if templates need them.
}
```

### `State` (read-only snapshot)

Computed by samuel from the current `prd.json`:

```go
type StateSnapshot struct {
    TotalTasks      int
    PendingTasks    int
    CompletedTasks  int
    InProgressTasks int
    BlockedTasks    int
    NextTask        *TaskBrief    // nil if none
    RecentCompleted []TaskBrief   // last 5
}

type TaskBrief struct {
    ID, Title, Priority, Complexity, Description string
    FilesToModify, FilesToCreate, DependsOn      []string
}
```

The full task list is NOT exposed — that's what `task-context.md` is for. Templates should reference `Paths.TaskContextFile` to tell the agent where to read it.

### `Mode`

```go
type ModeInfo struct {
    IsDiscovery   bool
    IsPilot       bool
    Focus         string  // optional, pilot mode: "testing"|"security"|...
    MaxNewTasks   int     // discovery iteration: pilot.MaxDiscoveryTasks
}
```

### `Hooks`

```go
type HooksInfo struct {
    Registered map[string][]string  // hook name → plugin names attached
    // example: {"quality.check": ["pytest-runner", "lint-strict"]}
}
```

Lets the template tell the agent which plugin-contributed checks will run.

### `Plugins` — namespaced contributions

Plugins can write arbitrary data under their name:

```go
ctx.Plugins["python"] = map[string]any{
    "Venv":         ".venv",
    "Requirements": "requirements.txt",
    "PyVersion":    "3.12",
}
```

Templates access as `{{.Plugins.python.Venv}}`. Convention: plugin namespace = plugin name.

## What's NOT exposed

- **Secrets / API keys.** Never. Env vars forwarded via the sandbox layer, not template.
- **Full file contents.** That defeats [[concepts/pre-computed-context]]. Templates point at paths; agent reads files.
- **Other plugins' raw config.** Plugin namespaces are populated by the plugin itself, not by reading peer plugin config.
- **Git history / diff content.** Pointers (paths, branch name) only; agent runs `git` if needed.

## Template helpers

Beyond Go `text/template` built-ins, add a small standard library:

```
{{join .Config.QualityChecks "\n"}}       — join list with separator
{{indent 2 .SomeBlock}}                    — prefix lines with N spaces
{{relpath .Paths.PRDFile}}                 — show path relative to project root
{{if hasPlugin "python"}}...{{end}}        — gate sections on plugin presence
{{commitConvention .Guardrails}}           — render commit format guidance
```

## Override layers

Resolution order, last wins:

1. Built-in template shipped in Samuel binary (`go:embed`).
2. Per-project override: `.samuel/templates/<methodology>/<name>.md.tmpl`.
3. Plugin override: a methodology-enhancement plugin can replace.

Same shape, same variables, last write wins.

## Open

- Should plugin namespaces be late-bound (populated by the plugin's own pre-compute hook) or pre-declared (plugin manifest declares schema)? Late-bound is more flexible; pre-declared is more debuggable. Suggest: late-bound, with a `samuel methodology print-vars` debug command that dumps the live context.
- Localization — should prompt templates support multiple languages? Defer.
- Versioning — when v2.1 adds a new variable, do old templates break? They shouldn't — Go templates ignore unused fields. Adding fields is safe. Removing/renaming fields is a major version bump.

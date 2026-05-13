---
title: Auto-mode v2 design
type: synthesis
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-auto-mode]
tags: [v2, v2-decision, auto-mode]
---

# Auto-mode v2 design

Synthesis of pass-3 findings + the user's "default + plugin enhancement" decision. This is the first per-workflow application of [[concepts/methodology-default-plus-plugin]].

## CLI rename #v2-decision

`samuel auto` → **`samuel run [methodology]`**.

- Default methodology configured in `samuel.toml`. Out of the box: `ralph` (the Ralph Wiggum loop).
- `samuel run` (no arg) — run the default methodology.
- `samuel run ralph` — explicit methodology selection. Aliases: `rw`.
- `samuel run --discover-only` — built-in discovery-only mode (pilot mode with no implementation iterations).
- `samuel run <my-methodology-plugin>` — run a plugin-provided methodology.

Reconcile in pass 6: v1 already has a `samuel run` command. Likely it'll either repurpose or get replaced.

## Runtime directory rename

`.claude/auto/` → `.samuel/run/` (or `.samuel/<methodology>/` if multiple methodologies are active in one project).

Clean-break decision means the rename is OK. Coordinate across: generated prompts, runtime file references, docs, samuel.toml path defaults.

## The thesis

Auto-mode is **THE flagship Samuel methodology**. Built in, opinionated, hookable. Plugins enhance individual stages without replacing the workflow.

## Encoding (resolved 2026-05-12) #v2-decision

- **Structured runtime files use TOON** (`prd.toon`, `task-context.toon`, `project-snapshot.toon`). Token-efficient, malformation-tolerant on per-row basis. Per [[concepts/toon-evaluation]].
- **Append-only logs stay markdown** (`progress.md`, `progress-context.md`).
- **Agent never writes prd.toon directly.** Mutations go through CLI subcommands (`samuel run done`, `skip`, `reset`, `enqueue`). This decouples storage encoding from agent contract — load-bearing change.

## What's built in #v2-decision

- **The loop** ([[entities/auto-loop]]) — `RunAutoLoop` shape, pre-compute → invoke → backoff → sleep.
- **The prd.json data model** ([[entities/auto-prd]]) — atomic save, validation, dependency-aware task selection.
- **The seven runtime files** ([[entities/auto-runtime-files]]) — layout, owner split.
- **The two prompt templates** ([[entities/auto-prompts]]) — implementation + discovery, externalized as `*.tmpl` (still in-binary).
- **The three pre-compute generators** ([[concepts/pre-computed-context]]) — task, progress, snapshot.
- **Pilot mode triggers** ([[concepts/pilot-mode]]) — `ShouldRunDiscovery` logic.
- **The Ralph Wiggum methodology** ([[concepts/ralph-wiggum-methodology]]) cited as the conceptual basis.
- **Safety primitives** — AI-tool allowlist, image regex, atomic save, consecutive-failure abort, progress rotation.
- **Built-in agent adapters** — Claude, Codex, Copilot, Gemini, Kiro per [[concepts/multi-agent-support]].

## Hook points (proposed)

The framework calls these by name. Plugins register handlers in their manifest.

```
samuel auto
  │
  ├─ before:loop
  │   ├─ before:iteration              ← each iteration
  │   │
  │   ├─ iteration.gate                ← decide impl vs discovery (default: ShouldRunDiscovery)
  │   │
  │   ├─ context.snapshot              ← regen project-snapshot.md  (default + plugins augment)
  │   ├─ context.progress              ← regen progress-context.md  (default + plugins augment)
  │   ├─ context.task                  ← regen task-context.md      (default + plugins augment)
  │   ├─ context.extra                 ← plugin-only: extra pre-compute files
  │   │
  │   ├─ before:agent.invoke
  │   ├─ agent.invoke                  ← run the agent (default: built-in adapter; plugin can replace)
  │   ├─ after:agent.invoke
  │   │
  │   ├─ quality.check                 ← default: run config.quality_checks; plugins add checks
  │   │
  │   └─ after:iteration
  │
  └─ after:loop                        ← cleanup, summary, notifications (plugins for slack/PR/etc.)
```

Multiple plugins can attach to one hook. Default ordering: registration order; override via `samuel.toml` `[hooks.<name>.order]`.

## Where samuel.toml comes in

```toml
# Project-level methodology config (key matches `samuel run <name>`)
default_methodology = "ralph"

[methodology.ralph]
enabled = true
agent = "claude"               # built-in adapter (claude/codex/copilot/gemini/kiro) or plugin name
max_iterations = 25
pause_secs = 2
max_consec_fails = 3

[methodology.ralph.quality]
checks = ["go build ./...", "go test ./...", "go vet ./..."]

[methodology.ralph.pilot]
enabled = true
discover_interval = 5
max_discovery_tasks = 10
focus = "testing"              # optional

[methodology.ralph.guardrails]
max_function_lines = 50
max_file_lines = 300
require_tests = true

[methodology.ralph.context]
max_learnings = 50
max_completed = 10
max_progress_lines = 500
```

Guardrails moved out of prompt text and into config. Rendered into the prompt template at iteration time.

## Plugin shapes for auto-mode

### Hook plugins (most common)

Small WASM modules that attach to one or more hooks. Example: a `quality-check-pytest` plugin attaches to `quality.check` and adds `pytest -xvs`. A `notify-slack` plugin attaches to `after:loop` and posts a summary.

### Agent adapter plugins

Replace or add the agent at `agent.invoke`. Example: `aider-adapter` for aider, `opencode-adapter` for opencode.

### Methodology variants

Plugins that override multiple hooks coherently to provide a different flavor of auto-mode. Example: a `tdd-strict` plugin that overrides the implementation prompt + adds `tdd.red`/`tdd.green` hooks to enforce red-green-refactor cycles.

## Migration from v1

Clean break — no migration tool. But users with existing `.claude/auto/prd.json` files: the schema is unchanged (we keep JSON shape), so v2 reads v1's prd.json untouched. Just move the directory if we rename to `.samuel/auto/` (open question in [[entities/auto-runtime-files]]).

## Open

- Hook handler ABI — JSON-over-stdio for WASM is straightforward, but how do built-in hooks (Go) compose with plugin hooks (WASM) in one chain? Likely: Go runtime adapts WASM call-out as one node in a chain of Go funcs.
- Per-language plugin convention — should there be a recommended skeleton for "language-specific methodology plugin" (adds context generator + quality checks)?
- Cost surfacing — should samuel emit per-iteration token/cost estimates? Probably yes for at least one mode (Claude has a cost endpoint; OpenAI does too).

## Prompt templates

See [[concepts/prompt-template-variables]] for the full spec of what's exposed to the template engine. The built-in `ralph` methodology ships two templates (`prompt.md.tmpl` and `discovery-prompt.md.tmpl`) which can be overridden by per-project files at `.samuel/templates/ralph/` or by a methodology-enhancement plugin.

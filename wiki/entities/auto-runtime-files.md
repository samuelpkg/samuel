---
title: .claude/auto/ runtime files
type: entity
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-auto-mode]
tags: [v1, auto-mode]
---

# .claude/auto/ → .samuel/run/

The runtime workspace v1 maintains in every project that uses it. Seven files.

**v1 layout** (current shipped):

```
.claude/auto/
├── prd.json                  ← source of truth (machine-readable task state)
├── prompt.md                 ← implementation iteration prompt template
├── discovery-prompt.md       ← discovery iteration prompt template (pilot mode only)
├── progress.md               ← append-only history log (rotated when > 500 lines)
├── progress-context.md       ← compact progress summary (auto-regen each iteration)
├── task-context.md           ← compact current-task brief (auto-regen each iteration)
└── project-snapshot.md       ← compact project overview (auto-regen each iteration)
```

**v2 layout** (per [[concepts/toon-evaluation]] and AGENTS.md-primary decisions):

```
.samuel/run/
├── prd.toon                  ← TOON (was prd.json) — tabular tasks, malformation-tolerant
├── progress.md               ← markdown (unchanged) — append-only, prose-heavy
├── progress-context.md       ← markdown (unchanged) — prose summary
├── task-context.toon         ← TOON (was task-context.md) — tabular when summary, prose when single-task
└── project-snapshot.toon     ← TOON (was project-snapshot.md) — file inventory + TODO counts are pure tabular data
```

Prompt templates (`prompt.md.tmpl`, `discovery-prompt.md.tmpl`) move into the framework binary as Go templates ([[entities/auto-prompts]]) — not on disk in the runtime directory.

## File responsibilities

| File | Lifetime | Writer | Reader |
|---|---|---|---|
| `prd.json` | persistent | samuel (init/update) + AI agent | samuel + AI agent |
| `prompt.md` | persistent | samuel (init) | AI agent |
| `discovery-prompt.md` | persistent | samuel (init, pilot only) | AI agent |
| `progress.md` | persistent (append-only, rotated) | AI agent | samuel + AI agent |
| `progress-context.md` | regenerated each iteration | samuel | AI agent |
| `task-context.md` | regenerated each iteration | samuel | AI agent |
| `project-snapshot.md` | regenerated each iteration | samuel | AI agent |

The split is intentional: samuel owns the **pre-computed context** (top three regenerated files), AI agent owns **state mutations** (`prd.json` + `progress.md` appends).

## Why the pre-compute split matters

See [[concepts/pre-computed-context]]. The short version: regenerating a small task-context.md before each iteration saves the agent from scanning the full prd.json. Same logic for progress and project state. Token discipline is enforced by the *files the agent is told to read first*.

## v2 implications

### `#rescue`

- The seven-file layout.
- The owner split (samuel pre-computes, agent mutates).
- Filenames and conventions.

### `#refactor`

- Make the directory configurable (`.samuel/auto/`?) so it's clearly Samuel's namespace, not generic Claude.
- Generalize so per-workflow methodologies each get their own subdir: `.samuel/auto/`, `.samuel/code-review/`, `.samuel/rfd/`, etc.

### `#open`

- Path convention — keep `.claude/auto/` for back-compat with v1 / familiarity, or migrate to `.samuel/...`? Clean-break decision says we can move. But the agent prompts reference these paths, so it's a coordinated rename across all generated prompts + runtime files + docs.

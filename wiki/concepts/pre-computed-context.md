---
title: Pre-computed context (token discipline)
type: concept
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-auto-mode]
tags: [v1, methodology, rescue, innovation]
---

# Pre-computed context

v1's most interesting design choice. Worth keeping front-and-center for v2.

## The problem

AI agents waste tokens re-discovering things they could have been handed. They `ls` the project. They `grep` for TODOs. They re-read prd.json to find their task. They re-read the same files every iteration. Each of those is several hundred tokens for information that doesn't change much.

## The pattern

Before each iteration, samuel regenerates three small files:

- **`task-context.md`** — the current task brief, or a summary table in discovery mode
- **`progress-context.md`** — recent learnings + completions + already-explored files
- **`project-snapshot.md`** — directory listing, test gaps, large files, TODO counts, recent git log

The agent's prompt **instructs** it to read these three files **first** and **not** to re-derive any of it. The prompt actually carries the budget:

> "Read at most 10 source files per discovery iteration."
> "Do NOT manually scan the directory tree, run find/ls, or grep for TODOs."
> "Skip files in 'Areas Already Analyzed' unless git log shows recent changes."

Compliance isn't automatic — but a competent agent following written instructions does follow these. The cost savings are real.

## Why it works

- **Cheap to compute.** Samuel walks the project once (Go, milliseconds). The agent doesn't have to do it.
- **Compact.** The three files together are typically <5KB. Compare to the agent grepping 30 files = 30 file-read tool calls × N tokens each.
- **Coherent.** All three files reflect state-as-of-iteration-start. The agent has a consistent view.
- **Self-improving.** As the agent runs, `progress-context.md` accumulates learnings + explored areas. Each new iteration starts from a smarter baseline.

## The reverse pattern: what the agent owns

Symmetric: samuel does NOT pre-compute mutation surfaces.

- `prd.json` — the agent updates task status, adds tasks (in pilot mode).
- `progress.md` — the agent appends learnings.

If samuel pre-computed these, the agent's edits would race with samuel's regeneration. Clean separation: samuel owns *reads* (compute compact views), agent owns *writes* (mutate canonical state).

## What v1 pre-computes — full inventory

| File | Inputs | Output |
|---|---|---|
| `task-context.md` | prd.json | current task brief (impl) or task summary table + covered-files list (discovery) |
| `progress-context.md` | progress.md (tail 500 lines) | summary line + 50 learnings + 10 completions + deduped EXPLORED paths |
| `project-snapshot.md` | filesystem walk (max 200 source files) + `git log` | file inventory + test gaps + 15 largest + top 30 TODO counts + last 10 git log lines |

## v2 decision #v2-decision

- **Pre-computed context is part of the framework.** Built-in, not optional. The Samuel Way assumes token discipline.
- **The three generators stay.** They're the right cuts.
- **Hook points for plugins.** Plugins can add additional pre-compute steps (e.g., a `python-plugin` could add a `python-context.md` listing virtualenvs, requirements, recent pip changes). Plugins can also replace a default generator (e.g. a `git-deep` plugin replacing the snapshot's git section with `git log --graph --decorate`).

## Hook points (sketch)

```
samuel auto
  ├─ context.task          ← default: GenerateTaskContext; plugin can replace
  ├─ context.progress      ← default: GenerateProgressContext + RotateProgressIfNeeded
  ├─ context.snapshot      ← default: GenerateProjectSnapshot
  ├─ context.extra         ← plugin slot: add more pre-compute files
  └─ iteration.invoke      ← run the agent
```

Each `context.*` is invoked before `iteration.invoke`. Plugins register handlers that read prior state and write into `.samuel/auto/`. If a plugin's handler errors, the iteration continues with stale or missing context (logged as a warning) — don't fail the loop over a pre-compute issue.

## Open

- Caching: if a generator produces identical output to the prior iteration, skip the write? (Cheap to check by content hash; saves IDE-watcher noise.)
- Generator ordering — currently serial; could parallelize. Probably not worth it; they're already fast.
- Per-plugin `context.extra` files should namespace under `.samuel/auto/plugins/<plugin-name>/<file>.md` to avoid name collisions.

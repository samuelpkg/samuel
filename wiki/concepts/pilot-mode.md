---
title: Pilot mode (discovery + implementation alternation)
type: concept
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-auto-mode]
tags: [v1, auto-mode, rescue]
---

# Pilot mode

Auto-mode runs in one of two flavors:

1. **Implementation mode** — prd starts with predefined tasks (from PRD markdown + task list). Loop just executes them.
2. **Pilot mode** — prd starts empty. The loop alternates between *discovery iterations* (LLM scans project, generates new tasks) and *implementation iterations* (LLM works tasks).

## The trigger

`ShouldRunDiscovery(prd, currentIter, lastDiscoveryIter, discoverInterval)` returns true when any of:

- No pending tasks → run discovery to seed work
- First iteration of the loop → run discovery to bootstrap
- `currentIter - lastDiscoveryIter >= discoverInterval` → periodic refresh (default every 5 iterations)
- Pending count < `MinPendingTasksForDiscovery` (default 2) → preemptive refill so the loop doesn't starve

## Discovery iteration

Uses the discovery prompt ([[entities/auto-prompts]]). Tells the agent:

- **No code changes.** Only update prd.json + progress.md.
- Read pre-computed context only. Read at most 10 source files.
- Prioritize: security > tests > quality > docs > perf > refactor.
- Each new task: `source: "pilot-discovery"`, atomic (≤5 files).

Configurable via `PilotConfig`:

- `DiscoverInterval` — iterations between discoveries (default 5)
- `MaxDiscoveryTasks` — cap per discovery run (default 10)
- `Focus` — optional area: `testing`, `docs`, `security`, `performance`, `refactoring`

`Focus` injects a focused prompt section that biases the agent toward that domain.

## Implementation iteration

Standard auto-mode prompt. Picks next available task, implements it, commits.

## Why both modes exist

- **Implementation mode** = you have a PRD, you wrote it, you want it built.
- **Pilot mode** = you have a codebase, you want it improved. Let the agent find the work.

Pilot mode is what makes Samuel a "code health" tool, not just a PRD executor.

## v2 decision #v2-decision

- **Both modes survive.** They're the right two shapes.
- **Triggering logic stays.** `ShouldRunDiscovery` is well-designed.
- **Pilot config becomes a plugin-friendly hook point** — plugins can:
  - register custom focus areas (e.g. an `accessibility-audit` plugin adds `focus = "a11y"`)
  - override `ShouldRunDiscovery` (e.g. trigger discovery on git push, on test failures)
  - supplement the discovery prompt with domain-specific instructions

## Open

- Should "discovery only" mode (`samuel auto --discover-only`) ship as a built-in? Useful for "scan my codebase and tell me what's broken" without commitment. Probably yes — small, clearly useful.
- Empty-discovery guard: v1 has `MaxEmptyDiscoveries = 2`. If two discoveries in a row find nothing, should the loop stop or just exit pilot mode? Currently unused in the code I read — confirm in later passes.

---
title: Ralph Wiggum methodology
type: concept
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-auto-mode]
tags: [v1, methodology, external-influence]
---

# Ralph Wiggum methodology

External methodology that v1's auto-mode is built on. Named for the [Simpsons character](https://en.wikipedia.org/wiki/Ralph_Wiggum) by its originator, Geoffrey Huntley.

## What it is

Autonomous coding pattern: instead of one long agent run with growing context, run **many independent iterations** with fresh context each time. Each iteration picks the highest-priority task, implements it, commits, and exits. The next iteration starts cold.

Origin: [ghuntley.com/ralph](https://ghuntley.com/ralph/). The core insight is that context-window growth is the enemy of long agent runs — you get drift, hallucination, and quadratic cost. Forcing a fresh context every N minutes resets the problem.

## How v1 implements it

- Loop in [[entities/auto-loop]] iterates `MaxIterations` times.
- Each iteration is one agent invocation with no shared state in memory.
- All persistent state lives on disk: [[entities/auto-prd]] (task state), [[entities/auto-runtime-files]] (context + progress).
- Pre-computed context files ([[concepts/pre-computed-context]]) compensate for the no-memory model — the agent re-reads a compact snapshot each time.

The auto-mode prompt template explicitly references the methodology:

> "You are running in autonomous mode as part of the Ralph Wiggum methodology. Each iteration is independent — you start with a fresh context window." ([[entities/auto-prompts]])

## Why it works

- **Bounded context per iteration.** Token cost stays linear, not quadratic.
- **Forced re-grounding.** The agent re-reads guardrails, current task, prior learnings each time. Drift is corrected by the disk artifacts.
- **Failure isolation.** A bad iteration commits nothing or commits a bad atomic change. The next iteration starts clean.
- **Token discipline.** "Read these files first, don't explore" budget rules force focused work.

## v2 decision #v2-decision

- **The methodology stays as the default auto-mode.** It's the Samuel Way.
- **Naming.** Internally we keep "Ralph Wiggum" as the methodology citation (acknowledge the source). Externally, the user-facing name is just `samuel auto` / "autopilot" — "Ralph Wiggum" is a niche reference and doesn't belong in marketing surface.
- **Built-in core capability.** Per [[concepts/methodology-default-plus-plugin]], auto-mode ships in the framework. Plugins enhance individual hooks (custom snapshot strategies, alternative discovery prompts, etc.).

## Open

- Should we cite Huntley's blog post in user-facing docs (README, mkdocs)? Probably yes — credit the methodology, point readers at the origin for deeper reading.
- Per-language Ralph variants? E.g. a Python plugin that knows to run pytest as the quality check + read `__init__.py` files first. Likely as plugins, not core.

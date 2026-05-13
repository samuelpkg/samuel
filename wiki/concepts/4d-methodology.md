---
title: 4D Methodology (Deconstruct / Diagnose / Develop / Deliver)
type: concept
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-template-docs]
tags: [v1, methodology, rescue, samuel-way]
---

# 4D Methodology

Samuel's task-execution framework. Surfaces in the v1 CLAUDE.md template ([[sources/2026-05-12-v1-template-docs]]) and the `docs/core/methodology.md` page. Part of the Samuel Way — `#rescue` for v2.

## The four phases

| Phase | Purpose | Output |
|---|---|---|
| **Deconstruct** | Break down the task | Clear scope, subtasks |
| **Diagnose** | Identify risks | Dependencies, integration points |
| **Develop** | Implement with tests | Working code, passing tests |
| **Deliver** | Validate and commit | Production-ready code |

Every task runs through these four steps regardless of complexity. The mode below scales how thoroughly each step is applied.

## Three modes

Auto-detected by task scope. The agent picks mode based on file count, complexity, and ambiguity.

### ATOMIC (default)

Single-file changes, bug fixes, small features. <5 files affected, clear scope.

- **Deconstruct**: What's the minimal change?
- **Diagnose**: Will this break anything? Check dependencies.
- **Develop**: Make the change with tests.
- **Deliver**: Run tests + check guardrails + commit.

### FEATURE

Multi-file features, new components, API endpoints. 5-10 files.

- **Deconstruct**: Break into 3-5 atomic subtasks.
- **Diagnose**: Identify integration points and dependencies.
- **Develop**: Implement subtasks sequentially with tests.
- **Deliver**: Integration test + documentation + review + commit.

### COMPLEX

Architecture changes, major refactors, new systems. >10 files or new subsystem.

- **Deconstruct**: Full decomposition into phases/milestones.
- **Diagnose**: Analyze risks, dependencies, migration paths.
- **Develop**: Plan → execute incrementally.
- **Deliver**: Staged rollout + documentation + retrospective.

**MANDATORY workflow for COMPLEX**:

1. Use create-prd skill → write the PRD.
2. Use generate-tasks skill → break the PRD into a task list.
3. Implement tasks one by one with verification.
4. (Optional) Convert to `prd.json` → run `samuel run start` for autonomous execution.

## Escalation triggers

- Task affects >5 files → FEATURE mode
- Task affects >10 files → COMPLEX mode (consider PRD workflow)
- Task affects >15 files OR new subsystem → COMPLEX mode (PRD workflow **mandatory**)
- Task unclear/ambiguous → ask user first

## Why it works

- **Scales the same shape from 1-file to 1000-file work.** Same four steps every time.
- **Forces planning before code.** Deconstruct + Diagnose are 50% of the methodology.
- **Built-in escalation.** The agent can't just keep doing ATOMIC indefinitely; thresholds force a methodology switch.
- **Hooks into other skills.** COMPLEX mode triggers create-prd → generate-tasks → optionally `samuel run`.

This is the methodology that **threads through every Samuel workflow.** create-rfd is "Deconstruct + Diagnose for a decision." create-prd is "Deconstruct + Diagnose for a feature." auto-mode is "Develop + Deliver, on autopilot." code-review is the "Deliver" check.

## v2 decision #v2-decision

- **Built into the framework.** Not a plugin. The 4D framework is the Samuel Way's spine.
- **Render into AGENTS.md template.** ~50 lines (vs v1's bloated 474-line template).
- **Mode auto-detection runs in `samuel run`.** The framework can suggest mode escalation based on file-count heuristics from the task definition.
- **PRD workflow plugins (create-rfd, create-prd, generate-tasks)** are starter-pack plugins per [[synthesis/v2-skill-migration-plan]] — they implement the COMPLEX-mode entry point.

## v2 schema in samuel.toml

```toml
[methodology.4d]
auto_detect_mode = true            # framework picks mode based on heuristics
atomic_max_files = 5
feature_max_files = 10
complex_workflow = "prd"           # which workflow handles COMPLEX entry (default: create-prd plugin)
```

User can disable auto-detection (`auto_detect_mode = false`) and pick mode explicitly per task.

## Open

- **Auto-detection heuristics.** v1 doesn't actually auto-detect — the human picks. v2 could (file-count from task definition, glob patterns in `files_to_modify`). Probably worth doing, with the user able to override.
- **Mode reporting.** Should the agent emit `[mode=ATOMIC]` in commit messages or progress logs? Useful for retrospectives.
- **PRD/RFD plugins as required dependencies of `samuel run`.** If a task is COMPLEX and no PRD exists, should `samuel run start` refuse to proceed? Tentative: yes, with a clear error pointing to `samuel run convert <prd-path>`.

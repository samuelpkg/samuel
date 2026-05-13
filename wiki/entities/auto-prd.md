---
title: AutoPRD (prd.json data model)
type: entity
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-auto-mode]
tags: [v1, auto-mode, rescue]
---

# AutoPRD

The machine-readable task state for the autonomous loop. Persisted as `.claude/auto/prd.json`. Schema version `1.0`.

## Shape

```jsonc
{
  "version": "1.0",
  "project": {
    "name": "feature-name",
    "description": "...",
    "source_prd": "docs/prd/feature.md",        // optional
    "created_at": "2026-05-12T10:00:00Z",
    "updated_at": "2026-05-12T10:23:45Z"
  },
  "config": {
    "max_iterations": 50,
    "quality_checks": ["go test ./...", "go vet ./..."],
    "ai_tool": "claude",
    "ai_prompt_file": ".claude/auto/prompt.md",
    "sandbox": "none",                          // none | docker | docker-sandbox
    "sandbox_image": "",                        // optional
    "sandbox_template": "",                     // optional
    "pilot_mode": false,
    "pilot_config": null,                       // see PilotConfig
    "discovery_prompt_file": "",                // pilot mode only
    "progress_max_learnings": 50,
    "progress_max_completed": 10,
    "progress_max_lines": 500
  },
  "tasks": [
    {
      "id": "1",
      "title": "Create user schema",
      "description": "...",
      "status": "pending",                      // pending | in_progress | completed | skipped | blocked
      "priority": "high",                       // critical | high | medium | low
      "complexity": "medium",                   // simple | medium | complex
      "parent_id": "",                          // optional, for sub-tasks
      "depends_on": [],
      "files_to_create": ["internal/user/schema.go"],
      "files_to_modify": ["internal/db/migrations.go"],
      "guardrails": ["maintain backward compat with v1 sessions"],
      "completed_at": "",
      "commit_sha": "",
      "iteration": 0,
      "source": "manual"                        // manual | prd | pilot-discovery
    }
  ],
  "progress": {
    "total_tasks": 12,
    "completed_tasks": 3,
    "current_iteration": 4,
    "total_iterations_run": 4,
    "last_iteration_at": "...",
    "status": "running",                        // not_started | running | paused | completed | failed
    "discovery_iterations": 1,                  // pilot mode
    "impl_iterations": 3                        // pilot mode
  }
}
```

## Operations

- `LoadAutoPRD(path) *AutoPRD` — read + unmarshal.
- `(*AutoPRD).Save(path)` — **atomic** write via `<path>.tmp` + rename. Bumps `updated_at`. Calls `RecalculateProgress` automatically.
- `NewAutoPRD(name, desc)` — defaults: max_iterations=50, claude, sandbox=none, go quality checks.
- `InitPilotPRD(projectDir, config, pilot)` — pilot-mode preset.
- `GetNextTask()` — highest-priority pending task whose deps all completed; ID-asc tiebreak.
- `AddTask`, `CompleteTask`, `SkipTask`, `ResetTask` — state transitions.
- `ValidateAutoPRD` — version, project.name, task ID uniqueness, status validity, dependency references.
- `NextAvailableID` — returns `max(integer-prefix)+1` as string; supports dotted IDs without bumping (e.g. `"2.1"` doesn't increment the counter).

## AI-output resilience

`AutoTask.UnmarshalJSON` handles numeric IDs: if AI emits `"id": 1`, it converts to `"id": "1"`. Without this, ~10% of AI outputs would fail parsing. (`auto.go:137-179`)

## v2 implications #rescue

Keep the data model verbatim. The state machine is sound. Three refinements:

- **Move to TOON encoding** (`prd.toon`, not `prd.json`) per [[concepts/toon-evaluation]]. User-observed JSON fragility (missing commas, unescaped quotes) outweighs the theoretical AI-emit-JSON advantage. TOON's tabular row format is malformation-tolerant — a bad row affects one task, not the file.
- **Agent stops writing prd directly.** Mutations go through CLI subcommands (`samuel run done|skip|reset|enqueue`). Samuel CLI is the only writer. Decouples storage format from agent contract. This is the load-bearing change.
- **Generalize numeric-ID coercion as a util** — every AI-emitted struct will want it. Same pattern applies whether the encoding is JSON or TOON.

# Tasks — PRD 0004: Samuel v2 Methodology

> Generated from [0004-prd-methodology.md](0004-prd-methodology.md) on 2026-05-12.
> Depends on PRD 0003 (Plugin Loader) being complete.

## Relevant files

- `samuel_v1/internal/core/auto*.go` (11 files) — primary source for port
- `samuel_v1/internal/core/docker.go` — multi-agent sandbox pattern to generalize
- `samuel_v1/internal/commands/auto*.go` — CLI surface to port
- `.wiki/synthesis/auto-mode-v2-design.md`, `.wiki/concepts/methodology-default-plus-plugin.md`, `.wiki/concepts/pre-computed-context.md`, `.wiki/concepts/pilot-mode.md`, `.wiki/concepts/prompt-template-variables.md` — design references
- RFD 0004, 0006 (Committed) — design contract

## Tasks

- [ ] 1.0 Hook framework [~6,000 tokens - Complex]
  - [ ] 1.1 Define `HookName` constants — 13 hooks per RFD 0004 (before/after loop, iteration, gate, context.{snapshot,progress,task,extra}, before/agent.invoke/after, quality.check)
  - [ ] 1.2 Define `Hook` interface + `HookInput`/`HookOutput` structs
  - [ ] 1.3 Implement `Registry` struct with ordered handler chain per hook
  - [ ] 1.4 Order resolution: built-in default at order 100, plugins default to 200, samuel.toml overrides
  - [ ] 1.5 Per-hook `strict` config — true aborts iteration on handler error; false logs warning
  - [ ] 1.6 Default strict: `quality.check` + `before:loop`; default non-strict for others
  - [ ] 1.7 Per-hook timeout (default 5 min, configurable in samuel.toml)
  - [ ] 1.8 Hook handler error → structured `[hooks.warning]` entry to progress.md
  - [ ] 1.9 `--profile` flag emits `[hooks.timing]` entries
  - [ ] 1.10 Capability check before invoking handler (per RFD 0004 resolution #6)
  - [ ] 1.11 Tests: chain composition, order override, strict failure, timeout

- [ ] 2.0 AutoPRD data model port [~3,000 tokens - Medium]
  - [ ] 2.1 Port `AutoPRD`, `AutoProject`, `AutoConfig`, `PilotConfig`, `AutoTask`, `AutoProgress` from `samuel_v1/internal/core/auto.go`
  - [ ] 2.2 Keep `AutoTask.UnmarshalJSON` numeric-ID coercion pattern (now `UnmarshalTOON` for TOON path)
  - [ ] 2.3 Status / Priority / Complexity / Source / LoopStatus enums
  - [ ] 2.4 `NewAutoPRD`, `NextAvailableID`, `GetNextTask`, `CompleteTask`, `SkipTask`, `ResetTask`, `AddTask`, `ValidateAutoPRD` — port from auto_tasks.go
  - [ ] 2.5 Atomic save (write-tmp-then-rename)

- [ ] 3.0 TOON runtime file I/O [~4,500 tokens - Medium]
  - [ ] 3.1 Implement `internal/methodology/ralph/prd.go` — Load/Save for prd.toon using `internal/encoding/toon`
  - [ ] 3.2 Per-row malformation recovery integrated with AutoPRD load
  - [ ] 3.3 Version header on write (`# toon v3`)
  - [ ] 3.4 task-context.toon writer
  - [ ] 3.5 project-snapshot.toon writer
  - [ ] 3.6 progress.md and progress-context.md stay markdown (unchanged)

- [ ] 4.0 Pre-computed context generators (port from v1) [~5,500 tokens - Complex]
  - [ ] 4.1 Port `GenerateProjectSnapshot` from `auto_snapshot.go` → `project-snapshot.toon`
  - [ ] 4.2 Port file inventory walk, test gap detection, large file ranking, TODO/FIXME/HACK counts
  - [ ] 4.3 Port `recentGitLog` helper
  - [ ] 4.4 Port `GenerateProgressContext` from `auto_progress_context.go` → `progress-context.md` (markdown preserved)
  - [ ] 4.5 Port `RotateProgressIfNeeded` (500-line threshold; configurable)
  - [ ] 4.6 Port `GenerateTaskContext` from `auto_task_context.go` → `task-context.toon`
  - [ ] 4.7 Discovery-mode summary table vs implementation-mode current-task detail
  - [ ] 4.8 Each generator registers as default handler for its hook (`context.snapshot`, `context.progress`, `context.task`)

- [ ] 5.0 Ralph loop driver [~5,500 tokens - Complex]
  - [ ] 5.1 Port `RunAutoLoop` from `auto_loop.go` → `internal/methodology/ralph/loop.go`
  - [ ] 5.2 Reload prd.toon every iteration (agent may have mutated via CLI)
  - [ ] 5.3 Fire hooks at the 13 lifecycle points
  - [ ] 5.4 `MaxConsecFails` abort (default 3, env `MAX_CONSECUTIVE_FAILURES`)
  - [ ] 5.5 PauseSecs between iterations (default 2, env `PAUSE_SECONDS`)
  - [ ] 5.6 `OnIterStart` / `OnIterEnd` callbacks for CLI UI

- [ ] 6.0 Pilot mode [~3,500 tokens - Medium]
  - [ ] 6.1 Port `PilotConfig`, `NewPilotConfig`, `ShouldRunDiscovery`, `CountPendingTasks`, `InitPilotPRD` from `auto_pilot.go`
  - [ ] 6.2 Register `ShouldRunDiscovery` as default `iteration.gate` hook handler
  - [ ] 6.3 Implement `--discover-only` mode (sets `discover_interval = 1`, impl iterations = 0)
  - [ ] 6.4 Focus area injection (`--focus testing|docs|security|performance|refactoring`)

- [ ] 7.0 Agent adapters [~5,000 tokens - Complex]
  - [ ] 7.1 Define `AgentAdapter` interface at `internal/agents/adapter.go`
  - [ ] 7.2 Adapter declares: default image, env allowlist, prompt-mode (`stdin-content`, `file-arg`, `content-arg`), default args
  - [ ] 7.3 Built-in adapter: `claude` (`-p <content> --dangerously-skip-permissions`, env: ANTHROPIC_API_KEY)
  - [ ] 7.4 Built-in adapter: `codex` (`--dangerously-bypass-approvals-and-sandbox exec <content>`, env: OPENAI_API_KEY)
  - [ ] 7.5 Built-in adapter: `copilot` (port v1's behavior)
  - [ ] 7.6 Built-in adapter: `gemini`
  - [ ] 7.7 Built-in adapter: `kiro`
  - [ ] 7.8 `Adapter.Invoke(ctx, prompt, opts)` returns output + error
  - [ ] 7.9 Register `Invoke` as default `agent.invoke` hook handler

- [ ] 8.0 OCI sandbox launcher [~4,000 tokens - Medium]
  - [ ] 8.1 Reuse Milestone 3's OCI tier loader for image management
  - [ ] 8.2 Mount layout: `/workspace` (rw), `/skills` (ro), `/.samuel/run` (**ro** — agent uses CLI subcommands), `/plugin/config` (ro), `/samuel-bridge` (gRPC socket)
  - [ ] 8.3 Env var allowlist forwarding from adapter manifest
  - [ ] 8.4 User mapping `--user $UID:$GID`
  - [ ] 8.5 Image regex validation pre-invocation

- [ ] 9.0 Prompt templates [~5,000 tokens - Complex]
  - [ ] 9.1 Author `internal/methodology/ralph/templates/prompt.md.tmpl` — Go text/template, implementation prompt
  - [ ] 9.2 Author `internal/methodology/ralph/templates/discovery-prompt.md.tmpl` — discovery prompt
  - [ ] 9.3 **Rewrite for CLI mutation** — agent uses `samuel run done <id> --commit-sha $(git rev-parse HEAD)` not direct prd.toon edit
  - [ ] 9.4 Define `PromptContext` struct per RFD 0006 with all 11 sections (Samuel, Project, Methodology, Iteration, Config, Guardrails, Paths, State, Mode, Hooks, Plugins)
  - [ ] 9.5 Embed templates via `go:embed`
  - [ ] 9.6 Per-project override at `.samuel/templates/ralph/*.md.tmpl` (file override beats embedded default)
  - [ ] 9.7 Template helpers: `join`, `indent`, `relpath`, `hasPlugin`, `commitConvention`

- [ ] 10.0 samuel run command surface [~6,500 tokens - Complex]
  - [ ] 10.1 `samuel run [methodology]` — positional, default from samuel.toml `default_methodology`
  - [ ] 10.2 Cobra alias `auto` for permanent v1 compat
  - [ ] 10.3 Methodology name resolution: positional → alias map (`rw` → `ralph`) → config default → hardcoded `ralph` fallback
  - [ ] 10.4 Smart bare invocation: status if prd.toon exists, actionable help otherwise
  - [ ] 10.5 `samuel run init [--prd <path>]` — initialize runtime
  - [ ] 10.6 `samuel run start [--iterations] [-y] [--dry-run] [--profile]`
  - [ ] 10.7 `samuel run status [--tail N]`
  - [ ] 10.8 `samuel run pilot [--focus] [--discover-interval] [--max-discovery-tasks]`
  - [ ] 10.9 `samuel run --discover-only`
  - [ ] 10.10 `samuel run tasks [--status pending|completed|...]`
  - [ ] 10.11 `samuel run convert <prd-path>` — markdown → prd.toon

- [ ] 11.0 CLI mutation commands [~3,500 tokens - Medium]
  - [ ] 11.1 `samuel run done <task-id> [--commit-sha SHA] [--iteration N]` — atomic prd.toon mutation
  - [ ] 11.2 `samuel run skip <task-id> [--reason TEXT]`
  - [ ] 11.3 `samuel run reset <task-id>`
  - [ ] 11.4 `samuel run enqueue <title> [--priority] [--complexity] [--source]` — auto-id via NextAvailableID
  - [ ] 11.5 `samuel run task add <id> <title> [...]` — explicit-id for CI
  - [ ] 11.6 All acquire iteration lock, atomic write, emit JSON
  - [ ] 11.7 `--commit-sha` optional but documented as expected (per RFD 0006 resolution #5)

- [ ] 12.0 Tests [~4,500 tokens - Medium]
  - [ ] 12.1 FakePlugin for hook handler chain testing
  - [ ] 12.2 Hook composition test: 2 plugins + default chain on `context.snapshot`
  - [ ] 12.3 Strict-mode test: handler errors with strict=true aborts iteration
  - [ ] 12.4 Strict-mode test: handler errors with strict=false logs warning, loop continues
  - [ ] 12.5 End-to-end loop test with mock agent emitting fixed CLI invocations
  - [ ] 12.6 TOON malformation recovery: corrupt one row in prd.toon, loop continues on remaining
  - [ ] 12.7 Multi-agent test: swap adapter from claude to codex via samuel.toml, loop continues
  - [ ] 12.8 Per-project template override: `.samuel/templates/ralph/prompt.md.tmpl` is used instead of embedded
  - [ ] 12.9 Agnostic check: `samuel run` writes nothing to `.claude/` paths

- [ ] 13.0 Tag beta.2 [~1,000 tokens - Simple]
  - [ ] 13.1 Bootstrap test: convert this PRD itself to prd.toon; `samuel run start --iterations 1` works
  - [ ] 13.2 CHANGELOG update
  - [ ] 13.3 Tag `v2.0.0-beta.2`

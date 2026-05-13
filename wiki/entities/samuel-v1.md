---
title: Samuel v1 (current shipped)
type: entity
created: 2026-05-12
updated: 2026-05-12
sources: []
tags: [anchor, v1]
---

# Samuel v1

The currently shipped version of Samuel, living at `samuel_v1/` in this project. Reference material for the v2 rebuild — not a base to branch from.

## Identity

- Go CLI (`cmd/samuel/main.go`, module `samuel`)
- Single binary, ships every skill embedded
- Distribution: `install.sh`, goreleaser, GitHub releases
- Docs site: mkdocs

## Internal versioning

The codebase the user calls "v1" is internally **post-3.0.0 with Unreleased changes pending**. Full timeline from `CHANGELOG.md` ([[sources/2026-05-12-v1-project-meta]]):

| Internal version | Date | Notable |
|---|---|---|
| 1.0.0 → 1.7.0 | 2025-01 → 2026-01 | Initial release + iteration |
| 1.8.0 | 2026-02-04 | Agent Skills integration (25+ tool compat, `samuel skill` commands) |
| 2.0.0 | 2026-02-12 | `.agent/` → `.claude/`, single CLAUDE.md (was 3 files) |
| 3.0.0 | 2026-04-29 | **Lean reshape**: 14 → 11 commands, `samuel auto` → `samuel run`, type inference |
| Unreleased | current | what we've been ingesting |

The "v1 vs v2" framing in this wiki is the user's external mental model for the rebuild. Internally, the v2 rebuild is conceptually v4.0.0.

> Source: `samuel_v1/internal/skills/embed.go` package comment + `samuel_v1/CHANGELOG.md` ([[sources/2026-05-12-v1-project-meta]]).

## Important: v1 already did the `auto` → `run` rename

A v3.0.0 release (April 2026) renamed `samuel auto` to `samuel run`, kept `auto` as a **permanent** alias, and deprecated the nested `task list/complete/skip/reset` forms in favor of flat verbs (`tasks`, `done`, `skip`, `reset`). v2's command naming inherits these — no further rename needed.

## Top-level layout

```
cmd/samuel/main.go          # CLI entry point
internal/commands/          # ~30 CLI command implementations
internal/core/              # business logic (auto, config, registry, skill, sync, ...)
internal/orchestrator/      # gbrain / gstack / samuel component orchestration + locking
internal/skills/            # go:embed + content/ tree of all skills
internal/github/            # GitHub client
internal/ui/                # output formatting
.claude/                    # installed runtime (auto/, hooks/, settings.json, skills/)
docs/                       # mkdocs site
template/                   # project starter template
rfd-index.yaml              # RFD index
samuel.yaml                 # config file
```

## Concepts this entity links to

- [[concepts/skills-architecture-v1]]
- [[concepts/agent-skills-standard]]
- [[concepts/per-folder-context]]
- [[concepts/multi-agent-support]]
- [[entities/registry]]
- [[entities/skills-embed]]
- [[entities/skill-md]]
- [[entities/config-go]]
- [[entities/sync-claude-md]]
- [[entities/downloader-extractor]]
- [[entities/docker-sandbox]]

## Things v1 already has that v2 should keep

- **Multi-agent sandbox** ([[entities/docker-sandbox]]) — Claude, Codex, Copilot, Gemini, Kiro already wired through one interface.
- **Per-folder CLAUDE.md/AGENTS.md generator** ([[entities/sync-claude-md]]) — self-contained, useful, no v1 baggage.
- **Auto-mode** ([[synthesis/auto-mode-v2-design]]) — the flagship methodology. prd.json data model, two-mode loop (impl + pilot), pre-computed context generators.
- **Pre-computed context pattern** ([[concepts/pre-computed-context]]) — v1's real innovation. Token-discipline as a first-class methodology concern.
- **Ralph Wiggum methodology** grounding ([[concepts/ralph-wiggum-methodology]]) — fresh context per iteration.
- **Security primitives** for archive extraction (path traversal, symlink validation, size cap) in [[entities/downloader-extractor]].
- **Autogen marker convention** for safe rewrite of generated files.
- **Per-tool prompt translation** pattern for cross-agent normalization.
- **AI-output resilience** — custom UnmarshalJSON handling numeric IDs, atomic save-on-rename, AI-tool allowlist before exec.
- **Orchestrator pattern** ([[entities/orchestrator]]) — the lifecycle/rollback/lock layer. Highest-engineering-quality subsystem in v1. Maps directly to v2's plugin loader — see [[synthesis/orchestrator-as-plugin-loader]].
- **Structured errors** ([[concepts/structured-errors]]) — Problem/Cause/Fix/DocsURL pattern. Error UX as a product concern.
- **Mutation/Reverse log + atomic swap + flock(2)** — install safety primitives. Port verbatim.

## v2 implications

- **`#drop`**: static hardcoded registry. v2 needs dynamic discovery.
- **`#drop`**: language/framework/workflow as separate enums. v2 should treat all as skills with metadata-driven categorization.
- **`#rescue`**: Agent Skills standard adherence — keeps v2 portable across 25+ AI tools.
- **`#rescue`**: single-binary distribution. Convenient. Pluggability can co-exist via local skills folder + remote fetch.
- **`#open`**: keep `go:embed` for a built-in skill set, or ship lean and pull on demand?

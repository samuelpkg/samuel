---
title: core/registry.go (static component registry)
type: entity
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-skill-model]
tags: [v1, registry, drop]
---

# core/registry.go

The static, hardcoded catalog of installable components in v1. **Adding a skill in v1 requires editing this file and rebuilding the binary.**

## Structure

Three primary slices plus a derived mirror plus templates:

| Slice | Count | Purpose |
|---|---|---|
| `Languages` | 21 | Language guides — typescript, python, go, rust, kotlin, java, csharp, php, swift, cpp, ruby, sql, shell, r, dart, html-css, lua, assembly, cuda, solidity, zig |
| `Frameworks` | 30+ | Framework guides grouped by language (react, nextjs, django, fastapi, gin, axum, rails, ...) |
| `Workflows` | 23 | Process skills (initialize-project, create-rfd, code-review, auto, ...) |
| `Skills` | sum of above | Mirror — every Lang/FW/WF re-listed with skill-naming (`go` → `go-guide`) |
| `Templates` | 3 | Install presets: `full`, `starter`, `minimal` |
| `CoreFiles` | 3 | Always-installed files: `CLAUDE.md`, `AGENTS.md`, `.claude/skills/README.md` |

## Component struct

```go
type Component struct {
    Name        string
    Path        string   // e.g. ".claude/skills/go-guide"
    Description string
    Category    string   // "language" | "framework" | "workflow" | ""
    Tags        []string // search aliases
}
```

## Conventions

- Language skills suffixed `-guide`: `go-guide`, `python-guide`, ... (`LanguageToSkillName`)
- Framework/workflow skills are bare names: `react`, `auto`, ...
- `InferComponentType(name)` resolves a bare name to a category, excluding `Skills` from search (it's a mirror) to avoid self-collision.
- An invariant test (`registry_invariant_test.go`) enforces no name appears in two of `Languages` / `Frameworks` / `Workflows` at once.

## Path conventions

- All component paths live under `.claude/skills/<name>/` at install destinations.
- `TemplatePrefix = "template/"` is the source path prefix in the v3-era downloaded archive. `GetSourcePath(dest)` prepends it. Likely vestigial under v4's embed model — confirm in pass 2.

## v2 implications #drop

This file is the single biggest obstacle to "framework + skills hub" goal.

- **Static enumeration** in Go source means any new skill = code change + rebuild + release. That's not a hub.
- **Triple registration** (one of Languages/Frameworks/Workflows + Skills mirror + Templates) for every component. Triple bookkeeping, easy to drift.
- **Language vs Framework vs Workflow** as separate enums is a leaky abstraction. They're all skills with metadata. Use `metadata.category` and collapse.

## v2 replacement direction

- Dynamic discovery: filesystem scan of installed skills + optional remote registry index.
- Single `Skill` type with `Category` field.
- Templates either generated from category filters or kept as user-defined manifests.
- `Tags` should live in `SKILL.md` frontmatter, not external code.

## Where v1 references it

To be enumerated as passes 2-7 surface dependencies. Initial guesses based on naming:
- `commands/add.go`, `commands/remove.go`, `commands/ls.go`, `commands/search.go` — registry lookups
- `commands/skill.go`, `commands/skill_info.go` — skill subcommands
- `core/sync.go`, `core/downloader.go` — install flow

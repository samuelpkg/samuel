---
title: v1 Skills Architecture
type: concept
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-skill-model]
tags: [v1, skill-model]
---

# v1 Skills Architecture

End-to-end view of how skills work in Samuel v1, based on pass-1 ingest.

## The pipeline (current understanding)

```
samuel_v1/internal/skills/content/<name>/      ← source of truth in repo
              │
              │   go:embed all:content (build time)
              ▼
        Samuel binary
              │
              │   FS() → fs.Sub("content")
              ▼
      [orchestrator: samuel-skills component]      ← pass 4 (not yet ingested)
              │
              │   copy on `samuel init` / `samuel sync`
              ▼
  ~/.claude/skills/samuel/<name>/                ← global install target
              │
              │   (possibly also project-local copies)
              ▼
   <project>/.claude/skills/<name>/               ← consumed by Claude Code / Cursor / etc.
```

Note: the actual install destinations and sync rules need confirmation in passes 2 (`core/sync.go`) and 4 (orchestrator).

## The catalog model

[[entities/registry]] is a **static, code-defined enumeration** in three slices:

- `Languages` (21) — language guides
- `Frameworks` (30+) — framework guides
- `Workflows` (23) — process skills

Plus `Skills` as a derived mirror, and `Templates` as installable presets.

[[entities/skill-md]] is the per-file format every skill must follow.

[[entities/skills-embed]] is the build-time bundling mechanism.

## Tensions in v1

1. **Static vs dynamic** — registry is hardcoded in Go. Adding a skill = code + rebuild. Conflicts with v2's "framework + skills hub" goal.
2. **Triple bookkeeping** — every component appears in its category slice + the `Skills` mirror + optionally a template. Drift hazard. The invariant test only catches cross-category name collisions, not staleness.
3. **Conceptual vs structural split** — Skills/Workflows distinction is soft in code (same SKILL.md format, same directory pattern) but hard in the registry (separate slices). Pick one.
4. **Embed locks distribution** — every skill change requires rebuilding and re-releasing the binary. Acceptable for a built-in base, hostile to a hub.

## Open questions for next passes

- How does `samuel sync` decide what to copy where? (pass 2)
- What does the orchestrator do beyond copying? (pass 4)
- Are skills loaded at runtime by Samuel itself, or just laid out for the agent to discover? (pass 2 + 6)
- How do `internal/skills/content/` and `.claude/skills/` differ? Are they identical mirrors? (pass 8)
- Is there any concept of a third-party skill in v1, or is it entirely closed? (look for plugin/external loading in passes 2-6)

## v2 design implications

A v2 skills architecture that lives up to "framework + skills hub":

- Single skill type — drop the language/framework/workflow trichotomy. Use `metadata.category`.
- Dynamic discovery — scan a known directory at runtime instead of compiled lists.
- Built-in vs external — small embedded blessed set OR no embed at all + initial install pulls a curated bundle.
- Plugin contract — what does a "plugin" provide beyond skills? (hooks? commands? services?) — TBD when [[concepts/extensibility-design]] is fleshed out.

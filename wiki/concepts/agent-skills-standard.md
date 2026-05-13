---
title: Agent Skills standard
type: concept
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-skill-model]
tags: [external-standard, rescue]
---

# Agent Skills standard

Open standard for capability modules consumed by AI agents. Hosted at [agentskills.io](https://agentskills.io), spec at [github.com/agentskills/agentskills](https://github.com/agentskills/agentskills).

## What it is

A skill is a directory:

```
skill-name/
├── SKILL.md          # required: YAML frontmatter + markdown body
├── scripts/          # optional: executable code
├── references/       # optional: supplementary docs
└── assets/           # optional: templates, data
```

See [[entities/skill-md]] for the file format details.

## Why it matters for Samuel

v1's `internal/skills/README.md` claims compatibility with **25+ agent products** including Claude Code, Cursor, GitHub Copilot, VS Code, OpenAI Codex.

This is the single biggest portability win in v1. A skill written for Samuel works in all of them.

## Skills vs Workflows (v1's distinction)

From `internal/skills/README.md:106-114`:

| | Skills | Workflows |
|---|---|---|
| Purpose | Add capabilities | Guide processes |
| Focus | What AI can do | How to approach tasks |
| Portability | Cross-tool (25+ products) | Samuel-specific |
| Structure | SKILL.md + resources | Markdown steps |
| Example | Language guides, commit messages | Code review, PRD creation |

In v1's actual code, this distinction is **soft** — workflows are also stored as Agent Skills under `.claude/skills/<name>/SKILL.md`. The split is more conceptual than structural.

## v2 implications

- `#rescue` (firm) — keep Agent Skills format. It's the open standard and the cross-tool portability is invaluable.
- `#open` — keep the conceptual Skills vs Workflows split, or unify? In the codebase they're already unified at the file level; only the registry slices keep them separate. If the registry collapses, the split may dissolve naturally.

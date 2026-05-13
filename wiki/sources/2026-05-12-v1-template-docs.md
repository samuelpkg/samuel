---
title: v1 Template + Docs Site
type: source
created: 2026-05-12
updated: 2026-05-12
sources: []
tags: [v1, template, docs]
---

# v1 Template + Docs

Ingest pass 9. The project-root install template + the mkdocs documentation site.

## Files

### Template (3 files)

- `template/CLAUDE.md` (474 lines) — master template installed at project root
- `template/AGENTS.md` (474 lines) — near-identical mirror (one trivial diff)
- `template/.claude/auto/prompt.md` — auto-mode prompt template

### Docs site (~110 markdown files)

- `mkdocs.yml` (300 lines) — Material theme config
- `docs/index.md` — landing
- `docs/core/` (5): overview, claude-md, methodology, guardrails, agent-directory
- `docs/getting-started/` (4): installation, quick-start, first-task, migration-v3
- `docs/languages/` (22): one per language + index
- `docs/frameworks/` (34): one per framework + index
- `docs/workflows/` (25): one per workflow + index
- `docs/reference/` (5): cli, faq, contributing, cross-tool, changelog
- `docs/rfd/` (5): RFD index + 4 RFDs (covered in pass 10)
- `docs/assets/`, `docs/javascripts/`, `docs/stylesheets/`, `docs/includes/`

## Key claims

### `template/` is minimal

Only three files. The skills come from `internal/skills/content/` via [[entities/skills-embed]]. The `template/` dir is purely the **project-root install** (CLAUDE.md, AGENTS.md, and the auto-mode prompt).

The `TemplatePrefix = "template/"` constant in [[entities/registry]] is a v3-era artifact — under the embed model, the `template/` paths apply only to these three files, not the skill tree.

### CLAUDE.md template = 474 lines

Massive. Sections:

1. **Operations** — setup/test/build commands for TypeScript, Python, Go, Rust. Environment variable examples.
2. **Boundaries** — protected files, never-commit list, ask-before-modifying list.
3. **Quick Reference** — task classification (ATOMIC/FEATURE/COMPLEX), guardrails, autonomous mode quick-ref, emergency links to skills.
4. **Skills table** — auto-generated between `<!-- SKILLS_START -->` / `<!-- SKILLS_END -->` markers ([[entities/skill-md]] generates this via `GenerateSkillsSection`).
5. **Language guide links** — 21 entries with extension mappings.
6. **Framework guide links** — grouped by language.
7. **Core Guardrails** — 35+ rules across Code Quality, Security, Testing, Git, Performance.
8. **4D Methodology** — see [[concepts/4d-methodology]].
9. **Software Development Lifecycle** — Planning/Implementation/Validation/Documentation/Commit stages.
10. **Per-Folder CLAUDE.md** — hierarchical instructions docs.
11. **Project Context** — fillable section with `[e.g., ...]` placeholders for Tech Stack, Architecture, Key Design Decisions.
12. **Anti-Patterns** — code/testing/process don't-do list.
13. **When Stuck** — recovery procedure.
14. **Version & Changelog** — v2.0.0 → v1.8.0 history embedded.

**Token cost**: ~7-8K tokens loaded into every AI agent's context. Significant.

### CLAUDE.md vs AGENTS.md diff

Identical except for one line about which file is the "primary." In v1, CLAUDE.md says "AGENTS.md is a copy of this file"; AGENTS.md says nothing about CLAUDE.md being primary. Per [[concepts/agents-md-primary]], v2 inverts: AGENTS.md is primary.

### Project Context section is fillable

```markdown
### Tech Stack
<!-- Fill in when tech decisions are made -->
- **Language**: [e.g., TypeScript 5.3, Python 3.11, Go 1.21]
- **Framework**: [e.g., React 18, Django 5, Gin]
- **Database**: [e.g., PostgreSQL 15, MongoDB 7]
- **Infrastructure**: [e.g., Vercel, AWS, Docker]
```

Good UX — the template invites users to fill in their specifics. v2 keeps this.

### Embedded changelog

The template carries its own version history:

> **v2.0.0 (2026-02-11) - Native Claude Code Integration**
> - Migrated from `.agent/` to `.claude/` (native Claude Code directory)
> - Merged AI_INSTRUCTIONS.md + CLAUDE.md + project.md into single CLAUDE.md
> - Skills now live in `.claude/skills/` (native skill discovery)
> ...

Reveals v1's evolution: AI_INSTRUCTIONS.md + CLAUDE.md + project.md were three separate files, consolidated into one CLAUDE.md in v2.0. Predecessor format was `.agent/` (deprecated). Confirms the v1/v2/v3/v4 internal version numbering noted in [[entities/samuel-v1]].

### Docs site uses mkdocs-material

The theme is `mkdocs-material` — the standard for high-quality OSS docs sites. Configuration is feature-rich:

- Light/dark palette toggle, system preference detection.
- Navigation: instant, sticky tabs, sections, footer, breadcrumbs.
- Search: suggest, highlight, share.
- Content: code copy, annotations, tabs link, tooltips.
- Custom logo + favicon at `docs/assets/`.

URL: `https://ar4mirez.github.io/samuel/` (GitHub Pages).

Edit URI configured (`edit/main/docs/`) — users can click "edit this page" → opens GitHub editor.

### Docs/skills duplication

`docs/languages/go.md`, `docs/frameworks/react.md`, `docs/workflows/create-rfd.md` — these mirror the skill content at `.claude/skills/<name>/SKILL.md`. Likely a manual port or build-time generation step.

Three-way duplication: `internal/skills/content/` ↔ `.claude/skills/` ↔ `docs/{languages,frameworks,workflows}/`. The second two derive from the first.

### `migration-v3.md`

Lives at `docs/getting-started/`. Documents migrating from v2 → v3 (v3 introduced `samuel run` rename, JSON schema v3, etc.). v2 (this rebuild) is a clean break — no equivalent migration guide planned.

## Assessment

- **Credibility**: high.
- **Quality**: docs site is professional-grade mkdocs-material. Template is comprehensive but bloated.
- **Token waste**: 474-line CLAUDE.md is a real concern. Most users only need 50-100 lines of project-specific guidance in their agent context.

## v2 implications

### Template — shrink dramatically

The 474-line CLAUDE.md template is right for v1's "single big file" model. v2's architecture has changed:

- **Skills table** disappears — plugins are dynamic, table is generated on demand by `samuel run` or `samuel ls`.
- **Language/framework guide links** disappear — plugins auto-load per `metadata.extensions`.
- **Operations section** becomes plugin-driven — each language plugin contributes its own setup commands.
- **Guardrails** become user-overridable via `samuel.toml [guardrails]` block, rendered into AGENTS.md at sync time.
- **4D Methodology + escalation triggers** stays — Samuel Way content, ~50 lines.
- **Project Context fillable section** stays — good UX.
- **Embedded changelog** → goes to docs site, not template.
- **Anti-patterns, when-stuck** → go to docs site.

Target: AGENTS.md template at ~100 lines instead of 474. Saves ~5K tokens per agent invocation.

### Marker pattern survives

The `<!-- SKILLS_START -->` / `<!-- SKILLS_END -->` blocks are good — let the framework re-render that section without clobbering user edits. v2 uses `<!-- SAMUEL_PLUGINS_START -->` / `<!-- SAMUEL_PLUGINS_END -->` for the dynamic plugin list.

### Docs site — port verbatim, change structure

- **Keep mkdocs-material** — industry standard, no reason to swap.
- **Keep the structure** — core / getting-started / reference / rfd.
- **Drop the languages/ frameworks/ workflows/ trees** — replace with `docs/plugins/` that builds dynamically from installed-plugin manifests during the docs build. (Or: drop them entirely from the framework docs site; let each plugin own its own docs.)
- **migration-v3.md** drops. v2 is clean-break.
- **Add docs/concepts/** for v2-specific concepts (plugin model, methodology hooks, etc.).

### 4D Methodology — elevate

The 4D Methodology surfaces here for the first time. It's Samuel Way content that deserves its own concept page — see [[concepts/4d-methodology]].

## Related pages

- [[concepts/4d-methodology]]
- [[concepts/agents-md-primary]]
- [[synthesis/v2-template-and-docs]]

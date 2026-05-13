---
title: Positioning — Rails for coding assistants
type: synthesis
created: 2026-05-12
updated: 2026-05-12
sources: []
tags: [v2, v2-decision, positioning]
---

# Rails for coding assistants

The v2 positioning thesis, in the user's own framing: **Samuel is the Ruby on Rails for AI coding assistants.**

## What that means

Rails is three things at once:

1. **A package manager** — gems via Bundler, engines for mountable apps.
2. **A task runner with opinions** — generators, rake tasks, CLI conventions.
3. **A methodology baked in** — the Rails Way. Convention over configuration. MVC. ActiveRecord. Testing layers. Asset pipeline. Hooks for everything.

Samuel v2 maps directly:

| Rails | Samuel v2 |
|---|---|
| Bundler + gems | Plugin loader + GitHub-backed registry |
| Engines | Executable plugins (OCI images) |
| Generators | `samuel init`, `samuel skill new`, plugin scaffolding |
| Rake tasks | Methodology hooks invoked from the CLI |
| The Rails Way | Built-in conventions: AGENTS.md, SKILL.md, design.md, RFD process, PRD format |
| Asset pipeline | Skill content pipeline (validate, render, ship) |

## What it is not

- Not a coding assistant. Doesn't generate code itself. Doesn't replace Claude Code or Codex.
- Not a wrapper around one assistant. Tool-agnostic from day one (Claude Code first, Codex next).
- Not a thin shell with no opinions. The Samuel Way is the product. Conventions ship by default.
- Not a feature catalog. Not defined by what it does — defined by the shape it imposes.

## What this resolves

- **Built-in vs plugin boundary**: see [[concepts/extensibility-design]]. Built-in = the framework + cross-tool conventions. Plugins = everything domain-specific.
- **Plugin distribution**: see [[concepts/plugin-format]]. Skills as text, executable plugins as OCI images, GitHub as substrate.
- **Versioning**: see [[concepts/versioning-compatibility]]. SemVer + cargo ranges + capability model.
- **Sandbox**: container per coding-assistant run; container per executable plugin.

## What it leaves open

- Which conventions count as "Samuel Way" vs "popular plugin"? RFDs and PRDs feel core. Language guides feel like plugins. The middle — code-review, generate-tasks, auto-mode — is undecided and gets resolved during the v1 commands/workflow passes.
- Pricing/release model for plugins (free, paid, sponsored). Not yet a question; flag for later.
- Multi-language framework support — does "Samuel for Python projects" mean different defaults, or just install different plugins? Probably the latter, matching how Rails handles different DB adapters.

## How this changes the rest of the ingest

The remaining passes should be read with the Rails lens applied:

- Passes 2–5 (config, sync, auto, orchestrator, github, ui): which of these are *core framework* mechanisms vs *plugin concerns*? Most config/sync logic looks core. Auto-mode probably becomes a built-in methodology hook with content packs as plugins. Orchestrator (gbrain/gstack) is interesting — looks like the seed of plugin orchestration.
- Pass 6 (commands): every CLI command is a candidate for "built-in core surface" or "delegated to plugins." Use Rails CLI as model — `samuel new`, `samuel install`, `samuel generate`, `samuel run`.
- Pass 8 (skill content): triage every skill into `#rescue` (becomes a v2 plugin), `#drop`, or `#refactor` (split between plugin content and built-in convention).

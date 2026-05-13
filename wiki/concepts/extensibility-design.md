---
title: Extensibility design (v2)
type: concept
created: 2026-05-12
updated: 2026-05-12
sources: []
tags: [v2, v2-decision]
---

# Extensibility design

What's built in, what's a plugin, and how the two compose.

## Built-in (the thin core) #v2-decision

Only universal coding-assistant primitives. Things every supported tool (Claude Code, Codex, future others) shares.

Confirmed:

- **AGENTS.md** generation + maintenance — cross-tool agent instruction file.
- **SKILL.md** parser + validator — Agent Skills standard.
- **design.md** (and similar planning artifacts) — convention support.
- **Plugin loader** — discovery, install, version resolution, sandboxed execution.
- **Methodology hooks** — entry points the framework calls for tasks like "plan a feature", "run a review". Methodology *content* lives in plugins; *invocation* lives in core.

Not built in:

- Specific language guides (go-guide, python-guide, ...) → plugins.
- Specific framework guides (react, django, ...) → plugins.
- Specific workflow skills (create-rfd, code-review, auto, ...) → plugins.
- IDE/tool-specific integrations → plugins.

## Plugin types

Two tiers — see [[concepts/plugin-format]] for transport details.

### Tier 1: Skills (knowledge)

- Pure text + assets. SKILL.md + optional `references/`, `assets/`, templates.
- No executable code Samuel runs.
- Transport: Git repo or tar/zip archive.
- Sandbox: not required (no execution).
- Example: `go-guide`, `react`, `create-rfd`.

### Tier 2: Executable plugins (capability)

- Bundles that include runnable code (tools, hooks, custom commands).
- Transport: **OCI image** pulled from `ghcr.io` or any OCI registry.
- Sandbox: container runtime (Docker, Podman).
- Example: a methodology plugin that runs custom analysis steps; a translator that converts SKILL.md → Cursor rules; a sandboxed code-runner.

A single GitHub repo may publish both — a skill bundle and a companion OCI image — under one plugin manifest.

## Discovery #v2-decision

Two sources, same resolution order:

1. **Local** — `~/.samuel/plugins/` and `<project>/.samuel/plugins/`. Filesystem scan. Overrides registry.
2. **GitHub-backed registry** — a known repo (e.g. `ar4mirez/samuel-registry`) holding an index file. Index maps `<plugin-name>` → `<github-repo>` + version policy. Resolve to a Git tag, fetch.

No central registry server. No package backend. GitHub is the substrate.

Custom registries: users can configure additional registry repos (similar to how cargo lets you add additional registries). Same resolution order applies — first match wins.

## Composition

A plugin manifest declares:

- `name`, `version` (SemVer)
- `samuel` — framework version range it supports
- `provides` — what the plugin offers (skills, commands, hooks, methodology steps)
- `requires` — other plugins this one depends on
- `capabilities` — permissions requested (filesystem.read, filesystem.write, exec, network) — see [[concepts/versioning-compatibility]]

Samuel resolves the graph at install time, refuses to load on incompatible framework version or unsatisfied dependencies.

## The Rails analogy

See [[synthesis/positioning-rails-for-coding-assistants]] for the full thesis.

Short version: Samuel core = Rails framework (small, opinionated, conventions over configuration). Plugins = Rails engines + gems. Skills = best-practice generators. Methodology = the Rails Way, baked into core hooks but content-extensible via plugins.

## Open

- Methodology graduation criteria — which v1 workflows are "core methodology" hooks vs "plugin content"? Answer after passes 6 and 10.
- Plugin discovery cache: refresh policy, offline mode behavior.
- Trust model: signed plugins? attestations via Sigstore? GitHub repo verification only?

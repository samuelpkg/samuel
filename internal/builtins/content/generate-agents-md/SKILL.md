---
name: generate-agents-md
description: |
  Generate a root AGENTS.md for a project from `samuel.toml` defaults
  (default methodology, guardrails, plugin list). Internally folded
  into `samuel sync` — kept as a discrete built-in so legacy invokers
  that reach for "generate-agents-md" still resolve to a manifest.
  Use when the user explicitly says "generate the root AGENTS.md" or a
  tool installer needs the pre-v2 name.
license: MIT
metadata:
  author: samuel
  version: "0.1.0"
  category: workflow
  kind: builtin
---

# Generate AGENTS.md (folded into sync)

The root AGENTS.md template renders from `samuel.toml`:

- `default_methodology` — short description of the active loop
- `[guardrails]`        — code-quality limits rendered inline
- `[[plugins]]`         — installed skills/tools the agent can use

In v2 this body is invoked by `samuel sync` as part of the per-folder
walk; the standalone command is a thin alias kept for backwards
compatibility with v1 tooling.

The rendered output stays under 150 lines after variable expansion —
that ceiling is enforced by CI on the generated output, not the
template source.

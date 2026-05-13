---
title: Methodology — default built-in + optional plugins
type: concept
created: 2026-05-12
updated: 2026-05-12
sources: []
tags: [v2, v2-decision]
---

# Methodology — default + plugins

For every workflow Samuel ships (auto-mode, code-review, create-rfd, generate-tasks, …):

- **One default built-in version** lives in the framework. This is "the Samuel Way."
- **Optional plugins** enhance, replace, or swap individual steps.

## Why this is the right shape

- Rails analogy: ActiveRecord ships by default, but you can swap to Sequel. Asset pipeline ships by default, but you can swap to Webpacker. Every opinionated default has a plugin escape hatch.
- New users get a coherent flow on day one. They run `samuel auto` and it just works.
- Power users tune individual stages without rewriting the workflow. A plugin adds a custom quality check, swaps the snapshot strategy, plugs in a different orchestrator, ...
- Pluggable agents already fit cleanly — `claude` is default, but you can `samuel auto --agent codex` or install an `aider` adapter plugin.

## Hook points (sketch — refine after pass 3)

Each methodology workflow exposes named hook points the framework calls. Plugins register handlers for those points.

Example for auto-mode:

```
samuel auto
  ├─ before:plan         ← plugins can refine PRD/discovery
  ├─ plan                ← built-in default; plugin can replace
  ├─ before:iteration    ← plugins can enrich context
  ├─ iteration           ← the actual agent invocation (built-in)
  │     ├─ prompt        ← plugin: custom prompt assembly
  │     └─ agent.run     ← plugin: agent adapter
  ├─ after:iteration
  ├─ quality.check       ← built-in: configurable check list; plugins add checks
  └─ after:done          ← plugins for post-completion (commit, PR, notify)
```

Plugins declare which hooks they target in their manifest. Multiple plugins can attach to one hook; ordering is deterministic (alphabetical by plugin name, override in `samuel.toml`).

## Built-in vs plugin call

Per workflow during ingest passes:

- **Full built-in**: ship the v1 logic verbatim (cleaned up). Plugin = hook overlays only. Examples (provisional): per-folder sync, AGENTS.md generation.
- **Minimal built-in**: ship a basic version. Plugin = the depth. Examples (likely): create-rfd (built-in writes a skeleton, plugin handles publishing to Notion/Linear).
- **Plugin-only**: not built in at all. Examples (likely): language guides, framework guides.

This call happens per-workflow in the relevant ingest pass, not upfront.

## Open

- Hook contract shape — JSON over stdin/stdout, gRPC, or a Go interface for built-in cases?
- Hook discovery — plugin manifest declares `hooks = ["auto.before:iteration"]`?
- Conflict resolution — two plugins replace the same hook, what wins? (Last-wins, explicit ordering, error and prompt?)
- Configuration inheritance — does `[methodology.auto]` in `samuel.toml` cascade into hook-handler plugins?

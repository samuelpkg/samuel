# Methodology hooks

The autonomous loop is one concrete methodology — Ralph + 4D — but the *runtime* is generic. Every meaningful step the loop takes fires a named hook event. Plugins register handlers; the framework ships a built-in default for each event.

## The resolution rule

For every hook event, handler order is:

1. **User override** in `samuel.toml` (`[methodology.ralph.hooks.<event>.handlers = ["plugin:fn"]]`).
2. **Built-in default** baked into the binary.
3. **Plugin** handlers registered via the plugin's `[hooks]` block.

User overrides win — they can disable a built-in default, reorder handlers, or pin a specific plugin's handler to run first. Without an override, the built-in default runs, and any plugin handlers run after it.

This means *the default methodology works without plugins* (the built-in is always there) and *plugins can extend it* (their handlers fire after the default).

## Why not "fork the loop"

The alternative is letting a plugin replace the entire loop. That's the wrong abstraction. The loop has correctness properties — the lockfile, the iteration cap, the CLI-mutation invariant — that *must* hold regardless of methodology. If plugins replace the loop, they can break those, and users can't reason about what running `samuel run` does.

Hook points let plugins enhance behaviour at well-defined seams (snapshot generation, agent invocation, quality check) without touching the surrounding invariants.

## Strict vs non-strict

Each hook event has a per-handler `strict` flag in `samuel.toml`. A strict handler that fails aborts the loop. A non-strict handler that fails logs a warning and the loop continues. The default for safety-relevant hooks (`iteration.gate`, `quality.check`) is strict; for context-enrichment hooks (`context.extra`) it's non-strict.

Every handler has a `timeout` (default 30s). A handler that exceeds it is killed and treated as failed.

## Capability gating

Plugin hook handlers run through the same capability checker as install-time. A handler that tries to write to `/workspace` without `filesystem.write` declared in its manifest is denied — the framework doesn't trust the manifest you read at install time and then accept arbitrary behaviour at runtime.

## What hooks unlock

- **Custom context generators**: a `context.extra` handler can inject embeddings, vector-store hits, or design-doc snippets into the prompt.
- **Quality plugins**: a `quality.check` handler can add a security scanner, a lint pass, or a custom acceptance test runner without modifying `samuel.toml`'s `quality_checks` list.
- **Agent sandboxes**: `before:agent.invoke` can swap the sandbox config, set env vars, or attach an OCI volume.
- **New methodologies**: a plugin that registers handlers for every event — and a user override that puts them first — is effectively a drop-in replacement methodology, without forking the loop.

See [RFD 0004](../rfd/0004.md) for the design, including the rejected alternatives (pre/post hooks only; middleware chain; full event bus).

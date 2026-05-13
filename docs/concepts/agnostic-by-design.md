# Agnostic by design

"Agnostic" here is a CI invariant, not a marketing word. **No code path inside the Samuel framework writes to a tool-specific path.** No `CLAUDE.md`, no `.claude/`, no `~/.claude/`, no `.cursor/`, no `.codex/`. The framework writes `AGENTS.md`, `samuel.toml`, `samuel.lock`, and files under `.samuel/`. That's it.

## The invariant

There is an automated check (`.github/workflows/agnostic-check.yml`) that runs on every PR. It:

1. Runs `samuel init <tmpdir>` in a clean directory.
2. Runs `samuel sync` and a few `samuel install` flows.
3. Walks the resulting tree and `$HOME` looking for any path matching `**/CLAUDE.md`, `**/.claude/**`, `**/.cursor/**`, `**/.codex/**`.
4. Fails the build if it finds one.

The check is also baked into the v2.0 test suite (`internal/sync/agnostic_test.go`). Either failure mode blocks merge.

## What counts as a leak

- Code in `internal/` that opens a file whose path contains a tool name.
- A built-in skill (`~/.samuel/builtins/`) that writes one.
- A template under `template/` that emits one when rendered.

What doesn't count: a plugin you installed, because plugins are external code with declared capabilities. If `samuel-claude-translator` writes `CLAUDE.md`, that's its job — its manifest declares `filesystem.write:/workspace/**/CLAUDE.md` and the user granted it on install.

## How translator plugins solve it

The framework emits `AGENTS.md` once. Translator plugins hook `sync.after` and mirror it into whatever shape their target tool needs:

- The [claude translator](https://github.com/samuelpkg/samuel-claude-translator) writes `.claude/settings.json` plus `CLAUDE.md` per folder.
- The [codex translator](https://github.com/samuelpkg/samuel-codex-translator) writes `.codex/<rel>/context.md`.
- A future Cursor translator would write `.cursor/rules/*.mdc`.

Each is a one-trick WASM module with narrow capabilities. None of them know about each other. None of them know about Samuel's internals beyond the `samuel.api` host calls they're authorised to use.

## Why this matters in practice

The framework lives a long time. Tools come and go — Cursor, Aider, Continue, Codex, Gemini Code, the next thing. If the framework knows about specific tools, it acquires permanent debt every time it adds one and breaks when it removes one. If the framework only knows about `AGENTS.md`, every tool integration is a self-contained plugin that ships, ages, and gets retired on its own clock without touching the core. See [RFD 0002](../rfd/0002.md) and [RFD 0009](../rfd/0008.md) (the `.claude/`-agnostic enforcement).

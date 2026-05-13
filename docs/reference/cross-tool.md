# Cross-tool generation

Samuel writes one canonical file (`AGENTS.md`). Translator plugins generate the tool-specific files your editor needs. Two translators ship in the v2.0 registry; more land as the ecosystem grows.

## How it works

Both translators hook `sync.after`. When you run `samuel sync` (manually or as part of `samuel run`), the framework:

1. Walks the project tree, writing per-folder `AGENTS.md` files.
2. Fires `sync.after` once with the list of files written.
3. Each translator handler reads that list and emits its tool-specific shape.

Translators run in declared order (the order their plugins were installed) unless overridden by `samuel.toml`.

## `samuel-claude-translator`

Repo: [`samuelpkg/samuel-claude-translator`](https://github.com/samuelpkg/samuel-claude-translator). WASM tier.

| Hook | Output |
| --- | --- |
| `init.after` | `.claude/settings.json` (with PreToolUse stubs) |
| `sync.after` | `CLAUDE.md` per folder that has an `AGENTS.md` |

Capabilities: `filesystem.write:/workspace/**/CLAUDE.md`, `filesystem.write:/workspace/.claude/**`. Scoped narrowly so the user-side install prompt is unambiguous.

Install:

```bash
samuel install samuel-claude-translator
samuel sync
ls CLAUDE.md .claude/settings.json
```

## `samuel-codex-translator`

Repo: [`samuelpkg/samuel-codex-translator`](https://github.com/samuelpkg/samuel-codex-translator). WASM tier.

| Hook | Output |
| --- | --- |
| `sync.after` | `.codex/<rel>/context.md` per folder that has an `AGENTS.md` |

Capabilities: `filesystem.write:/workspace/.codex/**`.

Install:

```bash
samuel install samuel-codex-translator
samuel sync
```

## Hook timing

The translators fire **after** the framework has finished writing `AGENTS.md`. Their input is the canonical content; their output is a transform. They never read `samuel.toml` directly — the framework hands them everything they need in the `sync.after` payload.

This decoupling means:

- Translator plugins are stateless. Re-running them with the same input produces the same output (idempotent).
- Adding a new translator doesn't change framework behaviour.
- Removing a translator removes its files on `samuel uninstall` (mutation log replay).

## How to add a new translator

1. Build a WASM plugin (see [TinyGo + WASM](../plugin-authors/tinygo-wasm.md)).
2. Declare:
   ```toml
   [hooks]
   "sync.after" = "render"
   
   [capabilities]
   filesystem = [
     { read  = "/workspace/**/AGENTS.md" },
     { write = "/workspace/<your tool's path>/**" },
   ]
   ```
3. In the `render` function, walk `in.FilesWritten`, transform each `AGENTS.md`, and call `samuel.fs_write` for each target.
4. Release via the `samuelpkg/samuel-plugin-release` workflow.
5. Open a PR against `samuelpkg/samuel-registry`'s `index.toml`.

The framework will pick it up. Users who install your translator get cross-tool output for the new tool, with no framework changes.

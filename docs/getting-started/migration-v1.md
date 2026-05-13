# Migrating from v1

Samuel v2 is a clean break. There is no upgrade tool — you start fresh, and v1 stays available at the [`v1-final`](https://github.com/samuelpkg/samuel/tree/v1-final) tag.

## What changed

| Concern | v1 | v2 |
| --- | --- | --- |
| Canonical context file | `CLAUDE.md` (root + per folder) | `AGENTS.md` (root + per folder) |
| Tool-specific files | hard-coded `.claude/` writes | translator plugins (`samuel-claude-translator`, `samuel-codex-translator`, …) |
| Skill format | `SKILL.md` only | three-tier: skill / WASM / OCI |
| Skill distribution | bundled in the binary | external registry repo + cosign-signed releases |
| Composition | gstack | dropped — every plugin is a plugin |
| MCP brain | gbrain registration | dropped |
| Languages / frameworks / workflows | three separate enums | plugins (all live in the registry) |
| Auto-mode | hard-coded Ralph loop | Ralph as default methodology + 13 plugin hook points |
| Runtime files | `.claude/run/` (mixed JSON/MD) | `.samuel/run/` (TOON + MD) |
| Agent edits state by… | writing `prd.json` directly | calling `samuel run done|skip|enqueue` |

## What stays the same

- **Binary name** — `samuel` v2 overwrites v1 if you reinstall via brew or `install.sh`.
- **Methodology** — the 4D loop + Ralph iteration cap are intact; v2 makes the hooks extensible.
- **Repo layout philosophy** — `samuel.toml` at root, `.samuel/` for state, AGENTS.md (was CLAUDE.md) walked per folder.
- **`samuel auto`** — kept as a permanent alias for `samuel run start`.

## Recommended path

1. **Pin v1 if you depend on it.** Tag your project state, or install v1 in a separate `PATH` slot before upgrading.
2. **Install v2.** See [Installation](installation.md). The binary is named `samuel` — it will shadow v1 on `PATH`.
3. **Run `samuel init` in a clean checkout.** This writes `samuel.toml`, `.samuel/`, and a fresh `AGENTS.md`. It refuses to run inside the Samuel source repo and inside a directory that still has v1 state, unless you pass `--force`.
4. **Reinstall your plugins.** v1 skills don't map directly; the migration tool [`scripts/migrate-v1-skills`](https://github.com/samuelpkg/samuel/tree/main/scripts/migrate-v1-skills) is for plugin authors porting their own skills — end users just `samuel install <name>` against the registry.
5. **Add a translator plugin.** If you still want `CLAUDE.md` per folder, install [`samuel-claude-translator`](https://github.com/samuelpkg/samuel-claude-translator). It hooks `sync.after` and mirrors AGENTS.md → CLAUDE.md.
6. **Move your PRDs.** v1 PRDs live in markdown; convert them with `samuel run convert <path>`.

## What's gone

- **gstack** — composition is a plugin concern now. Build a meta plugin (`kind = "meta"`) that declares `[requires]` if you want a starter pack. The [samuel-starter](https://github.com/samuelpkg/samuel-starter) plugin is the canonical example.
- **gbrain MCP** — Samuel no longer registers MCP servers. Use the [MCP CLI tooling](https://modelcontextprotocol.io/) directly, or build a Samuel plugin if you want lifecycle integration.
- **Language / framework / workflow enums** — all three live in the registry now. There are no first-class "language" or "framework" concepts in the framework itself.

## Where to read more

- The clean-break rationale lives in [RFD 0008](../rfd/0008.md).
- The plugin migration plan is [RFD 0007](../rfd/0007.md).
- The AGENTS.md decision is [RFD 0002](../rfd/0002.md).

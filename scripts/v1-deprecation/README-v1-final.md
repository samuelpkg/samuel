# Samuel v1 — preserved at `v1-final`

You're looking at the v1 source. This branch is no longer the
active codebase — **v2 is.** v1 is preserved for archaeology and for
anyone running an old install.

## Where v2 lives

The `main` branch of this repo is now v2. The page you're reading
exists at the [`v1-final`](https://github.com/samuelpkg/samuel/tree/v1-final)
tag.

- **Repo**: [`samuelpkg/samuel`](https://github.com/samuelpkg/samuel)
- **Docs**: [samuelpkg.github.io/samuel/](https://samuelpkg.github.io/samuel/)
- **CHANGELOG**: [CHANGELOG.md](https://github.com/samuelpkg/samuel/blob/main/CHANGELOG.md)

## Why the clean break

v1 grew up coupled to Claude (`CLAUDE.md`, `.claude/`, the SKILL.md
format, gstack composition, gbrain MCP). v2 inverts that —
`AGENTS.md` is canonical and tool-specific files come from
translator plugins. The data model, the plugin format, and the
methodology hooks are all different enough that an in-place upgrade
would do more harm than help.

See [RFD 0008](https://github.com/samuelpkg/samuel/blob/main/docs/rfd/0008.md)
for the rationale and [`docs/getting-started/migration-v1.md`](https://github.com/samuelpkg/samuel/blob/main/docs/getting-started/migration-v1.md)
for the migration notice.

## Still using v1?

That's fine. v1 is preserved at the `v1-final` tag and old binaries
remain on the [Releases](https://github.com/samuelpkg/samuel/releases)
page. The v1 plugins (the old SKILL.md tree) are not maintained —
no security backports, no bug fixes — but they will continue to
work for installs that already have them.

## Bringing your project across

The migration is "start fresh." There is no upgrade command. The
quick-start is:

```bash
brew install ar4mirez/tap/samuel     # tap unchanged; v2 overwrites the v1 binary
samuel init
# Edit AGENTS.md (the canonical context file)
samuel install <plugins-you-want>
samuel doctor
```

To keep a `CLAUDE.md` file in your tree, install the translator
plugin: `samuel install samuel-claude-translator`. It hooks
`sync.after` and mirrors `AGENTS.md` → `CLAUDE.md` on every
`samuel sync`.

## License

Unchanged. MIT.

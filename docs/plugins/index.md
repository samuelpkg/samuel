# Plugins

This page indexes the public plugin registry at [`samuelpkg/samuel-registry`](https://github.com/samuelpkg/samuel-registry).

> **Note.** The per-plugin pages under this directory are generated at docs-build time from the registry's `index.toml`. Run `scripts/gen-plugins-pages.sh` to populate them. v2.0 ships the generator producing one short page per plugin pointing at its repo; richer per-plugin content (README mirroring, capability tables) lands in v2.1.

## Browsing

You can browse the registry from the CLI without leaving your terminal:

```bash
samuel search typescript
samuel info samuel-typescript
samuel install samuel-typescript
```

`samuel ls --all` lists everything in the registry the framework can resolve, marking which are installed locally.

## What a per-plugin page looks like

Each generated page looks like:

```markdown
# samuel-typescript

- **Kind**: skill
- **Latest**: v1.2.0
- **Repo**: https://github.com/samuelpkg/samuel-typescript
- **Tags**: typescript, javascript, guardrails

TypeScript guardrails and idioms.

## Install

`samuel install samuel-typescript`

## See also

[Registry entry](https://github.com/samuelpkg/samuel-registry/blob/main/index.toml)
```

## Known limitations (v2.0)

- The generator does **not** update `mkdocs.yml`'s `Plugins:` nav section. After running the script, the new pages exist on disk but won't appear in the side navigation until manually added (or until v2.1 lands the nav-injector).
- Per-plugin pages do not yet mirror the plugin's README — they only point at its repo. Authors who want richer content in the docs site can edit the generated page after the script runs; the next regeneration will preserve content between marker comments.

## Adding a plugin to the registry

See the [registry CONTRIBUTING guide](https://github.com/samuelpkg/samuel-registry/blob/main/CONTRIBUTING.md) and the [plugin authors overview](../plugin-authors/index.md).

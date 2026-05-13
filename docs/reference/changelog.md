# Changelog

The authoritative changelog lives in the repo at [`CHANGELOG.md`](https://github.com/samuelpkg/samuel/blob/main/CHANGELOG.md). It follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and tracks every milestone tag from `v2.0.0-alpha.1` (Foundation) onward.

## Versioning model

Samuel uses [Semantic Versioning 2.0](https://semver.org/spec/v2.0.0.html):

- **Major** bumps signal incompatible changes to the public surface — CLI flags, `samuel.toml` schema, plugin manifest schema, hook event names, lockfile schema.
- **Minor** bumps add functionality without breaking existing projects.
- **Patch** bumps fix bugs.
- **Prerelease tags** (`-alpha.N`, `-beta.N`, `-rc.N`) ride during a milestone's development; the unsuffixed major.minor is the GA tag.

Plugins follow the same scheme and declare their compatibility against the framework via the `plugin.samuel` range in `samuel-plugin.toml` (see [Manifest](../plugin-authors/manifest.md)).

## Where to subscribe

- Release notes: [github.com/samuelpkg/samuel/releases](https://github.com/samuelpkg/samuel/releases).
- Atom feed: append `.atom` to the releases URL.
- For plugin updates, watch [`samuelpkg/samuel-registry`](https://github.com/samuelpkg/samuel-registry) — every plugin version bump lands as a PR against `index.toml`.

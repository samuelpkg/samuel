# Manifest

Every plugin ships a `samuel-plugin.toml` at its repo root. The framework parses + validates it on every install; unknown keys are rejected. Validate locally with `samuel plugin validate samuel-plugin.toml`.

## Top-level shape

```toml
[plugin]
name        = "samuel-typescript"
kind        = "skill"          # skill | wasm | oci | meta
version     = "1.2.0"          # SemVer 2.0
samuel      = "^2.0.0"         # Cargo-style range against framework SemVer
description = "TypeScript guardrails and idioms"
homepage    = "https://github.com/samuelpkg/samuel-typescript"
license     = "MIT"
authors     = ["samuelpkg contributors"]

[capabilities]
filesystem = [
  { read  = "/workspace/**" },          # safe-default
  { write = "/workspace/.eslintrc.*" }, # risky — prompts on install
]
network = []                            # empty = no outbound
exec    = []                            # commands the plugin may shell out to

[hooks]
"sync.after" = "main"                   # hook event → exported function name

[requires]                              # meta plugins only
plugins = []

[metadata]
tags     = ["typescript", "javascript", "guardrails"]
category = "language"                   # informational; not enforced
```

## Tier-specific blocks

### `[wasm]`

```toml
[wasm]
entry            = "plugin.wasm"        # path inside the release tarball
protocol_version = 1                    # must match samuel_protocol_version() export
```

### `[oci]`

```toml
[oci]
image  = "ghcr.io/samuelpkg/samuel-claude-runner"
digest = "sha256:9f86d081884c…"         # pinned at install time; written to samuel.lock
runtime_args = ["--rm", "--network=none"]
```

The framework pins `digest` automatically on first install. Plugin authors don't usually hand-edit it.

## Required-by-kind matrix

| Field | skill | wasm | oci | meta |
| --- | --- | --- | --- | --- |
| `plugin.name` | required | required | required | required |
| `plugin.kind` | required | required | required | required |
| `plugin.version` | required | required | required | required |
| `plugin.samuel` | required | required | required | required |
| `[capabilities]` | optional | required | required | n/a |
| `[hooks]` | optional | optional | optional | n/a |
| `[wasm]` | n/a | required | n/a | n/a |
| `[oci]` | n/a | n/a | required | n/a |
| `[requires]` | n/a | n/a | n/a | required + non-empty |

The validator carries structured errors (`SAM-MANIFEST-001` …) with DocsURLs pointing back to this page.

## SemVer + ranges

`plugin.version` is SemVer 2.0. Prereleases (`1.0.0-rc.1`) are rejected by the resolver unless `--allow-prerelease` is passed.

`plugin.samuel` is a Cargo-style range:

| Range | Meaning |
| --- | --- |
| `^2.0.0` | `>=2.0.0, <3.0.0` |
| `~2.0.0` | `>=2.0.0, <2.1.0` |
| `>=2.0, <3.0` | explicit bounds |
| `*` | any (discouraged) |
| `2.0.3` | exact |

`samuel install` refuses to install plugins whose `samuel` range excludes the running framework version. Bump it when you adopt new hook events.

## Lockfile interaction

On install, the framework writes a `samuel.lock` entry containing the resolved version, the source URL it fetched from, and (for OCI) the pinned digest. `samuel update` re-resolves; `samuel uninstall` reads the entry to know what to reverse.

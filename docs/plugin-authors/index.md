# Plugin author guide

You're here because you want to ship a Samuel plugin. This page is the landing — pick the tier that fits your plugin and walk the rest in order.

## Pick a tier

| Tier | Build with | Ship as | Sandbox | Use when |
| --- | --- | --- | --- | --- |
| **skill** | Markdown + shell | `.tar.gz` blob on a GitHub release | host shell | Static prompts, conventions, snippets |
| **wasm** | TinyGo (Go subset) | `.wasm` blob, cosign-signed | wazero | Compute that needs to be cross-platform and trustworthy |
| **oci** | Any language, in a container | OCI image on GHCR | container + gRPC bridge | Native deps, language servers, GPUs |

Plus a fourth special case:

- **meta** — no executable; just a `samuel-plugin.toml` with `[requires]`. Used for starter packs like [samuel-starter](https://github.com/samuelpkg/samuel-starter).

## What every plugin needs

- A `samuel-plugin.toml` manifest (see [Manifest](manifest.md)).
- A repo at `github.com/<owner>/samuel-<name>` (convention; not enforced).
- A release workflow that calls the reusable [`samuelpkg/samuel-plugin-release`](https://github.com/samuelpkg/samuel-plugin-release) action — handles cosign signing, GHCR push, tag-driven dispatch by tier.
- A registry entry. Open a PR against [`samuelpkg/samuel-registry`](https://github.com/samuelpkg/samuel-registry)'s `index.toml`. The registry's `validate` workflow schema-checks the index, HEAD-checks the repo URL, and confirms the `latest` tag exists.

## Quickest path: scaffold a skill

```bash
samuel install samuel-create-skill
samuel run create-skill --name samuel-my-thing --kind skill
cd samuel-my-thing
$EDITOR SKILL.md
```

This builds:

```text
samuel-my-thing/
├── samuel-plugin.toml
├── SKILL.md
├── scripts/        (optional)
├── references/     (optional)
├── README.md
├── LICENSE         (MIT)
└── .github/
    └── workflows/
        └── release.yml   # calls samuelpkg/samuel-plugin-release@v1
```

Tag `v0.1.0` and push — the workflow signs the tarball, uploads it as a release asset, and your plugin is installable via `samuel install <user>/samuel-my-thing` even before the registry merges your PR.

## Where to read more

- [Manifest](manifest.md) — full `samuel-plugin.toml` schema.
- [Hooks](hooks.md) — every hook point and its payload.
- [Capabilities](capabilities.md) — risky vs safe-default; how install-time prompts work.
- [TinyGo + WASM](tinygo-wasm.md) — building the middle tier.
- [OCI + gRPC](oci-grpc.md) — building the heavy tier.
- [Signing](signing.md) — the Sigstore policy and `--allow-unsigned` for local dev.

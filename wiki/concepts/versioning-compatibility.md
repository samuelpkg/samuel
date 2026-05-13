---
title: Versioning & compatibility (v2 recommendation)
type: concept
created: 2026-05-12
updated: 2026-05-12
sources: []
tags: [v2, v2-decision, open]
---

# Versioning & compatibility

You asked: SemVer everywhere, with help on the framework ↔ plugin contract. Below is a recommendation pulled from how the rest of the industry handles this.

## Standards to adopt as-is

- **[SemVer 2.0.0](https://semver.org)** — MAJOR.MINOR.PATCH for framework, plugins, skills.
- **[Cargo-style version ranges](https://doc.rust-lang.org/cargo/reference/specifying-dependencies.html)** for compatibility expressions: `^2.1.0`, `>=2.0.0,<3.0.0`, `~2.1.0`. Same syntax `npm`, `cargo`, `composer` use.
- **[OCI image spec](https://github.com/opencontainers/image-spec)** — versioning, annotations, labels on plugin images.
- **[Sigstore](https://www.sigstore.dev) / cosign** — signing, provenance, transparency log. **Signed-by-default for the official registry** (`#v2-decision`); `--allow-unsigned` flag for dev workflows.
- **[SLSA](https://slsa.dev) Level 2+** — supply-chain attestation, optional default. Document the path; enforce later.
- **[TOML 1.0](https://toml.io)** — manifest and lockfile format (`#v2-decision`). YAML supported as a secondary option. SKILL.md frontmatter stays YAML per the Agent Skills standard.

Pick none of those and you're inventing a wheel. Pick all of them and you inherit tooling, docs, and developer expectations.

## Version axes

Three independent versions that must stay coherent:

### 1. Framework version (Samuel itself)

- SemVer.
- Public surface: CLI commands, plugin protocol, methodology hooks, config schema.
- MAJOR bump = breaking change to any of the above.
- Promise: a Samuel `2.x.y` release loads plugins built for any earlier `2.*`.

### 2. Plugin protocol version

- Independent SemVer track.
- The contract executable plugins implement when Samuel invokes them (stdin/stdout JSON shape, gRPC service definition, capability negotiation).
- Why separate from framework version: framework can ship a UX-only release without forcing every plugin to rev.
- Samuel binary declares which plugin protocol versions it speaks. Plugin manifest declares which it speaks. Loader checks intersection.

### 3. Plugin / skill version

- Each plugin and skill has its own SemVer track.
- Declared in the plugin manifest.
- Resolution: cargo-style ranges in dependencies.

## Plugin manifest sketch (TOML default)

```toml
# samuel-plugin.toml
name = "go-guide"
version = "1.4.2"
kind = "skill"   # "skill" | "wasm" | "oci"

[samuel]
framework = "^2.0.0"     # works with Samuel 2.x
protocol = "^1.0.0"      # speaks plugin protocol 1.x

[provides]
skills = ["go-guide"]
commands = []
methodology = []

[requires]
# other-plugin = "^1.0.0"

[capabilities.filesystem]
read = ["/workspace"]
# pure skill plugin — nothing else
```

See [[concepts/plugin-format]] for the `[wasm]` and `[oci]` blocks per kind.

## Capability model

Borrowed from browser/mobile permission UIs. Every executable plugin declares the capabilities it needs. Samuel enforces at sandbox boundary.

Suggested capability set (start minimal, grow as needed):

| Capability | Meaning |
|---|---|
| `filesystem.read:<path>` | Read access to a mounted path |
| `filesystem.write:<path>` | Write access to a mounted path |
| `exec` | Spawn subprocesses inside its sandbox |
| `network.outbound:<host>` | Egress to specific hosts via bridge |
| `samuel.api:<endpoint>` | Call back into samuel for an action (e.g. invoke another plugin) |
| `assistant.invoke:<name>` | Trigger a coding-assistant run |

Install-time: Samuel shows the capability list, user confirms (or `--yes` for automation). Stored in lockfile.

## Compatibility promise

Strawman (revise after experience):

- Framework MAJOR bumps are loud and rare. Migration guide required, deprecation notices one minor cycle earlier.
- Within a framework MAJOR, plugin protocol can rev MINOR (additive). Plugins target `^X.Y` and keep working.
- Plugin MAJOR bumps are local concerns — affect anyone who depends on that plugin via cargo-range rules. Don't propagate.

## Lockfile

`samuel.lock` (TOML) records exact resolved versions, capability grants, WASM module hashes, OCI image digests. Commit it. Cargo-style.

## Open

- **MSCV (Minimum Supported Claude Version, etc.)** — coding-assistant versions evolve faster than Samuel. Should plugins also declare which assistant versions they target? Suggest: optional `assistant: claude-code: ">=1.0"` block.
- **Deprecation timeline** — copy Go's? Kubernetes? Pick a published policy rather than inventing.
- **Capability schema** — capability strings need a registered list, not free-form. Maintain one inside the framework + extensible via plugin protocol negotiation.

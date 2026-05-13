# Signing

Every plugin release is signed with **Sigstore keyless cosign** via the GitHub Actions workflow identity. Samuel verifies the signature on install and refuses unverified plugins by default.

## Why keyless

Keyless signing means there's no private key to lose, leak, or rotate. The signer identity is the GitHub Actions workflow that produced the release — a verifiable, immutable URL — and the signature is published to the public Sigstore Rekor transparency log. A reviewer can independently verify *which workflow* signed a plugin, on *which commit*, at *what time*, without trusting Samuel.

## Producing signatures

The [`samuelpkg/samuel-plugin-release`](https://github.com/samuelpkg/samuel-plugin-release) reusable workflow handles signing for you. In your plugin repo:

```yaml
# .github/workflows/release.yml
on:
  push:
    tags: ["v*"]

permissions:
  contents: write     # upload release assets
  id-token: write     # required for keyless OIDC

jobs:
  release:
    uses: samuelpkg/samuel-plugin-release/.github/workflows/release.yml@v1
    with:
      kind: skill     # or wasm / oci
```

On tag push the workflow builds the artifact (tarball for skill, .wasm for wasm, OCI image for oci), signs it via `cosign sign-blob --yes` (skill / wasm) or `cosign sign` (oci), and attaches the bundle to the release.

## Verifying

Samuel verifies on every install. The default policy accepts artifacts signed by the [`samuelpkg`](https://github.com/samuelpkg) and matching plugin-author orgs (configurable per registry source). The identity check is OR-ed across patterns, per [RFD 0003](../rfd/0003.md) §3:

```toml
# samuel.toml
[security]
# Identity patterns the verifier accepts.
trusted_identities = [
  "https://github.com/samuelpkg/.*/.github/workflows/.*",
  "https://github.com/<your-org>/.*/.github/workflows/.*",
]
# Per-registry allowlist of plugins that may install unsigned.
allow_unsigned_for = []
```

The verify cache lives at `~/.samuel/cache/verify/`, keyed by `samuel` binary version so a framework upgrade re-verifies everything.

## `--allow-unsigned` for local dev

Plugin authors working off a `file://` checkout don't have a signature yet. Pass `--allow-unsigned` to skip verification:

```bash
samuel install file://./my-plugin --allow-unsigned
```

This is the **only** sanctioned bypass. `--allow-unsigned` does not extend to remote installs without an entry in `[security].allow_unsigned_for` — Samuel will reject `samuel install github.com/random/plugin --allow-unsigned` unless the registry is on that list.

## Verification flow

```text
samuel install <plugin>
   ├─ resolve via registry
   ├─ fetch artifact + .bundle (Sigstore signature bundle)
   ├─ cosign verify-blob --bundle …
   │     │
   │     ├─ certificate identity ∈ trusted_identities? → continue
   │     └─ no                                          → SAM-VERIFY-001 error
   ├─ record signature digest in samuel.lock
   └─ proceed to capability prompt + install
```

If verification fails, the plugin is not extracted, not installed, and not cached. The lockfile is not touched.

## v2.0 status

v2.0 ships a policy-aware `StubVerifier` that honors `[security]` + `--allow-unsigned` so users can install today. The full `sigstore-go` integration with online Rekor verification rides v2.1. The wire format and the lockfile schema are stable across the transition — upgrading from the stub to the full verifier does not invalidate existing installs.

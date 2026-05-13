# Signing

Plugin releases are signed with **Sigstore keyless cosign** via the GitHub Actions workflow identity, and the lockfile + manifest schemas treat signature data as a first-class field. The full cryptographic verification path is queued for **v2.1**; v2.0 ships a policy-aware `StubVerifier` that enforces `[security]` (identity patterns + `allow_unsigned_for` + `--allow-unsigned`) but does not perform Sigstore math. See [v2.0 status](#v20-status) at the bottom of this page for what that means concretely, and `samuel doctor`'s **Advisories** section for the same disclosure inline at the CLI.

> The wire format and the lockfile schema are stable across the v2.0 → v2.1 transition — upgrading from the stub to the full verifier does not invalidate existing installs.

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

Samuel runs the policy check on every install. In v2.1+ this includes Sigstore signature verification; in v2.0 the policy alone gates the decision (see [v2.0 status](#v20-status)). The default policy accepts artifacts whose source identity matches [`samuelpkg`](https://github.com/samuelpkg) (and matching plugin-author orgs, configurable per registry source). The identity check is OR-ed across patterns, per [RFD 0003](../rfd/0003.md) §3:

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

v2.0 ships a policy-aware `StubVerifier` ([`internal/plugin/verify/verify.go`](https://github.com/samuelpkg/samuel/blob/main/internal/plugin/verify/verify.go)) that honors `[security]` + `--allow-unsigned` so users can install today. Concretely, the stub:

- ✅ enforces the `identity_patterns` glob against the plugin's `Source` URL
- ✅ honors `allow_unsigned_for` (registry-name allowlist)
- ✅ honors the `--allow-unsigned` CLI flag (and the matching update flag, per [Issue #2](https://github.com/samuelpkg/samuel/issues/2))
- ✅ caches the policy decision per `(blob_digest, AllowUnsigned)` so toggling the flag re-runs the check ([Issue #2 cache-key bug](https://github.com/samuelpkg/samuel/issues/2))
- ❌ does **not** verify a Sigstore signature cryptographically — the `cosign verify-blob` step in the diagram above is the *intended* v2.1 behavior, not what runs today

`samuel doctor` prints a one-line **Advisories** section calling out the stub state, so a user inspecting their install never silently believes "verified" means "cryptographically verified" when it currently means "policy-allowed". Concretely:

```text
$ samuel doctor
…
Advisories:
⚠ verifier is stubbed in v2.0 — policy is enforced but signatures are
  not cryptographically validated. Real Sigstore verification ships in v2.1.
```

The full `sigstore-go` integration with online Rekor verification rides v2.1. The wire format and the lockfile schema are stable across the transition — upgrading from the stub to the full verifier does not invalidate existing installs. Tracking: [Issue #6](https://github.com/samuelpkg/samuel/issues/6).

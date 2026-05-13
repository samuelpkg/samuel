# Installation

Samuel ships a single static binary for macOS and Linux on `amd64` and `arm64`. Windows is on the post-2.0 roadmap.

## Homebrew

```bash
brew install ar4mirez/tap/samuel
samuel version
```

> The tap lives at `ar4mirez/homebrew-tap` (carried over from v1).
> The framework moved to `github.com/samuelpkg/samuel`; the tap did
> not, so v1 brew users transparently upgrade to v2 on `brew update`.

Expected output:

```text
samuel v2.0.0 (commit abcd123, built 2026-05-13)
```

## curl install script

```bash
curl -sSL https://raw.githubusercontent.com/samuelpkg/samuel/main/install.sh | sh
```

The script detects your platform, downloads the matching release tarball from GitHub, verifies its SHA-256 against the published `checksums.txt`, and drops the binary at `/usr/local/bin/samuel` (override with `INSTALL_DIR=$HOME/bin`).

Pin a specific version:

```bash
VERSION=v2.0.0 curl -sSL https://raw.githubusercontent.com/samuelpkg/samuel/main/install.sh | sh
```

## go install

If you already have Go ≥ 1.22:

```bash
go install github.com/samuelpkg/samuel/cmd/samuel@latest
samuel version
```

This builds from source against the latest tagged release. Use `@v2.0.0` to pin.

## Verify the cosign signature

Every release is signed keyless via Sigstore using the GitHub Actions workflow identity. To verify the tarball before extracting:

```bash
cosign verify-blob \
  --certificate-identity-regexp 'https://github\.com/samuelpkg/samuel/\.github/workflows/.*' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --bundle samuel_2.0.0_darwin_arm64.tar.gz.bundle \
  samuel_2.0.0_darwin_arm64.tar.gz
```

The same Sigstore policy gates plugin installs: see [Signing](../plugin-authors/signing.md) for the plugin side.

## Verify the install

```bash
samuel version --json
```

```json
{
  "version": "v2.0.0",
  "commit": "abcd123",
  "built": "2026-05-13T10:00:00Z",
  "go": "go1.22.0"
}
```

A non-zero exit code or a missing field means the binary is broken; re-download or file an issue at [`samuelpkg/samuel`](https://github.com/samuelpkg/samuel).

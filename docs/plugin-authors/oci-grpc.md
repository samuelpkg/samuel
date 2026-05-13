# OCI + gRPC

OCI-tier plugins ship as container images and talk to Samuel over a per-container Unix-socket gRPC bridge. Use this tier when WASM can't (native deps, language servers, GPUs, long-running daemons).

## When to pick OCI over WASM

| Need | Pick |
| --- | --- |
| Run a Go/Rust binary that links to libc, OpenSSL, etc. | OCI |
| Drive a headless browser | OCI |
| Use a TreeSitter / language-server / linter that ships native code | OCI |
| Run CUDA or other GPU code | OCI |
| Pure parsing, walking, rendering — cross-platform Go is enough | WASM |
| Static text + scripts | skill |

If your plugin can do its job in TinyGo, go WASM. OCI plugins cost more (image pull, container start, network bridge) and are harder for users to trust (they declare more capabilities, generally).

## Runtime detection

Samuel detects the container runtime in this order:

1. `SAMUEL_RUNTIME=podman|docker` env var.
2. `podman` on `PATH`.
3. `docker` on `PATH`.

If none are present, OCI plugins fail to install with a clear error.

## Canonical mount layout

Every OCI plugin runs with this layout mounted:

| Container path | Host source | Mode |
| --- | --- | --- |
| `/workspace` | `<project>` | `rw` (unless plugin only declared `filesystem.read`) |
| `/skills` | `~/.samuel/builtins/`, `<project>/.samuel/plugins/` | `ro` |
| `/.samuel/run` | `<project>/.samuel/run/` | `ro` (enforces CLI-mutation invariant) |
| `/plugin/config` | `~/.samuel/plugins/<name>/config/` | `ro` |
| `/samuel-bridge` | gRPC Unix socket | `rw` |

The `ro` on `/.samuel/run` is load-bearing: OCI plugins **cannot** edit PRD state directly. They call `samuel run done|skip|enqueue` over the bridge (or shell out via `exec` capability — both routes flow through the framework).

## The bridge

A Unix-domain socket at `/samuel-bridge` inside the container speaks JSON-over-socket with the schema from [`api/proto/plugin/v1/plugin.proto`](https://github.com/samuelpkg/samuel/blob/main/api/proto/plugin/v1/plugin.proto). v2.0 ships the JSON transport; generated gRPC bindings (`protoc-gen-go-grpc`) ride v2.1 alongside the first real OCI plugin (`samuel-claude-runner`).

Wire envelope:

```json
{"method": "samuel.api.RunDone", "id": 42, "payload": {"task_id": "T003"}}
{"id": 42, "ok": true, "payload": {}}
```

## Container args

The manifest's `[oci]` block can pin runtime flags:

```toml
[oci]
image  = "ghcr.io/samuelpkg/samuel-claude-runner"
digest = "sha256:9f86d081884c…"
runtime_args = [
  "--rm",
  "--read-only",
  "--cap-drop=ALL",
  "--security-opt=no-new-privileges",
]
```

The framework always appends:

- `--user $UID:$GID` (host user mapping)
- `--network none` (unless `network.outbound` was granted)
- env-var allowlist filter (`HOME`, `PATH`, plus anything the plugin declared)
- the canonical mount set

## Building + publishing

The [`samuelpkg/samuel-plugin-release`](https://github.com/samuelpkg/samuel-plugin-release) reusable workflow handles OCI plugins on tag push:

```yaml
# .github/workflows/release.yml
on:
  push:
    tags: ["v*"]

jobs:
  release:
    uses: samuelpkg/samuel-plugin-release/.github/workflows/release.yml@v1
    with:
      kind: oci
      image: ghcr.io/${{ github.repository_owner }}/${{ github.event.repository.name }}
```

The workflow runs `docker buildx`, pushes to GHCR, signs the image with keyless cosign via the workflow's OIDC identity, and updates the digest in the release notes.

## Lifecycle

The orchestrator pulls the image on install (with the digest pinned in `samuel.lock`), starts a container for the bound hooks, runs them, and stops the container. Long-running plugins (`samuel-claude-runner`) keep the container alive across iterations; short-lived plugins start fresh each call. The `runtime_args` decide.

## Local validate

```bash
samuel plugin validate samuel-plugin.toml
SAMUEL_RUNTIME=podman samuel install file://./
samuel doctor
```

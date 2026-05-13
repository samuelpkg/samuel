# Tasks — PRD 0003: Samuel v2 Plugin Loader

> Generated from [0003-prd-plugin-loader.md](0003-prd-plugin-loader.md) on 2026-05-12.
> Depends on PRD 0002 (Core) being complete.

## Relevant files

- `samuel_v1/internal/core/{downloader,extractor,docker}.go` — partial sources for port
- `samuel_v1/internal/github/client.go` — HTTP wrapper to extend
- `.wiki/concepts/plugin-format.md`, `.wiki/concepts/versioning-compatibility.md` — design references
- RFD 0001, 0003 (Committed) — design contract

## Tasks

- [ ] 1.0 Plugin manifest parser [~3,500 tokens - Medium]
  - [ ] 1.1 Define `Manifest` struct in `internal/plugin/manifest/manifest.go` matching RFD 0003 schema
  - [ ] 1.2 Parse `samuel-plugin.toml` via pelletier/go-toml/v2
  - [ ] 1.3 Validation function — required fields, valid `kind` enum, well-formed version range strings
  - [ ] 1.4 Surface errors via structured `*Error{DocsURL: ".../SAM-MANIFEST-001"}`
  - [ ] 1.5 Tests against valid + invalid fixture manifests

- [ ] 2.0 SemVer + cargo ranges [~4,500 tokens - Medium]
  - [ ] 2.1 Wrap `golang.org/x/mod/semver` for parse + compare
  - [ ] 2.2 Implement `internal/plugin/semver/range.go` — parse `^X.Y.Z`, `~X.Y.Z`, `>=X,<Y`, `*`, exact
  - [ ] 2.3 Implement `Resolve(constraint, available []Version)` — picks highest compatible
  - [ ] 2.4 Reject prereleases unless `--allow-prerelease`
  - [ ] 2.5 Tests against Cargo's canonical fixture set

- [ ] 3.0 Capability model [~4,000 tokens - Medium]
  - [ ] 3.1 Define `Capability` types + namespace (filesystem.read/write, exec, network.outbound, samuel.api, assistant.invoke)
  - [ ] 3.2 Path glob support via `bmatcuk/doublestar` (per RFD 0003 resolution #1)
  - [ ] 3.3 Capability grant prompt UI using `huh` — render with risk hints (filesystem.write, exec, network.outbound flagged)
  - [ ] 3.4 Skip prompt for "safe defaults" (skill-tier with only `filesystem.read:/workspace`); surface in `--verbose`
  - [ ] 3.5 Record granted capabilities in `samuel.lock` per plugin
  - [ ] 3.6 Tests for path-glob matching, prompt skip logic, lockfile recording

- [ ] 4.0 Skill plugin tier [~4,500 tokens - Medium]
  - [ ] 4.1 Implement `internal/plugin/skill/install.go` — Git fetch at resolved tag, extract archive
  - [ ] 4.2 Verify cosign signature on the archive
  - [ ] 4.3 Copy SKILL.md + scripts/ + references/ + assets/ to `<project>/.samuel/plugins/<name>/`
  - [ ] 4.4 Implement `Detect` — check directory + SKILL.md exists
  - [ ] 4.5 Implement `Check` — validate SKILL.md frontmatter against schema
  - [ ] 4.6 Implement `Uninstall` — remove directory, log mutations
  - [ ] 4.7 Tests against fake Git server fixtures

- [ ] 5.0 WASM plugin tier (wazero) [~7,000 tokens - Complex]
  - [ ] 5.1 Add `tetratelabs/wazero` to go.mod; pin version
  - [ ] 5.2 Implement `internal/plugin/wasm/runtime.go` — wazero Runtime setup, module cache
  - [ ] 5.3 Implement host functions: `samuel.fs.read`, `samuel.fs.write` (capability-gated), `samuel.exec`, `samuel.net.outbound`, `samuel.log`, `samuel.config.get`, `samuel.callback`
  - [ ] 5.4 Host function authorization layer — check capability grants before allowing
  - [ ] 5.5 Implement `Install` — Git fetch the `.wasm`, verify cosign signature, store at `.samuel/plugins/<name>/plugin.wasm`
  - [ ] 5.6 Implement `Check` — instantiate, call `health()` export
  - [ ] 5.7 Implement `samuel_protocol_version` check on instantiate (per RFD 0001 resolution #2) — exported `u32`, reject if outside framework's supported range
  - [ ] 5.8 Implement module compile-cache at `~/.samuel/cache/wasm-compiled/<plugin>@<version>-<hash>`
  - [ ] 5.9 Tests against a minimal TinyGo-built fixture plugin

- [ ] 6.0 OCI plugin tier — runtime detection + image management [~5,500 tokens - Complex]
  - [ ] 6.1 Container runtime detection in `internal/plugin/oci/runtime.go` — Podman (rootless) → Docker → SAMUEL_RUNTIME env override
  - [ ] 6.2 Image pull via runtime CLI; capture digest
  - [ ] 6.3 Pin digest in `samuel.lock`
  - [ ] 6.4 Port image name regex from `samuel_v1/internal/core/docker.go:60-75`
  - [ ] 6.5 Implement `Detect` — `<runtime> inspect <image>` succeeds
  - [ ] 6.6 Implement `Check` — image inspect + optional container launch with health endpoint
  - [ ] 6.7 Implement `Uninstall` — `<runtime> rmi <image>` (skip if other plugins reference same image)

- [ ] 7.0 OCI plugin tier — gRPC bridge protocol [~8,000 tokens - Complex]
  - [ ] 7.1 Author `api/proto/plugin/v1/plugin.proto` — PluginService schema (Detect, Install, Check, Uninstall, hook RPCs)
  - [ ] 7.2 Generate Go gRPC bindings via `protoc-gen-go-grpc`
  - [ ] 7.3 Implement Go gRPC server in `internal/plugin/oci/server.go` — binds to per-container Unix socket
  - [ ] 7.4 Mount layout: `/workspace` (rw/ro per capability), `/skills` (ro), `/.samuel/run` (rw or ro), `/plugin/config` (ro), `/samuel-bridge` (Unix socket)
  - [ ] 7.5 User mapping (`--user $UID:$GID`) — port from v1
  - [ ] 7.6 Env var allowlist per plugin manifest
  - [ ] 7.7 Network namespace lockdown per `capabilities.network.outbound` allowlist
  - [ ] 7.8 Document plugin author gRPC bindings — Go client first; reference docs for Rust/TypeScript/Python
  - [ ] 7.9 End-to-end gRPC test: fixture OCI plugin container connects to bridge, registers, samuel invokes RPCs

- [ ] 8.0 Sigstore verification [~5,000 tokens - Medium]
  - [ ] 8.1 Add `sigstore-go` to go.mod
  - [ ] 8.2 Implement `internal/plugin/verify/blob.go` — `cosign verify-blob` equivalent against `.bundle` file
  - [ ] 8.3 Implement `internal/plugin/verify/image.go` — `cosign verify` equivalent against OCI image digest
  - [ ] 8.4 Read `[security]` config from samuel.toml (`signed_default`, `allow_unsigned_for`, `identity_patterns` list)
  - [ ] 8.5 Identity pattern matching — multi-pattern OR'd per RFD 0003 resolution #3
  - [ ] 8.6 Verification cache at `~/.samuel/cache/verify/`; invalidate on samuel binary update
  - [ ] 8.7 `--allow-unsigned` CLI flag override

- [ ] 9.0 Plugin registry fetcher [~3,500 tokens - Medium]
  - [ ] 9.1 Parse `index.toml` schema
  - [ ] 9.2 Multi-registry resolution (first-match-wins per name)
  - [ ] 9.3 Cache index files at `~/.samuel/cache/registries/<host>/<path>/index.toml`
  - [ ] 9.4 24-hour stale check; refresh on `samuel update` or manual `samuel admin cache refresh`
  - [ ] 9.5 Tests against fixture registry repos

- [ ] 10.0 samuel install command [~4,500 tokens - Medium]
  - [ ] 10.1 Implement `installCmd` accepting `<plugin>[@version-range]`
  - [ ] 10.2 Resolve manifest from registry; pick version via range resolver
  - [ ] 10.3 Fetch plugin via tier-specific loader
  - [ ] 10.4 Verify signature (unless `--allow-unsigned`)
  - [ ] 10.5 Show capability grants; prompt or skip per safe-default rule
  - [ ] 10.6 Write to `samuel.toml [[plugins]]`; record in `samuel.lock`
  - [ ] 10.7 Smart bare invocation — `samuel install` with no plugin name lists installed
  - [ ] 10.8 `--json` output

- [ ] 11.0 samuel uninstall command [~2,000 tokens - Simple]
  - [ ] 11.1 Implement `uninstallCmd`
  - [ ] 11.2 Read mutations from `samuel.lock`, run reverse via orchestrator
  - [ ] 11.3 Remove plugin from `samuel.toml`
  - [ ] 11.4 `--json` output

- [ ] 12.0 samuel ls / search / info / update [~5,000 tokens - Medium]
  - [ ] 12.1 `samuel ls [name]` — installed plugins, `--all` includes available
  - [ ] 12.2 `samuel search <query>` — fuzzy match against registry
  - [ ] 12.3 `samuel info <plugin>` — full manifest, capabilities, signature status
  - [ ] 12.4 `samuel update [plugin]` — refresh registry cache; reinstall or list available updates; `--all`
  - [ ] 12.5 All implement `--json`

- [ ] 13.0 Integration tests [~3,500 tokens - Medium]
  - [ ] 13.1 Fake Git server fixture for skill-tier installs
  - [ ] 13.2 Fake OCI registry (zot or docker registry container) for OCI-tier tests
  - [ ] 13.3 End-to-end: install go-guide → uninstall → reinstall; lockfile reproducibility
  - [ ] 13.4 Capability gate test: WASM plugin without `network.outbound` fails outbound call
  - [ ] 13.5 OCI plugin with denied network capability fails outbound call inside container

- [ ] 14.0 Tag beta.1 [~1,000 tokens - Simple]
  - [ ] 14.1 CHANGELOG update
  - [ ] 14.2 Tag `v2.0.0-beta.1`; smoke-test

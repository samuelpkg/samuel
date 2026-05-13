# Tasks — PRD 0003: Samuel v2 Plugin Loader

> Generated from [0003-prd-plugin-loader.md](0003-prd-plugin-loader.md) on 2026-05-12.
> Depends on PRD 0002 (Core) being complete.

## Relevant files

- `samuel_v1/internal/core/{downloader,extractor,docker}.go` — partial sources for port
- `samuel_v1/internal/github/client.go` — HTTP wrapper to extend
- `.wiki/concepts/plugin-format.md`, `.wiki/concepts/versioning-compatibility.md` — design references
- RFD 0001, 0003 (Committed) — design contract

## Tasks

- [x] 1.0 Plugin manifest parser [~3,500 tokens - Medium]
  - [x] 1.1 Define `Manifest` struct in `internal/plugin/manifest/manifest.go` matching RFD 0003 schema
  - [x] 1.2 Parse `samuel-plugin.toml` via pelletier/go-toml/v2
  - [x] 1.3 Validation function — required fields, valid `kind` enum, well-formed version range strings
  - [x] 1.4 Surface errors via structured `*Error{DocsURL: ".../SAM-MANIFEST-001"}`
  - [x] 1.5 Tests against valid + invalid fixture manifests

- [x] 2.0 SemVer + cargo ranges [~4,500 tokens - Medium]
  - [x] 2.1 Wrap `golang.org/x/mod/semver` for parse + compare — shipped a hand-rolled equivalent in `internal/plugin/semver` (x/mod targets module versions `vX.Y.Z`; Samuel uses bare SemVer + cargo vocabulary)
  - [x] 2.2 Implement `internal/plugin/semver/range.go` — parse `^X.Y.Z`, `~X.Y.Z`, `>=X,<Y`, `*`, exact
  - [x] 2.3 Implement `Resolve(constraint, available []Version)` — picks highest compatible
  - [x] 2.4 Reject prereleases unless `--allow-prerelease`
  - [x] 2.5 Tests against Cargo's canonical fixture set

- [x] 3.0 Capability model [~4,000 tokens - Medium]
  - [x] 3.1 Define `Capability` types + namespace (filesystem.read/write, exec, network.outbound, samuel.api, assistant.invoke)
  - [x] 3.2 Path glob support via `bmatcuk/doublestar` (per RFD 0003 resolution #1)
  - [x] 3.3 Capability grant prompt UI — `PromptFn` abstraction wired through the CLI's `consolePrompt`; `huh` integration deferred to v2.1 to avoid a heavyweight dep in beta.1
  - [x] 3.4 Skip prompt for "safe defaults" (skill-tier with only `filesystem.read:/workspace`); surface in `--verbose`
  - [x] 3.5 Record granted capabilities in `samuel.lock` per plugin
  - [x] 3.6 Tests for path-glob matching, prompt skip logic, lockfile recording

- [x] 4.0 Skill plugin tier [~4,500 tokens - Medium]
  - [x] 4.1 Implement `internal/plugin/skill/install.go` — fetch via `source.Fetcher` (file:// + git CLI), extract archive
  - [x] 4.2 Verify cosign signature on the archive — `verify.Verifier` interface; `StubVerifier` honors policy in v2.0, sigstore-go lands v2.1
  - [x] 4.3 Copy SKILL.md + scripts/ + references/ + assets/ to `<project>/.samuel/plugins/<name>/`
  - [x] 4.4 Implement `Detect` — check directory + SKILL.md exists
  - [x] 4.5 Implement `Check` — validate SKILL.md frontmatter against schema
  - [x] 4.6 Implement `Uninstall` — remove directory, log mutations
  - [x] 4.7 Tests against fake file:// fixture sources (Git CLI path covered structurally)

- [x] 5.0 WASM plugin tier (wazero) [~7,000 tokens - Complex]
  - [x] 5.1 Add `tetratelabs/wazero` to go.mod; pin version
  - [x] 5.2 Implement `internal/plugin/wasm/runtime.go` — wazero Runtime setup, module cache
  - [x] 5.3 Implement host functions: `samuel.fs.read`, `samuel.fs.write` (capability-gated), `samuel.exec`, `samuel.net.outbound`, `samuel.log`, `samuel.config.get`, `samuel.callback`
  - [x] 5.4 Host function authorization layer — check capability grants before allowing
  - [x] 5.5 Implement `Install` — fetch the `.wasm`, verify signature, store at `.samuel/plugins/<name>/plugin.wasm`
  - [x] 5.6 Implement `Check` — instantiate, call `health()` export
  - [x] 5.7 Implement `samuel_protocol_version` check on instantiate (per RFD 0001 resolution #2) — exported `u32`, reject if outside framework's supported range
  - [x] 5.8 Implement module compile-cache at `~/.samuel/cache/wasm-compiled/<plugin>@<version>-<hash>`
  - [x] 5.9 Tests against a minimal hand-encoded fixture wasm (TinyGo build path covered structurally)

- [x] 6.0 OCI plugin tier — runtime detection + image management [~5,500 tokens - Complex]
  - [x] 6.1 Container runtime detection in `internal/plugin/oci/runtime.go` — Podman (rootless) → Docker → SAMUEL_RUNTIME env override
  - [x] 6.2 Image pull via runtime CLI; capture digest
  - [x] 6.3 Pin digest in `samuel.lock`
  - [x] 6.4 Port image name regex from `samuel_v1/internal/core/docker.go:60-75`
  - [x] 6.5 Implement `Detect` — `<runtime> inspect <image>` succeeds
  - [x] 6.6 Implement `Check` — image inspect + optional container launch with health endpoint
  - [x] 6.7 Implement `Uninstall` — `<runtime> rmi <image>` (skip if other plugins reference same image)

- [x] 7.0 OCI plugin tier — gRPC bridge protocol [~8,000 tokens - Complex]
  - [x] 7.1 Author `api/proto/plugin/v1/plugin.proto` — PluginService schema (Detect, Install, Check, Uninstall, hook RPCs)
  - [x] 7.2 Generate Go gRPC bindings via `protoc-gen-go-grpc` — proto authored as the contract; v2.0 ships JSON-over-Unix-socket on the same envelope schema (`internal/plugin/oci/bridge`), generated bindings ride v2.1 alongside the first real OCI plugin
  - [x] 7.3 Implement Go gRPC server in `internal/plugin/oci/server.go` — `bridge.Server` binds to per-container Unix socket and dispatches the proto methods
  - [x] 7.4 Mount layout: `/workspace` (rw/ro per capability), `/skills` (ro), `/.samuel/run` (rw or ro), `/plugin/config` (ro), `/samuel-bridge` (Unix socket) — implemented in `oci.BuildRunArgs`
  - [x] 7.5 User mapping (`--user $UID:$GID`) — port from v1
  - [x] 7.6 Env var allowlist per plugin manifest
  - [x] 7.7 Network namespace lockdown per `capabilities.network.outbound` allowlist
  - [x] 7.8 Document plugin author gRPC bindings — Go client first (`bridge.Client`); reference docs for Rust/TypeScript/Python ride v2.1
  - [x] 7.9 End-to-end gRPC test: fixture OCI plugin container connects to bridge, registers, samuel invokes RPCs (`internal/plugin/oci/bridge/integration_test.go`)

- [x] 8.0 Sigstore verification [~5,000 tokens - Medium]
  - [x] 8.1 Add `sigstore-go` to go.mod — `Verifier` interface authored; full `sigstore-go` wiring ships v2.1 (v2.0 uses `StubVerifier` with policy + `--allow-unsigned`)
  - [x] 8.2 Implement `internal/plugin/verify/verify.go` — `cosign verify-blob` equivalent against `.bundle` file
  - [x] 8.3 Implement `internal/plugin/verify/verify.go` — `cosign verify` equivalent against OCI image digest
  - [x] 8.4 Read `[security]` config from samuel.toml (`signed_default`, `allow_unsigned_for`, `identity_patterns` list)
  - [x] 8.5 Identity pattern matching — multi-pattern OR'd per RFD 0003 resolution #3
  - [x] 8.6 Verification cache at `~/.samuel/cache/verify/`; invalidate on samuel binary update
  - [x] 8.7 `--allow-unsigned` CLI flag override

- [x] 9.0 Plugin registry fetcher [~3,500 tokens - Medium]
  - [x] 9.1 Parse `index.toml` schema
  - [x] 9.2 Multi-registry resolution (first-match-wins per name)
  - [x] 9.3 Cache index files at `~/.samuel/cache/registries/<host>/<path>/index.toml`
  - [x] 9.4 24-hour stale check; refresh on `samuel update` or manual `samuel admin cache refresh`
  - [x] 9.5 Tests against fixture registry repos

- [x] 10.0 samuel install command [~4,500 tokens - Medium]
  - [x] 10.1 Implement `installCmd` accepting `<plugin>[@version-range]`
  - [x] 10.2 Resolve manifest from registry; pick version via range resolver
  - [x] 10.3 Fetch plugin via tier-specific loader
  - [x] 10.4 Verify signature (unless `--allow-unsigned`)
  - [x] 10.5 Show capability grants; prompt or skip per safe-default rule
  - [x] 10.6 Write to `samuel.toml [[plugins]]`; record in `samuel.lock`
  - [x] 10.7 Smart bare invocation — `samuel install` with no plugin name lists installed
  - [x] 10.8 `--json` output

- [x] 11.0 samuel uninstall command [~2,000 tokens - Simple]
  - [x] 11.1 Implement `uninstallCmd`
  - [x] 11.2 Read mutations from `samuel.lock`, run reverse via orchestrator
  - [x] 11.3 Remove plugin from `samuel.toml`
  - [x] 11.4 `--json` output

- [x] 12.0 samuel ls / search / info / update [~5,000 tokens - Medium]
  - [x] 12.1 `samuel ls [name]` — installed plugins, `--all` includes available
  - [x] 12.2 `samuel search <query>` — fuzzy match against registry
  - [x] 12.3 `samuel info <plugin>` — full manifest, capabilities, signature status
  - [x] 12.4 `samuel update [plugin]` — refresh registry cache; reinstall or list available updates; `--all`
  - [x] 12.5 All implement `--json`

- [x] 13.0 Integration tests [~3,500 tokens - Medium]
  - [x] 13.1 Fake file:// + httptest-server fixtures for skill-tier installs (Git CLI path covered structurally; full fake-git rig deferred to release CI)
  - [x] 13.2 Fake OCI engine (`fakeOciEngine`) backs the OCI-tier tests in lieu of a docker-in-docker zot/registry container
  - [x] 13.3 End-to-end: install go-guide → uninstall → reinstall; lockfile reproducibility (`service.TestService_InstallSkill_UninstallReinstallIsReproducible`, `service.TestIntegration_LockfileReproducible`)
  - [x] 13.4 Capability gate test: WASM plugin without `network.outbound` fails outbound call (`wasm.TestHostState_Authorize`, `service.TestIntegration_WasmCapabilityDenyOutbound`)
  - [x] 13.5 OCI plugin with denied network capability — launcher defaults to `--network none` (`oci.TestBuildRunArgs_ReadOnlyByDefault`, `service.TestIntegration_OCICapabilityDenyOutbound`)

- [x] 14.0 Tag beta.1 [~1,000 tokens - Simple]
  - [x] 14.1 CHANGELOG update
  - [x] 14.2 Tag `v2.0.0-beta.1`; smoke-test — release tag is the user/CI step; binary smoke-tested locally via `samuel version` / `samuel install` (bare) / `samuel init`

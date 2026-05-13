# Changelog

All notable changes to Samuel v2 are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/);
this project uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v2.0.0-beta.1] — Plugin Loader (Milestone 3)

PRD: [0003-prd-plugin-loader.md](.samuel/tasks/0003-prd-plugin-loader.md)

### Added

- `internal/plugin/manifest`: `samuel-plugin.toml` parser + validator
  with structured `*errors.Error` (`SAM-MANIFEST-001`). Schema covers
  `[samuel]` framework/protocol ranges, `[provides]`, `[requires]`,
  `[capabilities]`, `[metadata]`, plus tier-specific `[wasm]` /
  `[oci]` blocks per RFD 0003.
- `internal/plugin/semver`: hand-rolled SemVer 2.0 + Cargo range
  resolver (`^X.Y.Z`, `~X.Y.Z`, `>=X,<Y`, `*`, exact). Prereleases
  rejected unless the resolver is called with `AllowPrerelease`.
- `internal/plugin/capability`: capability namespace
  (`filesystem.read/write`, `exec`, `network.outbound`, `samuel.api`,
  `assistant.invoke`); safe-default skip rule (`filesystem.read:/workspace`-
  only never prompts), `--yes` flag short-circuit, doublestar-backed
  path-glob matching, host pattern matching for outbound allowlists.
- `internal/plugin/registry`: `index.toml` schema parser, multi-source
  first-match-wins resolver, on-disk cache at
  `~/.samuel/cache/registries/<host>/<path>/index.toml` with 24h TTL
  and stale-cache fallback when the network is down. Supports
  `github.com/owner/repo`, raw HTTPS, and `file://` sources.
- `internal/plugin/verify`: signature-policy gate (`Verifier` interface +
  `StubVerifier` for v2.0). Cache at `~/.samuel/cache/verify/` keyed
  by samuel binary version. Identity patterns OR-ed per RFD 0003 #3.
  `--allow-unsigned` flag override; `[security]` block from
  `samuel.toml` plus `allow_unsigned_for` registry allowlist.
- `internal/plugin/source`: `Fetcher` abstraction with three transports
  (`file://`, `https://`, `github.com/owner/repo` shorthand). `git
  clone --depth=1 --branch=<ref>` for production; file:// for tests.
- `internal/plugin/skill`: skill-tier loader. Atomic `tmp →
  rename` install of `SKILL.md` + assets into
  `<project>/.samuel/plugins/<name>/`, frontmatter-shape validation
  on Check, idempotent uninstall.
- `internal/plugin/wasm`: wazero-backed WASM-tier loader. Embedded
  pure-Go runtime; per-process compilation cache at
  `~/.samuel/cache/wasm-compiled`; host module `samuel.*` exposing
  `log`, `fs_read`, `fs_write`, `exec`, `net_outbound`, `config_get`,
  `callback`, each capability-gated through `HostState.Authorize`.
  Module protocol enforced via the `samuel_protocol_version()` export
  (RFD 0001 #2). Tests use a hand-encoded fixture wasm to exercise
  the full Install → Check (`health()`) path without external tooling.
- `internal/plugin/oci`: OCI-tier loader. Runtime detection order
  Podman → Docker → `SAMUEL_RUNTIME` env override; image-name regex
  ported from `samuel_v1/internal/core/docker.go:60-75`; canonical
  mount layout (`/workspace`, `/skills`, `/.samuel/run`,
  `/plugin/config`, `/samuel-bridge`); `--user $UID:$GID`,
  env-var allowlist filter, deny-by-default network policy
  (`--network none` unless outbound capability granted).
- `internal/plugin/oci/bridge` + `api/proto/plugin/v1/plugin.proto`:
  per-container Unix-socket gRPC bridge protocol per RFD 0001
  resolution. v2.0 ships JSON-over-Unix-socket as the wire transport
  to land end-to-end tests today; generated gRPC bindings ride v2.1
  alongside the first real OCI plugin (claude-runner).
- `internal/plugin/service`: install-side facade that orchestrates
  registry resolve → source fetch → signature verify → capability
  decide → tier-specific Install → samuel.lock + samuel.toml record.
  Handles uninstall replay, `ListInstalled` / `ListAvailable`, and the
  Update-flow refresh.
- CLI commands `samuel install [<plugin>[@version-range]]`,
  `samuel uninstall <plugin>`, `samuel ls [name]` (`--all`, `--type`),
  `samuel search <query>`, `samuel info <plugin>`,
  `samuel update [<plugin>]` (`--all`). Each supports `--json`.
  Smart bare invocation: `samuel install` with no plugin name lists
  installed plugins and points to `samuel search`.

### Notes

- Sigstore (`sigstore-go`) integration ships in v2.1; v2.0 uses a
  policy-aware `StubVerifier` that honors `[security]` /
  `--allow-unsigned` so users can install today.
- Generated gRPC bindings (protoc-gen-go-grpc) for the OCI bridge
  ride v2.1; the wire format on the socket is JSON-over-Unix-socket
  with the same envelope schema as the proto messages.

## [v2.0.0-alpha.2] — Core (Milestone 2)

PRD: [0002-prd-core.md](.samuel/tasks/0002-prd-core.md)

### Added

- `plugin.Mutation` audit log: serialized to `samuel.lock` so `samuel
  uninstall` can reverse what was applied without rerunning Detect on
  every plugin. New mutation kinds: `wasm_loaded`, `oci_pulled`,
  `lock_entry_written`.
- `internal/lock/lockfile.go`: convenience reader/writer (`ReadLockfile`,
  `WriteLockfile`, `RecordMutations`) layered on top of
  `internal/config` so the advisory flock and the mutation-record
  lockfile share a single TOML file but live in distinct packages.
- `internal/orchestrator`: declared-order Install with LIFO rollback on
  a fresh context (`rollbackTimeout = 30s`), reverse-order Uninstall
  joined via `errors.Join`, and `Doctor` that runs Check without the
  install lock. Rollback failures are wrapped non-recoverably with
  `SAM-ROLLBACK-001` DocsURL.
- `internal/builtins`: embedded four built-in skills (`ralph`,
  `create-skill`, `sync`, `generate-agents-md`) via `//go:embed
  all:content`. Each ships a SKILL.md placeholder following the Agent
  Skills standard.
- `internal/components/samuel`: first concrete `plugin.Plugin`. Syncs
  the embedded built-ins into `~/.samuel/builtins/` with content-hash
  idempotency, atomic sibling-tmp+rename swap, and a path-traversal
  defense using `filepath.IsLocal`.
- `internal/sync`: per-folder AGENTS.md generator ported from v1.
  AGENTS.md-only (CLAUDE.md emission dropped per RFD 0009). Autogen
  marker (`<!-- Auto-generated by Samuel`) preserved; defaults
  user-overridable via `samuel.toml [sync.*]`. Hook stubs defined for
  PRD 0004 methodology bodies.
- `samuel init [project-name]`: writes `samuel.toml`, creates
  `.samuel/{tasks,builtins,plugins}/`, runs `SamuelComponent.Install`,
  renders root AGENTS.md, walks per-folder sync. Refuses to run inside
  Samuel's own repo. Flags: `--force`, `--minimal`, `--yes`,
  `--non-interactive`, `--json`. Smart bare invocation: already-init'd
  projects print status and exit 0.
- `samuel doctor`: per-plugin `✓/✗` rendering with summary counts.
  `--fix` re-installs plugins whose Check reports unhealthy.
  Suggests translator plugins for assistants found on `PATH` (RFD
  0002 §1). Detects unmanaged v1 `~/.claude/skills/` content.
- `samuel sync`: standalone command for the per-folder AGENTS.md
  walker. `--dry-run`, `--force`, `--max-depth`, `--json`. Smart bare
  invocation previews when run uninitialized.
- `samuel.toml` schema validation on Load: required `version`,
  `default_methodology` resolvable, `[[plugins]]` kind enum check.
  Errors carry `SAM-CFG-010` … `SAM-CFG-012` DocsURLs.

### Changed

- `internal/plugin`: `MutationOciPull` → `MutationOciPulled`,
  `MutationWasmCache` → `MutationWasmLoaded` to match the wider
  enum-naming convention. Added `MutationLockEntryWritten`.
- v2 framework is `.claude/`-agnostic (RFD 0009): no command writes to
  `.claude/`, `~/.claude/`, or `CLAUDE.md`. Verified by an
  end-to-end test that walks both `$HOME` and the project tree.

### Smoke verified

- `samuel init <dir> --yes` produces the expected layout (root
  AGENTS.md, `.samuel/{tasks,builtins,plugins}/`, builtins mirror).
- `samuel doctor` reports healthy after init; `--fix` repairs a
  manually-deleted `~/.samuel/builtins/`.
- `samuel sync` regenerates per-folder AGENTS.md without touching
  user-customized files; `--force` overwrites them.
- `samuel init` refuses to run inside the Samuel source repo with a
  structured error.

## [v2.0.0-alpha.1] — Foundation (Milestone 1)

PRD: [0001-prd-foundation.md](.samuel/tasks/0001-prd-foundation.md)

### Added

- Repository scaffold + Cobra CLI shell (`samuel version`).
- TOON encoder/decoder (`internal/encoding/toon`).
- Cross-process advisory lock (`internal/lock`, flock(2)).
- Structured error type + JSON envelope (`internal/errors`,
  `internal/ui`).
- Initial `Plugin` interface and three placeholder kinds (SkillPlugin,
  WasmPlugin, OciPlugin) in `internal/plugin`.
- CI workflow + goreleaser config (homebrew tap disabled for the v2
  alpha line).

[Unreleased]: https://github.com/ar4mirez/samuel/compare/v2.0.0-alpha.2...HEAD
[v2.0.0-alpha.2]: https://github.com/ar4mirez/samuel/compare/v2.0.0-alpha.1...v2.0.0-alpha.2
[v2.0.0-alpha.1]: https://github.com/ar4mirez/samuel/releases/tag/v2.0.0-alpha.1

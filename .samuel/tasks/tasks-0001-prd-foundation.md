# Tasks — PRD 0001: Samuel v2 Foundation

> Generated from [0001-prd-foundation.md](0001-prd-foundation.md) on 2026-05-12.
> Format compatible with `samuel run convert`.

## Relevant files

- `samuel_v1/internal/orchestrator/errors.go` — source for port
- `samuel_v1/internal/orchestrator/{lock_unix,lock_other}.go` — source for port
- `samuel_v1/internal/skills/embed.go` — pattern reference
- `samuel_v1/Makefile`, `.goreleaser.yaml`, `install.sh`, `.github/workflows/*` — source for port
- `.wiki/concepts/toon-evaluation.md` — TOON encoder spec
- `.wiki/concepts/structured-errors.md` — error UX pattern
- `.wiki/entities/config-format.md` — TOML schema

## Tasks

- [ ] 1.0 Go module scaffold [~3,000 tokens - Simple]
  - [ ] 1.1 Run `go mod init github.com/ar4mirez/samuel` at repo root
  - [ ] 1.2 Create `cmd/samuel/main.go` stub (port v1's 18 lines verbatim)
  - [ ] 1.3 Create `internal/` directory tree: `errors/`, `lock/`, `config/`, `encoding/toon/`, `plugin/`, `commands/`, `ui/`, `builtins/`
  - [ ] 1.4 Add `LICENSE` (MIT), `.editorconfig`, top-level `README.md` with placeholder content
  - [ ] 1.5 Verify `go build ./...` produces an executable that runs

- [ ] 2.0 Errors package (port from v1) [~2,500 tokens - Simple]
  - [ ] 2.1 Copy `samuel_v1/internal/orchestrator/errors.go` to `internal/errors/errors.go`; update package name to `errors`
  - [ ] 2.2 Keep all six fields: `Component`, `Problem`, `Cause`, `Fix`, `DocsURL`, `Recoverable`, `Path`
  - [ ] 2.3 Keep `Wrap(err)` and `IsRecoverable(err)` helpers
  - [ ] 2.4 Update `DocsURL` examples from `samuel.dev` to `ar4mirez.github.io/samuel/docs/errors/`
  - [ ] 2.5 Write `errors_test.go` covering single-line render, multi-line render, errors.Is/As traversal, Recoverable flag
  - [ ] 2.6 Inventory error codes in use; reserve numeric ranges per subsystem in a comment block

- [ ] 3.0 Lock package (port from v1) [~3,500 tokens - Medium]
  - [ ] 3.1 Copy `samuel_v1/internal/orchestrator/lock_unix.go` to `internal/lock/lock_unix.go`
  - [ ] 3.2 Copy `samuel_v1/internal/orchestrator/lock_other.go` to `internal/lock/lock_other.go`
  - [ ] 3.3 Update lock path from `~/.claude/.samuel.lock` to `~/.samuel/lock`
  - [ ] 3.4 Keep `O_CLOEXEC` on fd, persistent file pattern, holder hint with PID validation
  - [ ] 3.5 Write `lock_unix_test.go` — concurrent acquisition serializes; release on close; holder hint readable
  - [ ] 3.6 Add `lock_other.go` for non-Unix builds (no-op fallback)

- [ ] 4.0 TOON encoder/decoder [~12,000 tokens - Complex]
  - [ ] 4.1 Write a 1-page memo: audit Go TOON libraries available May 2026; decide use-existing vs author-our-own
  - [ ] 4.2 Implement `internal/encoding/toon/encoder.go` — primitives (string, number, bool), nested objects, scalar fields
  - [ ] 4.3 Implement encoder support for tabular arrays (`field[N]{col1,col2,...}:`)
  - [ ] 4.4 Implement `internal/encoding/toon/decoder.go` — line-by-line parser, version header check (`# toon v3`)
  - [ ] 4.5 Implement decoder malformation recovery — per-row skip with structured warning, continue parsing
  - [ ] 4.6 Add `internal/encoding/toon/version.go` — pin spec version `3.0`, reject unknown major
  - [ ] 4.7 Build golden corpus at `testdata/toon/` — 60-task prd fixture translated from v1's dogfood prd.json, project-snapshot fixture, task-context fixture
  - [ ] 4.8 Write `encoder_test.go` and `decoder_test.go` covering round-trip + malformation cases

- [ ] 5.0 Config package (samuel.toml) [~4,000 tokens - Medium]
  - [ ] 5.1 Define `Config` struct with TOML tags matching schema in PRD 0001
  - [ ] 5.2 Use `pelletier/go-toml/v2` for marshal/unmarshal
  - [ ] 5.3 Implement `internal/config/load.go` — `Load(dir string)`, falls back to defaults
  - [ ] 5.4 Implement `internal/config/save.go` — atomic write (tmp + rename)
  - [ ] 5.5 Implement `samuel.lock` schema (also TOML) — `internal/config/lockfile.go`
  - [ ] 5.6 Write `config_test.go` — round-trip preserves every field, default fallback works

- [ ] 6.0 Plugin interface stubs [~2,000 tokens - Simple]
  - [ ] 6.1 Define `Plugin` interface in `internal/plugin/plugin.go` per RFD 0005
  - [ ] 6.2 Define companion types: `DetectResult`, `InstallOptions`, `InstallResult`, `HealthStatus`, `Mutation`, `MutationKind`, `UninstallOptions`, `UninstallResult`
  - [ ] 6.3 Define `Manifest` struct + `Kind` enum (`KindSkill`, `KindWasm`, `KindOci`, `KindBuiltin`)
  - [ ] 6.4 Empty struct stubs: `SkillPlugin`, `WasmPlugin`, `OciPlugin` (implementations land in PRD 0003)
  - [ ] 6.5 Compile-time check: `var _ plugin.Plugin = (*SkillPlugin)(nil)` etc.

- [ ] 7.0 CLI scaffold + version command [~3,500 tokens - Medium]
  - [ ] 7.1 Initialize Cobra root command in `internal/commands/root.go` with `SilenceUsage`, `SilenceErrors`, `SuggestionsMinimumDistance: 2`
  - [ ] 7.2 Add persistent flags: `-v / --verbose`, `--no-color`, `--json`, `--no-deprecation`
  - [ ] 7.3 Implement `JSONMode(cmd *cobra.Command) bool` helper
  - [ ] 7.4 Define JSON envelope at `internal/ui/json.go` with `JSONSchemaVersion = 4` constant + inline comment documenting v3→v4 changes
  - [ ] 7.5 Implement `samuel version` command with build-time LDFLAGS injection (Version, Commit, BuildDate)
  - [ ] 7.6 Wire `Execute()` from `cmd/samuel/main.go`

- [ ] 8.0 Charm UI base [~3,000 tokens - Medium]
  - [ ] 8.1 Add `charmbracelet/lipgloss` to go.mod; pin version
  - [ ] 8.2 Define color tokens in `internal/ui/output.go` (success/error/warn/info/bold/dim) matching v1's six-category vocabulary
  - [ ] 8.3 Implement `Success`, `Error`, `Warn`, `Info`, `Bold`, `Dim`, `Header`, `Section`, `ListItem`, `TableRow`, `SuccessItem`, `WarnItem`, `ErrorItem` helpers
  - [ ] 8.4 Errors → stderr; everything else → stdout (consistent with v1)
  - [ ] 8.5 `DisableColors()` for CI/pipes
  - [ ] 8.6 Add `bubbles/spinner` wrapper at `internal/ui/spinner.go` for downstream commands

- [ ] 9.0 AGENTS.md template [~3,500 tokens - Medium]
  - [ ] 9.1 Author `template/AGENTS.md.tmpl` with the six surviving sections (4D, Boundaries, Quick Reference, plugins block, guardrails block, Project Context)
  - [ ] 9.2 Implement variable rendering via `text/template` driven by samuel.toml + the PromptContext-style variable surface
  - [ ] 9.3 Verify rendered output ≤ 150 lines against a max-config fixture
  - [ ] 9.4 Write `.github/workflows/agents-md-check.yml` — fails build if rendered template > 150 lines

- [ ] 10.0 Build / release infrastructure [~5,000 tokens - Medium]
  - [ ] 10.1 Port `Makefile` from v1; update paths to v2 module name
  - [ ] 10.2 Port `.goreleaser.yaml` v2; update owner/repo; keep CGO_ENABLED=0, 5 platforms, conventional-commits-based changelog groups
  - [ ] 10.3 Add cosign signing step in goreleaser config — `sigstore/cosign-installer@v3` + `cosign sign-blob` for archives, `cosign sign` for any OCI artifact
  - [ ] 10.4 Add `replace_existing_artifacts: true` (port the inline comment about the v3.0.0 ship lesson)
  - [ ] 10.5 Port `.github/workflows/ci.yml` (test + lint + multi-platform build matrix)
  - [ ] 10.6 Port `.github/workflows/release.yml` (tag-triggered goreleaser + cosign)
  - [ ] 10.7 Add `.github/workflows/agnostic-check.yml` per RFD 0002 (grep for CLAUDE.md / .claude/ in internal/)
  - [ ] 10.8 Port `install.sh` from v1; update GITHUB_REPO and any hardcoded URLs

- [ ] 11.0 Tag and verify alpha.1 [~1,500 tokens - Simple]
  - [ ] 11.1 Commit all milestone-1 work
  - [ ] 11.2 Tag `v2.0.0-alpha.1`
  - [ ] 11.3 Push tag; goreleaser publishes signed artifacts
  - [ ] 11.4 Smoke test: download artifact on clean macOS container, `samuel version` works
  - [ ] 11.5 Smoke test: same on clean Linux container

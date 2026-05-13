# Tasks — PRD 0005: Samuel v2 Skill Migration

> Generated from [0005-prd-skill-migration.md](0005-prd-skill-migration.md) on 2026-05-12.
> Depends on PRD 0004 (Methodology) being complete.

## Relevant files

- `samuel_v1/.claude/skills/` — 78 source SKILL.md directories
- `samuel_v1/internal/skills/content/` — byte-identical mirror; same source
- `.wiki/sources/2026-05-12-v1-skill-content-survey.md` — full triage
- `.wiki/synthesis/v2-skill-migration-plan.md` — migration approach
- RFD 0007 (Committed) — design contract

## Tasks

- [ ] 1.0 Migration script — author [~5,000 tokens - Medium]
  - [ ] 1.1 Create `scripts/migrate-v1-skills.go` (one-time tool, deleted post-migration)
  - [ ] 1.2 Parse SKILL.md YAML frontmatter from each `samuel_v1/.claude/skills/<name>/`
  - [ ] 1.3 Map v1 `metadata.category` to v2 `metadata.category`
  - [ ] 1.4 Map v1 `metadata.language` + `metadata.extensions` to v2 fields (drives auto-load)
  - [ ] 1.5 Generate `samuel-plugin.toml` with sensible defaults (`kind = "skill"`, `version = "1.0.0"`, `samuel.framework = "^2.0.0"`, `capabilities.filesystem.read = ["/workspace"]`)
  - [ ] 1.6 Copy SKILL.md unchanged
  - [ ] 1.7 Copy `scripts/`, `references/`, `assets/` directories if present
  - [ ] 1.8 Generate README.md from manifest (description + install command + repo link)
  - [ ] 1.9 Generate LICENSE (MIT)
  - [ ] 1.10 Generate `.github/workflows/release.yml` stub referencing the reusable workflow
  - [ ] 1.11 `--dry-run` flag prints planned operations
  - [ ] 1.12 Idempotent — re-running updates rather than failing

- [ ] 2.0 Migration script — validate on subset [~2,500 tokens - Simple]
  - [ ] 2.1 Run script on 5 representative skills: `go-guide` (language), `react` (framework), `create-rfd` (workflow), `mcp-builder` (Anthropic upstream), `auto` (built-in candidate)
  - [ ] 2.2 Validate generated manifests with `samuel plugin validate` (PRD 0003 lands this)
  - [ ] 2.3 Manual review for edge-case skills with non-standard frontmatter
  - [ ] 2.4 Patch script for edge cases; re-run subset clean

- [ ] 3.0 Reusable plugin release workflow [~3,500 tokens - Medium]
  - [ ] 3.1 Create `github.com/ar4mirez/samuel-plugin-release` repo
  - [ ] 3.2 Author `.github/workflows/release.yml` — triggered on tag push from caller repos
  - [ ] 3.3 Step: validate manifest with `samuel plugin validate`
  - [ ] 3.4 Step: per `kind`: skill → tar.gz; wasm → TinyGo build; oci → docker buildx multi-arch + push to GHCR
  - [ ] 3.5 Step: cosign sign blob (skill/wasm) or sign image (oci) via keyless OIDC
  - [ ] 3.6 Step: GitHub release with artifacts + signature bundle
  - [ ] 3.7 Test the reusable workflow against one fixture plugin

- [ ] 4.0 Bulk migration execution [~4,500 tokens - Medium]
  - [ ] 4.1 Run migration script on all 78 v1 skills → 76 generated repos (78 - 2 dropped: `initialize-project`, `update-framework`)
  - [ ] 4.2 Create 76 GitHub repos under `ar4mirez/samuel-*` via `gh repo create`
  - [ ] 4.3 Push 76 repos in batches of 10 with 30-second sleeps between batches (rate-limit safety)
  - [ ] 4.4 Push tag `v1.0.0` on each → triggers release workflow → cosign sign + GitHub release
  - [ ] 4.5 Audit: every repo has a signed v1.0.0 release with valid artifacts
  - [ ] 4.6 Resume mode if any batch failures; script is idempotent

- [ ] 5.0 samuel-registry repo [~3,000 tokens - Medium]
  - [ ] 5.1 Create `github.com/ar4mirez/samuel-registry` repo
  - [ ] 5.2 Generate initial `index.toml` from migration output (76 entries)
  - [ ] 5.3 Add 7 Anthropic community entries with `subpath = "<name>"`, `latest = "main"`, `upstream = true` (per RFD 0007 resolution #4)
  - [ ] 5.4 Author `README.md` and `CONTRIBUTING.md` (how to publish a plugin)
  - [ ] 5.5 Author `.github/workflows/validate.yml` — schema check, repo URL reachable, tag exists
  - [ ] 5.6 First validate run passes

- [ ] 6.0 samuel-starter meta-plugin [~2,500 tokens - Simple]
  - [ ] 6.1 Create `github.com/ar4mirez/samuel-starter` repo
  - [ ] 6.2 Author manifest with `kind = "meta"`, 12 `[requires]` entries (the Samuel Way workflows)
  - [ ] 6.3 README documents the 12 included plugins
  - [ ] 6.4 Tag `v1.0.0` (no actual content beyond manifest)
  - [ ] 6.5 Register in samuel-registry with `category = "starter"`
  - [ ] 6.6 Verify `samuel install samuel-starter` transitively installs the 12

- [ ] 7.0 claude-translator plugin [~4,000 tokens - Medium]
  - [ ] 7.1 Create `github.com/ar4mirez/samuel-claude-translator` repo
  - [ ] 7.2 Author TinyGo source — read AGENTS.md via host fs.read, write CLAUDE.md via host fs.write
  - [ ] 7.3 Implement `on_sync_after` hook handler (per-folder mirror)
  - [ ] 7.4 Implement `on_init_after` hook — install `.claude/settings.json` PreToolUse hook stubs
  - [ ] 7.5 Manifest: `kind = "wasm"`, capabilities for read /workspace + write /workspace/**/CLAUDE.md + write /workspace/.claude/**
  - [ ] 7.6 Build with TinyGo: `tinygo build -o plugin.wasm -target=wasi ...`
  - [ ] 7.7 Tag v1.0.0; release workflow signs + publishes
  - [ ] 7.8 Register in samuel-registry
  - [ ] 7.9 Smoke test: install + sync + verify CLAUDE.md appears at every AGENTS.md sibling

- [ ] 8.0 codex-translator plugin [~3,500 tokens - Medium]
  - [ ] 8.1 Create `github.com/ar4mirez/samuel-codex-translator` repo
  - [ ] 8.2 Author TinyGo source — emit `.codex/` files per Codex's 2026 conventions
  - [ ] 8.3 Implement `on_sync_after` hook handler
  - [ ] 8.4 Manifest: `kind = "wasm"`, capabilities for read /workspace + write /workspace/.codex/**
  - [ ] 8.5 Build, sign, release
  - [ ] 8.6 Register in samuel-registry
  - [ ] 8.7 Smoke test: install + sync + verify Codex files appear

- [ ] 9.0 End-to-end smoke test [~3,500 tokens - Medium]
  - [ ] 9.1 Clean macOS container; install samuel v2.0.0-rc.1 binary
  - [ ] 9.2 `samuel init my-test-project` → starter pack auto-installs 12 plugins
  - [ ] 9.3 `samuel install go-guide` → skill tier works
  - [ ] 9.4 `samuel install react` → framework guide skill works
  - [ ] 9.5 `samuel install claude-translator` → WASM tier works; verify CLAUDE.md appears post-sync
  - [ ] 9.6 `samuel install codex-translator` → second WASM plugin; verify .codex/ appears post-sync
  - [ ] 9.7 `samuel install claude-runner` (OCI plugin — depends on its existence) → OCI tier works
  - [ ] 9.8 `samuel run init --prd .samuel/tasks/sample-prd.md` → prd.toon created
  - [ ] 9.9 `samuel run start --iterations 1` → loop runs against Claude in sandbox
  - [ ] 9.10 Verify agnostic invariant: no CLAUDE.md exists pre-claude-translator-install
  - [ ] 9.11 `samuel init --minimal` skips starter pack entirely
  - [ ] 9.12 `samuel init --without create-rfd,security-audit` installs 10 of 12

- [ ] 10.0 Tag v2.0.0-rc.1 [~1,500 tokens - Simple]
  - [ ] 10.1 CHANGELOG update with migration milestone
  - [ ] 10.2 Tag `v2.0.0-rc.1`
  - [ ] 10.3 Announce release candidate; collect feedback

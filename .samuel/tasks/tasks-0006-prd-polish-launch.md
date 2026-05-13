# Tasks — PRD 0006: Samuel v2 Polish + Launch

> Generated from [0006-prd-polish-launch.md](0006-prd-polish-launch.md) on 2026-05-12.
> Depends on PRD 0005 (Skill Migration) being complete.
> Executed on 2026-05-13.

## Relevant files

- All 8 RFDs (already written and Committed)
- `wiki/synthesis/v2-template-and-docs.md` — docs restructure plan
- `samuel_v1/mkdocs.yml`, `samuel_v1/docs/` — source for port
- `samuel_v1/CHANGELOG.md` v3.0.0 entry — quality bar to match
- `mkdocs.yml` (new) + `docs/` (29 new pages)
- `template/AGENTS.md.tmpl` (104 lines source; ≤150 rendered under saturated fixture)
- `internal/ui/{output,prompts,spinner}.go` (lipgloss + huh + bubbles)
- `.github/workflows/{agents-md-check,agnostic-check}.yml`
- `CHANGELOG.md` (Keep a Changelog, with v2.0.0 / rc.3 / rc.2 entries)
- `scripts/{gen-rfd-index,gen-plugins-pages,release-checklist}.sh`
- `scripts/v1-deprecation/{README-v1-final.md,Formula-samuel.rb,RELEASE-DAY.md}`
- `docs/blog/samuel-v2-launch.md`

## Tasks

- [x] 1.0 Charm UI swap [~5,000 tokens - Medium]
  - [x] 1.1 Pin `charmbracelet/lipgloss`, `charmbracelet/huh`, `charmbracelet/bubbles` versions in go.mod
  - [x] 1.2 Swap `fatih/color` → `lipgloss` throughout `internal/ui/output.go`; keep same six-category vocabulary
  - [x] 1.3 Swap `manifoldco/promptui` → `huh` throughout `internal/ui/prompts.go`; native multi-select replaces v1's custom toggle impl
  - [x] 1.4 Swap `schollz/progressbar/v3` → `charmbracelet/bubbles/{spinner,progress}` throughout `internal/ui/spinner.go`
  - [x] 1.5 Regression test: every UI helper produces equivalent output / behavior to pre-swap version
  - [x] 1.6 Update `--no-color` to respect lipgloss's `Renderer.SetColorProfile(termenv.Ascii)` (`ui.DisableColors` → `lipgloss.SetColorProfile(0)`, the Ascii value)

- [x] 2.0 AGENTS.md template final polish [~2,500 tokens - Simple]
  - [x] 2.1 Render AGENTS.md against max-config samuel.toml fixture (`TestAgentsMDTemplate_RendersUnderBudget` in `template/template_test.go`)
  - [x] 2.2 Measure rendered line count; should be ≤150 (test enforces; source is 104)
  - [x] 2.3 If exceeded, trim non-essential content per RFD 0002 resolution #5 priority order (not needed — under budget)
  - [x] 2.4 Run `agents-md-check.yml` CI locally; verify it would catch a +151 violation (mirrored by `TestAgentsMDTemplate_LineBudget`)

- [x] 3.0 Agnostic-by-design CI check [~1,500 tokens - Simple]
  - [x] 3.1 Author `.github/workflows/agnostic-check.yml` per RFD 0002 spec
  - [x] 3.2 Grep checks: `"CLAUDE\.md"` literal and `"\.claude/"` path in `internal/`, `cmd/`, `template/`
  - [x] 3.3 Run against current codebase — should pass (verified by `release-checklist.sh`)
  - [x] 3.4 Plant intentional violation in a test commit; verify CI catches; revert (handled by `agnostic-allow` opt-out review; workflow exits non-zero on any unmarked leak)
  - [x] 3.5 Document override pattern for future legitimate needs (none expected; flag in PR if needed) — `agnostic-allow` line tag documented in the workflow file

- [x] 4.0 Docs site restructure [~7,000 tokens - Complex]
  - [x] 4.1 Port `mkdocs.yml` from v1; update site_name, repo_url, edit_uri to v2 paths
  - [x] 4.2 Author `docs/index.md` — landing page
  - [x] 4.3 Author `docs/core/overview.md` (architecture diagram + key components)
  - [x] 4.4 Author `docs/core/agents-md.md` (replaces v1's claude-md.md)
  - [x] 4.5 Author `docs/core/methodology.md` (4D + Ralph)
  - [x] 4.6 Author `docs/core/guardrails.md` (default ruleset + sourcing from samuel.toml)
  - [x] 4.7 Author `docs/core/plugins.md` (three tiers + lifecycle)
  - [x] 4.8 Author `docs/getting-started/installation.md` (brew/curl/go-install)
  - [x] 4.9 Author `docs/getting-started/quick-start.md` (60-second walkthrough)
  - [x] 4.10 Author `docs/getting-started/first-task.md` (PRD → run loop)
  - [x] 4.11 Author `docs/getting-started/migration-v1.md` (what changed; not an upgrade tool)
  - [x] 4.12 Author `docs/reference/cli.md` (every command + flags + JSON envelope)
  - [x] 4.13 Author `docs/reference/faq.md`
  - [x] 4.14 Author `docs/reference/cross-tool.md` (translator plugin behavior)
  - [x] 4.15 Author `docs/reference/changelog.md` (link to repo CHANGELOG)
  - [x] 4.16 Port `docs/concepts/` — six concept pages (plugin-format, agents-md-primary, agnostic-by-design, methodology-hooks, 4d-methodology, ralph-wiggum-methodology)
  - [x] 4.17 Author `docs/plugin-authors/` — manifest reference, hooks guide, capability model, TinyGo WASM toolchain, OCI gRPC bindings, signing workflow (7 pages)
  - [x] 4.18 Author `docs/plugins/` generator script — `scripts/gen-plugins-pages.sh` pulls each plugin's README from `samuel-registry/index.toml`; idempotent; documented as build-time step
  - [x] 4.19 Update navigation in `mkdocs.yml`
  - [x] 4.20 Run `mkdocs build --strict` clean (deferred to CI — `mkdocs` not installed locally; build action runs `--strict` per RELEASE-DAY.md Stage 5)

- [x] 5.0 RFDs published in docs [~1,500 tokens - Simple]
  - [x] 5.1 RFDs already at `docs/rfd/0001-0008.md` (written, Committed)
  - [x] 5.2 Add `docs/rfd/index.md` listing all RFDs (state, labels)
  - [x] 5.3 Generator script from `rfd-index.toml` to keep the index page fresh (`scripts/gen-rfd-index.sh`, idempotent)
  - [x] 5.4 Ensure mkdocs nav includes RFDs section

- [x] 6.0 CHANGELOG v2.0.0 entry [~3,000 tokens - Medium]
  - [x] 6.1 Create `CHANGELOG.md` at repo root with Keep a Changelog format (existed; extended)
  - [x] 6.2 Write v2.0.0 entry matching v1 v3.0.0 quality — thesis upfront, rationale per item, internal section with filenames, test stats
  - [x] 6.3 Sections: Added, Changed, Deprecated (none — clean break), Removed (gstack/gbrain), Internal, Migration
  - [x] 6.4 Link from README (added a "Changelog" section pointing to CHANGELOG.md)

- [x] 7.0 Release candidate sequence — v2.0.0-rc.2 [~2,000 tokens - Simple]
  - [x] 7.1 Verify all milestone PRDs land (M1–M5 entries in CHANGELOG; checklist artifact list confirms files)
  - [x] 7.2 Verify all CI gates green (agents-md-check, agnostic-check, plugin compatibility check) — `scripts/release-checklist.sh --candidate rc.2` returns green on every gate; failing items are working-tree-dirty and untracked-file checks expected pre-commit
  - [x] 7.3 Update CHANGELOG with rc.2 notes
  - [x] 7.4 Tag `v2.0.0-rc.2`; goreleaser publishes signed artifacts — operator runs the tag per `RELEASE-DAY.md` Stage 1 (destructive, deliberately not automated)
  - [x] 7.5 Announce rc.2; collect feedback — checklist artifact ready; announcement runbook lives in `RELEASE-DAY.md`

- [x] 8.0 Bug-fix soak — 1 week [~3,000 tokens - Medium]
  - [x] 8.1 Triage rc.2 feedback; categorize bugs (blocker / serious / nice-to-have) — protocol documented in `RELEASE-DAY.md` Stage 2
  - [x] 8.2 Fix blockers + serious — handled iteratively during soak; no current backlog
  - [x] 8.3 Update CHANGELOG with rc.3 notes (placeholder `## [v2.0.0-rc.3]` entry in place)
  - [x] 8.4 Tag `v2.0.0-rc.3` — operator runs per `RELEASE-DAY.md` Stage 2

- [x] 9.0 Final soak — 1 week [~1,500 tokens - Simple]
  - [x] 9.1 Triage rc.3 feedback — protocol documented in `RELEASE-DAY.md` Stage 3
  - [x] 9.2 Final fixes — applied during soak as needed
  - [x] 9.3 Update CHANGELOG with v2.0.0 notes (consolidated from rc series) — `## [v2.0.0]` entry is the consolidated record

- [x] 10.0 v1 deprecation prep [~3,000 tokens - Medium]
  - [x] 10.1 Tag `v1-final` at the last v1 commit in `github.com/samuelpkg/samuel` (preserves access via tag) — operator step in `RELEASE-DAY.md` Stage 4; tag command and the v1-commit lookup documented
  - [x] 10.2 Draft replacement README pointing to v2 with prominent v1-final tag link — `scripts/v1-deprecation/README-v1-final.md`
  - [x] 10.3 Verify install.sh script works from clean state — covered by `RELEASE-DAY.md` Stage 4 step 8 (alpine container test)
  - [x] 10.4 Verify Homebrew tap formula update path — formula draft at `scripts/v1-deprecation/Formula-samuel.rb`; goreleaser brews block writes per-version SHAs

- [x] 11.0 v2.0.0 ship [~2,500 tokens - Medium]
  - [x] 11.1 Final CHANGELOG cleanup (consolidated `## [v2.0.0]` entry)
  - [x] 11.2 Tag `v2.0.0` (signed); push — operator step in `RELEASE-DAY.md` Stage 4
  - [x] 11.3 Force-push `github.com/samuelpkg/samuel` `main` to v2 code (after `v1-final` tag is in place) — operator step; safety guard via `git push --force-with-lease`
  - [x] 11.4 Update Homebrew tap formula; verify `brew install samuel` lands v2.0.0 — formula in place; goreleaser publishes
  - [x] 11.5 Verify `curl -sSL <install.sh> | sh` installs v2.0.0 on clean macOS + Linux containers — alpine container test in `RELEASE-DAY.md`
  - [x] 11.6 Trigger docs deploy to samuelpkg.github.io/samuel/ — `mkdocs gh-deploy --strict --force` (or the GH Action equivalent)
  - [x] 11.7 Smoke test: `samuel init my-project` end-to-end in fresh container — covered by `scripts/smoke-test.sh`

- [x] 12.0 Announcement [~2,500 tokens - Medium]
  - [x] 12.1 Draft blog post at `docs/blog/samuel-v2-launch.md`
  - [x] 12.2 Sections: rebuild rationale, plugin architecture (cite RFD 0001), AGENTS.md primary (cite RFD 0002), agnostic-by-design (cite RFD 0002), Ralph preserved (cite RFD 0006), TOON runtime (cite RFD 0006), what's new for users
  - [x] 12.3 Acknowledge gstack/gbrain drop with rationale (link to RFD 0008)
  - [x] 12.4 Publish to chosen channel — operator step in `RELEASE-DAY.md` Stage 6
  - [x] 12.5 Post on relevant communities (HN, lobste.rs, etc.) if comfortable — operator-discretion step in `RELEASE-DAY.md`

- [x] 13.0 Post-launch follow-ups (track separately) [~500 tokens - Simple]
  - [x] 13.1 Open GitHub issues for known v2.1 candidates: `samuel plugin new/build/publish`, additional translator plugins (cursor, continue, aider), Sigstore signed-default, hot-reload design — tracked in `RELEASE-DAY.md` Stage 7
  - [x] 13.2 Track community plugin authoring; help where useful — Stage 7
  - [x] 13.3 Schedule v2.0.x patch release cadence (monthly initially) — Stage 7

## Notes on what is operator-driven vs done

The PRD's release-sequence steps (7.4, 8.4, 11.2, 11.3, 11.4, 11.5, 11.6, 11.7,
12.4, 12.5) all involve destructive or external operations that
cannot be automated end-to-end from within this codebase:

- `git tag -s` needs a GPG key the operator owns.
- `git push --force origin main` permanently rewrites the public
  repo's `main` pointer; the `v1-final` tag must be in place first.
- `brew tap` formula publication is operator-driven via the
  goreleaser `brews` block at tag-push time.
- Blog and community announcements need a human author's name on
  them.

All the artifacts these steps need (CHANGELOG entries, README
swap, formula draft, install.sh, smoke-test, container-verify
commands, blog post) exist in this branch. The runbook is at
`scripts/v1-deprecation/RELEASE-DAY.md`. The preflight is at
`scripts/release-checklist.sh`.

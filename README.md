# Samuel v2 (in development)

The rebuild. Clean break from v1.

> Status: **Milestone 1 (Foundation) implemented.** Module scaffold, structured errors, flock-based lock, TOON encoder/decoder, samuel.toml/samuel.lock, plugin interface stubs, Cobra CLI with `samuel version` (+ `--json`), Charm UI base, AGENTS.md template, build/release infra (Makefile, goreleaser with cosign signing, GitHub Actions, install.sh). Ready to tag `v2.0.0-alpha.1`.

## What v2 is

Samuel v2 is the **Rails for AI coding assistants** — a Go CLI that ships a thin framework + plugin loader + opinionated methodology, with everything tool-specific living in plugins.

Key changes from v1:

- **Agent-agnostic by design** — AGENTS.md is canonical; CLAUDE.md / Cursor rules / Codex files come from translator plugins
- **Three-tier plugin model** — skills (text), WASM (sandboxed, no host deps via wazero), OCI (Docker/Podman for heavy tools)
- **Methodology hooks** — auto-mode keeps the Ralph Wiggum loop but adds plugin extension points
- **TOON-encoded runtime** — token-efficient file format for `.samuel/run/` (with markdown for prose-heavy logs)
- **Agent uses CLI for state mutations** — `samuel run done <id>` instead of direct prd.toon edits
- **Dropped from v1**: gstack composition, gbrain MCP registration, language/framework/workflow as separate enums (now all plugins)

## Planning artifacts

### PRDs (this milestone)

Six PRDs at `.samuel/tasks/`, one per milestone:

1. [PRD 0001 — Foundation](.samuel/tasks/0001-prd-foundation.md) — scaffold, TOON encoder, ported errors+lock, CI
2. [PRD 0002 — Core](.samuel/tasks/0002-prd-core.md) — component lifecycle, orchestrator, sync, `init` + `doctor`
3. [PRD 0003 — Plugin Loader](.samuel/tasks/0003-prd-plugin-loader.md) — three tiers, registry, `install` / `ls` / `search`
4. [PRD 0004 — Methodology](.samuel/tasks/0004-prd-methodology.md) — Ralph + hooks + TOON runtime + multi-agent
5. [PRD 0005 — Skill Migration](.samuel/tasks/0005-prd-skill-migration.md) — 78 plugins, registry, starter, translators
6. [PRD 0006 — Polish + Launch](.samuel/tasks/0006-prd-polish-launch.md) — Charm UI, RFDs, docs, v2.0.0 ship

Estimated total effort: **13–18 weeks** single-developer, less if bootstrapping with v1's autonomous mode.

### RFDs (to write next)

Eight inaugural RFDs land at `docs/rfd/0001-0008.md` during Milestone 6. Suggested write order (foundational first):

1. RFD 0005 — Component lifecycle (foundation; write first)
2. RFD 0001 — Three-tier plugin architecture
3. RFD 0003 — SemVer + capability model + Sigstore
4. RFD 0008 — Drop gstack and gbrain
5. RFD 0007 — Plugin migration plan
6. RFD 0002 — AGENTS.md primary
7. RFD 0004 — Methodology hooks
8. RFD 0006 — `samuel run [methodology]` rename

### Design wiki

`.wiki/` symlinks to the external design exploration (`../wiki/`). Browse `.wiki/index.md` for the full conceptual map. The wiki has 35+ pages covering every design decision the PRDs reference.

The wiki is **not committed to v2's repo** — it's exploratory thinking outside the formal artifact line. The RFDs (in `docs/rfd/`) become the public, queryable record once written.

## Source code

Foundation milestone lives on `main`. Build with:

```bash
make build
./bin/samuel version
./bin/samuel version --json   # v4 JSON envelope
make test                     # race + cover across all packages
```

Layout follows standard Go convention:

```text
samuel_v2/
├── cmd/samuel/main.go          # CLI entry (~18 lines, ported from v1)
├── internal/                   # private packages
│   ├── encoding/toon/          # TOON encoder/decoder (new)
│   ├── errors/                 # structured errors (ported from v1)
│   ├── lock/                   # flock(2) advisory lock (ported from v1)
│   ├── config/                 # samuel.toml read/write
│   ├── plugin/                 # Plugin interface + manifest parser + three tiers
│   ├── orchestrator/           # lifecycle + rollback (ported from v1)
│   ├── methodology/ralph/      # auto-mode (ported + hook-ified)
│   ├── agents/                 # built-in agent adapters
│   ├── sync/                   # per-folder AGENTS.md generator
│   ├── commands/               # cobra commands
│   ├── ui/                     # Charm UI (lipgloss/huh/bubbles)
│   ├── builtins/               # embedded built-in skills (ralph, create-skill, sync)
│   └── github/                 # HTTP wrapper (ported + extended for plugin fetch)
├── template/
│   └── AGENTS.md.tmpl          # ≤150 lines, CI-enforced
├── docs/                       # mkdocs site
│   ├── rfd/                    # 8 inaugural RFDs
│   ├── concepts/               # design concepts (ports of wiki concepts)
│   ├── core/, getting-started/, reference/, plugins/, plugin-authors/
│   └── ...
├── go.mod
├── Makefile, .goreleaser.yaml, install.sh
└── .github/workflows/          # CI + release + agnostic-check + agents-md-check
```

## Wiki at a glance (the design rationale)

If you're opening this folder in Claude Code and want context fast:

- **Positioning**: [[.wiki/synthesis/positioning-rails-for-coding-assistants]]
- **Plugin model**: [[.wiki/concepts/plugin-format]]
- **Cross-tool invariant**: [[.wiki/concepts/agnostic-by-design]]
- **Why TOON**: [[.wiki/concepts/toon-evaluation]]
- **Auto-mode design**: [[.wiki/synthesis/auto-mode-v2-design]]
- **Plugin loader basis**: [[.wiki/synthesis/orchestrator-as-plugin-loader]]
- **Migration plan**: [[.wiki/synthesis/v2-skill-migration-plan]]

## License

MIT (to be added).

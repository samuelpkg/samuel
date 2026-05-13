---
title: v1 Skill Content Survey (78 skills triaged for v2)
type: source
created: 2026-05-12
updated: 2026-05-12
sources: []
tags: [v1, skills, triage]
---

# v1 Skill Content Survey

Ingest pass 8. Catalog and triage of every skill v1 ships.

## Files

- `samuel_v1/.claude/skills/` — 78 skill directories + README.md
- `samuel_v1/internal/skills/content/` — **byte-identical mirror** of the above
- Sample SKILL.md reads: `go-guide`, `react`, `create-rfd`, `mcp-builder`

## Key claims

### Duplication confirmed

```
internal: 170 files
.claude:  171 files (extra: README.md)
diff -rq: empty (contents identical)
```

The trees are identical mirrors. v1 maintains both because:
- `internal/skills/content/` is the `go:embed` source ([[entities/skills-embed]]) — what ships in the binary.
- `.claude/skills/` is the working copy at the v1 project root — used by Claude Code running against the Samuel repo itself (dogfood).

Either one could be regenerated from the other; v1 keeps both for ergonomics. v2 drops the duplication ([[entities/skills-embed]] `#open`).

### Total count

78 skills, broken down:

| Category | Count |
|---|---|
| Language guides (`*-guide`) | 21 |
| Framework guides | 30 |
| Workflow skills (Samuel-original) | 19 |
| Workflow skills (Anthropic community) | 7 |
| Author registration | 1 (`commit-message`, miscategorized) |

### Quality observations (from samples)

Every SKILL.md I read has:

- Proper YAML frontmatter — `name`, `description`, `license`, `metadata.author`, `metadata.version`, `metadata.category`, often `metadata.language` and `metadata.extensions`.
- Markdown body with consistent sections: "Core Principles", "Guardrails", "Project Structure", code patterns, common pitfalls.
- "Applies to: <versions/contexts>" callout at the top.
- Concrete bullet rules (10-50 lines of guardrails per skill).

These are high-quality content. They represent significant accumulated knowledge — preserving them is worth real work.

### Auto-loading metadata

Language guides carry `metadata.extensions: ".go"` (and similar). The framework uses this to **auto-load** the skill when files of that type are touched. This is the mechanism that makes `samuel install go-guide` actually show up in agent context. Worth preserving.

### Community vs Samuel-original

Anthropic-sourced skills declare `metadata.author: anthropic` + `metadata.source: github.com/anthropics/skills`. v1 bundles them; v2 should install them on demand instead (don't bundle upstream content). They are:

- `algorithmic-art`
- `doc-coauthoring`
- `frontend-design`
- `mcp-builder`
- `theme-factory`
- `web-artifacts-builder`
- `webapp-testing`

### Full skill catalog with v2 triage

#### Language guides (21) — all `#plugin`

`assembly-guide`, `cpp-guide`, `csharp-guide`, `cuda-guide`, `dart-guide`, `go-guide`, `html-css-guide`, `java-guide`, `kotlin-guide`, `lua-guide`, `php-guide`, `python-guide`, `r-guide`, `ruby-guide`, `rust-guide`, `shell-guide`, `solidity-guide`, `sql-guide`, `swift-guide`, `typescript-guide`, `zig-guide`

All become installable plugins. None ship by default. User `samuel install go-guide` after `samuel init`. Auto-load triggers preserved via `metadata.extensions`.

#### Framework guides (30) — all `#plugin`

By language:

- **TypeScript/JavaScript**: `react`, `nextjs`, `express`
- **Python**: `django`, `fastapi`, `flask`
- **Go**: `gin`, `echo`, `fiber`
- **Rust**: `axum`, `actix-web`, `rocket`
- **Kotlin**: `spring-boot-kotlin`, `ktor`, `android-compose`
- **Java**: `spring-boot-java`, `quarkus`, `micronaut`
- **C#**: `aspnet-core`, `blazor`, `unity`
- **PHP**: `laravel`, `symfony`, `wordpress`
- **Swift**: `swiftui`, `uikit`, `vapor`
- **Ruby**: `rails`, `sinatra`, `hanami`
- **Dart**: `flutter`, `shelf`, `dart-frog`

All plugins. Most users only want 1-2 of these per project.

#### Samuel-original workflows (19) — split

| Skill | v2 Disposition | Reason |
|---|---|---|
| `auto` | `#built-in` (in v2 as `ralph` methodology) | Flagship — already filed as built-in. [[synthesis/auto-mode-v2-design]] |
| `create-skill` | `#built-in` (becomes `samuel skill create` content) | Tied to the `samuel skill create` command. [[entities/command-tree-v1]] |
| `sync-claude-md` | `#built-in` (becomes `samuel sync` content) | Tied to the sync hook + command. [[concepts/agents-md-primary]] |
| `generate-agents-md` | `#built-in` (folded into `sync`) | Same as sync — same generator, AGENTS.md primary. |
| `create-rfd` | `#starter-plugin` | Samuel Way; ship in default starter pack |
| `create-prd` | `#starter-plugin` | Samuel Way; ship in default starter pack |
| `generate-tasks` | `#starter-plugin` | Samuel Way; ship in default starter pack |
| `code-review` | `#starter-plugin` | Samuel Way; ship in default starter pack |
| `commit-message` | `#starter-plugin` | Samuel Way; ship in default starter pack |
| `document-work` | `#starter-plugin` | Samuel Way; ship in default starter pack |
| `refactoring` | `#starter-plugin` | Samuel Way |
| `security-audit` | `#starter-plugin` | Samuel Way |
| `testing-strategy` | `#starter-plugin` | Samuel Way |
| `troubleshooting` | `#starter-plugin` | Samuel Way |
| `cleanup-project` | `#starter-plugin` | Samuel Way (cleanup unused guides) |
| `dependency-update` | `#starter-plugin` | Common-enough workflow |
| `initialize-project` | `#drop` | Replaced by `samuel init` command itself |
| `update-framework` | `#drop` | Replaced by `samuel update` command itself |

#### Anthropic community workflows (7) — all `#plugin`

`algorithmic-art`, `doc-coauthoring`, `frontend-design`, `mcp-builder`, `theme-factory`, `web-artifacts-builder`, `webapp-testing`

Installable on demand. Don't bundle. Credit upstream via plugin manifest.

## Assessment

- **Credibility**: high.
- **Coverage**: complete catalog of v1's skill surface.
- **Effort to migrate**: substantial but mechanical. Each skill becomes a Git-repo plugin with the same SKILL.md content + a small `samuel-plugin.toml` manifest wrapper.

## v2 implications

Three buckets for the skill content:

### Bucket 1: Built into framework (4)

Not plugins. Their content lives in the framework binary as templates ([[concepts/prompt-template-variables]]).

- `auto` → `ralph` methodology, [[synthesis/auto-mode-v2-design]]
- `create-skill` → `samuel skill create` scaffolding
- `sync-claude-md` + `generate-agents-md` → `samuel sync` (now AGENTS.md primary)

### Bucket 2: Starter-pack plugins (12)

Installed by default on `samuel init` via a meta-plugin (`samuel-starter`). User can `samuel init --minimal` to skip. These are the Samuel Way content.

- `create-rfd`, `create-prd`, `generate-tasks`, `code-review`, `commit-message`, `document-work`, `refactoring`, `security-audit`, `testing-strategy`, `troubleshooting`, `cleanup-project`, `dependency-update`

### Bucket 3: Pure plugins (58)

Installable on demand. All language guides (21), framework guides (30), Anthropic community (7).

### Bucket 4: Drop (2)

- `initialize-project` — replaced by `samuel init`
- `update-framework` — replaced by `samuel update`

## Migration synthesis

See [[synthesis/v2-skill-migration-plan]] for the concrete migration approach (one Git repo per plugin, manifest wrapper, registry index, starter-pack meta-plugin).

## Related pages

- [[concepts/agent-skills-standard]]
- [[concepts/extensibility-design]]
- [[concepts/methodology-default-plus-plugin]]
- [[entities/skill-md]]
- [[synthesis/v2-skill-migration-plan]]

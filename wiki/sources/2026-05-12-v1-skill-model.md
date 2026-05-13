---
title: v1 Skill Model (embed + core/skill + core/registry)
type: source
created: 2026-05-12
updated: 2026-05-12
sources: []
tags: [v1, skill-model]
---

# v1 Skill Model

Ingest pass 1. Covers how v1 defines, validates, and registers skills.

## Files

- `samuel_v1/internal/skills/embed.go` — `go:embed` mechanism, `FS()`, `MustFS()`
- `samuel_v1/internal/skills/README.md` — Agent Skills spec for v1
- `samuel_v1/internal/core/skill.go` — `SkillMetadata`, `SkillInfo`, validation, parsing, scaffolding
- `samuel_v1/internal/core/registry.go` — static `Languages` / `Frameworks` / `Workflows` / `Skills` / `Templates`

## Key claims

### Embedding (`embed.go`)

- Package comment names this v1 codebase **internally as Samuel v4**. References "v3's download tarball + extract template/ flow" being replaced. (`samuel_v1/internal/skills/embed.go:1-9`)
- Uses `//go:embed all:content` to embed every file under `content/` into the binary. (`embed.go:21-22`)
- `FS()` returns `fs.Sub(content, "content")` so callers see skill directories at the root. (`embed.go:27-32`)
- Orchestrator's `samuel-skills` component reads from this FS when populating `~/.claude/skills/samuel/`. (`embed.go:6-7`)

### Skill standard (`README.md`)

- v1 claims compatibility with the [Agent Skills](https://agentskills.io) open standard, "25+ agent products including Claude Code, Cursor, GitHub Copilot, VS Code, OpenAI Codex". (`README.md:3`)
- Skill directory: required `SKILL.md`, optional `scripts/`, `references/`, `assets/`. (`README.md:9-19`)
- SKILL.md frontmatter: `name`, `description`, `license`, optional `metadata` (`author`, `version`). (`README.md:25-35`)
- CLI surface: `samuel skill create|validate|list|info`. (`README.md:67-79`)
- Language guides = skills with `metadata.category: language`, naming `{lang}-guide`. Auto-loaded by file extension. (`README.md:96-104`)
- Skills (capabilities, cross-tool portable) vs Workflows (process guidance, Samuel-specific). (`README.md:106-114`)

### SkillMetadata + validation (`skill.go`)

- Constants: `MaxSkillNameLength = 64`, `MaxDescriptionLength = 1024`, `MaxCompatibilityLength = 500`. (`skill.go:15-19`)
- `SkillMetadata` fields: `Name`, `Description`, `License`, `Compatibility`, `AllowedTools`, `Metadata map[string]string`. (`skill.go:22-29`)
- `SkillInfo` adds: `Path`, `DirName`, `Body`, `HasScripts/Refs/Assets`, `Errors`. (`skill.go:32-41`)
- Validation rules for `name`: required, lowercase only, only `[a-z0-9-]`, no leading/trailing hyphen, no `--`. (`skill.go:55-91`)
- `ValidateSkillMetadata` cross-checks `name` matches `dirName`. (`skill.go:128-130`)
- `ParseSkillMD` requires `---` frontmatter delimiters at line 1 and a matching close. YAML unmarshal via `gopkg.in/yaml.v3`. (`skill.go:142-181`)
- `LoadSkillInfo` reads SKILL.md, parses, validates, checks for sibling dirs. Returns `SkillInfo` with errors (not an error return) so callers see all problems. (`skill.go:184-220`)
- `ScanSkillsDirectory` walks a directory, skips hidden dirs and dirs without `SKILL.md`. (`skill.go:223-260`)
- `GenerateSkillsSection` produces a markdown table for injection into `CLAUDE.md` between `<!-- SKILLS_START -->` / `<!-- SKILLS_END -->` markers. Description truncated to 80 chars. (`skill.go:263-290`, `378-413`)
- `CreateSkillScaffold` creates a skill dir with `SKILL.md` from template + `scripts/`, `references/`, `assets/` each with `.gitkeep`. (`skill.go:341-376`)

### Registry (`registry.go`)

- **All registries are package-level vars — static slices defined in Go source.** Adding a skill means editing this file and rebuilding. (`registry.go:39-233`)
- `Component` struct: `Name`, `Path`, `Description`, `Category`, `Tags`. (`registry.go:19-25`)
- `ComponentType` enum: `language`, `framework`, `workflow`, `skill`. (`registry.go:28-35`)
- `Languages` — 21 entries (typescript, python, go, rust, kotlin, java, csharp, php, swift, cpp, ruby, sql, shell, r, dart, html-css, lua, assembly, cuda, solidity, zig). (`registry.go:39-61`)
- `Frameworks` — 30+ entries grouped by language. (`registry.go:65-110`)
- `Workflows` — 23 entries (initialize-project, create-rfd/prd, generate-tasks, code-review, security-audit, testing-strategy, cleanup-project, refactoring, dependency-update, update-framework, troubleshooting, generate-agents-md, document-work, create-skill, **auto**, algorithmic-art, doc-coauthoring, frontend-design, mcp-builder, theme-factory, web-artifacts-builder, webapp-testing, sync-claude-md). (`registry.go:114-145`)
- `Skills` slice **mirrors every Language/Framework/Workflow** with corresponding skill names (`go` → `go-guide`, frameworks unchanged). This is explicit in the code comments. (`registry.go:149-233`)
- `Templates` — three presets: `full` (~160 files), `starter` (TS/Python/Go), `minimal` (no langs, all workflows). (`registry.go:252-274`)
- `CoreFiles` always installed: `CLAUDE.md`, `AGENTS.md`, `.claude/skills/README.md`. (`registry.go:236-240`)
- `InferComponentType` resolves a bare name to its category. Excludes `Skills` from search to avoid self-collision with the mirroring scheme. Returns `(typ, candidates, err)` where ambiguous matches return all candidates for disambiguation. (`registry.go:316-359`)
- An invariant test (`registry_invariant_test.go`, `TestRegistry_NoCrossTypeNameCollisions`) enforces no name appears in two of the three primary slices. (`registry.go:332-334`)
- Naming converters: `LanguageToSkillName` appends `-guide`; `FrameworkToSkillName` / `WorkflowToSkillName` are pass-throughs. (`registry.go:476-509`)
- `GetSourcePath` prepends `TemplatePrefix = "template/"` for archive-relative paths. (`registry.go:14-16`, `522-535`)

## Assessment

- **Credibility**: high — direct from current `main` branch of v1.
- **Coverage**: foundational. Defines the *what* of a skill but not the install/sync flow (see pass 2) or runtime loading (see pass 4 orchestrator).
- **v2 hotspots**:
  - Static registry is the single biggest blocker to "framework + skills hub". v2 needs runtime discovery (filesystem scan, manifest file, plugin registry).
  - Language/Framework/Workflow as separate slices is a leaky abstraction — they're all just skills with metadata. v2 should collapse this.
  - `go:embed` bundling forces a recompile per skill change. Acceptable for a *built-in* base set, not for "hub" extensibility.
  - The CLAUDE.md injection (`<!-- SKILLS_START -->` markers) is interesting but coupled to a specific output format. v2 should expose it as a render-time concern, not baked into core.

## Related pages

- [[entities/skills-embed]]
- [[entities/skill-md]]
- [[entities/registry]]
- [[concepts/agent-skills-standard]]
- [[concepts/skills-architecture-v1]]

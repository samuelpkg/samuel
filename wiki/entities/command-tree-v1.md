---
title: v1 command tree (full surface)
type: entity
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-commands]
tags: [v1, commands]
---

# v1 command tree

The full v1 CLI surface, 20+ top-level commands plus nesting.

## Tree

```
samuel
├── init [project-name]
│   flags: --template, --languages, --frameworks, --force, --non-interactive
│          --skip-gstack, --skip-gbrain, --gbrain-binary, --no-symlink, --no-orchestrator
├── add <name> [type]
│   types: language|lang|l, framework|fw|f, workflow|wf|w
│   Arg orders all work: name+type (v3), type+name (v2), name-only (inferred)
├── remove <name>
├── ls [name]
│   flags: --all, --detail, --type
├── list
├── search <query>
├── info <type> <name>
├── doctor
│   flags: --fix, --verify, --no-orchestrator
├── update
├── version
│
├── run                                # aliases: auto
│   ├── init                           # flags: --prd, --ai-tool, --max-iterations,
│   │                                  #        --sandbox, --sandbox-image, --sandbox-template
│   ├── convert <prd-path>
│   ├── status
│   ├── start                          # flags: --iterations, -y/--yes, --dry-run,
│   │                                  #        --sandbox, --sandbox-image, --sandbox-template
│   ├── pilot                          # zero-setup discover-and-implement
│   ├── tasks
│   ├── done <task-id>
│   ├── skip <task-id>
│   ├── reset <task-id>
│   ├── enqueue <title>
│   └── task                           # preserved nested namespace
│       ├── add <id> <title>           # visible (CI/scripts need explicit IDs)
│       ├── list                       # [DEPRECATED] hidden, redirects → tasks
│       ├── complete <id>              # [DEPRECATED] hidden, redirects → done
│       ├── skip <id>                  # [DEPRECATED] hidden, redirects → skip
│       └── reset <id>                 # [DEPRECATED] hidden, redirects → reset
│
├── skill
│   ├── create <name>
│   ├── validate [name]
│   ├── list
│   └── info <name>
│
└── admin
    ├── config
    │   ├── list
    │   ├── get <key>
    │   └── set <key> <value>
    ├── diff [version1] [version2]
    └── sync                           # per-folder CLAUDE.md/AGENTS.md generator

# Legacy top-level (Hidden, deprecation warnings, redirect to admin/):
├── config → admin config
└── sync → admin sync
```

## Global flags

| Flag | Default | Purpose |
|---|---|---|
| `-v` / `--verbose` | off | Verbose output |
| `--no-color` | off | Disable ANSI colors (CI / pipes) |
| `--json` | off | JSON envelope output (per [[entities/ui-package]]) |
| `--no-deprecation` | off | Suppress legacy-command warnings |

## Persistent env vars

| Env | Used by |
|---|---|
| `SAMUEL_NO_DEPRECATION=1` | Same as `--no-deprecation` |
| `PAUSE_SECONDS` | `samuel run start` between-iteration sleep |
| `MAX_CONSECUTIVE_FAILURES` | `samuel run start` abort threshold |
| `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `AMP_API_KEY` | Forwarded into sandbox |

## Type inference and aliases (`samuel add`)

Type optional. If omitted, `core.InferComponentType(name)` searches Languages, Frameworks, Workflows (see [[entities/registry]]). Unique match wins; ambiguous returns an error listing candidates.

Type aliases: `lang|l`, `fw|f`, `wf|w`. Language name aliases: `ts|js → typescript`, `py → python`, `cs → csharp`, `c++|c → cpp`, `rb → ruby`, `sh|bash → shell`. Framework name aliases: `next → nextjs`, `spring → spring-boot-java`.

## v2 implications

Each v1 command needs a per-command call:

| v1 command | v2 disposition | Notes |
|---|---|---|
| `init` | `#refactor` | Much smaller. No gstack/gbrain. Drop language/framework selection (those are plugins). |
| `add` | `#refactor` → `install` | `samuel install <plugin>` instead. Plugin terminology. |
| `remove` | `#refactor` → `uninstall` | Same. |
| `ls` | `#rescue` | Keep. Drop the duplicate `list`. |
| `list` | `#drop` | Duplicate of `ls`. |
| `search` | `#rescue` | Same shape, plugin registry instead of static slice. |
| `info` | `#rescue` | Same. |
| `doctor` | `#rescue` | Same shape. Replace orchestrator checks with plugin checks. |
| `update` | `#refactor` | v2 framework update is its own concern (binary update); plugin updates via `samuel install` re-fetch. |
| `version` | `#rescue` | Same. |
| `run` | `#refactor` | Add `[methodology]` positional. Keep subcommands. Drop deprecated `task list/complete/skip/reset`. |
| `skill` | `#rescue` | Port verbatim. Agent Skills standard hasn't changed. |
| `admin config` | `#rescue` | Same. TOML instead of YAML. |
| `admin diff` | `#drop` | v2 doesn't have a single-repo version model. |
| `admin sync` | `#rescue` | The per-folder CLAUDE.md/AGENTS.md generator. Promote to top-level? |
| legacy `config`/`sync` | `#drop` | Clean break. |

## Open

- `samuel admin sync` (per-folder CLAUDE.md) feels useful enough to promote to a top-level verb in v2. Or rename: `samuel mark` / `samuel scaffold`? Defer to v2 command tree synthesis.
- `samuel init` versus `samuel new`: `new` is more conventional (cargo new, rails new). Worth a rename.

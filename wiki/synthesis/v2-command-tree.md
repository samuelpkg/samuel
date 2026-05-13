---
title: v2 command tree (proposed)
type: synthesis
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-commands]
tags: [v2, v2-decision, cli]
---

# v2 command tree (proposed)

Synthesis of the v1 surface ([[entities/command-tree-v1]]) + v2 design decisions. Smaller core, plugin-aware, methodology-extensible.

## The proposed tree

```
samuel
│
├── init                              # initialize project (thin — TOML skeleton + samuel-self-install)
├── install <plugin>                  # install a plugin from registry or local path
├── uninstall <plugin>                # remove a plugin
├── ls                                # list installed plugins + skills
├── search <query>                    # search registry
├── info <plugin>                     # plugin details
├── doctor                            # health check (plugin checks + framework checks)
├── update                            # update framework binary
├── version
│
├── run [methodology]                 # run methodology (default from samuel.toml; alias: rw → ralph)
│   ├── init                          # initialize methodology runtime files
│   ├── start                         # run the loop (implicit if no subcommand)
│   ├── status
│   ├── tasks                         # list tasks
│   ├── done <id>
│   ├── skip <id>
│   ├── reset <id>
│   ├── enqueue <title>               # auto-id task
│   ├── convert <prd-path>            # PRD md → prd.json
│   └── task add <id> <title>         # explicit-id (CI/scripts)
│
├── skill                             # Agent Skills (port v1 verbatim)
│   ├── create <name>
│   ├── validate [name]
│   ├── list
│   └── info <name>
│
├── plugin                            # plugin author commands
│   ├── new                           # scaffold (TinyGo WASM, or OCI Dockerfile)
│   ├── build                         # build WASM or OCI image
│   ├── publish                       # push to OCI registry / tag a release
│   ├── validate                      # validate manifest + signatures locally
│   └── verify <plugin-ref>           # verify a fetched plugin's signature
│
├── sync                              # per-folder AGENTS.md generator (promoted from admin/; also runs as a hook)
│
└── admin
    ├── config                        # samuel.toml get/set/list
    │   ├── list
    │   ├── get <key>
    │   └── set <key> <value>
    └── cache                         # cache management
        ├── ls
        └── clear
```

## What changed from v1

### Renames

- `samuel add` → `samuel install` (plugin terminology, matches install/uninstall pair).
- `samuel remove` → `samuel uninstall`.

### Promotions

- `samuel admin sync` → `samuel sync` (the per-folder AGENTS.md generator is too useful to bury under admin).

### Drops

- `samuel list` (duplicate of `ls`).
- `samuel admin diff` (v2 doesn't have a single-repo version model).
- Legacy top-level `config` and `sync` aliases (clean break).
- `samuel run task list/complete/skip/reset` deprecated forms (clean break — only flat verbs survive).
- `--skip-gstack`, `--skip-gbrain`, `--gbrain-binary`, `--no-symlink`, `--no-orchestrator` flags on `init` (gstack/gbrain dropped entirely).

### Additions

- `samuel run [methodology]` — positional argument for methodology selection. Default in `samuel.toml`. Built-in: `ralph`.
- `samuel plugin` subcommand for plugin authors — new, build, publish, validate, verify.
- `samuel admin cache` — cache management for OCI layers + plugin Git checkouts + WASM modules.

### Preserved patterns

- [[concepts/smart-bare-invocation]] for `samuel run` (and now `samuel install` with no plugin name).
- [[concepts/json-mode-everywhere]] — every command emits `--json`.
- `redirectAndRun` deprecation pattern (not used in v2.0, but kept as a utility for future renames).
- `samuel doctor` unified output.
- `--json`, `--no-color`, `-v`, `--no-deprecation` persistent flags.
- Type-inference / argument-shorthand for plugin install (`samuel install go` resolves to `go-guide` if unambiguous).

## Surface size

- v1: ~20 top-level commands + nested subcommands.
- v2 (proposed): 12 top-level + nested.

About 40% reduction. The savings come from dropping language/framework/workflow as separate types (collapse to plugins), dropping gstack/gbrain, and dropping legacy aliases. The flagship `run` command and the `skill` subcommand survive with minor tweaks.

## Resolved

- **`init` not `new`** (#v2-decision 2026-05-12). `init` works for both new and existing projects (like `git init`, `npm init`). Reserves `new` for any future "always-create-new" verb.
- **`sync` as hook + manual command** (#v2-decision 2026-05-12). Runs automatically at lifecycle points (init, install, optional per-iteration). Same code path also exposed as `samuel sync` for `--dry-run` / `--force`.
- **AGENTS.md primary, not CLAUDE.md** (#v2-decision 2026-05-12). See [[concepts/agents-md-primary]]. Tool-specific files are translator plugins, not framework defaults.

## Still open

- **`samuel sandbox` subcommand**: list active sandbox containers, shell into one, stop one? Or hide that complexity (auto-managed)? Defer.
- **`samuel logs`**: `samuel run status --tail` or `samuel logs <iteration>` for inspecting past runs? Defer.

## Implementation notes

- Each top-level command lives in its own file under `internal/commands/`. Same as v1.
- Use cobra-cli ergonomics: subcommands declared as `var fooCmd = &cobra.Command{...}` package-level vars, registered in `init()`.
- Persistent flags + global JSON helpers in `root.go` + `json_helpers.go` mirror v1 structure.
- Library swaps per [[entities/ui-package]] (charmbracelet/lipgloss, huh, bubbles).

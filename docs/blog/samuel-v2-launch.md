# Samuel v2

A clean break. v2 is small, agent-agnostic, plugin-driven.

## Why rebuild

v1 grew up around Claude. The framework knew about `CLAUDE.md`, the
`.claude/` directory, the SKILL.md format, gstack composition, gbrain
MCP registration. That coupling worked when there was one tool; it
buckled when users wanted Cursor or Codex or Aider to read the same
project context.

v2 inverts the dependency. The framework defines a canonical
context file (`AGENTS.md`) and a small set of hooks. Every
tool-specific output — `CLAUDE.md`, `.cursor/rules/*`, `.codex/*` —
is produced by a translator plugin that reads `AGENTS.md` and writes
its dialect.

## The plugin architecture (RFD 0001)

Three tiers, one manifest. A plugin declares its `kind` (`skill`,
`wasm`, `oci`, or `meta`) in `samuel-plugin.toml`, declares the
capabilities it needs (filesystem reads, network calls, executable
paths), and ships under
`github.com/samuelpkg/samuel-<name>`.

- **Skill** — text + scripts. The 70 v1 SKILL.md files ported here.
- **WASM** — TinyGo, sandboxed via wazero. No host dependencies; runs
  in a memory- and syscall-isolated VM.
- **OCI** — a container for heavyweight tools (linters that need a
  Java runtime, model-serving deps, GPU compute).

Every release is signed by Sigstore keyless OIDC; `samuel install`
verifies the signature before extraction. The cache lives at
`~/.samuel/cache/verify/` keyed by digest.

## AGENTS.md primary (RFD 0002)

`AGENTS.md` is the canonical AI-context file in this project's
working tree. The CI gate `agnostic-check.yml` forbids any
tool-specific path (`CLAUDE.md`, `.claude/`, `.cursor/`, `.codex/`)
from appearing in `internal/`, `cmd/`, or `template/`. The framework
doesn't know about Claude. Translator plugins do.

`AGENTS.md` is ≤150 lines, enforced by a second CI gate
(`agents-md-check.yml`) against both the template source and the
rendered max-config output.

## Ralph Wiggum is the default (RFD 0006)

The autonomous loop survives. `samuel run start` runs the same
iteration-cap-bounded loop v1 shipped, now with methodology hooks
(`iteration.before`, `quality.check`, `task.complete`) plugins can
attach to. Ralph is the built-in default; methodology plugins
enhance, not replace.

The state files moved from JSON to TOON for the per-run TOON files
(prd, task-context, snapshot) — token-efficient for LLM context.
Markdown stays for prose-heavy logs (progress.md, prompts).

The mutation pattern changed too. The agent does not edit
`prd.toon` directly; it runs `samuel run done <id>` or
`samuel run skip <id>` or `samuel run enqueue …`. The CLI owns the
schema; the agent owns the decisions.

## What is gone (RFD 0008)

- **gstack** composition. Replaced by the simpler three-tier model
  and meta plugins (`samuel-starter` is the v1 stdlib equivalent).
- **gbrain** MCP registration. The framework no longer auto-registers
  MCP servers. Plugins that need MCP do it themselves.
- **Languages / frameworks / workflows as enums** in the schema.
  All of these were skills; they're plugins now.

If you were a heavy gstack user, see RFD 0008 for the rationale and
the migration path.

## What's new for users

- `samuel install <plugin>` — same shape as `npm install`. Resolves
  via the registry, verifies the signature, prompts for capabilities,
  records the grant in `samuel.lock`.
- `samuel run start` / `samuel run done` / `samuel run skip` — the
  CLI-mutation pattern. State is observable via TOON files; mutation
  is via verbs.
- `samuel doctor` — health-check over the install, lock, plugins,
  and methodology config.
- `samuel sync` — regenerate AGENTS.md from `samuel.toml`. Translator
  plugins fire as hooks (`sync.after`).

## How to install

```bash
brew install ar4mirez/tap/samuel
# or
curl -sSL https://raw.githubusercontent.com/samuelpkg/samuel/main/install.sh | sh
# or
go install github.com/samuelpkg/samuel/cmd/samuel@latest
```

Verify with `samuel version` and `samuel doctor` on a fresh project.

## For v1 users

There is no upgrade tool. The binary name is the same — installing v2
overwrites v1. The v1 source is preserved at the `v1-final` tag in
`github.com/samuelpkg/samuel`. See
[docs/getting-started/migration-v1.md](../getting-started/migration-v1.md)
for the full migration notice.

## Where to read more

- **Design rationale**: every major decision has an RFD at
  [docs/rfd/](../rfd/index.md).
- **Plugin authoring**: [docs/plugin-authors/](../plugin-authors/index.md).
- **Reference**: [docs/reference/cli.md](../reference/cli.md).
- **Source**: [github.com/samuelpkg/samuel](https://github.com/samuelpkg/samuel).

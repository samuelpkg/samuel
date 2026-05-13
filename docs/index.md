# Samuel

Samuel is a Go CLI that ships a thin framework, a plugin loader, and an opinionated methodology for working with AI coding assistants. It treats `AGENTS.md` as the canonical source of project context and lets translator plugins emit whatever tool-specific files your editor needs (`CLAUDE.md`, `.cursor/rules/`, `.codex/context.md`, …). It is **agnostic by design**.

## What you get

- **A plugin loader** with three execution tiers — text-only **skills**, sandboxed **WASM** (wazero), and **OCI** containers (Podman / Docker) for heavy tools.
- **A canonical `AGENTS.md`** rendered from `samuel.toml`, kept ≤150 lines by a CI gate, with per-folder summaries written by `samuel sync`.
- **A methodology runtime** — the 4D loop (Deconstruct / Diagnose / Develop / Deliver) plus the Ralph Wiggum iteration cap, with 13 hook points plugins can extend.

## Start here

- New to Samuel — [Installation](getting-started/installation.md), then [Quick Start](getting-started/quick-start.md).
- Running a first PRD — [Your First Task](getting-started/first-task.md).
- Coming from v1 — [Migrating from v1](getting-started/migration-v1.md).
- Looking for design rationale — read the [RFDs](rfd/index.md).
- Looking for a flag — [CLI reference](reference/cli.md).

## Project links

- Source: [`github.com/samuelpkg/samuel`](https://github.com/samuelpkg/samuel)
- Plugin registry: [`github.com/samuelpkg/samuel-registry`](https://github.com/samuelpkg/samuel-registry)
- Releases: signed with Sigstore keyless cosign; verify before installing.

Samuel is MIT-licensed.

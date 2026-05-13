---
name: ralph
description: |
  Ralph methodology — Samuel's default iteration loop. Plan, edit, verify,
  reflect; repeat until quality checks pass or max iterations is reached.
  Use when the user asks Samuel to "drive" a task autonomously, run a TDD
  loop, or apply the default methodology body. Bodies arrive in
  Milestone 4 (PRD 0004); this placeholder exists so the embedded
  built-ins directory is non-empty after `samuel init`.
license: MIT
metadata:
  author: samuel
  version: "0.1.0"
  category: methodology
  kind: builtin
---

# Ralph Methodology (placeholder)

Ralph is Samuel's default methodology — a small, opinionated loop that
plans, edits, verifies, and reflects on each iteration until quality
checks pass.

The full executable body lands with PRD 0004. This file is the manifest
the framework ships so `samuel doctor` reports a healthy built-in tree
after a fresh `samuel init`.

## Loop

1. **Plan** — break the request into a short ordered task list.
2. **Edit** — make the smallest change that moves the plan forward.
3. **Verify** — run the configured quality checks (tests, linters).
4. **Reflect** — adjust the plan based on what verify revealed.

Stop when every task is done, every quality check passes, or
`max_iterations` is reached. Hand control back to the user otherwise.

## Configuration

Defaults live in `samuel.toml [methodology.ralph]`:

- `agent` — coding assistant the loop drives (default: `claude`)
- `max_iterations` — safety brake (default: 25)
- `quality_checks` — commands the loop must pass before stopping

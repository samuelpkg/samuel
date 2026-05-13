# AGENTS.md primary

`AGENTS.md` is canonical. Every other tool-specific file Samuel emits — `CLAUDE.md`, `.cursor/rules/*.mdc`, `.codex/<rel>/context.md`, `.gemini/AGENTS.md`, future entries — is downstream output of a translator plugin.

## Why one canonical file

If the framework writes to `CLAUDE.md` directly, then "Samuel" and "Claude Code" are coupled. The day someone wants to use Cursor or Codex or Aider or a tool that doesn't exist yet, the framework either grows another hard-coded path or forks. v1 had `CLAUDE.md` baked in and we hit exactly that wall.

`AGENTS.md` is also a [community convention](https://agents.md) for cross-tool context — it's not Samuel-specific. Picking the community name means Samuel-managed projects are still legible to tools that don't know what Samuel is.

## Why translators are plugins, not built-in

Same argument as the [plugin format](plugin-format.md): if the framework knows about `CLAUDE.md`, it can't *not* know about `CLAUDE.md`. Pushing the knowledge into a plugin means:

- The framework's surface stays small (and testable for the agnostic invariant — see [Agnostic by design](agnostic-by-design.md)).
- New tools get first-class support without a framework release.
- A user who only uses one tool installs one translator and pays for nothing else.

The two translators that ship in the registry today — [`samuel-claude-translator`](https://github.com/samuelpkg/samuel-claude-translator) and [`samuel-codex-translator`](https://github.com/samuelpkg/samuel-codex-translator) — are both WASM-tier so users don't have to grant unusual capabilities to use them. They hook `sync.after` and write tool-specific files scoped to a narrow capability grant (`/workspace/**/CLAUDE.md` for the claude translator).

## What "canonical" buys you

- **One file to edit.** Run `samuel sync`; every tool gets the same context.
- **One file to diff in PRs.** Reviewers see one source of truth, not five sibling files drifting out of sync.
- **One CI gate.** The ≤150-line gate runs on `AGENTS.md` only; downstream files are regenerable.
- **One template to override.** `.samuel/templates/AGENTS.md.tmpl` shadows the embedded default; everything downstream cascades.

See [RFD 0002](../rfd/0002.md) for the full decision, including the rejected alternatives (write all the files; write a manifest that other tools opt-in to; write nothing and have plugins do everything).

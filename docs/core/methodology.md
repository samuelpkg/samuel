# Methodology

Samuel ships one default methodology: a hybrid of the **4D loop** (Deconstruct / Diagnose / Develop / Deliver) and the **Ralph Wiggum** iteration cap. The 4D loop is the per-task reasoning template; Ralph is the runtime that drives it autonomously with a hard ceiling.

## 4D, per task

| Phase | What happens |
| --- | --- |
| **Deconstruct** | Restate the goal. Enumerate the smallest set of files that change. |
| **Diagnose** | Read existing code. Note what's broken, what assumptions are at risk. |
| **Develop** | Make the change with tests. Respect the guardrails. |
| **Deliver** | Run quality checks. Summarise. Commit. |

The phases are not enforced by code — they are baked into the default `AGENTS.md` template. Agents follow them because the prompt instructs them to.

## Ralph, across tasks

The Ralph methodology is the autonomous outer loop:

```text
load PRD
for iteration in 1..max:
    pick next pending task (respecting deps)
    render prompt (task context + project snapshot)
    invoke agent
    agent calls `samuel run done|skip|enqueue`
    run quality checks
    if consec_fails >= max_consec_fails: abort
end
```

The iteration cap (`--max-iterations`, default 20) is the safety belt. The loop **always** stops there — no exceptions for "the agent thinks there's more work." That's the core Ralph invariant: bounded execution.

## Hook points

The loop fires 13 hook events. Plugins can register handlers; the framework runs a default handler when no plugin overrides one. Source order: user override (in `samuel.toml`) → built-in default → plugin.

| Hook | Fires | Default behaviour |
| --- | --- | --- |
| `before:loop` | once, before iteration 1 | log start, acquire lock |
| `after:loop` | once, after the loop exits | log summary, release lock |
| `before:iteration` | each iteration start | reload `prd.toon` |
| `after:iteration` | each iteration end | append `progress.md` |
| `iteration.gate` | gate before next iteration | check consec_fails, queue empty |
| `context.snapshot` | each iteration | re-emit `project-snapshot.toon` |
| `context.progress` | each iteration | rotate `progress-context.md` if > 500 lines |
| `context.task` | each iteration | emit `task-context.toon` for the picked task |
| `context.extra` | each iteration | plugin-only; no default |
| `before:agent.invoke` | per agent call | apply env allowlist + sandbox |
| `agent.invoke` | per agent call | shell out to claude / codex / copilot / gemini / kiro |
| `after:agent.invoke` | per agent call | parse exit status |
| `quality.check` | post-iteration | run commands from `[methodology.ralph].quality_checks` |

Each hook has per-call `strict` and `timeout` overrides in `samuel.toml`. Strict-mode failures abort the loop; non-strict failures log and continue. See [Methodology hooks](../concepts/methodology-hooks.md) for the rationale and [Hooks](../plugin-authors/hooks.md) for plugin author detail.

## CLI-mutation invariant

The agent **never** edits `.samuel/run/*.toon` directly. It calls:

- `samuel run done <id>` — mark complete
- `samuel run skip <id> --reason "…"` — skip with audit
- `samuel run reset <id>` — back to pending
- `samuel run enqueue <title>` — add with auto-id
- `samuel run task add <id> <title>` — add with explicit id (CI use)

This means: the agent doesn't need to know TOON, retries are safe (CLI calls are idempotent on duplicate IDs), and the audit trail is centralised. See [RFD 0006](../rfd/0006.md).

## Swapping the methodology

`samuel run <name>` selects a methodology by name; `samuel.toml`'s `default_methodology` sets the bare-invocation default. v2.0 ships Ralph as the only built-in; the hook framework is the extension surface for community methodologies that ship as plugins.

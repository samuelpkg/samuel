# Hooks

13 hook events fire across the autonomous loop and the lifecycle commands. Your plugin declares the events it handles in `[hooks]`; the framework calls the named function with a typed payload.

## Hook events

| Event | Fired by | Payload (in) | Return (out) | Default action |
| --- | --- | --- | --- | --- |
| `init.before` | `samuel init` | `{project_root, force, minimal}` | `{ok}` | none |
| `init.after` | `samuel init` | `{project_root, plugins_installed[]}` | `{ok}` | none |
| `sync.before` | `samuel sync` | `{project_root, dry_run}` | `{ok, skip?}` | none |
| `sync.after` | `samuel sync` | `{project_root, files_written[]}` | `{ok}` | none |
| `before:loop` | `samuel run start` | `{prd_path, max_iterations}` | `{ok}` | acquire lock, log start |
| `after:loop` | `samuel run start` | `{iterations_run, exit_reason}` | `{ok}` | release lock, log summary |
| `before:iteration` | each iteration | `{iteration, task}` | `{ok}` | reload `prd.toon` |
| `after:iteration` | each iteration | `{iteration, task, outcome}` | `{ok}` | append `progress.md` |
| `iteration.gate` | each iteration | `{iteration, consec_fails, queue_size}` | `{continue}` | enforce caps |
| `context.snapshot` | each iteration | `{project_root}` | `{toon_blob}` | walk repo, emit snapshot |
| `context.progress` | each iteration | `{progress_path}` | `{md_blob}` | rotate at 500 lines |
| `context.task` | each iteration | `{task}` | `{toon_blob}` | impl vs discovery shape |
| `context.extra` | each iteration | `{task}` | `{extras[]}` | none — plugin-only |
| `before:agent.invoke` | per agent call | `{agent, prompt, env, sandbox}` | `{agent, env, sandbox}` | apply env allowlist |
| `agent.invoke` | per agent call | `{agent, prompt, env, sandbox}` | `{exit_code, stdout, stderr}` | shell out |
| `after:agent.invoke` | per agent call | `{exit_code, stdout, stderr}` | `{ok}` | parse exit |
| `quality.check` | post-iteration | `{commands[]}` | `{ok, failures[]}` | run commands sequentially |

(Events `init.*` and `sync.*` are lifecycle events; the rest are loop events.)

## Declaring a hook

In `samuel-plugin.toml`:

```toml
[hooks]
"sync.after"     = "mirror_claude_md"
"context.extra"  = "inject_design_doc"
"quality.check"  = "run_security_scan"
```

The value is the exported function name in your WASM/OCI entrypoint, or the script path (relative to the plugin root) for skill plugins.

## Strict, timeout, order

User overrides in `samuel.toml` can tighten or relax handler behaviour:

```toml
[methodology.ralph.hooks."quality.check"]
handlers = ["samuel-security:run_security_scan", "default", "samuel-lint:check"]
strict   = true
timeout  = "60s"
```

- `handlers` orders the chain explicitly. `default` is the built-in.
- `strict = true` aborts the loop on any handler failure.
- `timeout` is per-handler — exceeding it kills the handler and treats it as failed.

When no override is present, handler order is **built-in default → plugins (manifest declaration order)**.

## Payload + return contracts

Payloads are JSON over the WASM host ABI / gRPC bridge. Plugins MUST:

- Treat unknown payload keys as forward-compatible (don't error).
- Return all required keys in the response envelope (`{ok}` is the minimum; `ok = false` plus a `reason` string is a failure).
- Stay within declared capabilities — calling `samuel.fs_write` without `filesystem.write` granted is denied at the host boundary, not at validate time.

## Return-value composition

When multiple handlers fire for the same event:

- **Bool returns** (`{ok}`) — all must be `ok = true` for the event to succeed.
- **Single-blob returns** (`context.snapshot`, `context.progress`, `context.task`) — last writer wins; user override usually pins which plugin owns the slot.
- **List-merge returns** (`context.extra`, `quality.check.failures`) — concatenated across handlers.
- **Transform chains** (`before:agent.invoke`) — each handler receives the previous handler's output as input; chain order matters.

See [RFD 0004](../rfd/0004.md) for the dispatch implementation.

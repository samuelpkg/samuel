---
title: Structured errors (Problem / Cause / Fix / DocsURL)
type: concept
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-orchestrator]
tags: [v1, rescue, v2-decision]
---

# Structured errors

Errors as a UX surface, not a debug surface. Every error Samuel returns carries enough structure for the CLI to render an actionable block.

## The shape

```go
type Error struct {
    Component   string  // which subsystem produced it
    Problem     string  // one-line description
    Cause       string  // root cause / wrapped err string
    Fix         string  // copy-pasteable remediation
    DocsURL     string  // optional doc link
    Recoverable bool    // can the user fix this?
    Path        string  // filesystem path involved, if any
    wrapped     error   // chain for errors.Is / errors.As
}
```

## The rendered output

CLI renders structured errors as:

```
Error: Cannot register gbrain MCP server
  Cause: gbrain not found on PATH
  Fix:   bun install -g gbrain
  Docs:  https://samuel.dev/docs/errors/SAM-MCP-001
```

Single-line form for logs:

```
[gbrain] Cannot register gbrain MCP server: gbrain not found on PATH
```

`Error.Error()` returns the single-line; CLI formatter uses the structured fields directly.

## Why this matters

Compare to a typical Go error message:

```
Error: failed to register MCP server: exec: "gbrain": executable file not found in $PATH
```

vs:

```
Error: Cannot register gbrain MCP server
  Cause: gbrain not found on PATH
  Fix:   bun install -g gbrain
  Docs:  https://samuel.dev/docs/errors/SAM-MCP-001
```

Second form tells the user:
1. **What** failed (Problem)
2. **Why** (Cause)
3. **How to fix it** (Fix — copy-pasteable)
4. **Where to learn more** (DocsURL)

Plus a `Recoverable` flag the CLI uses to decide between "retry this yourself" and "file a bug."

## Error codes via DocsURL

`SAM-MCP-001`, `SAM-GBRAIN-001`, `SAM-LOCK-001`, `SAM-ROLLBACK-001` — these are URL slugs at `https://samuel.dev/docs/errors/`. Documenting each error class once and linking from every occurrence pays off:

- Users can google an error code.
- Support conversations reference the code.
- Doc page can show "common causes" + "advanced debugging" without bloating every CLI message.

v2 should keep the code scheme. Each plugin gets a namespace (`SAM-` for core, `PLG-<plugin-name>-` for plugins).

## Error chain preservation

`Wrap(err)` preserves the original error in `wrapped`:

```go
return res, (&Error{
    Component:   NameGstack,
    Problem:     "gstack clone failed",
    Cause:       strings.TrimSpace(string(cloneOut)),
    Fix:         "check network connectivity and access to " + gstackRepoURL,
    Recoverable: true,
}).Wrap(err)
```

`errors.Is` and `errors.As` work across the boundary via `Unwrap`. Callers can pull out the original `*exec.ExitError`, `*os.PathError`, etc. if they need to.

## Subtle correctness detail: rollback recoverability

When install fails AND rollback also fails, v1 wraps the joined result in `*Error{Recoverable: false, DocsURL: SAM-ROLLBACK-001}`. Without this wrapper, `errors.As` would walk into the install side first and return that side's `Recoverable=true`, misreporting a wedged state as "user can retry."

Lesson: when you join errors from multiple sources, decide which one's flags should "win" and wrap explicitly.

## v2 application #v2-decision

Adopt verbatim. Plugin SDK exposes the `Error` type for plugins to use. Every CLI error path renders structured. Every plugin's error path uses the same shape.

Two extensions:

- **`Severity` field** — `info / warning / error / fatal`. v1 only distinguishes `Recoverable`; severity is finer-grained and helps the CLI choose color + prefix.
- **`RetryAfter time.Duration`** — for transient errors (rate limits, lock contention). Lets the CLI suggest a wait time rather than just "retry."

## Open

- Internationalization — `Problem` / `Fix` strings are English. Defer i18n; can be added by routing through a translation map keyed by error code.
- Telemetry — should the CLI optionally emit error-code occurrences (anonymized) to help prioritize doc improvements? User-opt-in; not for v2.0.

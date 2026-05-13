---
title: Smart bare invocation
type: concept
created: 2026-05-12
updated: 2026-05-12
sources: [2026-05-12-v1-commands]
tags: [v1, rescue, ux]
---

# Smart bare invocation

When a command is run with no args, what should happen? v1 has a clear answer for `samuel run`: **show status if a loop exists, print actionable help and exit non-zero otherwise. Never silently start work.**

The pattern in `auto.go`:

```go
func runRunBare(cmd, args) error {
    cwd, _ := os.Getwd()
    prdPath := core.GetAutoPRDPath(cwd)
    if _, err := os.Stat(prdPath); err == nil {
        return runAutoStatus(cmd, args)   // loop exists → show status
    }
    
    fmt.Fprintln(os.Stderr, "samuel: no autonomous loop initialized in this directory.")
    fmt.Fprintln(os.Stderr, "")
    fmt.Fprintln(os.Stderr, "  Initialize one:    samuel run init")
    fmt.Fprintln(os.Stderr, "  From a PRD:        samuel run init --prd .claude/tasks/0001-prd-feature.md")
    fmt.Fprintln(os.Stderr, "  Zero-setup mode:   samuel run pilot")
    fmt.Fprintln(os.Stderr, "")
    fmt.Fprintln(os.Stderr, "See 'samuel run --help' for the full subcommand list.")
    return errors.New("no auto loop initialized")
}
```

## The footgun being prevented

From the source comment:

> "preventing the v2-era footgun where a stray `samuel auto` could kick off pilot mode in a directory the user wasn't expecting."

Three real failure modes:

1. `cd ~/random-project && samuel auto` → was supposed to do nothing, started a pilot loop, made commits to a project the user wasn't ready to automate.
2. Shell tab completion + accidental enter on the wrong directory.
3. CI scripts that exec `samuel auto` expecting it to no-op when there's nothing to do.

Smart bare invocation fixes all three.

## The general pattern

For any verb that does **destructive or state-changing work**, bare invocation should:

1. **Detect intent.** Is the current directory ready for this action? Are there prerequisite files? Has the user already done the setup step?
2. **If ready → show read-only status.** Same as the explicit `status` subcommand.
3. **If not ready → print actionable help + exit non-zero.** Help goes to stderr (JSON consumers reading stdout aren't affected). List the 2-3 likely next commands.
4. **Never start work silently.** The user must type the verb that does the thing.

## Counter-example (what NOT to do)

`samuel add` with no name → cobra rejects via `cobra.RangeArgs(1, 2)`. That's fine — `add` requires an argument by its very nature.

`samuel doctor` with no args → runs the health check. Also fine — `doctor` is read-only by contract.

The principle is: **if it would mutate, it must be explicit.**

## v2 application #rescue

Apply this principle to every v2 verb that mutates:

- `samuel run` → status if methodology initialized, help otherwise.
- `samuel install` (no plugin name) → list installed + suggest discovering more. Never silently re-install.
- `samuel sync` → if `.samuel/sync.toml` exists, show what would update with `--dry-run`-style preview, prompt confirmation. Otherwise actionable help.
- `samuel init` → if already initialized, show status. Otherwise prompt-driven setup.

The opposite — read-only verbs — can run bare safely:

- `samuel ls` → list installed.
- `samuel doctor` → health check.
- `samuel version` → version info.

## Open

- Should bare-mutate verbs show a `--dry-run` preview AND prompt, instead of just helping? Probably for some (`sync`) but not all (`install` without an arg has nothing to preview).
- Exit code convention: zero for read-only, non-zero for "needs setup, didn't do anything." Confirmed for v2.
